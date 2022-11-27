package main

import (
	"encoding/gob"
	"fmt"
	"net/url"
	"os"
	"path"
	"time"

	"github.com/mattn/go-mastodon"
	"github.com/rmrfslashbin/mastopost/pkg/mastoclient"
	"github.com/rmrfslashbin/mastopost/pkg/rssfeed"
	"github.com/rmrfslashbin/mastopost/pkg/utils"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// Bootstrap configs
var (
	log           zerolog.Logger
	cfgFile       string
	homeConfigDir string
)

// FeedConfig is the configuration for a single RSS feed
type FeedConfig struct {
	FeedURL       string     `json:"feedurl"`
	LastUpdated   *time.Time `json:"lastupdated"`
	LastPublished *time.Time `json:"lastpublished"`
}

// Config is the configuration for all RSS feeds
type Config struct {
	Feeds map[string]FeedConfig `json:"feeds"`
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "mastopost",
	Short: "RSS feed items to Mastodoon status post",
	Long:  "Post RSS feed items to Mastodon status posts. ",
	Run: func(cmd *cobra.Command, args []string) {
		// Catch errors
		var err error
		defer func() {
			if err != nil {
				log.Fatal().Err(err).Msg("Error")
			}
		}()
		if err := mastopost(); err != nil {
			log.Fatal().Err(err).Msg("Error posting to Mastodon")
		}
	},
}

// init is called before main
func init() {
	var err error

	log = zerolog.New(os.Stderr).With().Timestamp().Logger()

	// Find home directory.
	homeConfigDir, err = os.UserConfigDir()
	cobra.CheckErr(err)
	homeConfigDir = path.Join(homeConfigDir, "mastopost")

	configFile := path.Join(homeConfigDir, "config.yaml")
	lastUpdateFile := path.Join(homeConfigDir, "lastupdate.gob")

	cobra.OnInitialize(initConfig)

	// Define flags/params
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", configFile, "Location of config file")
	rootCmd.PersistentFlags().String("feedname", "", "Name of the RSS feed to configure/run")
	rootCmd.PersistentFlags().String("lastupdate", lastUpdateFile, "Location of last update file")
	rootCmd.PersistentFlags().String("url", "", "URL of the RSS feed to parse")
	rootCmd.PersistentFlags().String("instance", "", "Mastodon instance to post to")
	rootCmd.PersistentFlags().String("clientid", "", "Mastodon app client id")
	rootCmd.PersistentFlags().String("clientsec", "", "Mastodon app client secret")
	rootCmd.PersistentFlags().String("token", "", "Mastodon app client access token")
	rootCmd.PersistentFlags().Bool("dryrun", false, "Dry run, don't post to Mastodon")

	// Require feedname
	rootCmd.MarkPersistentFlagRequired("feedname")

	// Bind flags to viper
	//viper.BindPFlags(rootCmd.PersistentFlags())
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Search config in home directory
		viper.AddConfigPath(homeConfigDir)
		viper.SetConfigType("yaml")
		viper.SetConfigName("config")
	}

	// read in environment variables that match
	viper.AutomaticEnv()

	// If a config file is found, read it in.
	viper.ReadInConfig()
}

// main is the entry point
func main() {
	rootCmd.Execute()
}

// Execute the mastopost command
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

// loadGOB loads the last update config data
func loadGOB(gobFile string) (Config, error) {
	var config Config
	file, err := os.Open(gobFile)
	if err != nil {
		// if the file doesn't exist, return an empty config
		if os.IsNotExist(err) {
			return config, nil
		}
		return config, err
	}
	defer file.Close()

	decoder := gob.NewDecoder(file)
	err = decoder.Decode(&config)
	if err != nil {
		return config, err
	}
	return config, nil
}

// saveGOB saves the last update config data
func saveGOB(gobFile string, config Config) error {
	file, err := os.Create(gobFile)
	if err != nil {
		return err
	}
	defer file.Close()
	enc := gob.NewEncoder(file)
	err = enc.Encode(config)
	if err != nil {
		return err
	}
	return nil
}

// mastopost is the main function
func mastopost() error {
	config, err := loadGOB(viper.GetString("lastupdate"))
	if err != nil {
		return err
	}

	if _, ok := config.Feeds[viper.GetString("feedname")]; !ok {
		if config.Feeds == nil {
			config.Feeds = make(map[string]FeedConfig)
		}
		config.Feeds[viper.GetString("feedname")] = FeedConfig{}
	}
	feedData := config.Feeds[viper.GetString("feedname")]

	feedUrl, err := url.Parse(viper.GetString("url"))
	if err != nil {
		return err
	}

	// Set up a new feed parser
	feed, err := rssfeed.New(
		rssfeed.WithLogger(log),
		rssfeed.WithURL(feedUrl),
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
	log.Info().
		Str("lastupdate", feed.GetLastUpdated().String()).
		Str("lastpublished", feed.GetLastPublished().String()).
		Msgf("Found %d new items", len(newItems))

	// Bail out if there's nothing new
	if len(newItems) < 1 {
		log.Info().Msg("no new posts")
		return nil
	}

	// Are we doing a dry run?
	if viper.GetBool("dryrun") {
		log.Info().Msg("dryrun mode. not posting to Mastodon")
		return nil
	}

	// Set up the Mastodon client
	instanceUrl, err := url.Parse(viper.GetString("instance"))
	if err != nil {
		return err
	}
	client, err := mastoclient.New(
		mastoclient.WithLogger(log),
		mastoclient.WithInstance(instanceUrl),
		mastoclient.WithClientID(viper.GetString("clientid")),
		mastoclient.WithClientSecret(viper.GetString("clientsec")),
		mastoclient.WithToken(viper.GetString("token")),
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
			Str("toInstance", instanceUrl.String()).
			Msg("posted to Mastodon")
	}
	close(ch)

	// Update state/config
	feedData.LastPublished = feed.GetLastPublished()
	feedData.LastUpdated = feed.GetLastUpdated()
	config.Feeds[viper.GetString("feedname")] = feedData
	err = saveGOB(viper.GetString("lastupdate"), config)
	if err != nil {
		return err
	}

	return nil
}
