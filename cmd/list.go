package cmd

import (
	"fmt"

	"github.com/iancoleman/strcase"
	"github.com/rmrfslashbin/mastopost/pkg/ssmparams"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List RSS feed parser/poster configurations",
	Run: func(cmd *cobra.Command, args []string) {
		// Catch errors
		var err error
		defer func() {
			if err != nil {
				log.Fatal().Err(err).Msg("main crashed")
			}
		}()
		if err := listConfigs(); err != nil {
			log.Fatal().Err(err).Msg("error listing")
		}
	},
}

var listCmdViper = viper.New()

func init() {
	initViper(listCmdViper)

	rootCmd.AddCommand(listCmd)

	listCmd.Flags().String("feedname", "", "Feed name")
	listCmd.Flags().String("awsprofile", "default", "AWS profile to use for credentials")
	listCmd.Flags().String("awsregion", "us-east-1", "AWS profile to use for credentials")

	listCmdViper.BindPFlags(listCmd.Flags())
}

func listConfigs() error {
	params, err := ssmparams.New(
		ssmparams.WithLogger(log),
		ssmparams.WithProfile(listCmdViper.GetString("awsprofile")),
		ssmparams.WithRegion(listCmdViper.GetString("awsregion")),
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

	path := "/mastopost/"
	if listCmdViper.GetString("feedname") != "" {
		name := strcase.ToCamel(listCmdViper.GetString("feedname"))
		path += name + "/"
	}

	fmt.Printf("Listing parameters for path %s\n", path)
	var nextToken *string
	for {
		opt, err := params.ListAllParams(path, nextToken)
		if err != nil {
			return err
		}

		for _, p := range opt.Parameters {
			fmt.Printf("Name:    %s\n", *p.Name)
			fmt.Printf("Value:   %s\n", *p.Value)
			fmt.Printf("mtime:   %s\n", *p.LastModifiedDate)
			fmt.Printf("Version: %d\n", p.Version)
			fmt.Printf("ARN:     %s\n", *p.ARN)
			fmt.Println()
		}

		nextToken = opt.NextToken
		if nextToken == nil {
			break
		}
	}

	return nil
}
