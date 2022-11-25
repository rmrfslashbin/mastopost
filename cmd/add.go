/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"strings"

	"github.com/mattn/go-mastodon"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

var (
	log    *zap.Logger
	client *mastodon.Client
)

// addCmd represents the add command
var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Add new RSS feed parser/poster configuration",
	Run: func(cmd *cobra.Command, args []string) {
		// Catch errors
		var err error
		defer func() {
			if err != nil {
				log.Fatal("main crashed", zap.Error(err))
			}
		}()
		if err := addNewConfig(); err != nil {
			log.Fatal("error adding new config", zap.Error(err))
		}
	},
}

func init() {
	log, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}
	defer log.Sync()
	rootCmd.AddCommand(addCmd)

	addCmd.PersistentFlags().String("url", "", "URL of the RSS feed to parse")
	addCmd.PersistentFlags().String("instance", "", "Mastodon instance to post to")
	addCmd.PersistentFlags().String("clientid", "", "Mastodon app client id")
	addCmd.PersistentFlags().String("clientsec", "", "Mastodon app client secret")
	addCmd.PersistentFlags().String("token", "", "Mastodon app client access token")
	addCmd.PersistentFlags().String("name", "", "App name")
	addCmd.PersistentFlags().String("cron", "", "Cron configuration for posting")
	addCmd.PersistentFlags().String("awsprofile", "default", "AWS profile to use for credentials (default: default)")
	addCmd.PersistentFlags().String("awsregion", "us-east-1", "AWS profile to use for credentials (default: us-east-1")
	addCmd.PersistentFlags().Bool("confirm", false, "Confirm adding new config")

	addCmd.MarkPersistentFlagRequired("url")
	addCmd.MarkPersistentFlagRequired("instance")
	addCmd.MarkPersistentFlagRequired("clientid")
	addCmd.MarkPersistentFlagRequired("clientsec")
	addCmd.MarkPersistentFlagRequired("token")
	addCmd.MarkPersistentFlagRequired("name")
	addCmd.MarkPersistentFlagRequired("cron")

	viper.BindPFlags(addCmd.PersistentFlags())
}

func addNewConfig() error {
	config := Config{
		URL:        viper.GetString("url"),
		Instance:   viper.GetString("instance"),
		ClientID:   viper.GetString("clientid"),
		ClientSec:  viper.GetString("clientsec"),
		Token:      viper.GetString("token"),
		Name:       viper.GetString("name"),
		Cron:       viper.GetString("cron"),
		AWSProfile: viper.GetString("awsprofile"),
		AWSRegion:  viper.GetString("awsregion"),
	}

	if !viper.GetBool("confirm") {
		fmt.Println("Confirm adding new config:")
		fmt.Printf("App name:                %s\n", config.Name)
		fmt.Printf("RSS feed URL:            %s\n", config.URL)
		fmt.Printf("Mastodon instance:       %s\n", config.Instance)
		fmt.Printf("Mastodon client id:      %s\n", config.ClientID)
		fmt.Printf("Mastodon client secret:  %s\n", config.ClientSec)
		fmt.Printf("Mastodon access token:   %s\n", config.Token)
		fmt.Printf("Cron configuration:      %s\n", config.Cron)
		fmt.Printf("AWS profile:             %s\n", config.AWSProfile)
		fmt.Printf("AWS region:              %s\n", config.AWSRegion)
		fmt.Print("Confirm adding new config? (y/n): ")
		var confirm string
		fmt.Scanln(&confirm)
		if strings.ToLower(confirm) != "y" {
			return nil
		}
	}

	return nil
}
