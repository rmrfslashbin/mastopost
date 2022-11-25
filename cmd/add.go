package cmd

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/aws/aws-sdk-go-v2/service/ssm/types"
	"github.com/iancoleman/strcase"
	"github.com/rmrfslashbin/mastopost/pkg/ssmparams"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
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
	addCmd.PersistentFlags().String("feedname", "", "Feed name")
	addCmd.PersistentFlags().String("cron", "", "Cron configuration for posting")
	addCmd.PersistentFlags().String("awsprofile", "default", "AWS profile to use for credentials")
	addCmd.PersistentFlags().String("awsregion", "us-east-1", "AWS profile to use for credentials")
	addCmd.PersistentFlags().Bool("confirm", false, "Confirm adding new config")

	addCmd.MarkPersistentFlagRequired("url")
	addCmd.MarkPersistentFlagRequired("instance")
	addCmd.MarkPersistentFlagRequired("clientid")
	addCmd.MarkPersistentFlagRequired("clientsec")
	addCmd.MarkPersistentFlagRequired("token")
	addCmd.MarkPersistentFlagRequired("feedname")
	addCmd.MarkPersistentFlagRequired("cron")

	viper.BindPFlags(addCmd.PersistentFlags())
}

func addNewConfig() error {
	fmt.Printf("feedname: %s\n", viper.GetString("feedname"))
	config := Config{
		URL:        viper.GetString("url"),
		Instance:   viper.GetString("instance"),
		ClientID:   viper.GetString("clientid"),
		ClientSec:  viper.GetString("clientsec"),
		Token:      viper.GetString("token"),
		Name:       strcase.ToCamel(viper.GetString("feedname")),
		Cron:       viper.GetString("cron"),
		AWSProfile: viper.GetString("awsprofile"),
		AWSRegion:  viper.GetString("awsregion"),
	}

	if !viper.GetBool("confirm") {
		fmt.Println("Confirm adding new config:")
		fmt.Printf("Feed name:               %s\n", config.Name)
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

	params, err := ssmparams.New(
		ssmparams.WithLogger(log),
		ssmparams.WithProfile(config.AWSProfile),
		ssmparams.WithRegion(config.AWSRegion),
	)
	if err != nil {
		return err
	}

	/*
		/mastopost/${feedname}/mastodon/instanceUrl
		/mastopost/${feedname}/mastodon/clientId
		/mastopost/${feedname}/mastodon/clientSecret
		/mastopost/${feedname}/mastodon/accessToken
		/mastopost/${feedname}/rss/feedUrl
		/mastopost/${feedname}/runtime/lastUpdated
		/mastopost/${feedname}/runtime/lastPublished
	*/

	var paramNames []*ssm.PutParameterInput

	paramNames = append(paramNames, &ssm.PutParameterInput{
		Name:      aws.String(fmt.Sprintf("/mastopost/%s/mastodon/instanceUrl", config.Name)),
		Value:     aws.String(config.Instance),
		Type:      types.ParameterTypeString,
		Overwrite: aws.Bool(true),
	})

	paramNames = append(paramNames, &ssm.PutParameterInput{
		Name:      aws.String(fmt.Sprintf("/mastopost/%s/mastodon/clientId", config.Name)),
		Value:     aws.String(config.ClientID),
		Type:      types.ParameterTypeString,
		Overwrite: aws.Bool(true),
	})

	paramNames = append(paramNames, &ssm.PutParameterInput{
		Name:      aws.String(fmt.Sprintf("/mastopost/%s/mastodon/clientSecret", config.Name)),
		Value:     aws.String(config.ClientSec),
		Type:      types.ParameterTypeString,
		Overwrite: aws.Bool(true),
	})

	paramNames = append(paramNames, &ssm.PutParameterInput{
		Name:      aws.String(fmt.Sprintf("/mastopost/%s/mastodon/accessToken", config.Name)),
		Value:     aws.String(config.Token),
		Type:      types.ParameterTypeString,
		Overwrite: aws.Bool(true),
	})

	paramNames = append(paramNames, &ssm.PutParameterInput{
		Name:      aws.String(fmt.Sprintf("/mastopost/%s/rss/feedUrl", config.Name)),
		Value:     aws.String(config.URL),
		Type:      types.ParameterTypeString,
		Overwrite: aws.Bool(true),
	})

	paramNames = append(paramNames, &ssm.PutParameterInput{
		Name:      aws.String(fmt.Sprintf("/mastopost/%s/runtime/lastUpdated", config.Name)),
		Value:     aws.String("unset"),
		Type:      types.ParameterTypeString,
		Overwrite: aws.Bool(true),
	})

	paramNames = append(paramNames, &ssm.PutParameterInput{
		Name:      aws.String(fmt.Sprintf("/mastopost/%s/runtime/lastPublished", config.Name)),
		Value:     aws.String("unset"),
		Type:      types.ParameterTypeString,
		Overwrite: aws.Bool(true),
	})

	for _, param := range paramNames {
		_, err := params.PutParam(param)
		if err != nil {
			return err
		}
	}
	return nil
}
