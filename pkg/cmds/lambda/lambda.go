package lambda

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/aws/aws-sdk-go-v2/service/ssm/types"
	"github.com/davecgh/go-spew/spew"
	"github.com/iancoleman/strcase"
	"github.com/rmrfslashbin/mastopost/pkg/config"
	"github.com/rmrfslashbin/mastopost/pkg/events"
	"github.com/rmrfslashbin/mastopost/pkg/ssmparams"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// AWSRegionRequiredError is returned when AWS Region is not set
type AWSRegionRequiredError struct {
	Err error
}

// Error returns the error message
func (e *AWSRegionRequiredError) Error() string {
	if e.Err == nil {
		return "AWS Region is required. Use WithRegion() to set it."
	}
	return e.Err.Error()
}

// AWSConfigError is an error returned when there is an error with the AWS Config
type AWSConfigError struct {
	Err error
}

// Error returns the error message
func (e *AWSConfigError) Error() string {
	if e.Err == nil {
		return "AWS Config error"
	} else {
		return fmt.Sprintf("AWS Config error: %s", e.Err.Error())
	}
}

// FeedNameMissingError is returned when the feed name is missing
type FeedNameMissingError struct {
	Err error
}

// Error returns the error message
func (e *FeedNameMissingError) Error() string {
	if e.Err == nil {
		return "Feed name is required"
	}
	return e.Err.Error()
}

// NoLambdaFunction is returned when the lambda function name is not set
type NoLambdaFunction struct {
	Err error
}

// Error returns the error message
func (e *NoLambdaFunction) Error() string {
	if e.Err == nil {
		return "no lambda function name provided. use WithLambdaFunctionName()"
	}
	return e.Err.Error()
}

// NoConfigFile is returned when a filename is required but not provided
type NoConfigFile struct {
	Err error
}

// Error returns the error message
func (e *NoConfigFile) Error() string {
	if e.Err == nil {
		return "no config file provided. use WithConfigFile() to set the config file"
	}
	return e.Err.Error()
}

// NoFeedName is returned when a feed name is required but not provided
type NoFeedName struct {
	Err error
}

// Error returns the error message
func (e *NoFeedName) Error() string {
	if e.Err == nil {
		return "no feed name provided. use WithFeedName() to set the feed name"
	}
	return e.Err.Error()
}

// FeedLoadError is returned when a feed cannot be loaded
type FeedLoadError struct {
	Err error
	Msg string
}

// Error returns the error message
func (e *FeedLoadError) Error() string {
	if e.Msg == "" {
		e.Msg = "error loading feed"
	}
	if e.Err != nil {
		e.Msg += ": " + e.Err.Error()
	}
	return e.Msg
}

// FeedNotInConfig is returned when a feed is not in the config
type FeedNotInConfig struct {
	Err      error
	Msg      string
	feedname string
}

// Error returns the error message
func (e *FeedNotInConfig) Error() string {
	if e.Msg == "" {
		e.Msg = "feed not in config"
	}
	if e.feedname != "" {
		e.Msg += ": " + e.feedname
	}
	if e.Err != nil {
		e.Msg += ": " + e.Err.Error()
	}
	return e.Msg
}

type NoLambdaFunctionARN struct {
	Err                error
	Msg                string
	LambdaFunctionName *string
}

func (e *NoLambdaFunctionARN) Error() string {
	if e.Msg == "" {
		e.Msg = "can't find lambda function in config"
	}
	if e.LambdaFunctionName != nil {
		e.Msg += ": " + *e.LambdaFunctionName
	}
	if e.Err != nil {
		e.Msg += ": " + e.Err.Error()
	}
	return e.Msg
}

// LambdaOptions is a function that can be used to configure the LambdaConfig
type LambdaOptions func(config *LambdaConfig)

// LambdaConfig is the configuration for the lambda command set
type LambdaConfig struct {
	awsprofile         *string
	awsregion          *string
	configFile         *string
	dryrun             bool
	feedName           *string
	lambdaFunctionName *string
	log                *zerolog.Logger
}

// NewOneshotConfig creates a new LambdaOptions
func NewLambda(opts ...LambdaOptions) (*LambdaConfig, error) {
	cfg := &LambdaConfig{}

	// Default dryrun to false
	cfg.dryrun = false

	// apply the list of options to Config
	for _, opt := range opts {
		opt(cfg)
	}

	// Set up the default logger if not set
	if cfg.log == nil {
		log := zerolog.New(os.Stderr).With().Timestamp().Logger()
		cfg.log = &log
	}

	if cfg.awsregion == nil {
		return nil, &AWSRegionRequiredError{}
	}

	return cfg, nil
}

