/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
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

var (
	log *zap.Logger
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
	log, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}
	defer log.Sync()

	initViper(viper.GetViper())
}

func initViper(v *viper.Viper) {
	v.AutomaticEnv()
}
