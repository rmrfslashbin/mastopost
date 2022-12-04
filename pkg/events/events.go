package events

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge/types"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	lambdaTypes "github.com/aws/aws-sdk-go-v2/service/lambda/types"
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
	iam         *iam.Client
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

	iamSvc := iam.NewFromConfig(c)
	cfg.iam = iamSvc

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

// OpenLambdaZipError is an error that occurs when the zip file cannot be opened.
type OpenLambdaZipError struct {
	Err      error
	Msg      string
	Filename string
}

// Error returns the error message.
func (e *OpenLambdaZipError) Error() string {
	if e.Msg == "" {
		e.Msg = "error opening lambda zip file"
	}
	if e.Filename != "" {
		e.Msg += " (" + e.Filename + ")"
	}
	if e.Err != nil {
		e.Msg += ": " + e.Err.Error()
	}
	return e.Msg
}

// ReadLambdaZipError is an error that occurs when the zip file cannot be read.
type ReadLambdaZipError struct {
	Err      error
	Msg      string
	Filename string
}

// Error returns the error message.
func (e *ReadLambdaZipError) Error() string {
	if e.Msg == "" {
		e.Msg = "error reading lambda zip file"
	}
	if e.Filename != "" {
		e.Msg += " (" + e.Filename + ")"
	}
	if e.Err != nil {
		e.Msg += ": " + e.Err.Error()
	}
	return e.Msg
}

// CreateFunctionError is an error that occurs when the lambda function cannot be created.
type CreateFunctionError struct {
	Err error
	Msg string
}

// Error returns the error message.
func (e *CreateFunctionError) Error() string {
	if e.Msg == "" {
		e.Msg = "error creating lambda function"
	}
	if e.Err != nil {
		e.Msg += ": " + e.Err.Error()
	}
	return e.Msg
}

// CreateRoleError is an error that occurs when the lambda role cannot be created.
type CreateRoleError struct {
	Err error
	Msg string
}

// Error returns the error message.
func (e *CreateRoleError) Error() string {
	if e.Msg == "" {
		e.Msg = "error creating lambda role"
	}
	if e.Err != nil {
		e.Msg += ": " + e.Err.Error()
	}
	return e.Msg
}

// AttachRolePolicyError is an error that occurs when the lambda role policy cannot be attached.
type AttachRolePolicyError struct {
	Err error
	Msg string
}

// Error returns the error message.
func (e *AttachRolePolicyError) Error() string {
	if e.Msg == "" {
		e.Msg = "error attaching role policy"
	}
	if e.Err != nil {
		e.Msg += ": " + e.Err.Error()
	}
	return e.Msg
}

// CreatePolicyError is an error that occurs when the lambda policy cannot be created.
type CreatePolicyError struct {
	Err error
	Msg string
}

// Error returns the error message.
func (e *CreatePolicyError) Error() string {
	if e.Msg == "" {
		e.Msg = "error creating policy"
	}
	if e.Err != nil {
		e.Msg += ": " + e.Err.Error()
	}
	return e.Msg
}

// InstallLambdaFunctionError is an error that occurs when the lambda function cannot be installed.
type InstallLambdaFunctionError struct {
	Err error
	Msg string
}

// Error returns the error message.
func (e *InstallLambdaFunctionError) Error() string {
	if e.Msg == "" {
		e.Msg = "error installing lambda function"
	}
	if e.Err != nil {
		e.Msg += ": " + e.Err.Error()
	}
	return e.Msg
}

// DeleteFunctionError is an error that occurs when the lambda function cannot be deleted.
type DeleteFunctionError struct {
	Err error
	Msg string
}

// Error returns the error message.
func (e *DeleteFunctionError) Error() string {
	if e.Msg == "" {
		e.Msg = "error deleting lambda function"
	}
	if e.Err != nil {
		e.Msg += ": " + e.Err.Error()
	}
	return e.Msg
}

// DeletePolicyError is an error that occurs when the lambda policy cannot be deleted.
type DeletePolicyError struct {
	Err error
	Msg string
}

// Error returns the error message.
func (e *DeletePolicyError) Error() string {
	if e.Msg == "" {
		e.Msg = "error deleting policy"
	}
	if e.Err != nil {
		e.Msg += ": " + e.Err.Error()
	}
	return e.Msg
}

// UninstallLambdaFunctionError is an error that occurs when the lambda function cannot be uninstalled.
type UninstallLambdaFunctionError struct {
	Err error
	Msg string
}

// Error returns the error message.
func (e *UninstallLambdaFunctionError) Error() string {
	if e.Msg == "" {
		e.Msg = "error uninstalling lambda function"
	}
	if e.Err != nil {
		e.Msg += ": " + e.Err.Error()
	}
	return e.Msg
}

type InstallLambdaFunctionInput struct {
	FunctionZipFilename *string
	FunctionName        *string
}

type InstallLambdaFunctionOutput struct {
	FunctionArn  string
	FunctionName string
	PolicyArn    string
}

