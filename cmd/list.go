package cmd

import (
	"fmt"

	"github.com/iancoleman/strcase"
	"github.com/rmrfslashbin/mastopost/pkg/ssmparams"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
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
				log.Fatal("main crashed", zap.Error(err))
			}
		}()
		if err := listConfigs(); err != nil {
			log.Fatal("error listing configs", zap.Error(err))
		}
	},
}

func init() {
	log, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}
	defer log.Sync()
	rootCmd.AddCommand(listCmd)

	listCmd.PersistentFlags().String("feedname", "", "Feed name")
	listCmd.PersistentFlags().String("awsprofile", "default", "AWS profile to use for credentials")
	listCmd.PersistentFlags().String("awsregion", "us-east-1", "AWS profile to use for credentials")

	viper.BindPFlags(listCmd.PersistentFlags())
}

func listConfigs() error {
	params, err := ssmparams.New(
		ssmparams.WithLogger(log),
		ssmparams.WithProfile(viper.GetString("awsprofile")),
		ssmparams.WithRegion(viper.GetString("awsregion")),
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
	if viper.GetString("feedname") != "" {
		name := strcase.ToCamel(viper.GetString("feedname"))
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
