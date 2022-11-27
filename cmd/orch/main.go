package main

import (
	"os"

	"github.com/rmrfslashbin/mastopost/pkg/events"
	"github.com/rs/zerolog"
)

var (
	log zerolog.Logger
)

func init() {
	// Set up the logger
	log = zerolog.New(os.Stderr).With().Timestamp().Logger()
}

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
