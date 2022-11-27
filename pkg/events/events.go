package events

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge/types"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/rs/zerolog"
)

type RuleArn *string

type NewEvent struct {
	Name               string
	Description        string
	ScheduleExpression string
	State              bool
	Feedname           string
	LambdaArn          string
}

type EventPramsOptions func(config *EventPramsConfig)

type EventPramsConfig struct {
	log         zerolog.Logger
	region      string
	profile     string
	eventbridge *eventbridge.Client
	lambda      *lambda.Client
}

func New(opts ...func(*EventPramsConfig)) (*EventPramsConfig, error) {
	cfg := &EventPramsConfig{}

	// apply the list of options to Config
	for _, opt := range opts {
		opt(cfg)
	}

	if cfg.region == "" {
		return nil, fmt.Errorf("region is required")
	}

	c, err := config.LoadDefaultConfig(context.TODO(), func(o *config.LoadOptions) error {
		o.Region = cfg.region
		if cfg.profile != "" {
			o.SharedConfigProfile = cfg.profile
		}
		return nil
	})
	if err != nil {
		panic(err)
	}
	eventbridgeSvc := eventbridge.NewFromConfig(c)
	cfg.eventbridge = eventbridgeSvc

	lambdaSvc := lambda.NewFromConfig(c)
	cfg.lambda = lambdaSvc

	return cfg, nil
}

func WithLogger(logger zerolog.Logger) EventPramsOptions {
	return func(config *EventPramsConfig) {
		config.log = logger
	}
}

func WithProfile(profile string) EventPramsOptions {
	return func(config *EventPramsConfig) {
		config.profile = profile
	}
}

func WithRegion(region string) EventPramsOptions {
	return func(config *EventPramsConfig) {
		config.region = region
	}
}

func (e *EventPramsConfig) PutRule(newEvent *NewEvent) (RuleArn, error) {
	// Disable the rule by default
	var enabled types.RuleState
	enabled = types.RuleStateDisabled
	if newEvent.State {
		enabled = types.RuleStateEnabled
	}

	putRuleResp, err := e.eventbridge.PutRule(context.TODO(), &eventbridge.PutRuleInput{
		Name:               aws.String(newEvent.Name),
		Description:        aws.String(newEvent.Description),
		ScheduleExpression: aws.String(newEvent.ScheduleExpression),
		State:              enabled,
		Tags: []types.Tag{
			{Key: aws.String("app"), Value: aws.String("mastopsot")},
			{Key: aws.String("feedname"), Value: aws.String(newEvent.Feedname)},
		},
	})
	if err != nil {
		return nil, err
	}

	_, err = e.lambda.AddPermission(context.TODO(), &lambda.AddPermissionInput{
		Action:       aws.String("lambda:InvokeFunction"),
		FunctionName: &newEvent.LambdaArn,
		Principal:    aws.String("events.amazonaws.com"),
		SourceArn:    putRuleResp.RuleArn,
		StatementId:  aws.String("Rule" + newEvent.Name + "InvokeLambdaFunction"),
	})
	if err != nil {
		return nil, err
	}

	putRuleTagetResp, err := e.eventbridge.PutTargets(context.TODO(), &eventbridge.PutTargetsInput{
		Rule: aws.String(newEvent.Name),
		Targets: []types.Target{
			{
				Arn:   aws.String(newEvent.LambdaArn),
				Id:    aws.String(newEvent.Name),
				Input: aws.String(fmt.Sprintf(`{"feed_name":"%s"}`, newEvent.Feedname)),
			},
		},
	})
	if err != nil {
		e.log.Error().
			Int32("FailedEntryCount", putRuleTagetResp.FailedEntryCount).
			Err(err).
			Str("FailedEntry", fmt.Sprintf("%v", putRuleTagetResp.FailedEntries)).
			Msg("Error adding target to rule")
		return nil, err
	}

	return putRuleResp.RuleArn, nil
}
