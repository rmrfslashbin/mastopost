package config

import (
	"encoding/gob"
	"encoding/json"
	"io"
	"os"
	"time"
)

// FilenameRequired is returned when a filename is required but not provided
type FilenameRequired struct {
	Err error
}

// Error returns the error message
func (e *FilenameRequired) Error() string {
	if e.Err == nil {
		return "filename required"
	}
	return e.Err.Error()
}

// FileNotExist is returned when a file does not exist
type FileNotExist struct {
	Err      error
	Filename string
	Msg      string
}

// Error returns the error message
func (e *FileNotExist) Error() string {
	if e.Msg == "" {
		e.Msg = "file does not exist"
	}
	if e.Filename != "" {
		e.Msg += ": " + e.Filename
	}
	return e.Msg
}

// FeedConfig contains the configuration for a feed
type FeedConfig struct {
	// GOB file to store the last update time data
	LastUpdateFile string `json:"lastupdatefile"`

	// FeedURL is the URL of the RSS feed to
	FeedURL string `json:"feedurl"`

	// Instance is the URL of the Mastodon instance
	Instance string `json:"instance"`

	// ClientId is the Mastodon client ID
	ClientId string `json:"clientid"`

	// ClientSecret is the Mastodon client secret
	ClientSecret string `json:"clientsecret"`

	// AccessToken is the Mastodon access token
	AccessToken string `json:"accesstoken"`
}

// LambdaFunctionConfig contains the configuration for a lambda function
type LambdaFunctionConfig struct {
	// FunctionName is the name of the lambda function
	Name string `json:"name"`

	// FunctionArn is the ARN (Amazon Resource Name) of the Lambda function
	FunctionArn string `json:"functionArn"`
}

// Config contains the configuration for mastopost
type Config struct {
	// Feeds is a map of feed names to feed config
	Feeds map[string]FeedConfig `json:"feeds"`

	// LambdaFunctionConfig is the configuration for the Lambda function
	LambdaFunctionConfig map[string]LambdaFunctionConfig `json:"lambdaFunctions"`
}

// FeedLastUpdate is the configuration for a single RSS feed
type FeedLastUpdate struct {
	// FeedName is the name of the feed
	FeedName string `json:"feed_name"`

	// LastUpdated is the last date & time the RSS feed was updated
	LastUpdated *time.Time `json:"lastupdated"`

	// LastPublished is the last date an item was published
	LastPublished *time.Time `json:"lastpublished"`
}

// LastUpdates contains the last update time for each feed
type LastUpdates struct {
	// filename is the name of the file to save the last update data to
	filename string

	// FeedLastUpdate contains the feed's last update data
	FeedLastUpdate FeedLastUpdate `json:"feed"`
}

// NewConfig creates a new Config object
func NewConfig(filename string) (*Config, error) {
	if filename == "" {
		return nil, &FilenameRequired{}
	}

	// Load the file
	c := &Config{}
	if err := c.Load(filename); err != nil {
		return nil, err
	}
	return c, nil
}

// Load loads the config data
func (c *Config) Load(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		// If the file doens't exist, return an empty config
		if os.IsNotExist(err) {
			return nil
		} else {
			return err
		}
	}
	defer file.Close()

	reader, err := io.ReadAll(file)
	if err != nil {
		return err
	}

	return json.Unmarshal(reader, &c)
}

// NewLastUpdates creates a new LastUpdates object
func NewLastUpdates(filename string) (*LastUpdates, error) {
	if filename == "" {
		return nil, &FilenameRequired{}
	}

	// Load the file
	l := &LastUpdates{filename: filename}
	if err := l.Load(); err != nil {
		return nil, err
	}
	return l, nil
}

// loadGOB loads the last update config data
func (l *LastUpdates) Load() error {
	file, err := os.Open(l.filename)
	if err != nil {
		// if the file doesn't exist, return an empty config
		if os.IsNotExist(err) {
			return nil
		} else {
			return err
		}
	}
	defer file.Close()

	decoder := gob.NewDecoder(file)
	err = decoder.Decode(&l.FeedLastUpdate)
	return err
}

// saveGOB saves the last update config data
func (l *LastUpdates) Save() error {
	file, err := os.Create(l.filename)
	if err != nil {
		return err
	}
	defer file.Close()
	enc := gob.NewEncoder(file)
	err = enc.Encode(l.FeedLastUpdate)
	if err != nil {
		return err
	}
	return nil
}
