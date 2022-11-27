/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"os"
	"path"

	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type Config struct {
	URL        string `json:"url"`
	Instance   string `json:"instance"`
	ClientID   string `json:"clientid"`
	ClientSec  string `json:"clientsec"`
	Token      string `json:"token"`
	Name       string `json:"name"`
	Cron       string `json:"cron"`
	AWSProfile string `json:"awsprofile"`
	AWSRegion  string `json:"awsregion"`
}

const (
	LAMBDA_FUNCTION_NAME = "mastopost-rss-crossposter"
)

var (
	log           zerolog.Logger
	cfgFile       string
	configFile    string
	homeConfigDir string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "mastopost",
	Short: "A brief description of your application",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	var err error
	// Set up the logger
	log = zerolog.New(os.Stderr).With().Timestamp().Logger()

	cobra.OnInitialize(initConfig)

	// Find home directory.
	homeConfigDir, err = os.UserConfigDir()
	cobra.CheckErr(err)
	homeConfigDir = path.Join(homeConfigDir, "mastopost")

	configFile = path.Join(homeConfigDir, "config.yaml")

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", configFile, "Location of config file")
}

func initConfig() {
	initViper(viper.GetViper())
}

func initViper(v *viper.Viper) {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Search config in home directory
		v.AddConfigPath(homeConfigDir)
		v.SetConfigType("yaml")
		v.SetConfigName("config")
	}

	// read in environment variables that match
	v.AutomaticEnv()

	// If a config file is found, read it in.
	// If a config file is found, read it in.
	if err := v.ReadInConfig(); err == nil {
		if v == viper.GetViper() {
			log.Info().Str("config", v.ConfigFileUsed()).Msg("Using config file")
		}
	}
}
