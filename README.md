# mastopost
Post RSS feeds to Mastodon

## Description
Mastopost is a flexible tool to post RSS, ATOM, and JSON feed items to Mastodon. The goal of the project is to leverage an AWS Lambda function and other cloud native services to post RSS feeds to Mastodon at a regular interval. A second goal is to provide a "one shot" tool to run the function locally.

## Upstream Projects
Mastopost is built using the following projects:
- https://github.com/mmcdole/gofeed: "The gofeed library is a robust feed parser that supports parsing both RSS, Atom and JSON feeds."
- https://github.com/iancoleman/strcase: "strcase is a go package for converting string case to various cases (e.g. snake case or camel case)..."
- https://github.com/mattn/go-mastodon: go-mastodon is a Go client library for Mastodon API.
- https://github.com/rs/zerolog: "Zero Allocation JSON Logger".
- https://github.com/davecgh/go-spew: "Package spew implements functions to pretty-print Go values in a form that is useful for developers."
- https://github.com/aws/aws-sdk-go-v2: "The AWS SDK for Go V2 provides a client and types for working with AWS services."

## Setup from source
Skip this section if you plan to download and run the binaries from the [Releases](https://github.com/rmrfslashbin/mastopost/releases) section.
- Clone this repo to your local machine `git clone https://github.com/rmrfslashbin/mastopost.git`.
- Install and set up [Go](https://golang.org/doc/install).
- Build the main binary `make build`. This will compile the binary to `bin/mastopost-${os}-${arch}`. It will also compile the lambda functions to binaries and create a zip file for each function in the `bin/` directory.

## CLI
The CLI tool is a monolithic binary to run and manage the Mastopost application. Run `mastopost --help` for usage information.
### Setup
- Copy `config.DIST.json` to `config.json` and edit as needed.
- Move the json file to the default location (run `mastopost cfg` to see the defaults), or explicitly set the `--config` flag.

### Commands
- cfg: print the default location of the config file. This is the location the CLI will look for the config file, unless the `--config` flag is set.
- oneshot: Run the application once. This is useful to run the application locally or via a cron job.
- job: job management commands. Run `mastopost job --help` for usage information.
  - add: Add a job to AWS Event Bridge.
  - delete: Delete a job from AWS Event Bridge.
  - list: List jobs in AWS Event Bridge.
  - status: Get the status of a job in AWS Event Bridge. Also enable or disable a job.
- lambda: manage lambda functions (not yet implemented).

## AWS Setup
Configuring AWS is beyond the scope of this document, but these steps should get you started:
- Create an AWS account.
- Set up the AWS CLI and configure it with your credentials: https://aws.amazon.com/cli/.
- Install the Lambda function. `mastopost lambda install --functionname mastopost --zipfile bin/mastopost-lambda.zip`.
- Note the output function and policy ARNs and add them to the `lambdaFunctions` section of the config file.


### CLI Configuration
The CLI config file is JSON file with the following structure:

```json
{
    "feeds": {
        "Atlbeltline": {
            "lastupdatefile": "/Users/user/Library/Application Support/mastopost/atlbeltline.gob",
            "feedurl": "https://beltline.org/feed/",
            "clientid": "mastodon_client_id",
            "clientsecret": "mastodon_client_secret",
            "accesstoken": "mastodon_access_token",
            "instance": "https://mastodon.example.com",
            "schedule": "rate(30 minutes)"
        },
        "Arstechnica": {
            "lastupdatefile": "/Users/user/Library/Application Support/mastopost/arstechnica.gob",
            "feedurl": "http://feeds.arstechnica.com/arstechnica/index",
            "clientid": "mastodon_client_id",
            "clientsecret": "mastodon_client_secret",
            "accesstoken": "mastodon_access_token",
            "instance": "https://mastodon.example.com",
            "schedule": "rate(30 minutes)"
        }
    },
    "lambdaFunctions": {
        "mastopost-rss-crossposter": {
            "functionArn": "arn:aws:lambda:us-east-1:xxxxxxxxxxxx:function:mastopost-rss-crossposter",
            "policyArn": "arn:aws:iam::xxxxxxxxxxxx:policy/policy-mastopost-rss-crossposter"
        }
    }
}
```
- `feeds`: REQUIRED: A map of feed names to feed configuration. The name of the feed is arbitrary and is used to identify the feed in the config file.
  - `lastupdatefile`: (Optional if not using oneshot): The path to a file to store the last update time. This is used to determine if a feed item has been posted to Mastodon.
  - `feedurl`: The URL of the RSS feed.
  - `clientid`: The Mastodon client ID.
  - `clientsecret`: The Mastodon client secret.
  - `accesstoken`: The Mastodon access token.
  - `instance`: The Mastodon instance URL.
  - `schedule`: (Optional if only using oneshot): The AWS Event Bridge schedule expression. See [AWS Event Bridge Schedule Expressions](https://docs.aws.amazon.com/eventbridge/latest/userguide/scheduled-events.html) for more information.
- `lambdaFunctions`: OPTIONAL (if only running oneshot): A map of Lambda function names to Lambda function configuration. The name of the function is arbitrary and is used to identify the function in the config file.
  - `functionArn`: The ARN of the Lambda function.
