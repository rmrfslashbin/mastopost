package main

import (
	"fmt"
	"os"
	"path"

	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	LAMBDA_FUNCTION_NAME = "mastopost-rss-crossposter"
)

var (
	log           zerolog.Logger
	cfgFile       string
	homeConfigDir string
)

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
				log.Fatal().Err(err).Msg("main crashed")
			}
		}()
		if err := mastopost(); err != nil {
			log.Fatal().Err(err).Msg("error posting")
		}
	},
}

// init is called before main
func init() {
	var err error

	// Set up the logger
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

func mastopost() error {
	/*
		lambdaFuncs := viper.GetStringMap("lambdaFunctions")
		spew.Dump(lambdaFuncs)

			if _, ok := lambdaFuncs[LAMBDA_FUNCTION_NAME]; !ok {
				return errors.New("missing lambdaFunctions in config")
			}
			if _, ok := lambdaFuncs[LAMBDA_FUNCTION_NAME]["functionarn"]; !ok {
				return errors.New("missing lambdaFunctions in config")
			}
	*/

	arn := viper.GetString("lambdaFunctions." + LAMBDA_FUNCTION_NAME)
	if arn == "" {
		return fmt.Errorf("missing lambdaFunctions.%s in config", LAMBDA_FUNCTION_NAME)
	}
	fmt.Printf("%s Arn: %s\n", LAMBDA_FUNCTION_NAME, arn)
	return nil
}

/*
func main() {
	eb, err := events.New(
		events.WithLogger(log),
		events.WithProfile("default"),
		events.WithRegion("us-east-1"),
	)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create eventbridge client")
	}
	arn, err := eb.PutRule(&events.NewEvent{
		Name:               "test",
		Description:        "test",
		ScheduleExpression: "rate(5 minutes)",
		//State: false, // default is false
		Feedname:  "test",
		LambdaArn: "arn:aws:lambda:us-east-1:xxxxxxxxxxxx:function:mastopost-rss-crossposter",
	})
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create eventbridge rule")
	}
	log.Info().Str("arn", *arn).Msg("created eventbridge rule")
}
*/
