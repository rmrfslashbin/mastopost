package oneshot

import (
	"fmt"
	"net/url"
	"os"

	"github.com/mattn/go-mastodon"
	"github.com/rmrfslashbin/mastopost/pkg/config"
	"github.com/rmrfslashbin/mastopost/pkg/mastoclient"
	"github.com/rmrfslashbin/mastopost/pkg/rssfeed"
	"github.com/rmrfslashbin/mastopost/pkg/utils"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// NoConfigFile is returned when a filename is required but not provided
type NoConfigFile struct {
	Err error
}

// Error returns the error message
func (e *NoConfigFile) Error() string {
	if e.Err == nil {
		return "no config file provided. use WithConfigFile() to set the config file"
	}
	return e.Err.Error()
}

// NoFeedName is returned when a feed name is required but not provided
type NoFeedName struct {
	Err error
}

// Error returns the error message
func (e *NoFeedName) Error() string {
	if e.Err == nil {
		return "no feed name provided. use WithFeedName() to set the feed name"
	}
	return e.Err.Error()
}

// FeedLoadError is returned when a feed cannot be loaded
type FeedLoadError struct {
	Err error
	Msg string
}

// Error returns the error message
func (e *FeedLoadError) Error() string {
	if e.Msg == "" {
		e.Msg = "error loading feed"
	}
	if e.Err != nil {
		e.Msg += ": " + e.Err.Error()
	}
	return e.Msg
}

// LastUpdateLoadError is returned when the last update time cannot be loaded
type LastUpdateLoadError struct {
	Err error
	Msg string
}

// Error returns the error message
func (e *LastUpdateLoadError) Error() string {
	if e.Msg == "" {
		e.Msg = "error loading last update time"
	}
	if e.Err != nil {
		e.Msg += ": " + e.Err.Error()
	}
	return e.Msg
}

// FeedNotInConfig is returned when a feed is not in the config
type FeedNotInConfig struct {
	Err      error
	Msg      string
	feedname string
}

// Error returns the error message
func (e *FeedNotInConfig) Error() string {
	if e.Msg == "" {
		e.Msg = "feed not in config"
	}
	if e.feedname != "" {
		e.Msg += ": " + e.feedname
	}
	if e.Err != nil {
		e.Msg += ": " + e.Err.Error()
	}
	return e.Msg
}

type FeedUrlParseError struct {
	Err error
	Msg string
	Url string
}

func (e *FeedUrlParseError) Error() string {
	if e.Msg == "" {
		e.Msg = "error parsing feed url"
	}
	if e.Url != "" {
		e.Msg += ": " + e.Url
	}
	if e.Err != nil {
		e.Msg += ": " + e.Err.Error()
	}
	return e.Msg
}

// InstanceUrlParseError is returned when the instance url cannot be parsed
type InstanceUrlParseError struct {
	Err error
	Msg string
	Url string
}

// Error returns the error message
func (e *InstanceUrlParseError) Error() string {
	if e.Msg == "" {
		e.Msg = "error parsing instance url"
	}
	if e.Url != "" {
		e.Msg += ": " + e.Url
	}
	if e.Err != nil {
		e.Msg += ": " + e.Err.Error()
	}
	return e.Msg
}

// OneshotOptions is a function that can be used to configure the OneshotConfig
type OneshotOptions func(config *OneshotConfig)

// OneshotConfig is the configuration for the oneshot command
type OneshotConfig struct {
	log        *zerolog.Logger
	configFile *string
	feedName   *string
	dryrun     bool
}

// NewOneshotConfig creates a new OneshotConfig
func NewOneshot(opts ...OneshotOptions) (*OneshotConfig, error) {
	cfg := &OneshotConfig{}

	// Default dryrun to false
	cfg.dryrun = false

	// apply the list of options to Config
	for _, opt := range opts {
		opt(cfg)
	}

	// Set up the default logger if not set
	if cfg.log == nil {
		log := zerolog.New(os.Stderr).With().Timestamp().Logger()
		cfg.log = &log
	}

	return cfg, nil
}

// WithConfigFile sets the config file to use
func WithConfigFile(configFile *string) OneshotOptions {
	return func(config *OneshotConfig) {
		config.configFile = configFile
	}
}

// WithDryRun sets the dryrun flag
func WithDryrun(dryrun bool) OneshotOptions {
	return func(config *OneshotConfig) {
		config.dryrun = dryrun
	}
}

// WithFeedName sets the feed name
func WithFeedName(feedName *string) OneshotOptions {
	return func(config *OneshotConfig) {
		config.feedName = feedName
	}
}

// WithLogger sets the logger to use
func WithLogger(log *zerolog.Logger) OneshotOptions {
	return func(config *OneshotConfig) {
		config.log = log
	}
}

// Run runs the oneshot command
func (c *OneshotConfig) Run() error {
	c.log.Debug().Msg("Running oneshot")

	if c.configFile == nil {
		return &NoConfigFile{}
	}

	if c.feedName == nil {
		return &NoFeedName{}
	}

	// Load the config file
	cfg, err := config.NewConfig(*c.configFile)
	if err != nil {
		return &FeedLoadError{Err: err}
	}

	// Ensure the feed is in the config
	if _, ok := cfg.Feeds[*c.feedName]; !ok {
		return &FeedNotInConfig{feedname: *c.feedName}
	}

	// Easy access to the feed config
	feedConfig := cfg.Feeds[*c.feedName]

	// Load the last update time config
	lastUpdateConfig, err := config.NewLastUpdates(feedConfig.LastUpdateFile)
	if err != nil {
		return &LastUpdateLoadError{Err: err}
	}

	// Easy access to the last update data
	feedlastUpdateData := &lastUpdateConfig.FeedLastUpdate

	// Parse the feed url
	feedURL, err := url.Parse(feedConfig.FeedURL)
	if err != nil {
		return &FeedUrlParseError{Url: feedConfig.FeedURL, Err: err}
	}

	// Set up a new feed parser
	feed, err := rssfeed.New(
		rssfeed.WithLogger(c.log),
		rssfeed.WithURL(feedURL),
		rssfeed.WithLastUpdated(feedlastUpdateData.LastUpdated),
		rssfeed.WithLastPublished(feedlastUpdateData.LastPublished),
	)
	if err != nil {
		return err
	}

	// Get the new items
	newItems, err := feed.Parse()
	if err != nil {
		return err
	}

	// Log some info
	c.log.Info().
		Str("lastupdate", feed.GetLastUpdated().String()).
		Str("lastpublished", feed.GetLastPublished().String()).
		Str("feedname", *c.feedName).
		Msgf("Found %d new items", len(newItems))

	// Bail out if there's nothing new
	if len(newItems) < 1 {
		c.log.Info().Msg("no new posts")
		return nil
	}

	// Are we doing a dry run?
	if c.dryrun {
		c.log.Info().Msg("dryrun mode. not posting to Mastodon")
		return nil
	}

	instanceUrl, err := url.Parse(feedConfig.Instance)
	if err != nil {
		return &InstanceUrlParseError{Url: feedConfig.Instance, Err: err}
	}

	// Set up the Mastodon client
	client, err := mastoclient.New(
		mastoclient.WithLogger(c.log),
		mastoclient.WithInstance(instanceUrl),
		mastoclient.WithClientID(feedConfig.ClientId),
		mastoclient.WithClientSecret(feedConfig.ClientSecret),
		mastoclient.WithToken(feedConfig.AccessToken),
	)
	if err != nil {
		return err
	}

	// Set up a channel to receive the results
	ch := make(chan *mastodon.ID)

	for _, item := range newItems {
		// create a new post/toot
		newPost, err := utils.MakePost(item)
		if err != nil {
			return err
		}
		go func(newPost *mastodon.Toot) {
			id, err := client.Post(newPost)
			if err != nil {
				log.Error().Err(err).Msg("error posting to Mastodon")
			}
			ch <- id
		}(newPost)
	}

	for i := 0; i < len(newItems); i++ {
		status := <-ch
		log.Info().
			Str("id", fmt.Sprintf("%v", status)).
			Str("toInstance", instanceUrl.String()).
			Msg("posted to Mastodon")
	}
	close(ch)

	// Update state/config
	feedlastUpdateData.LastPublished = feed.GetLastPublished()
	feedlastUpdateData.LastUpdated = feed.GetLastUpdated()
	if feedlastUpdateData.FeedName == "" {
		feedlastUpdateData.FeedName = *c.feedName
	}

	if err := lastUpdateConfig.Save(); err != nil {
		return err
	}

	return nil
}
