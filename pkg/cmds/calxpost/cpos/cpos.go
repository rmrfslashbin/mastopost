package cpos

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	ics "github.com/arran4/golang-ical"
	"github.com/rmrfslashbin/mastopost/pkg/config"
	"github.com/rs/zerolog"
)

const (
	TSTAMP_FORMAT        = "20060102T150405Z"
	TSTAMP_SIMPLE_FORMAT = "20060102T150405"
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

	/*
		// Load the last update time config
		lastUpdateConfig, err := config.NewLastUpdates(feedConfig.LastUpdateFile)
		if err != nil {
			return &LastUpdateLoadError{Err: err}
		}

		// Easy access to the last update data
		feedlastUpdateData := &lastUpdateConfig.FeedLastUpdate
	*/

	// Parse the feed url
	feedURL, err := url.Parse(feedConfig.FeedURL)
	if err != nil {
		return &FeedUrlParseError{Url: feedConfig.FeedURL, Err: err}
	}

	resp, err := http.Get(feedURL.String())
	if err != nil {
		return &FeedLoadError{Err: err}
	}
	defer resp.Body.Close()

	// Set up a new cal feed parser
	cal, err := ics.ParseCalendar(resp.Body)
	if err != nil {
		return &FeedLoadError{Err: err}
	}

	for _, event := range cal.Events() {
		descr := strings.Replace(event.ComponentBase.GetProperty("DESCRIPTION").Value, `\n`, "\n", -1)
		sequence := event.ComponentBase.GetProperty("SEQUENCE").Value
		summary := event.ComponentBase.GetProperty("SUMMARY").Value
		uid := event.ComponentBase.GetProperty("UID").Value

		createDate, err := time.Parse(TSTAMP_FORMAT, event.ComponentBase.GetProperty("CREATED").Value)
		if err != nil {
			return err
		}

		endDate, err := time.Parse(TSTAMP_SIMPLE_FORMAT, event.ComponentBase.GetProperty("DTEND").Value)
		if err != nil {
			return err
		}

		if _, ok := event.ComponentBase.GetProperty("DTEND").ICalParameters["TZID"]; !ok {
			return errors.New("DTEND does not have a TZID")
		}
		endDateTZ := event.ComponentBase.GetProperty("DTSTART").ICalParameters["TZID"][0]

		dtStamp, err := time.Parse(TSTAMP_FORMAT, event.ComponentBase.GetProperty("DTSTAMP").Value)
		if err != nil {
			return err
		}

		startDate, err := time.Parse(TSTAMP_SIMPLE_FORMAT, event.ComponentBase.GetProperty("DTSTART").Value)
		if err != nil {
			return err
		}

		if _, ok := event.ComponentBase.GetProperty("DTSTART").ICalParameters["TZID"]; !ok {
			return errors.New("DTSTART does not have a TZID")
		}
		startDateTZ := event.ComponentBase.GetProperty("DTSTART").ICalParameters["TZID"][0]

		lastModified, err := time.Parse(TSTAMP_FORMAT, event.ComponentBase.GetProperty("LAST-MODIFIED").Value)
		if err != nil {
			return err
		}

		fmt.Printf("Description:    %s\n", descr)
		fmt.Printf("Created:        %s\n", createDate.Format(time.RFC3339))
		fmt.Printf("End Date:       %s %s\n", endDate.Format(time.RFC3339), endDateTZ)
		fmt.Printf("DT Stamp:       %s\n", dtStamp.Format(time.RFC3339))
		fmt.Printf("Start Date:     %s %s\n", startDate.Format(time.RFC3339), startDateTZ)
		fmt.Printf("Last Modified:  %s\n", lastModified.Format(time.RFC3339))
		fmt.Printf("Sequence:       %s\n", sequence)
		fmt.Printf("Summary:        %s\n", summary)
		fmt.Printf("UID:            %s\n", uid)
	}

	return nil
}
