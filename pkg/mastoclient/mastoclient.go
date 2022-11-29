package mastoclient

import (
	"context"
	"fmt"
	"net/url"

	"github.com/mattn/go-mastodon"
	"github.com/rs/zerolog"
)

type NoInstance struct {
	Err error
	Msg string
}

// Error returns the error message
func (e *NoInstance) Error() string {
	if e.Msg == "" {
		e.Msg = "no instance. use WithInstance()"
	}
	if e.Err != nil {
		e.Msg += ": " + e.Err.Error()
	}
	return e.Msg
}

// No client ID error
type NoClientID struct {
	Err error
	Msg string
}

// Error returns the error message
func (e *NoClientID) Error() string {
	if e.Msg == "" {
		e.Msg = "no client ID. use WithClientID()"
	}
	if e.Err != nil {
		e.Msg += ": " + e.Err.Error()
	}
	return e.Msg
}

// No client secret error
type NoClientSecret struct {
	Err error
	Msg string
}

// Error returns the error message
func (e *NoClientSecret) Error() string {
	if e.Msg == "" {
		e.Msg = "no client secret. use WithClientSecret()"
	}
	if e.Err != nil {
		e.Msg += ": " + e.Err.Error()
	}
	return e.Msg
}

// No token error
type NoToken struct {
	Err error
	Msg string
}

// Error returns the error message
func (e *NoToken) Error() string {
	if e.Msg == "" {
		e.Msg = "No token. use WithToken()"
	}
	if e.Err != nil {
		e.Msg += ": " + e.Err.Error()
	}
	return e.Msg
}

// Post failed error
type PostFailed struct {
	Err error
	Msg string
}

// Error returns the error message
func (e *PostFailed) Error() string {
	if e.Msg == "" {
		e.Msg = "post failed"
	}
	if e.Err != nil {
		e.Msg += ": " + e.Err.Error()
	}
	return e.Msg
}

// Options for the weather query
type Option func(c *Config)

// Config for the weather query
type Config struct {
	log       *zerolog.Logger
	instance  *url.URL
	clientid  string
	clientsec string
	token     string
}

// NewConfig creates a new Config
func New(opts ...Option) (*Config, error) {
	c := &Config{}

	// apply the list of options to Config
	for _, opt := range opts {
		opt(c)
	}

	return c, nil
}

// WithInstance sets the instance to use
func WithInstance(instance *url.URL) Option {
	return func(c *Config) {
		c.instance = instance
	}
}

// WithClientID sets the client ID to use
func WithClientID(clientid string) Option {
	return func(c *Config) {
		c.clientid = clientid
	}
}

// WithClientSecret sets the client secret to use
func WithClientSecret(clientsec string) Option {
	return func(c *Config) {
		c.clientsec = clientsec
	}
}

// WithToken sets the token to use
func WithToken(token string) Option {
	return func(c *Config) {
		c.token = token
	}
}

// WithLogger sets the logger to use
func WithLogger(log *zerolog.Logger) Option {
	return func(c *Config) {
		c.log = log
	}
}

func (c *Config) Post(toot *mastodon.Toot) (*mastodon.ID, error) {
	// Check set up
	if c.instance == nil {
		return nil, &NoInstance{}
	}

	if c.clientid == "" {
		return nil, &NoClientID{}
	}

	if c.clientsec == "" {
		return nil, &NoClientSecret{}
	}

	if c.token == "" {
		return nil, &NoToken{}
	}

	// Set up Mastodon client
	client := mastodon.NewClient(&mastodon.Config{
		Server:       c.instance.String(),
		ClientID:     c.clientid,
		ClientSecret: c.clientsec,
		AccessToken:  c.token,
	})

	// Post the toot
	if status, err := client.PostStatus(context.Background(), toot); err != nil {
		fmt.Println(err)
		return nil, err
	} else {
		return &status.ID, nil
	}
}
