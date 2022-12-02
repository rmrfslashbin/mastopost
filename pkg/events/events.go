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

// DescribeRuleError is an error returned when there is an error with the DescribeRule API call
type DescribeRuleError struct {
	Err error
}

// Error returns the error message
func (e *DescribeRuleError) Error() string {
	if e.Err == nil {
		return "EventBridge DescribeRule API call error"
	} else {
		return fmt.Sprintf("EventBridge DescribeRule API call error: %s", e.Err.Error())
	}
}

// PutRuleError is an error returned when there is an error with the PutRule call
type PutRuleError struct {
	Err error
	Msg string
}

// Error returns the error message
func (e *PutRuleError) Error() string {
	if e.Msg == "" {
		e.Msg = "error putting event bridge rule"
	}
	if e.Err != nil {
		e.Msg += ": " + e.Err.Error()
	}
	return e.Msg
}

// AddPermissionError is an error returned when there is an error with the AddPermission call
type AddPermissionError struct {
	Err error
	Msg string
}

// Error returns the error message
func (e *AddPermissionError) Error() string {
	if e.Msg == "" {
		e.Msg = "error adding IAM permission to Lambda function"
	}
	if e.Err != nil {
		e.Msg += ": " + e.Err.Error()
	}
	return e.Msg
}

// PutTargetsError is an error returned when there is an error with the PutTargets call
type PutTargetsError struct {
	Err              error
	Msg              string
	FailedEntryCount *int32
	FailedEntries    *[]types.PutTargetsResultEntry
}

// Error returns the error message
func (e *PutTargetsError) Error() string {
	if e.Msg == "" {
		e.Msg = "error adding event bridge rule target"
	}
	if e.FailedEntryCount != nil {
		e.Msg += fmt.Sprintf(": Failed Entry Count: %d", *e.FailedEntryCount)
	}
	if e.FailedEntries != nil {
		e.Msg += fmt.Sprintf(": Failed Entries: %v", e.FailedEntries)
	}
	if e.Err != nil {
		e.Msg += ": " + e.Err.Error()
	}
	return e.Msg
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
		StatementId:  aws.String("Rule-" + newEvent.Name + "-InvokeLambdaFunction"),
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
