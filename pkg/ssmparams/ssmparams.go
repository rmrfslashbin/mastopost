package ssmparams

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
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

// DeleteParametersError is an error returned when there is an error with the DeleteParameters call
type DeleteParametersError struct {
	Err error
	Msg string
}

// Error returns the error message
func (e *DeleteParametersError) Error() string {
	if e.Msg == "" {
		e.Msg = "error deleting AWS parameters"
	}
	if e.Err != nil {
		e.Msg += ": " + e.Err.Error()
	}
	return e.Msg
}

// GetParametersError is an error returned when there is an error with the GetParameters call
type GetParametersError struct {
	Err error
	Msg string
}

// Error returns the error message
func (e *GetParametersError) Error() string {
	if e.Msg == "" {
		e.Msg = "error getting AWS parameters"
	}
	if e.Err != nil {
		e.Msg += ": " + e.Err.Error()
	}
	return e.Msg
}

// GetParametersByPathError is an error returned when there is an error with the GetParametersByPath call
type GetParametersByPathError struct {
	Err error
	Msg string
}

// Error returns the error message
func (e *GetParametersByPathError) Error() string {
	if e.Msg == "" {
		e.Msg = "error getting AWS parameters by path"
	}
	if e.Err != nil {
		e.Msg += ": " + e.Err.Error()
	}
	return e.Msg
}

// PutParameterError is an error returned when there is an error with the PutParameter call
type PutParameterError struct {
	Err error
	Msg string
}

// Error returns the error message
func (e *PutParameterError) Error() string {
	if e.Msg == "" {
		e.Msg = "error putting AWS parameter"
	}
	if e.Err != nil {
		e.Msg += ": " + e.Err.Error()
	}
	return e.Msg
}

type SSMParamsOutput struct {
	Params            map[string]interface{}
	InvalidParameters []string
}

type SSMParamsOption func(config *SSMParamsConfig)

type SSMParamsConfig struct {
	log     *zerolog.Logger
	region  string
	profile string
	ssm     *ssm.Client
}

func New(opts ...func(*SSMParamsConfig)) (*SSMParamsConfig, error) {
	cfg := &SSMParamsConfig{}

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
	svc := ssm.NewFromConfig(c)
	cfg.ssm = svc

	return cfg, nil
}

func WithLogger(logger *zerolog.Logger) SSMParamsOption {
	return func(config *SSMParamsConfig) {
		config.log = logger
	}
}

func WithProfile(profile string) SSMParamsOption {
	return func(config *SSMParamsConfig) {
		config.profile = profile
	}
}

func WithRegion(region string) SSMParamsOption {
	return func(config *SSMParamsConfig) {
		config.region = region
	}
}

func (config *SSMParamsConfig) GetParams(paramNames []string) (*SSMParamsOutput, error) {
	params, err := config.ssm.GetParameters(context.TODO(), &ssm.GetParametersInput{
		Names: paramNames,
	})
	if err != nil {
		return nil, &GetParametersError{Err: err}
	}
	output := make(map[string]interface{}, len(params.Parameters))

	for _, v := range params.Parameters {
		output[*v.Name] = *v.Value
	}
	return &SSMParamsOutput{
		Params:            output,
		InvalidParameters: params.InvalidParameters,
	}, nil
}

func (config *SSMParamsConfig) PutParam(params *ssm.PutParameterInput) (*ssm.PutParameterOutput, error) {
	resp, err := config.ssm.PutParameter(context.TODO(), params)
	if err != nil {
		return nil, &PutParameterError{Err: err}
	}
	return resp, nil
}

func (config *SSMParamsConfig) ListAllParams(path string, nextToken *string) (*ssm.GetParametersByPathOutput, error) {
	resp, err := config.ssm.GetParametersByPath(context.TODO(), &ssm.GetParametersByPathInput{
		Path:      aws.String(path),
		Recursive: aws.Bool(true),
		NextToken: nextToken,
	})
	if err != nil {
		return nil, &GetParametersByPathError{Err: err}
	}

	return resp, nil
}

func (config *SSMParamsConfig) DeleteParams(paramNames []string) (*ssm.DeleteParametersOutput, error) {
	resp, err := config.ssm.DeleteParameters(context.TODO(), &ssm.DeleteParametersInput{
		Names: paramNames,
	})
	if err != nil {
		return nil, &DeleteParametersError{Err: err}
	}

	return resp, nil
}
