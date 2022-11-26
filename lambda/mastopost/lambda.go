package main

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/rmrfslashbin/mastopost/pkg/mastoclient"
	"github.com/rmrfslashbin/mastopost/pkg/rssfeed"
	"github.com/rmrfslashbin/mastopost/pkg/ssmparams"
	"github.com/rmrfslashbin/mastopost/pkg/utils"
	"go.uber.org/zap"
)

var (
	aws_region string
	log        *zap.Logger
)

type Message struct {
	FeedName string `json:"feed_name"`
}

type Config struct {
	feedUrl       *url.URL
	instance      *url.URL
	clientID      string
	clientSec     string
	token         string
	lastUpdated   *time.Time
	lastPublished *time.Time
}

func init() {
	log, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}
	defer log.Sync()
	aws_region = os.Getenv("AWS_REGION")
}

func main() {
	// Catch errors
	var err error
	defer func() {
		if err != nil {
			log.Fatal("main crashed", zap.Error(err))
		}
	}()
	lambda.Start(handler)
}

func handler(ctx context.Context, message Message) error {
	if message.FeedName == "" {
		return errors.New("feed_name is required")
	}

	params, err := ssmparams.New(
		ssmparams.WithLogger(log),
		ssmparams.WithRegion(aws_region),
	)
	if err != nil {
		return err
	}

	path := "/mastopost/" + message.FeedName + "/"
	config := &Config{}
	var nextToken *string
	for {
		opt, err := params.ListAllParams(path, nextToken)
		if err != nil {
			return err
		}

		for _, p := range opt.Parameters {
			fmt.Printf("Name:    %s\n", *p.Name)
			fmt.Printf("Value:   %s\n", *p.Value)
			fmt.Printf("mtime:   %s\n", *p.LastModifiedDate)
			fmt.Printf("Version: %d\n", p.Version)
			fmt.Printf("ARN:     %s\n", *p.ARN)
			key := strings.TrimPrefix(path, *p.Name)
			switch key {
			case "mastodon/instanceUrl":
				if instanceUrl, err := url.Parse(*p.Value); err != nil {
					return err
				} else {
					config.instance = instanceUrl
				}
			case "mastodon/clientId":
				config.clientID = *p.Value
			case "mastodon/clientSecret":
				config.clientSec = *p.Value
			case "mastodon/accessToken":
				config.token = *p.Value
			case "rss/feedUrl":
				if feedUrl, err := url.Parse(*p.Value); err != nil {
					return err
				} else {
					config.feedUrl = feedUrl
				}
			case "rss/lastUpdated":
				if t, err := time.Parse(time.RFC3339, *p.Value); err == nil {
					config.lastUpdated = &t
				}
			case "mastodon/lastPublished":
				if t, err := time.Parse(time.RFC3339, *p.Value); err == nil {
					config.lastPublished = &t
				}
			}
		}

		nextToken = opt.NextToken
		if nextToken == nil {
			break
		}
	}

	// Set up a new feed parser
	feed, err := rssfeed.New(
		rssfeed.WithLogger(log),
		rssfeed.WithURL(config.feedUrl),
		rssfeed.WithLastUpdated(config.lastUpdated),
		rssfeed.WithLastPublished(config.lastPublished),
	)
	if err != nil {
		return err
	}

	// Get the new items
	newItems, err := feed.Parse()
	if err != nil {
		return err
	}

	if len(newItems) < 1 {
		log.Info("No new items in feed",
			zap.String("feed name", message.FeedName),
		)
		return nil
	}

	client, err := mastoclient.New(
		mastoclient.WithLogger(log),
		mastoclient.WithInstance(config.instance),
		mastoclient.WithClientID(config.clientID),
		mastoclient.WithClientSecret(config.clientSec),
		mastoclient.WithToken(config.token),
	)
	if err != nil {
		return err
	}

	for _, item := range newItems {
		// create a new post/toot
		newPost, err := utils.MakePost(item)
		if err != nil {
			return err
		}
		errCh := client.AsyncPost(newPost)
	}

	return nil
}