func (e *EventPramsConfig) InstallLambdaFunction(input *InstallLambdaFunctionInput) (*InstallLambdaFunctionOutput, error) {
	if input.FunctionZipFilename == nil {
		return nil, &InstallLambdaFunctionError{Msg: "function zip filename is required"}
	}

	if input.FunctionName == nil {
		return nil, &InstallLambdaFunctionError{Msg: "function name is required"}
	}

	fh, err := os.Open(*input.FunctionZipFilename)
	if err != nil {
		return nil, &OpenLambdaZipError{Err: err, Filename: *input.FunctionZipFilename}
	}
	defer fh.Close()
	zipfile, err := ioutil.ReadAll(fh)
	if err != nil {
		return nil, &ReadLambdaZipError{Err: err}
	}
	e.log.Info().
		Str("zip file name", *input.FunctionZipFilename).
		Msg("lambda zip file read")

	roleName := "role-mastopost-lambda-" + *input.FunctionName
	role, err := e.iam.CreateRole(context.TODO(), &iam.CreateRoleInput{
		AssumeRolePolicyDocument: aws.String(`{"Version": "2012-10-17","Statement": [{"Effect": "Allow","Principal": {"Service": "lambda.amazonaws.com"},"Action": "sts:AssumeRole"}]}`),
		RoleName:                 aws.String(roleName),
		Description:              aws.String("Role for mastopost lambda function: " + *input.FunctionName),
	})
	if err != nil {
		return nil, &CreateRoleError{Err: err}
	}
	e.log.Info().
		Str("role name", roleName).
		Msg("iam role created")

	policy, err := e.iam.CreatePolicy(context.TODO(), &iam.CreatePolicyInput{
		Description:    aws.String("Policy for mastopost lambda function: " + *input.FunctionName),
		PolicyDocument: aws.String(`{"Version": "2012-10-17","Statement": [{"Action": ["ssm:GetParameter","ssm:GetParameters","ssm:GetParametersByPath","ssm:PutParameter"],"Resource": "arn:aws:ssm:us-east-1:150319663043:parameter/mastopost/*","Effect": "Allow"}]}`),
		PolicyName:     aws.String("policy-mastopost-lambda-" + *input.FunctionName),
	})
	if err != nil {
		return nil, &CreatePolicyError{Err: err}
	}

	e.log.Info().
		Str("policy arn", *policy.Policy.Arn).
		Msg("iam policy created")

	if _, err := e.iam.AttachRolePolicy(context.TODO(), &iam.AttachRolePolicyInput{
		PolicyArn: policy.Policy.Arn,
		RoleName:  aws.String(roleName),
	}); err != nil {
		return nil, &AttachRolePolicyError{Err: err}
	}
	e.log.Info().Msg("iam policy attached to role")

	if _, err := e.iam.AttachRolePolicy(context.TODO(), &iam.AttachRolePolicyInput{
		PolicyArn: aws.String("arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"),
		RoleName:  aws.String(roleName),
	}); err != nil {
		return nil, &AttachRolePolicyError{Err: err}
	}
	e.log.Info().Msg("AWSLambdaBasicExecutionRole iam policy attached to role")

	opt, err := e.lambda.CreateFunction(context.TODO(), &lambda.CreateFunctionInput{
		Code: &lambdaTypes.FunctionCode{
			ZipFile: zipfile,
		},
		FunctionName: aws.String("mastopost-" + *input.FunctionName),
		Role:         role.Role.Arn,
		Architectures: []lambdaTypes.Architecture{
			lambdaTypes.ArchitectureArm64,
		},
		Description: aws.String("Mastopost lambda function: " + *input.FunctionName),
		Handler:     aws.String("bootstrap"),
		PackageType: lambdaTypes.PackageTypeZip,
		Publish:     true,
		Runtime:     lambdaTypes.RuntimeProvidedal2,
		Timeout:     aws.Int32(30),
	})
	if err != nil {
		return nil, &CreateFunctionError{Err: err}
	}
	e.log.Info().
		Str("function name", *opt.FunctionName).
		Str("function arn", *opt.FunctionArn).
		Msg("lambda function created")

	return &InstallLambdaFunctionOutput{
		FunctionArn:  *opt.FunctionArn,
		FunctionName: *opt.FunctionName,
		PolicyArn:    *policy.Policy.Arn,
	}, nil
}

type UninstallLambdaFunctionInput struct {
	FunctionNameArn *string
	PolicyArn       *string
}

func (e *EventPramsConfig) UninstallLambdaFunction(input *UninstallLambdaFunctionInput) error {
	if input.FunctionNameArn == nil {
		return &UninstallLambdaFunctionError{Msg: "function arn is required"}
	}

	if input.PolicyArn == nil {
		return &UninstallLambdaFunctionError{Msg: "policy arn is required"}
	}

	// functionName := mastopost-rss-xpost-test

	if _, err := e.lambda.DeleteFunction(context.TODO(), &lambda.DeleteFunctionInput{
		FunctionName: input.FunctionNameArn,
	}); err != nil {
		return &DeleteFunctionError{Err: err}
	}

	var marker *string
	for {
		opt, err := e.iam.ListEntitiesForPolicy(context.TODO(), &iam.ListEntitiesForPolicyInput{
			PolicyArn: input.PolicyArn,
			Marker:    marker,
		})
		if err != nil {
			return err
		}
		for _, role := range opt.PolicyRoles {
			if _, err := e.iam.DetachRolePolicy(context.TODO(), &iam.DetachRolePolicyInput{
				PolicyArn: input.PolicyArn,
				RoleName:  role.RoleName,
			}); err != nil {
				return err
			}
			if _, err := e.iam.DeleteRole(context.TODO(), &iam.DeleteRoleInput{
				RoleName: role.RoleName,
			}); err != nil {
				return err
			}
		}

		marker = opt.Marker
		if marker == nil {
			break
		}
	}

	if _, err := e.iam.DeletePolicy(context.TODO(), &iam.DeletePolicyInput{
		PolicyArn: input.PolicyArn,
	}); err != nil {
		return &DeletePolicyError{Err: err}
	}

	return nil
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
