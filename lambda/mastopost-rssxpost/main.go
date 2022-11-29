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
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/aws/aws-sdk-go-v2/service/ssm/types"
	"github.com/mattn/go-mastodon"
	"github.com/rmrfslashbin/mastopost/pkg/mastoclient"
	"github.com/rmrfslashbin/mastopost/pkg/rssfeed"
	"github.com/rmrfslashbin/mastopost/pkg/ssmparams"
	"github.com/rmrfslashbin/mastopost/pkg/utils"
	"github.com/rs/zerolog"
)

var (
	aws_region string
	log        zerolog.Logger
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
	log = zerolog.New(os.Stderr).With().Timestamp().Logger()
	aws_region = os.Getenv("AWS_REGION")
}

func main() {
	// Catch errors
	var err error
	defer func() {
		if err != nil {
			log.Fatal().Err(err).Msg("main crashed")
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
			case "runtime/lastUpdated":
				if t, err := time.Parse(time.RFC3339, *p.Value); err == nil {
					config.lastUpdated = &t
				}
			case "runtime/lastPublished":
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
		rssfeed.WithLogger(&log),
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
		log.Info().
			Str("feedName", message.FeedName).
			Msg("no new items")
		return nil
	}

	client, err := mastoclient.New(
		mastoclient.WithLogger(&log),
		mastoclient.WithInstance(config.instance),
		mastoclient.WithClientID(config.clientID),
		mastoclient.WithClientSecret(config.clientSec),
		mastoclient.WithToken(config.token),
	)
	if err != nil {
		return err
	}

	ch := make(chan *mastodon.ID)

	for _, item := range newItems {
		// create a new post/toot
		newPost, err := utils.MakePost(item)
		if err != nil {
			return err
		}
		go func(newPost *mastodon.Toot) {
			id, err := client.Post(newPost)
			if err != nil {
				log.Error().Err(err).Msg("failed to post")
			}
			ch <- id
		}(newPost)
	}

	for i := 0; i < len(newItems); i++ {
		status := <-ch
		log.Info().
			Str("toInstance", config.instance.String()).
			Str("statusId", fmt.Sprintf("%v", status)).
			Msg("posted")
	}
	close(ch)

	// Update state/config
	var paramNames []*ssm.PutParameterInput

	paramNames = append(paramNames, &ssm.PutParameterInput{
		Name:      aws.String(fmt.Sprintf("%sruntime/lastUpdated", path)),
		Value:     aws.String(feed.GetLastUpdated().Format(time.RFC3339)),
		Type:      types.ParameterTypeString,
		Overwrite: aws.Bool(true),
	})

	paramNames = append(paramNames, &ssm.PutParameterInput{
		Name:      aws.String(fmt.Sprintf("%sruntime/lastPublished", path)),
		Value:     aws.String(feed.GetLastPublished().Format(time.RFC3339)),
		Type:      types.ParameterTypeString,
		Overwrite: aws.Bool(true),
	})

	for _, param := range paramNames {
		_, err := params.PutParam(param)
		if err != nil {
			return err
		}
	}

	return nil
}
