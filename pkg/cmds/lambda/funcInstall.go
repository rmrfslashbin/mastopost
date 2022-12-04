package lambda

import (
	"fmt"

	"github.com/rmrfslashbin/mastopost/pkg/events"
	"github.com/rs/zerolog/log"
)

func (l *LambdaConfig) Install() error {
	// TODO: code will force the function name to prefix with "mastopost-"

	eb, err := events.New(
		events.WithLogger(l.log),
		events.WithProfile(*l.awsprofile),
		events.WithRegion(*l.awsregion),
	)
	if err != nil {
		log.Error().Msg("failed to create eventbridge client")
		return err

	}

	opt, err := eb.InstallLambdaFunction(&events.InstallLambdaFunctionInput{
		FunctionName:        l.lambdaFunctionName,
		FunctionZipFilename: l.zipfilename,
	})
	if err != nil {
		log.Error().Msg("failed to install lambda function")
		return err
	}

	fmt.Printf("function name: %s\n", opt.FunctionName)
	fmt.Printf("function arn:  %s\n", opt.FunctionArn)
	fmt.Printf("policy arn:    %s\n", opt.PolicyArn)
	return nil
}
