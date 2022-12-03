package lambda

import (
	"os"

	"github.com/iancoleman/strcase"
	"github.com/rs/zerolog"
)

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
	zipfilename        *string
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

// WithZipFilename sets the zip filename to use
func WithZipFilename(zipfilename *string) LambdaOptions {
	return func(config *LambdaConfig) {
		config.zipfilename = zipfilename
	}
}
