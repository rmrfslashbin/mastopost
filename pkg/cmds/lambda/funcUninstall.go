package lambda

import (
	"github.com/rmrfslashbin/mastopost/pkg/config"
	"github.com/rmrfslashbin/mastopost/pkg/events"
	"github.com/rs/zerolog/log"
)

func (l *LambdaConfig) Uninstall() error {
	// Load the config file
	cfg, err := config.NewConfig(*l.configFile)
	if err != nil {
		return &FeedLoadError{Err: err}
	}

	// Get function config data
	lambdaFunction, err := cfg.GetFunction(l.lambdaFunctionName)
	if err != nil {
		return err
	}

	eb, err := events.New(
		events.WithLogger(l.log),
		events.WithProfile(*l.awsprofile),
		events.WithRegion(*l.awsregion),
	)
	if err != nil {
		log.Error().Msg("failed to create eventbridge client")
		return err
	}

	if err := eb.UninstallLambdaFunction(&events.UninstallLambdaFunctionInput{
		FunctionArn: &lambdaFunction.FunctionArn,
		PolicyArn:   &lambdaFunction.PolicyArn,
	}); err != nil {
		log.Error().Msg("failed to uninstall lambda function")
		return err
	}
	return nil
}
