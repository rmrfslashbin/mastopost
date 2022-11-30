package main

import (
	"fmt"
	"os"
	"path"

	"github.com/alecthomas/kong"
	"github.com/rmrfslashbin/mastopost/pkg/cmds/oneshot"
	"github.com/rs/zerolog"
)

type Context struct {
	log        *zerolog.Logger
	configFile *string
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

type AddCmd struct {
	Paths []string `arg:"" optional:"" name:"path" help:"Paths to list." type:"path"`
}

func (r *AddCmd) Run(ctx *Context) error {
	fmt.Println("add ", r.Paths)
	return nil
}

type RemoveRmd struct {
	Paths []string `arg:"" optional:"" name:"path" help:"Paths to list." type:"path"`
}

type ListCmd struct {
	Paths []string `arg:"" optional:"" name:"path" help:"Paths to list." type:"path"`
}

type StatusCmd struct {
	Paths []string `arg:"" optional:"" name:"path" help:"Paths to list." type:"path"`
}

type CLI struct {
	LogLevel   string  `name:"loglevel" env:"LOGLEVEL" default:"info" enum:"panic,fatal,error,warn,info,debug,trace" help:"Set the log level."`
	ConfigFile *string `name:"configfile" env:"CONFIG_FILE" help:"Path to the config file."`

	Oneshot OneshotCmd `cmd:"" help:"Run an RSS feed parser and post to Mastodon."`
	Add     AddCmd     `cmd:"" help:"Add a new Mastopost job."`
	Remove  RemoveRmd  `cmd:"" help:"Remove a Mastopost job."`
	List    ListCmd    `cmd:"" help:"List Mastopost jobs."`
	Status  StatusCmd  `cmd:"" help:"Show status of Mastopost jobs."`
}

var (
	homeConfigDir string
	configFile    string
)

const (
	APP_NAME    = "mastopost"
	CONFIG_FILE = "config.json"
)

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
	err = ctx.Run(&Context{log: &log, configFile: cli.ConfigFile})
	ctx.FatalIfErrorf(err)
}
