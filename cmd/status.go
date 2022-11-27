package cmd

import (
	"github.com/davecgh/go-spew/spew"
	"github.com/rmrfslashbin/mastopost/pkg/events"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// listCmd represents the list command
var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show status for RSS feed parser/poster configuration",
	Run: func(cmd *cobra.Command, args []string) {
		// Catch errors
		var err error
		defer func() {
			if err != nil {
				log.Fatal().Err(err).Msg("main crashed")
			}
		}()
		if err := statusConfigs(); err != nil {
			log.Fatal().Err(err).Msg("error getting status")
		}
	},
}

var statusCmdViper = viper.New()

func init() {
	initViper(statusCmdViper)

	rootCmd.AddCommand(statusCmd)

	statusCmd.PersistentFlags().String("feedname", "", "Feed name")
	statusCmd.Flags().String("awsprofile", "default", "AWS profile to use for credentials")
	statusCmd.Flags().String("awsregion", "us-east-1", "AWS profile to use for credentials")

	statusCmd.MarkFlagRequired("feedname")

	listCmdViper.BindPFlags(statusCmd.Flags())
	listCmdViper.BindPFlags(statusCmd.PersistentFlags())
}

func statusConfigs() error {
	spew.Dump(viper.AllSettings())
	eb, err := events.New(
		events.WithLogger(log),
		events.WithProfile(statusCmdViper.GetString("awsprofile")),
		events.WithRegion(statusCmdViper.GetString("awsregion")),
	)
	if err != nil {
		log.Error().Msg("failed to create eventbridge client")
		return err

	}

	status, err := eb.GetEventByName(statusCmdViper.GetString("feedname"))
	if err != nil {
		log.Error().Msg("failed to get event")
		return err
	}
	spew.Dump(status)

	return nil
}
