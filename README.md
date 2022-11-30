# mastopost
Post RSS feeds to Mastodon

## Description
Mastopost is a flexible tool to post RSS feed items to Mastodon. The goal of the project is to leverage an AWS Lambda function and other cloud native services to post RSS feeds to Mastodon at a regular interval. A second goal is to provide a "one shot" tool to run the function locally.

## Setup from source
Skip this section if you plan to download and run the binaries from the [Releases](https://github.com/rmrfslashbin/mastopost/releases) section.
- Clone this repo to your local machine `git clone https://github.com/rmrfslashbin/mastopost.git`.
- Install and set up [Go](https://golang.org/doc/install).
- Build the main binary `make build`. This will compile the binary to `bin/mastopost-os-arch`.

## AWS Setup
Configuring AWS is beyond the scope of this document, but these steps should get you started:
- Build the lambda binary `make build-lambda`
- Fix up the [AWS Cloudformation template] (/aws-cloudformation/template.yaml) as needed.
- Fix up the `Makefile` as needed to deploy to AWS.
- Deploy the AWS Cloudformation template `make deploy`.

## CLI
The CLI tool is a monolithic binary to run and manage the Mastopost application. Run `mastopost --help` for usage information.
- Copy `config.DIST.json` to `config.json` and edit as needed.
- Move the json file to the default location, or explicitly set the `--config` flag.
- cfg: print the default location of the config file. This is the location the CLI will look for the config file, unless the `--config` flag is set.
- oneshot: Run the application once. This is useful to run the application locally or via a cron job.
- more to come...

### CLI Configuration
The CLI config file is JSON file with the following structure:

```json
{
    "feeds": {
        "atlbeltline": {
            "lastupdatefile": "/Users/user/Library/Application Support/mastopost/atlbeltline.gob",
            "feedurl": "https://beltline.org/feed/",
            "clientid": "mastodon_client_id",
            "clientsecret": "mastodon_client_secret",
            "accesstoken": "mastodon_access_token",
            "instance": "https://mastodon.example.com"
        },
        "arstechnica": {
            "lastupdatefile": "/Users/user/Library/Application Support/mastopost/arstechnica.gob",
            "feedurl": "http://feeds.arstechnica.com/arstechnica/index",
            "clientid": "mastodon_client_id",
            "clientsecret": "mastodon_client_secret",
            "accesstoken": "mastodon_access_token",
            "instance": "https://mastodon.example.com"
        }
    },
    "lambdaFunctions": {
        "mastopost-rss-crossposter": {
            "functionArn": "arn:aws:lambda:us-east-1:xxxxxxxxxxxx:function:mastopost-rss-crossposter"
        }
    }
}
```
- `feeds`: A map of feed names to feed configuration. The name of the feed is arbitrary and is used to identify the feed in the config file.
  - `lastupdatefile`: The path to a file to store the last update time. This is used to determine if a feed item has been posted to Mastodon.
  - `feedurl`: The URL of the RSS feed.
  - `clientid`: The Mastodon client ID.
  - `clientsecret`: The Mastodon client secret.
  - `accesstoken`: The Mastodon access token.
  - `instance`: The Mastodon instance URL.
- `lambdaFunctions`: A map of Lambda function names to Lambda function configuration. The name of the function is arbitrary and is used to identify the function in the config file.
  - `functionArn`: The ARN of the Lambda function.