// WithAWSProfile sets the AWS profile to use
func WithAWSProfile(awsprofile *string) LambdaOptions {
	return func(config *LambdaConfig) {
		config.awsprofile = awsprofile
	}
}

// WithAWSRegion sets the AWS region to use
func WithAWSRegion(awsregion *string) LambdaOptions {
	return func(config *LambdaConfig) {
		config.awsregion = awsregion
	}
}

// WithConfigFile sets the config file to use
func WithConfigFile(configFile *string) LambdaOptions {
	return func(config *LambdaConfig) {
		config.configFile = configFile
	}
}

// WithDryRun sets the dryrun flag
func WithDryrun(dryrun bool) LambdaOptions {
	return func(config *LambdaConfig) {
		config.dryrun = dryrun
	}
}

// WithFeedName sets the feed name
func WithFeedName(feedName *string) LambdaOptions {
	if feedName != nil {
		return func(config *LambdaConfig) {
			camelCaseFeedName := strcase.ToCamel(*feedName)
			config.feedName = &camelCaseFeedName
		}
	}
	return func(config *LambdaConfig) {
		config.feedName = nil
	}
}

// WithLambdaFunctionName sets the lambda function name
func WithLambdaFunctionName(lambdaFunctionName *string) LambdaOptions {
	return func(config *LambdaConfig) {
		config.lambdaFunctionName = lambdaFunctionName
	}
}

// WithLogger sets the logger to use
func WithLogger(log *zerolog.Logger) LambdaOptions {
	return func(config *LambdaConfig) {
		config.log = log
	}
}

// List lists the lambda event configurations
func (l *LambdaConfig) List() error {
	params, err := ssmparams.New(
		ssmparams.WithLogger(l.log),
		ssmparams.WithProfile(*l.awsprofile),
		ssmparams.WithRegion(*l.awsregion),
	)
	if err != nil {
		return err
	}
	path := "/mastopost/"
	if l.feedName != nil {
		path = fmt.Sprintf("%s%s/", path, *l.feedName)
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

func (l *LambdaConfig) Add() error {
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

	fmt.Printf("feedname: %s\n", *l.feedName)
	fmt.Println("Confirm adding new config:")
	fmt.Printf("Feed name:               %s\n", *l.feedName)
	fmt.Printf("RSS feed URL:            %s\n", feedConfig.FeedURL)
	fmt.Printf("Mastodon instance:       %s\n", feedConfig.Instance)
	fmt.Printf("Mastodon client id:      %s\n", feedConfig.ClientId)
	fmt.Printf("Mastodon client secret:  %s\n", feedConfig.ClientSecret)
	fmt.Printf("Mastodon access token:   %s\n", feedConfig.AccessToken)
	//fmt.Printf("Cron configuration:      %s\n", config.Cron)
	fmt.Printf("AWS profile:             %s\n", *l.awsprofile)
	fmt.Printf("AWS region:              %s\n", *l.awsregion)
	fmt.Print("Confirm adding new config? (y/n): ")
	var confirm string
	fmt.Scanln(&confirm)
	if strings.ToLower(confirm) != "y" {
		return nil
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
		Name:               *l.feedName,
		Description:        "Mastopost cron for " + *l.feedName,
		ScheduleExpression: "rate(30 minutes)",
		//State: false, // default is false
		Feedname:  *l.feedName,
		LambdaArn: lambdaFunctionArn,
	})
	if err != nil {
		log.Error().Msg("failed to create eventbridge rule")
		return err
	}
	log.Info().Str("arn", *ruleArn).Msg("created eventbridge rule")

	return nil
}

func (l *LambdaConfig) Delete() error {
	/*
		params, err := ssmparams.New(
			ssmparams.WithLogger(log),
			ssmparams.WithProfile(deleteCmdViper.GetString("awsprofile")),
			ssmparams.WithRegion(deleteCmdViper.GetString("awsregion")),
		)
		if err != nil {
			return err
		}

		path := "/mastopost/" + strcase.ToCamel(deleteCmdViper.GetString("feedname")) + "/"

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

		if !deleteCmdViper.GetBool("confirm") {
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


	*/
	return nil
}
