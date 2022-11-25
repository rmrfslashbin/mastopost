#mastopost
Post RSS feeds to Mastodon

## Description
This repository contains a CLI tool to post RSS feeds to Mastodon and an AWS process to schedule cloud native posting.

## CLI
- Build the CLI tool with `go build -o mastopost cmd/mastopost/main.go`.
- Copy `config.yaml.DIST` to `config.yaml` and fill in the Mastodon credentials and feed settings.
- Place the config file in the default locationn (run `mastopost --help` to see the default location).
- Run the CLI (pass the path to the config file with the `--config` flag if needed).
- The param `--feedname` is required and must persist across multiple runs. It is used to keep track of the last post.

### CLI Configuration
The CLI config file is a YAML file with the following structure:

```yaml
lastupdate: "/Users/myuser/Library/Application Support/mastopost/lastupdate.gob"
url: https://example.com/feed/
clientid: mastodon_client_id
clientsec: mastodon_client_secret
token: mastodon_access_token
instance: https://example.com
```
- lastupdate: The path to the file where the last update time is stored.
- URL is the URL for the RSS feed to parse.
- clientid, clientsec and token are the credentials for the Mastodon account to post to.
- instance is the URL of the Mastodon instance to post to.
