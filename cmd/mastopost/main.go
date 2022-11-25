package main

import (
	"encoding/gob"
	"fmt"
	"net/url"
	"os"
	"path"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/iancoleman/strcase"
	"github.com/mattn/go-mastodon"
	"github.com/rmrfslashbin/mastopost/pkg/mastoclient"
	"github.com/rmrfslashbin/mastopost/pkg/rssfeed"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

var (
	log           *zap.Logger
	cfgFile       string
	homeConfigDir string
)

type FeedConfig struct {
	FeedURL       string     `json:"feedurl"`
	LastUpdated   *time.Time `json:"lastupdated"`
	LastPublished *time.Time `json:"lastpublished"`
}

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
				log.Fatal("main crashed", zap.Error(err))
			}
		}()
		if err := mastopost(); err != nil {
			log.Fatal("error posting", zap.Error(err))
		}
	},
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

// init is called before main
func init() {
	var err error

	// Set up the logger
	log, err = zap.NewProduction()
	if err != nil {
		panic(err)
	}
	defer log.Sync()

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
	viper.BindPFlags(rootCmd.PersistentFlags())
	spew.Dump(viper.AllSettings())
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	fmt.Println("InitConfig")
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

func loadGOB(gobFile string) (Config, error) {
	var config Config
	file, err := os.Open(gobFile)
	if err != nil {
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
	spew.Dump(viper.AllSettings())
	fmt.Printf("Last update fime is %s\n", viper.GetString("lastupdate"))
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
	)
	if err != nil {
		return err
	}

	// Set last updated and last published times, if provided
	if feedData.LastUpdated != nil {
		feed.SetLastUpdated(feedData.LastUpdated)
	}
	if feedData.LastPublished != nil {
		feed.SetLastPublished(feedData.LastPublished)
	}

	// Get the new items
	newItems, err := feed.Parse()
	if err != nil {
		return err
	}

	// Only run if there's new items
	if len(newItems) > 0 {
		posts := make([]*mastodon.Toot, len(newItems))

		for _, item := range newItems {
			// create a new post/toot
			newPost, err := MakePost(item)
			if err != nil {
				return err
			}
			posts = append(posts, newPost)
		}

		// Log some info
		log.Info("last update", zap.String("lastupdate", feed.GetLastUpdated().String()))
		log.Info("last published", zap.String("lastpublished", feed.GetLastPublished().String()))

		// Are we doing a dry run?
		if viper.GetBool("dryrun") {
			log.Info("dryrun mode. not posting to Mastodon")
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

		// Post the new items
		ids, err := client.Post(posts)
		if err != nil {
			return err
		}
		log.Info("posted to Mastodon", zap.Any("ids", ids))

		// Update state/config
		feedData.LastPublished = feed.GetLastPublished()
		feedData.LastUpdated = feed.GetLastUpdated()
		config.Feeds[viper.GetString("feedname")] = feedData
		err = saveGOB(viper.GetString("lastupdate"), config)
		if err != nil {
			return err
		}
	}
	return nil
}

// MakePost formats the RSS item into a Mastodon post
func MakePost(item rssfeed.NewItems) (*mastodon.Toot, error) {
	author := ""
	hashtags := ""

	if item.Author != nil {
		if item.Author.Name != "" {
			author = " by " + item.Author.Name
		}
		if item.Author.Email != "" {
			author += " (" + item.Author.Email + ")"
		}
	}

	if item.Categories != nil {
		hashtags = "\n"
		for _, cat := range item.Categories {
			hashtags += " #" + strcase.ToCamel(cat)
		}
	}

	newPost := &mastodon.Toot{
		Status: fmt.Sprintf("%s%s\n\n%s\n\n%s\n\n%s", item.Title, author, item.Published, item.Link, hashtags),
	}
	return newPost, nil
}
