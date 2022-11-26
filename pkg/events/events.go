package events

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge"
	"go.uber.org/zap"
)

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
	log         *zap.Logger
	region      string
	profile     string
	eventbridge *eventbridge.Client
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
	svc := eventbridge.NewFromConfig(c)
	cfg.eventbridge = svc

	return cfg, nil
}

func WithLogger(logger *zap.Logger) EventPramsOptions {
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

func (e *EventPramsConfig) PutRule(ruleInput *eventbridge.PutRuleInput) (*string, error) {
	putRuleResp, err := e.eventbridge.PutRule(context.TODO(), ruleInput)
	if err != nil {
		return nil, err
	}

	/*
		ruleName := *ruleInput.Name + "_target"
		putRuleTagetResp, err := e.eventbridge.PutTargets(context.TODO(), &eventbridge.PutTargetsInput{
			Rule: aws.String(ruleName),
			Targets: []eventbridge.Target{
		})
		if err != nil {
			return nil, err
		}
	*/

	return putRuleResp.RuleArn, nil
}
