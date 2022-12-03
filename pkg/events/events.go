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

type DeleteRuleInput struct {
	FunctionName *string
	FunctionArn  *string
	FeedName     *string
}

type RuleArn *string

type NewEvent struct {
	Name               string
	Description        string
	ScheduleExpression string
	State              bool
	Feedname           string
	LambdaArn          string
}

type EventDetails struct {
	Arn                string
	Description        string
	Name               string
	ScheduleExpression string
	State              bool
}

type EventPramsOptions func(config *EventPramsConfig)

type EventPramsConfig struct {
	log         *zerolog.Logger
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
		return nil, &AWSRegionRequiredError{}
	}

	c, err := config.LoadDefaultConfig(context.TODO(), func(o *config.LoadOptions) error {
		o.Region = cfg.region
		if cfg.profile != "" {
			o.SharedConfigProfile = cfg.profile
		}
		return nil
	})
	if err != nil {
		return nil, &AWSConfigError{Err: err}
	}
	eventbridgeSvc := eventbridge.NewFromConfig(c)
	cfg.eventbridge = eventbridgeSvc

	lambdaSvc := lambda.NewFromConfig(c)
	cfg.lambda = lambdaSvc

	return cfg, nil
}

func WithLogger(logger *zerolog.Logger) EventPramsOptions {
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

func (e *EventPramsConfig) DeleteRule(input *DeleteRuleInput) error {
	if _, err := e.lambda.RemovePermission(context.TODO(), &lambda.RemovePermissionInput{
		FunctionName: aws.String(*input.FunctionArn),
		StatementId:  aws.String("mastopost-" + *input.FeedName + "-InvokeLambdaFunction"),
	}); err != nil {
		return &RemovePermissionError{Err: err}
	}

	ruleName := "mastopost-" + *input.FeedName
	if opt, err := e.eventbridge.RemoveTargets(context.TODO(), &eventbridge.RemoveTargetsInput{
		Ids:   []string{ruleName},
		Rule:  aws.String(ruleName),
		Force: true,
	}); err != nil {
		return &RemoveTargetsError{
			Err:              err,
			FailedEntryCount: &opt.FailedEntryCount,
			FailedEntries:    &opt.FailedEntries,
		}
	}

	if _, err := e.eventbridge.DeleteRule(context.TODO(), &eventbridge.DeleteRuleInput{
		Name:  aws.String(ruleName),
		Force: true,
	}); err != nil {
		return &DeleteRuleError{Err: err}
	}

	return nil
}

func (e *EventPramsConfig) DisableRule(name string) error {
	if _, err := e.eventbridge.DisableRule(context.TODO(), &eventbridge.DisableRuleInput{
		Name: aws.String(name),
	}); err != nil {
		return &DisableRuleError{Err: err}
	}

	return nil
}

func (e *EventPramsConfig) EnableRule(name string) error {
	if _, err := e.eventbridge.EnableRule(context.TODO(), &eventbridge.EnableRuleInput{
		Name: aws.String(name),
	}); err != nil {
		return &EnableRuleError{Err: err}
	}

	return nil
}

func (e *EventPramsConfig) GetEventByName(name string) (*EventDetails, error) {
	resp, err := e.eventbridge.DescribeRule(context.TODO(), &eventbridge.DescribeRuleInput{
		Name: aws.String(name),
	})
	if err != nil {
		return nil, &DescribeRuleError{Err: err}
	}

	return &EventDetails{
		Arn:                *resp.Arn,
		Description:        *resp.Description,
		Name:               *resp.Name,
		ScheduleExpression: *resp.ScheduleExpression,
		State:              resp.State == types.RuleStateEnabled,
	}, nil
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
		return nil, &PutRuleError{Err: err}
	}

	_, err = e.lambda.AddPermission(context.TODO(), &lambda.AddPermissionInput{
		Action:       aws.String("lambda:InvokeFunction"),
		FunctionName: &newEvent.LambdaArn,
		Principal:    aws.String("events.amazonaws.com"),
		SourceArn:    putRuleResp.RuleArn,
		StatementId:  aws.String(newEvent.Name + "-InvokeLambdaFunction"),
	})
	if err != nil {
		return nil, &AddPermissionError{Err: err}
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
		return nil, &PutTargetsError{
			Err:              err,
			FailedEntryCount: &putRuleTagetResp.FailedEntryCount,
			FailedEntries:    &putRuleTagetResp.FailedEntries,
		}
	}

	return putRuleResp.RuleArn, nil
}
