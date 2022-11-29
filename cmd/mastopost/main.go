package main

import (
	"fmt"
	"net/url"
	"os"
	"path"

	"github.com/alecthomas/kong"
	"github.com/rmrfslashbin/mastopost/pkg/cmds/oneshot"
	"github.com/rs/zerolog"
)

type Context struct {
	log *zerolog.Logger
}

type OneshotCmd struct {
	LastUpdateFile *os.File `type:"existingfile" help:"File to store the last update time."`
	NoConfirm      bool     `name:"noconfirm" help:"Don't prompt for confirmation."`
	DryRun         bool     `name:"dryrun" help:"Don't actually post to Mastodon."`
	Feedname       string   `name:"feedname" env:"FEED_NAME" required:"" help:"Name of the feed to post."`
	StatsFile      string   `name:"statsfile" env:"STATS_FILE" required:"" type:"existingfile" help:"Path to the stats file."`
	FeedURL        *url.URL `name:"feedurl" env:"FEED_URL" required:"" help:"URL of the feed to post."`
	Instance       *url.URL `name:"instance" env:"INSTANCE" required:"" help:"URL of the Mastodon instance."`
	ClientId       string   `name:"clientid" env:"CLIENT_ID" required:"" help:"Mastodon client ID."`
	ClientSecret   string   `name:"clientsecret" env:"CLIENT_SECRET" required:"" help:"Mastodon client secret."`
	AccessToken    string   `name:"accesstoken" env:"ACCESS_TOKEN" required:"" help:"Mastodon access token."`
}

func (r *OneshotCmd) Run(ctx *Context) error {
	if foo, err := oneshot.NewOneshot(
		oneshot.WithLogger(ctx.log),
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
	LogLevel   string   `name:"loglevel" default:"info" enum:"panic,fatal,error,warn,info,debug,trace" help:"Set the log level."`
	ConfigFile *os.File `name:"configfile" env:"CONFIG_FILE" type:"existingfile" help:"Path to the config file."`

	Oneshot OneshotCmd `cmd:"" help:"Run an RSS feed parser and post to Mastodon."`
	Add     AddCmd     `cmd:"" help:"Add a new Mastopost job."`
	Remove  RemoveRmd  `cmd:"" help:"Remove a Mastopost job."`
	List    ListCmd    `cmd:"" help:"List Mastopost jobs."`
	Status  StatusCmd  `cmd:"" help:"Show status of Mastopost jobs."`
}

var (
	homeConfigDir  string
	configFile     string
	lastUpdateFile string
)

const (
	APP_NAME        = "mastopost"
	CONFIG_FILE     = "config.json"
	LAST_UPATE_FILE = "lastupdate.gob"
)

func init() {

}

func main() {
	// Set up the logger
	log := zerolog.New(os.Stderr).With().Timestamp().Logger()

	// Find home directory.
	homeConfigDir, err := os.UserConfigDir()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to find user config directory")
	}

	homeConfigDir = path.Join(homeConfigDir, APP_NAME)
	configFile = path.Join(homeConfigDir, CONFIG_FILE)
	lastUpdateFile = path.Join(homeConfigDir, LAST_UPATE_FILE)

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

	log.Debug().Msg("Starting up")
	log.Debug().
		Str("configfile", configFile).
		Str("lastupdatefile", lastUpdateFile).
		Str("homeconfigdir", homeConfigDir).
		Msg("config paths/files")

	// Call the Run() method of the selected parsed command.
	err = ctx.Run(&Context{log: &log})
	ctx.FatalIfErrorf(err)
}
