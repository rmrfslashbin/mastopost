package lambda

import (
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/aws/aws-sdk-go-v2/service/ssm/types"
	"github.com/rmrfslashbin/mastopost/pkg/config"
	"github.com/rmrfslashbin/mastopost/pkg/events"
	"github.com/rmrfslashbin/mastopost/pkg/ssmparams"
	"github.com/rs/zerolog/log"
)

type AddInput struct {
	Confirm *bool
	Enable  *bool
}

// Add adds a new job and config
func (l *LambdaConfig) Add(addInput *AddInput) error {
	confirm := false
	enable := false

	if addInput.Confirm != nil {
		confirm = *addInput.Confirm
	}

	if addInput.Enable != nil {
		enable = *addInput.Enable
	}

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
		fmt.Println("Confirm adding new config:")
		fmt.Printf("Feed name:               %s\n", *l.feedName)
		fmt.Printf("Schedule Expression:     %s\n", feedConfig.ScheduleExpression)
		fmt.Printf("RSS feed URL:            %s\n", feedConfig.FeedURL)
		fmt.Printf("Mastodon instance:       %s\n", feedConfig.Instance)
		fmt.Printf("Mastodon client id:      %s\n", feedConfig.ClientId)
		fmt.Printf("Mastodon client secret:  %s\n", feedConfig.ClientSecret)
		fmt.Printf("Mastodon access token:   %s\n", feedConfig.AccessToken)
		fmt.Printf("AWS profile:             %s\n", *l.awsprofile)
		fmt.Printf("AWS region:              %s\n", *l.awsregion)
		fmt.Printf("Lambda function name:    %s\n", *l.lambdaFunctionName)
		fmt.Printf("Lambda function ARN:     %s\n", lambdaFunctionArn)
		fmt.Printf("Enable:                  %t\n", enable)
		fmt.Print("Confirm adding new config? (y/n): ")
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

	/*
		/mastopost/${feedname}/mastodon/instanceUrl
		/mastopost/${feedname}/mastodon/clientId
		/mastopost/${feedname}/mastodon/clientSecret
		/mastopost/${feedname}/mastodon/accessToken
		/mastopost/${feedname}/rss/feedUrl
		/mastopost/${feedname}/runtime/lastUpdated
		/mastopost/${feedname}/runtime/lastPublished
	*/

	var paramNames []*ssm.PutParameterInput

	paramNames = append(paramNames, &ssm.PutParameterInput{
		Name:      aws.String(fmt.Sprintf("/mastopost/%s/mastodon/instanceUrl", *l.feedName)),
		Value:     aws.String(feedConfig.Instance),
		Type:      types.ParameterTypeString,
		Overwrite: aws.Bool(true),
	})

	paramNames = append(paramNames, &ssm.PutParameterInput{
		Name:      aws.String(fmt.Sprintf("/mastopost/%s/mastodon/clientId", *l.feedName)),
		Value:     aws.String(feedConfig.ClientId),
		Type:      types.ParameterTypeString,
		Overwrite: aws.Bool(true),
	})

	paramNames = append(paramNames, &ssm.PutParameterInput{
		Name:      aws.String(fmt.Sprintf("/mastopost/%s/mastodon/clientSecret", *l.feedName)),
		Value:     aws.String(feedConfig.ClientSecret),
		Type:      types.ParameterTypeString,
		Overwrite: aws.Bool(true),
	})

	paramNames = append(paramNames, &ssm.PutParameterInput{
		Name:      aws.String(fmt.Sprintf("/mastopost/%s/mastodon/accessToken", *l.feedName)),
		Value:     aws.String(feedConfig.AccessToken),
		Type:      types.ParameterTypeString,
		Overwrite: aws.Bool(true),
	})

	paramNames = append(paramNames, &ssm.PutParameterInput{
		Name:      aws.String(fmt.Sprintf("/mastopost/%s/rss/feedUrl", *l.feedName)),
		Value:     aws.String(feedConfig.FeedURL),
		Type:      types.ParameterTypeString,
		Overwrite: aws.Bool(true),
	})

	epoch := time.Date(1980, 1, 1, 0, 0, 0, 0, time.UTC).Format(time.RFC3339)

	paramNames = append(paramNames, &ssm.PutParameterInput{
		Name:      aws.String(fmt.Sprintf("/mastopost/%s/runtime/lastUpdated", *l.feedName)),
		Value:     aws.String(epoch),
		Type:      types.ParameterTypeString,
		Overwrite: aws.Bool(true),
	})

	paramNames = append(paramNames, &ssm.PutParameterInput{
		Name:      aws.String(fmt.Sprintf("/mastopost/%s/runtime/lastPublished", *l.feedName)),
		Value:     aws.String(epoch),
		Type:      types.ParameterTypeString,
		Overwrite: aws.Bool(true),
	})

	for _, param := range paramNames {
		_, err := params.PutParam(param)
		if err != nil {
			return err
		}
		log.Info().Str("name", *param.Name).Msg("put parameter")
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

	ruleArn, err := eb.PutRule(&events.NewEvent{
		Name:               "mastopost-" + *l.feedName,
		Description:        "Mastopost cron for " + *l.feedName,
		ScheduleExpression: feedConfig.ScheduleExpression,
		State:              enable,
		Feedname:           *l.feedName,
		LambdaArn:          lambdaFunctionArn,
	})
	if err != nil {
		log.Error().Msg("failed to create eventbridge rule")
		return err
	}
	log.Info().Str("arn", *ruleArn).Msg("created eventbridge rule")

	return nil
}
