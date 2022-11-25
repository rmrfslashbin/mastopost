package ssmparams

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"go.uber.org/zap"
)

type SSMParamsOutput struct {
	Params            map[string]interface{}
	InvalidParameters []string
}

type SSMParamsOption func(config *SSMParamsConfig)

type SSMParamsConfig struct {
	log     *zap.Logger
	region  string
	profile string
	ssm     *ssm.Client
}

func NewSSMParams(opts ...func(*SSMParamsConfig)) (*SSMParamsConfig, error) {
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

func SetLogger(logger *zap.Logger) SSMParamsOption {
	return func(config *SSMParamsConfig) {
		config.log = logger
	}
}

func SetProfile(profile string) SSMParamsOption {
	return func(config *SSMParamsConfig) {
		config.profile = profile
	}
}

func SetRegion(region string) SSMParamsOption {
	return func(config *SSMParamsConfig) {
		config.region = region
	}
}

func (config *SSMParamsConfig) GetParams(paramNames []string) (*SSMParamsOutput, error) {
	params, err := config.ssm.GetParameters(context.TODO(), &ssm.GetParametersInput{
		Names: paramNames,
	})
	if err != nil {
		config.log.Error("error getting parameters",
			zap.String("Action", "ssmparams::GetParameters"),
			zap.Error(err))
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

func (config *SSMParamsConfig) PutParams(params *ssm.PutParameterInput) (*ssm.PutParameterOutput, error) {
	resp, err := config.ssm.PutParameter(context.TODO(), params)
	if err != nil {
		config.log.Error("error putting parameters",
			zap.String("Action", "ssmparams::PutParameters"),
			zap.Error(err))
		return nil, err
	}
	return resp, nil
}
