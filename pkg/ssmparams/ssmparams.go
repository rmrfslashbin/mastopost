package ssmparams

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/rs/zerolog"
)

type SSMParamsOutput struct {
	Params            map[string]interface{}
	InvalidParameters []string
}

type SSMParamsOption func(config *SSMParamsConfig)

type SSMParamsConfig struct {
	log     zerolog.Logger
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
	svc := ssm.NewFromConfig(c)
	cfg.ssm = svc

	return cfg, nil
}

func WithLogger(logger zerolog.Logger) SSMParamsOption {
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
		config.log.Error().Str("action", "ssmparams::GetParameters").Msg("error getting parameters")
		return nil, err
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
		config.log.Error().Str("action", "ssmparams::PutParam").Msg("error putting parameter")
		return nil, err
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
		config.log.Error().Str("action", "ssmparams::ListAllParams").Msg("error listing parameters")
		return nil, err
	}

	return resp, nil
}

func (config *SSMParamsConfig) DeleteParams(paramNames []string) (*ssm.DeleteParametersOutput, error) {
	resp, err := config.ssm.DeleteParameters(context.TODO(), &ssm.DeleteParametersInput{
		Names: paramNames,
	})
	if err != nil {
		config.log.Error().Str("action", "ssmparams::DeleteParams").Msg("error deleting parameters")
		return nil, err
	}

	return resp, nil
}
