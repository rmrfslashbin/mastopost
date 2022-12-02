package lambda

import (
	"github.com/davecgh/go-spew/spew"
	"github.com/rmrfslashbin/mastopost/pkg/events"
	"github.com/rs/zerolog/log"
)

// Status prints the status of the lambda event
func (l *LambdaConfig) Status() error {
	eb, err := events.New(
		events.WithLogger(l.log),
		events.WithProfile(*l.awsprofile),
		events.WithRegion(*l.awsregion),
	)
	if err != nil {
		log.Error().Msg("failed to create eventbridge client")
		return err

	}

	status, err := eb.GetEventByName(*l.feedName)
	if err != nil {
		log.Error().Msg("failed to get event")
		return err
	}
	spew.Dump(status)

	return nil
}
