.DEFAULT_GOAL := build
.PHONY: build

stack_name = mastopost

build:
	@printf "  building mastopost:\n"
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
	@printf "  building mastopost-lambda-rssxpost:\n"
	@printf "    linux  :: arm64"
	@GOOS=linux GOARCH=arm64 go build -o bin/mastopost-lambda-rssxpost/bootstrap lambda/mastopost-rssxpost/main.go
	@printf " done.\n"
	@printf "  zipping mastopost-lambda-rssxpost to bin/mastopost-lambda-rssxpost.zip:\n"
	@zip -j bin/mastopost-lambda-rssxpost.zip bin/mastopost-lambda-rssxpost/bootstrap
	@printf " done.\n"

tidy:
	@echo "Making mod tidy"
	@go mod tidy

update:
	@echo "Updating $(stack_name)"
	@go get -u ./...
	@go mod tidy	

prune:
	@git gc --prune=now
	@git remote prune origin
