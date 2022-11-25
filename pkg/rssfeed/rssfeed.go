package rssfeed

import (
	"net/url"
	"time"

	"github.com/mmcdole/gofeed"
	"go.uber.org/zap"
)

type NoUpdates struct {
	Err error
	Msg string
	Url *url.URL
}

type ParserError struct {
	Err error
	Msg string
	Url *url.URL
}

type ConfigError struct {
	Err     error
	Msg     string
	Item    string
	SetWith string
}

func (e *NoUpdates) Error() string {
	if e.Msg == "" {
		e.Msg = "no updates"
	}
	if e.Url != nil {
		e.Msg += " for " + e.Url.String()
	}
	return e.Msg
}

func (e *ParserError) Error() string {
	if e.Msg == "" {
		e.Msg = "error parsing feed"
	}
	if e.Url != nil {
		e.Msg += " (" + e.Url.String() + ")"
	}
	if e.Err != nil {
		e.Msg += ": " + e.Err.Error()
	}
	return e.Msg
}

func (e *ConfigError) Error() string {
	if e.Msg == "" {
		e.Msg = "missing required configuration."
	}
	if e.Item != "" {
		e.Msg += " " + e.Item + " is required."
	}
	if e.SetWith != "" {
		e.Msg += " Set it with " + e.SetWith + "."
	}
	return e.Msg
}

type NewItems *gofeed.Item

// Options for the weather query
type Option func(c *Config)

// Config for the weather query
type Config struct {
	log           *zap.Logger
	url           *url.URL
	lastUpdated   *time.Time
	lastPublished *time.Time
}

// NewConfig creates a new Config
func New(opts ...Option) (*Config, error) {
	c := &Config{}

	// apply the list of options to Config
	for _, opt := range opts {
		opt(c)
	}

	// Set up default logger
	if c.log == nil {
		if log, err := zap.NewProduction(); err != nil {
			return nil, err
		} else {
			c.log = log
			defer c.log.Sync()
		}
	}

	if c.lastUpdated == nil {
		epoch := time.Date(1980, 1, 1, 0, 0, 0, 0, time.UTC)
		c.lastUpdated = &epoch
	}

	if c.lastPublished == nil {
		epoch := time.Date(1980, 1, 1, 0, 0, 0, 0, time.UTC)
		c.lastPublished = &epoch
	}

	return c, nil
}

// WithURL sets the URL for the RSS feed
func WithURL(u *url.URL) Option {
	return func(c *Config) {
		c.url = u
	}
}

// WithLastUpdated sets the last updated time for the RSS feed
func WithLastUpdated(t time.Time) Option {
	return func(c *Config) {
		c.lastUpdated = &t
	}
}

// WithLogger sets the logger for the RSS feed
func WithLogger(l *zap.Logger) Option {
	return func(c *Config) {
		c.log = l
	}
}

// WithLastPublished sets the last published time for the RSS feed
func WithLastPublished(t time.Time) Option {
	return func(c *Config) {
		c.lastPublished = &t
	}
}

// GetURL returns the URL for the RSS feed
func (c *Config) GetURL() *url.URL {
	return c.url
}

// GetLastUpdated returns the last updated time for the RSS feed
func (c *Config) GetLastUpdated() *time.Time {
	return c.lastUpdated
}

// GetLastPublished returns the last published time for the RSS feed
func (c *Config) GetLastPublished() *time.Time {
	return c.lastPublished
}

// SetLastUpdated sets the last updated time for the RSS feed
func (c *Config) SetLastUpdated(timestamp *time.Time) {
	c.lastUpdated = timestamp
}

// SetLastPublished sets the last published time for the RSS feed
func (c *Config) SetLastPublished(timestamp *time.Time) {
	c.lastPublished = timestamp
}

// Parse the RSS feed
func (c *Config) Parse() ([]NewItems, error) {
	// ensure we have a URL
	if c.url == nil {
		return nil, &ConfigError{Item: "url", SetWith: "WithURL"}
	}

	// Set up the RSS parser
	fp := gofeed.NewParser()
	// Parse the RSS feed
	feed, err := fp.ParseURL(c.url.String())
	if err != nil {
		return nil, &ParserError{Err: err, Url: c.url}
	}

	// Log info about the feed
	c.log.Info("Parsed feed",
		zap.String("title", feed.Title),
		zap.String("link", feed.Link),
		zap.String("feed last updated", feed.UpdatedParsed.String()),
		zap.Int("items", len(feed.Items)),
		zap.String("last updated", c.lastUpdated.String()),
		zap.String("last published", c.lastPublished.String()),
	)
	if c.lastUpdated.After(*feed.UpdatedParsed) {
		c.log.Info("No updates")
		return nil, &NoUpdates{Url: c.url}
	}
	c.lastUpdated = feed.UpdatedParsed

	var newItems []NewItems

	var lastPublished time.Time
	for _, item := range feed.Items {
		if item.PublishedParsed.After(*c.lastPublished) {
			c.log.Debug("Found new item",
				zap.String("title", item.Title),
				zap.String("link", item.Link),
				zap.Time("published", *item.PublishedParsed),
			)
			newItems = append(newItems, item)
			if item.PublishedParsed.After(lastPublished) {
				lastPublished = *item.PublishedParsed
			}
		}
	}
	if !lastPublished.IsZero() {
		c.lastPublished = &lastPublished
	}
	return newItems, nil
}
