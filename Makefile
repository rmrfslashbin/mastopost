.DEFAULT_GOAL := build
.PHONY: build

stack_name = mastopost
deploy_bucket = is-us-east-1-deployment
aws_profile = default

build:
	@printf "  building mastopost-cli:\n"
	@printf "    linux  :: arm64"
	@GOOS=linux GOARCH=arm64 go build -o bin/mastopost-linux-arm64 cmd/mastopost/main.go
	@printf " done.\n"
	@printf "    linux  :: amd64"
	@GOOS=linux GOARCH=amd64 go build -o bin/mastopost-linux-amd64 cmd/mastopost/main.go
	@printf " done.\n"
	@printf "    darwin :: amd64"
	@GOOS=darwin GOARCH=amd64 go build -o bin/mastopost-darwin-amd64 cmd/mastopost/main.go
	@printf " done.\n"
	@printf "    darwin :: arm64"
	@GOOS=darwin GOARCH=arm64 go build -o bin/mastopost-darwin-arm64 cmd/mastopost/main.go
	@printf " done.\n"

tidy:
	@echo "Making mod tidy"
	@go mod tidy

update:
	@echo "Updating $(stack_name)"
	@go get -u ./...
	@go mod tidy

deploy: lambda-build
	aws --profile $(aws_profile) cloudformation package --template-file aws-cloudformation/template.yaml --s3-bucket $(deploy_bucket) --output-template-file build/out.yaml
	aws --profile $(aws_profile) cloudformation deploy --template-file build/out.yaml --s3-bucket $(deploy_bucket) --stack-name $(stack_name) --capabilities CAPABILITY_NAMED_IAM

lambda-build:
	GOOS=linux GOARCH=arm64 go build -o bin/mastopost-lambda-rssxpost/bootstrap lambda/mastopost-rssxpost/main.go
	zip -j bin/mastopost-lambda-rssxpost.zip bin/mastopost-lambda-rssxpost/bootstrap

cfdescribe:
	aws --profile $(aws_profile) cloudformation describe-stack-events --stack-name $(stack_name)

prune:
	@git gc --prune=now
	@git remote prune origin
