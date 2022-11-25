package cmd

import (
	"fmt"
	"strings"

	"github.com/davecgh/go-spew/spew"
	"github.com/iancoleman/strcase"
	"github.com/rmrfslashbin/mastopost/pkg/ssmparams"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

// deleteCmd represents the delete command
var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete RSS feed parser/poster configurations",
	Run: func(cmd *cobra.Command, args []string) {
		// Catch errors
		var err error
		defer func() {
			if err != nil {
				log.Fatal("main crashed", zap.Error(err))
			}
		}()
		if err := deleteConfigs(); err != nil {
			log.Fatal("error deleting configs", zap.Error(err))
		}
	},
}

func init() {
	log, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}
	defer log.Sync()
	rootCmd.AddCommand(deleteCmd)

	deleteCmd.Flags().String("feedname", "", "Feed name")
	deleteCmd.Flags().String("awsprofile", "default", "AWS profile to use for credentials")
	deleteCmd.Flags().String("awsregion", "us-east-1", "AWS profile to use for credentials")
	deleteCmd.Flags().Bool("confirm", false, "Confirm delete")

	deleteCmd.MarkFlagRequired("feedname")
	viper.BindPFlags(deleteCmd.Flags())
	spew.Dump(deleteCmd.Flags())
}

func deleteConfigs() error {
	spew.Dump(viper.AllSettings())
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

	path := "/mastopost/" + strcase.ToCamel(viper.GetString("feedname")) + "/"

	var nextToken *string
	var paths []string
	fmt.Printf("Listing %s\n", path)
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
			paths = append(paths, *p.Name)
		}
		nextToken = opt.NextToken
		if nextToken == nil {
			break
		}
	}
	fmt.Printf("Got %d parameters for path %s\n", len(paths), path)

	if !viper.GetBool("confirm") {
		fmt.Print("Confirm delete? (y/n): ")
		var confirm string
		fmt.Scanln(&confirm)
		if strings.ToLower(confirm) != "y" {
			return nil
		}
	}

	delRes, err := params.DeleteParams(paths)
	if err != nil {
		return err
	}

	fmt.Printf("Deleted %d parameters\n", len(delRes.DeletedParameters))

	return nil
}
