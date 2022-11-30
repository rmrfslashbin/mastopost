package main

import (
	"fmt"
	"os"
	"path"

	"github.com/alecthomas/kong"
	"github.com/davecgh/go-spew/spew"
	"github.com/rmrfslashbin/mastopost/pkg/cmds/oneshot"
	"github.com/rmrfslashbin/mastopost/pkg/config"
	"github.com/rs/zerolog"
)

var (
	homeConfigDir string
	configFile    string
)

const (
	APP_NAME    = "mastopost"
	CONFIG_FILE = "config.json"
)

type Context struct {
	log           *zerolog.Logger
	configFile    *string
	homeConfigDir *string
}

type CfgCmd struct {
}

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

type LambdaAddCmd struct {
}

func (r *LambdaAddCmd) Run(ctx *Context) error {
	fmt.Println("add ")
	return nil
}

type LambdaInstallCmd struct {
}

func (r *LambdaInstallCmd) Run(ctx *Context) error {
	fmt.Println("install ")
	return nil
}

type LambdaListCmd struct {
}

func (r *LambdaListCmd) Run(ctx *Context) error {
	fmt.Println("list ")
	return nil
}

type LambdaRemoveRmd struct {
}

func (r *LambdaRemoveRmd) Run(ctx *Context) error {
	fmt.Println("remove ")
	return nil
}

type LambdaStatusCmd struct {
}

func (r *LambdaStatusCmd) Run(ctx *Context) error {
	fmt.Println("status ")
	return nil
}

type OneshotCmd struct {
	DryRun   bool   `name:"dryrun" help:"Don't actually post to Mastodon."`
	Feedname string `name:"feedname" env:"FEED_NAME" required:"" help:"Name of the feed to post."`
}

func (r *OneshotCmd) Run(ctx *Context) error {
	if foo, err := oneshot.NewOneshot(
		oneshot.WithLogger(ctx.log),
		oneshot.WithConfigFile(ctx.configFile),
		oneshot.WithFeedName(&r.Feedname),
		oneshot.WithDryrun(r.DryRun),
	); err != nil {
		return err
	} else {
		return foo.Run()
	}

}

type CLI struct {
	LogLevel   string  `name:"loglevel" env:"LOGLEVEL" default:"info" enum:"panic,fatal,error,warn,info,debug,trace" help:"Set the log level."`
	ConfigFile *string `name:"config" env:"CONFIG_FILE" help:"Path to the config file."`

	Oneshot OneshotCmd `cmd:"" help:"Run an RSS feed parser and post to Mastodon."`
	Lambda  struct {
		Add    LambdaAddCmd    `cmd:"" help:"Add a new Mastopost job."`
		Remove LambdaRemoveRmd `cmd:"" help:"Remove a Mastopost job."`
		List   LambdaListCmd   `cmd:"" help:"List Mastopost jobs."`
		Status LambdaStatusCmd `cmd:"" help:"Show status of Mastopost jobs."`
	} `cmd:"" help:"Manage the Lambda functions and events."`

	Cfg CfgCmd `cmd:"" help:"Show Mastopost config details."`
}

func main() {
	var err error
	// Set up the logger
	log := zerolog.New(os.Stderr).With().Timestamp().Logger()

	// Find home directory.
	homeConfigDir, err = os.UserConfigDir()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to find user config directory")
	}

	homeConfigDir = path.Join(homeConfigDir, APP_NAME)
	configFile = path.Join(homeConfigDir, CONFIG_FILE)

	var cli CLI
	ctx := kong.Parse(&cli)

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

	if cli.ConfigFile == nil {
		log.Debug().
			Str("configfile", configFile).
			Msg("Using default config file")
		cli.ConfigFile = &configFile
	}

	log.Debug().Msg("Starting up")
	log.Debug().
		Str("configfile", *cli.ConfigFile).
		Str("homeconfigdir", homeConfigDir).
		Msg("config paths/files")

	// Call the Run() method of the selected parsed command.
	err = ctx.Run(&Context{log: &log, configFile: cli.ConfigFile, homeConfigDir: &homeConfigDir})
	ctx.FatalIfErrorf(err)
}
