package main

import (
	"fmt"
	"os"
	"path"

	"github.com/alecthomas/kong"
	"github.com/davecgh/go-spew/spew"
	"github.com/rmrfslashbin/mastopost/pkg/cmds/lambda"
	"github.com/rmrfslashbin/mastopost/pkg/cmds/oneshot"
	"github.com/rmrfslashbin/mastopost/pkg/config"
	"github.com/rs/zerolog"
)

var (
	// homeConfigDir is the location of the user's config directory
	homeConfigDir string

	// configFile is the location of the user's config file
	configFile string
)

const (
	// APP_NAME is the name of the application
	APP_NAME = "mastopost"

	// CONFIG_FILE is the name of the config file
	CONFIG_FILE = "config.json"
)

// Context is used to pass context/global configs to the commands
type Context struct {
	// log is the logger
	log *zerolog.Logger

	// configFile is the location of the config file
	configFile *string

	// homeConfigDir is the location of the user's config directory
	homeConfigDir *string
}

// CfCmd prints the current config
type CfgCmd struct {
}

// Run is the entry point for the cfg command
func (r *CfgCmd) Run(ctx *Context) error {
	fmt.Printf("Home config directory: %s/\n", *ctx.homeConfigDir)
	fmt.Printf("Config file location:  %s\n\n", *ctx.configFile)

	c, err := config.NewConfig(*ctx.configFile)
	if err != nil {
		return err
	}
	spew.Dump(c)

	return nil
}

// JobAddCmd adds a new job
type JobAddCmd struct {
	AWSProfile         string `name:"profile" help:"AWS profile to use" default:"default"`
	AWSRegion          string `name:"region" help:"AWS region to use" default:"us-east-1"`
	Confirm            bool   `name:"confirm" default:"false" help:"Confirm the job add (don't prompt for confirmation)"`
	Enable             bool   `name:"enable" default:"false" help:"Enable the job after adding it"`
	FeedName           string `name:"feedname" required:"" help:"Feed name to use"`
	LambdaFunctionName string `name:"lambdafn" required:"" help:"Lambda function name to use"`
}

// Run is the entry point for the job add command
func (r *JobAddCmd) Run(ctx *Context) error {
	l, err := lambda.NewLambda(
		lambda.WithAWSProfile(&r.AWSProfile),
		lambda.WithAWSRegion(&r.AWSRegion),
		lambda.WithConfigFile(ctx.configFile),
		lambda.WithFeedName(&r.FeedName),
		lambda.WithLambdaFunctionName(&r.LambdaFunctionName),
		lambda.WithLogger(ctx.log),
	)
	if err != nil {
		return err
	}
	return l.Add(&lambda.AddInput{
		Confirm: &r.Confirm,
		Enable:  &r.Enable,
	})
}

// JobDeleteCmd deletes a feed job
type JobDeleteRmd struct {
	AWSProfile         string `name:"profile" help:"AWS profile to use" default:"default"`
	AWSRegion          string `name:"region" help:"AWS region to use" default:"us-east-1"`
	Confirm            bool   `name:"confirm" default:"false" help:"Confirm the job add (don't prompt for confirmation)"`
	FeedName           string `name:"feedname" required:"" help:"Feed name to use"`
	LambdaFunctionName string `name:"lambdafn" required:"" help:"Lambda function name to use"`
}

// Run is the entry point for the job delete command
func (r *JobDeleteRmd) Run(ctx *Context) error {
	l, err := lambda.NewLambda(
		lambda.WithAWSProfile(&r.AWSProfile),
		lambda.WithAWSRegion(&r.AWSRegion),
		lambda.WithConfigFile(ctx.configFile),
		lambda.WithFeedName(&r.FeedName),
		lambda.WithLambdaFunctionName(&r.LambdaFunctionName),
		lambda.WithLogger(ctx.log),
	)
	if err != nil {
		return err
	}
	return l.Delete(r.Confirm)
}

