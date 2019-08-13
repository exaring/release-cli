GIT_TAG := $(shell git tag --sort=-creatordate | head -n1)

.PHONY: default build test install

default: build

test:
	@printf "\033[01;33m>> Running tests\033[0m\n"
	go test -cover -race $$(go list ./... | grep -v /vendor/ | grep -v /integration | tr "\n" " ")

build: test
	@printf "\033[01;33m>> Running build\033[0m\n"
	go build -ldflags '-X main.Version=$(GIT_TAG)' -o bin/release ./cmd/release

install: test
	@printf "\033[01;33m>> Running install\033[0m\n"
	go install -ldflags '-X main.Version=$(GIT_TAG)' ./cmd/release
