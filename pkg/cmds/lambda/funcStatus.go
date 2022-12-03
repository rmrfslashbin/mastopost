package lambda

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/rmrfslashbin/mastopost/pkg/events"
	"github.com/rs/zerolog/log"
)

type StatusInput struct {
	Enable  *bool
	Disable *bool
}

// Status prints the status of the lambda event
func (l *LambdaConfig) Status(input *StatusInput) error {
	if l.feedName == nil {
		return &NoFeedName{}
	}

	/*
		if l.configFile == nil {
			return &NoConfigFile{}
		}

		// Load the config file
		cfg, err := config.NewConfig(*l.configFile)
		if err != nil {
			return &FeedLoadError{Err: err}
		}

		// Ensure the feed is in the config
		if _, ok := cfg.Feeds[*l.feedName]; !ok {
			return &FeedNotInConfig{feedname: *l.feedName}
		}

		// Easy access to the feed config
		feedConfig := cfg.Feeds[*l.feedName]
	*/

	eb, err := events.New(
		events.WithLogger(l.log),
		events.WithProfile(*l.awsprofile),
		events.WithRegion(*l.awsregion),
	)
	if err != nil {
		log.Error().Msg("failed to create eventbridge client")
		return err

	}

	ruleName := "mastopost-" + *l.feedName

	if input.Enable != nil && *input.Enable {
		err = eb.EnableRule(ruleName)
		if err != nil {
			return err
		}
	}

	if input.Disable != nil && *input.Disable {
		err = eb.DisableRule(ruleName)
		if err != nil {
			return err
		}
	}

	status, err := eb.GetEventByName(ruleName)
	if err != nil {
		log.Error().Msg("failed to get event")
		return err
	}

	state := "enabled"
	if !status.State {
		state = "disabled"
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', tabwriter.TabIndent)
	fmt.Fprintln(w, "Attrib\t Value\t")
	fmt.Fprintln(w, "Arn\t", status.Arn, "\t")
	fmt.Fprintln(w, "Description\t", status.Description, "\t")
	fmt.Fprintln(w, "Name\t", status.Name, "\t")
	fmt.Fprintln(w, "Schedule\t", status.ScheduleExpression, "\t")
	fmt.Fprintln(w, "State\t", state, "\t")
	w.Flush()
	return nil
}
