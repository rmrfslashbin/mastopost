package utils

import (
	"fmt"

	"github.com/iancoleman/strcase"
	"github.com/mattn/go-mastodon"
	"github.com/rmrfslashbin/mastopost/pkg/rssfeed"
)

// makePost formats the RSS item into a Mastodon post
func MakePost(item rssfeed.NewItems) (*mastodon.Toot, error) {
	author := ""
	hashtags := ""

	if item.Author != nil {
		if item.Author.Name != "" {
			author = " by " + item.Author.Name
		}
		if item.Author.Email != "" {
			author += " (" + item.Author.Email + ")"
		}
	}

	if item.Categories != nil {
		hashtags = "\n"
		for _, cat := range item.Categories {
			hashtags += " #" + strcase.ToCamel(cat)
		}
	}

	newPost := &mastodon.Toot{
		Status: fmt.Sprintf("%s%s\n\n%s\n\n%s\n\n%s", item.Title, author, item.Published, item.Link, hashtags),
	}
	return newPost, nil
}
