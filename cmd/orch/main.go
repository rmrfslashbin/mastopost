package main

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge/types"
	"github.com/rmrfslashbin/mastopost/pkg/events"
	"go.uber.org/zap"
)

var (
	log *zap.Logger
)

func init() {
	// Set up the logger
	log, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}
	defer log.Sync()
}

func main() {
	eb, err := events.New(
		events.WithLogger(log),
		events.WithProfile("default"),
		events.WithRegion("us-east-1"),
	)
	if err != nil {
		log.Fatal("error creating eventbridge client", zap.Error(err))
	}
	arn, err := eb.PutRule(&eventbridge.PutRuleInput{
		Name:               aws.String("test"),
		Description:        aws.String("Testing for mastopost orchestrtion"),
		ScheduleExpression: aws.String("rate(10 minutes)"),
		State:              types.RuleStateEnabled,
		Tags: []types.Tag{
			{Key: aws.String("app"), Value: aws.String("mastopsot")},
			{Key: aws.String("feedname"), Value: aws.String("orch_test")},
		},
	})
	if err != nil {
		log.Fatal("error putting rule", zap.Error(err))
	}
	log.Info("rule arn", zap.String("arn", *arn))
}
