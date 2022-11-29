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

type OneshotOptions func(config *OneshotConfig)

type OneshotConfig struct {
	log            *zerolog.Logger
	lastUpdateFile string
	feedName       string
	feedURL        *url.URL
	dryrun         bool
	instance       *url.URL
	clientid       string
	clientsec      string
	token          string
}

func NewOneshot(opts ...OneshotOptions) (*OneshotConfig, error) {
	cfg := &OneshotConfig{}

	// apply the list of options to Config
	for _, opt := range opts {
		opt(cfg)
	}

	if cfg.log == nil {
		log := zerolog.New(os.Stderr).With().Timestamp().Logger()
		cfg.log = &log
	}

	return cfg, nil
}

func WithLogger(log *zerolog.Logger) OneshotOptions {
	return func(config *OneshotConfig) {
		config.log = log
	}
}

func WithLastUpdateFile(lastUpdateFile string) OneshotOptions {
	return func(config *OneshotConfig) {
		config.lastUpdateFile = lastUpdateFile
	}
}

func WithFeedName(feedName string) OneshotOptions {
	return func(config *OneshotConfig) {
		config.feedName = feedName
	}
}

func WithDryrun(dryrun bool) OneshotOptions {
	return func(config *OneshotConfig) {
		config.dryrun = dryrun
	}
}

func WithInstance(instance *url.URL) OneshotOptions {
	return func(config *OneshotConfig) {
		config.instance = instance
	}
}

func WithClientID(clientid string) OneshotOptions {
	return func(config *OneshotConfig) {
		config.clientid = clientid
	}
}

func WithClientSecret(clientsec string) OneshotOptions {
	return func(config *OneshotConfig) {
		config.clientsec = clientsec
	}
}

func WithToken(token string) OneshotOptions {
	return func(config *OneshotConfig) {
		config.token = token
	}
}

func WithFeedURL(feedURL *url.URL) OneshotOptions {
	return func(config *OneshotConfig) {
		config.feedURL = feedURL
	}
}

func (c *OneshotConfig) Run() error {
	c.log.Debug().Msg("Running oneshot")
	lastUpdateConfig, err := config.NewLastUpdates(c.lastUpdateFile)
	if err != nil {
		return err
	}

	if _, ok := lastUpdateConfig.Feeds[c.feedName]; !ok {
		if lastUpdateConfig.Feeds == nil {
			lastUpdateConfig.Feeds = make(map[string]config.FeedConfig)
		}
		lastUpdateConfig.Feeds[c.feedName] = config.FeedConfig{}
	}
	feedData := lastUpdateConfig.Feeds[c.feedName]

	// Set up a new feed parser
	feed, err := rssfeed.New(
		rssfeed.WithLogger(c.log),
		rssfeed.WithURL(c.feedURL),
		rssfeed.WithLastUpdated(feedData.LastUpdated),
		rssfeed.WithLastPublished(feedData.LastPublished),
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

	// Set up the Mastodon client
	client, err := mastoclient.New(
		mastoclient.WithLogger(c.log),
		mastoclient.WithInstance(c.instance),
		mastoclient.WithClientID(c.clientid),
		mastoclient.WithClientSecret(c.clientid),
		mastoclient.WithToken(c.clientid),
	)
	if err != nil {
		return err
	}

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
			Str("toInstance", c.instance.String()).
			Msg("posted to Mastodon")
	}
	close(ch)

	// Update state/config
	feedData.LastPublished = feed.GetLastPublished()
	feedData.LastUpdated = feed.GetLastUpdated()
	lastUpdateConfig.Feeds[c.feedName] = feedData
	if err := lastUpdateConfig.Save(); err != nil {
		return err
	}

	return nil
}
