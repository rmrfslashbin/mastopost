package lambda

import (
	"fmt"
	"strings"

	"github.com/rmrfslashbin/mastopost/pkg/config"
	"github.com/rmrfslashbin/mastopost/pkg/events"
	"github.com/rmrfslashbin/mastopost/pkg/ssmparams"
	"github.com/rs/zerolog/log"
)

// Delete removes a job and configs
func (l *LambdaConfig) Delete(confirm bool) error {

	if l.configFile == nil {
		return &NoConfigFile{}
	}

	if l.feedName == nil {
		return &NoFeedName{}
	}

	if l.lambdaFunctionName == nil {
		return &NoLambdaFunction{}
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

	// Check for lambda function ARN
	if _, ok := cfg.LambdaFunctionConfig[*l.lambdaFunctionName]; !ok {
		return &NoLambdaFunctionARN{LambdaFunctionName: l.lambdaFunctionName}
	}
	lambdaFunctionArn := cfg.LambdaFunctionConfig[*l.lambdaFunctionName].FunctionArn
	if lambdaFunctionArn == "" {
		return &NoLambdaFunctionARN{LambdaFunctionName: l.lambdaFunctionName}
	}

	if !confirm {
		fmt.Printf("feedname: %s\n", *l.feedName)
		fmt.Println("Confirm job delete:")
		fmt.Printf("Feed name:               %s\n", *l.feedName)
		fmt.Printf("Schedule Expression:     %s\n", feedConfig.ScheduleExpression)
		fmt.Printf("RSS feed URL:            %s\n", feedConfig.FeedURL)
		fmt.Printf("Mastodon instance:       %s\n", feedConfig.Instance)
		fmt.Printf("AWS profile:             %s\n", *l.awsprofile)
		fmt.Printf("AWS region:              %s\n", *l.awsregion)
		fmt.Print("Confirm delete of config? (y/n): ")
		var userConfirm string
		fmt.Scanln(&userConfirm)
		if strings.ToLower(userConfirm) != "y" {
			return &NoConfirm{}
		}
	}

	params, err := ssmparams.New(
		ssmparams.WithLogger(l.log),
		ssmparams.WithProfile(*l.awsprofile),
		ssmparams.WithRegion(*l.awsregion),
	)
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

	/*
		/mastopost/${feedname}/mastodon/instanceUrl
		/mastopost/${feedname}/mastodon/clientId
		/mastopost/${feedname}/mastodon/clientSecret
		/mastopost/${feedname}/mastodon/accessToken
		/mastopost/${feedname}/rss/feedUrl
		/mastopost/${feedname}/runtime/lastUpdated
		/mastopost/${feedname}/runtime/lastPublished
	*/

	paramNames := []string{
		fmt.Sprintf("/mastopost/%s/mastodon/instanceUrl", *l.feedName),
		fmt.Sprintf("/mastopost/%s/mastodon/clientId", *l.feedName),
		fmt.Sprintf("/mastopost/%s/mastodon/clientSecret", *l.feedName),
		fmt.Sprintf("/mastopost/%s/mastodon/accessToken", *l.feedName),
		fmt.Sprintf("/mastopost/%s/rss/feedUrl", *l.feedName),
		fmt.Sprintf("/mastopost/%s/runtime/lastUpdated", *l.feedName),
		fmt.Sprintf("/mastopost/%s/runtime/lastPublished", *l.feedName),
	}

	if opt, err := params.DeleteParams(paramNames); err != nil {
		return &ParametersDeleteError{InvalidParameters: opt.InvalidParameters, Err: err}
	} else {
		l.log.Info().Msgf("Deleted %d parameters", len(paramNames))
	}

	if err := eb.DeleteRule(&events.DeleteRuleInput{
		FeedName:    l.feedName,
		FunctionArn: &lambdaFunctionArn,
	}); err != nil {
		return &EventBridgeDeleteError{Err: err}
	}
	l.log.Info().Msgf("Deleted event")

	return nil
}