// JobListCmd lists the jobs already set up
type JobListCmd struct {
	AWSProfile string  `name:"profile" help:"AWS profile to use" default:"default"`
	AWSRegion  string  `name:"region" help:"AWS region to use" default:"us-east-1"`
	FeedName   *string `name:"feedname" help:"Feed name to use"`
}

// Run is the entry point for the job list command
func (r *JobListCmd) Run(ctx *Context) error {
	l, err := lambda.NewLambda(
		lambda.WithLogger(ctx.log),
		lambda.WithAWSProfile(&r.AWSProfile),
		lambda.WithAWSRegion(&r.AWSRegion),
		lambda.WithFeedName(r.FeedName),
	)
	if err != nil {
		return err
	}
	return l.List()
}

// JobStatusCmd prints the status of a job
type JobStatusCmd struct {
	AWSProfile string `name:"profile" help:"AWS profile to use" default:"default"`
	AWSRegion  string `name:"region" help:"AWS region to use" default:"us-east-1"`
	FeedName   string `name:"feedname" required:"" help:"Feed name to use"`
	Enable     *bool  `name:"enable" xor:"Change State" group:"Change State" help:"Enable the job after adding it"`
	Disable    *bool  `name:"disable" xor:"Change State" group:"Change State" help:"Disable the job after adding it"`
}

// Run is the entry point for the job status command
func (r *JobStatusCmd) Run(ctx *Context) error {
	l, err := lambda.NewLambda(
		lambda.WithLogger(ctx.log),
		lambda.WithAWSProfile(&r.AWSProfile),
		lambda.WithAWSRegion(&r.AWSRegion),
		lambda.WithFeedName(&r.FeedName),
		lambda.WithConfigFile(ctx.configFile),
	)
	if err != nil {
		return err
	}
	return l.Status(&lambda.StatusInput{Enable: r.Enable, Disable: r.Disable})
}

// LambdaInstallCmd installs a new lambda function
type LambdaInstallCmd struct {
	AWSProfile   string `name:"profile" help:"AWS profile to use" default:"default"`
	AWSRegion    string `name:"region" help:"AWS region to use" default:"us-east-1"`
	FunctionName string `name:"functionname" required:"" help:"Lambda function name to use"`
	ZipFile      string `name:"zipfile" required:"" existingfile:"" help:"Zip file to use"`
}

// Run is the entry point for the lambda install command
func (r *LambdaInstallCmd) Run(ctx *Context) error {
	l, err := lambda.NewLambda(
		lambda.WithLogger(ctx.log),
		lambda.WithAWSProfile(&r.AWSProfile),
		lambda.WithAWSRegion(&r.AWSRegion),
		lambda.WithConfigFile(ctx.configFile),
		lambda.WithLambdaFunctionName(&r.FunctionName),
		lambda.WithZipFilename(&r.ZipFile),
	)
	if err != nil {
		return err
	}
	return l.Install()
}

// LambdaUninstallCmd uninstalls a lambda function
type LambdaUninstallCmd struct {
	AWSProfile   string `name:"profile" help:"AWS profile to use" default:"default"`
	AWSRegion    string `name:"region" help:"AWS region to use" default:"us-east-1"`
	FunctionName string `name:"functionname" required:"" help:"Lambda function name to use"`
}

// Run is the entry point for the lambda uninstall command
func (r *LambdaUninstallCmd) Run(ctx *Context) error {
	l, err := lambda.NewLambda(
		lambda.WithLogger(ctx.log),
		lambda.WithAWSProfile(&r.AWSProfile),
		lambda.WithAWSRegion(&r.AWSRegion),
		lambda.WithConfigFile(ctx.configFile),
		lambda.WithLambdaFunctionName(&r.FunctionName),
	)
	if err != nil {
		return err
	}
	return l.Uninstall()
}

