.PHONY: default build test vendor

default: build

build: test
	go build -o bin/release ./cmd/release

vendor:
	go get github.com/rancher/trash
	rm -rf vendor/
	trash -u
	trash