// OneshotCmd runs a single instance of the oneshot command to parse and post RSS feeds to Mastodon
type OneshotCmd struct {
	DryRun   bool   `name:"dryrun" help:"Don't actually post to Mastodon."`
	Feedname string `name:"feedname" env:"FEED_NAME" required:"" help:"Name of the feed to post."`
}

// Run is the entry point for the oneshot command
func (r *OneshotCmd) Run(ctx *Context) error {
	// Set up a new oneshot struct
	if foo, err := oneshot.NewOneshot(
		oneshot.WithLogger(ctx.log),
		oneshot.WithConfigFile(ctx.configFile),
		oneshot.WithFeedName(&r.Feedname),
		oneshot.WithDryrun(r.DryRun),
	); err != nil {
		return err
	} else {
		// Run the oneshot
		return foo.Run()
	}

}

// CLI is the main CLI struct
type CLI struct {
	// Global flags/args
	LogLevel   string  `name:"loglevel" env:"LOGLEVEL" default:"info" enum:"panic,fatal,error,warn,info,debug,trace" help:"Set the log level."`
	ConfigFile *string `name:"config" env:"CONFIG_FILE" help:"Path to the config file."`

	// Cfg commmand
	Cfg CfgCmd `cmd:"" help:"Show Mastopost config details."`

	// Job commands
	Job struct {
		Add    JobAddCmd    `cmd:"" help:"Add a new Mastopost job."`
		Delete JobDeleteRmd `cmd:"" help:"Deletes a Mastopost job."`
		List   JobListCmd   `cmd:"" help:"List Mastopost jobs."`
		Status JobStatusCmd `cmd:"" help:"Show status of Mastopost jobs."`
	} `cmd:"" help:"Manages jobs/events"`

	// Lambda commands
	Lambda struct {
		Install   LambdaInstallCmd   `cmd:"" help:"Install a new Mastopost lambda function."`
		Uninstall LambdaUninstallCmd `cmd:"" help:"Uninstall a Mastopost lambda function."`
	} `cmd:"" help:"Manage the Lambda functions and events."`

	// Oneshot command
	Oneshot OneshotCmd `cmd:"" help:"Run an RSS feed parser and post to Mastodon."`
}

// main is the entry point for the CLI
func main() {
	var err error

	// Set up the logger
	log := zerolog.New(os.Stderr).With().Timestamp().Logger()

	// Find home directory.
	homeConfigDir, err = os.UserConfigDir()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to find user config directory")
	}

	// Set up the home dir and config file locations
	homeConfigDir = path.Join(homeConfigDir, APP_NAME)
	configFile = path.Join(homeConfigDir, CONFIG_FILE)

	// Parse the command line
	var cli CLI
	ctx := kong.Parse(&cli)

	// Set up the logger's log level
	// Default to info via the CLI args
	switch cli.LogLevel {
	case "panic":
		log = log.Level(zerolog.PanicLevel)
	case "fatal":
		log = log.Level(zerolog.FatalLevel)
	case "error":
		log = log.Level(zerolog.ErrorLevel)
	case "warn":
		log = log.Level(zerolog.WarnLevel)
	case "info":
		log = log.Level(zerolog.InfoLevel)
	case "debug":
		log = log.Level(zerolog.DebugLevel)
	case "trace":
		log = log.Level(zerolog.TraceLevel)
	}

	// Set up the config file if not provided
	if cli.ConfigFile == nil {
		log.Debug().
			Str("configfile", configFile).
			Msg("Using default config file")
		cli.ConfigFile = &configFile
	}

	// Log some start up stuff for debugging
	log.Debug().Msg("Starting up")
	log.Debug().
		Str("configfile", *cli.ConfigFile).
		Str("homeconfigdir", homeConfigDir).
		Msg("config paths/files")

	// Call the Run() method of the selected parsed command.
	err = ctx.Run(&Context{log: &log, configFile: cli.ConfigFile, homeConfigDir: &homeConfigDir})

	// FatalIfErrorf terminates with an error message if err != nil
	ctx.FatalIfErrorf(err)
}
