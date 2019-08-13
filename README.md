# release cli
[![license](https://img.shields.io/badge/license-apache-red.svg?style=flat)](https://raw.githubusercontent.com/github.com/exaring/release-cli/blob/master/LICENSE) 
[![Go Report Card](https://goreportcard.com/badge/github.com/exaring/release-cli)](https://goreportcard.com/report/github.com/exaring/release-cli)
[![Build Status](https://travis-ci.org/exaring/release-cli.svg?branch=master)](https://travis-ci.org/exaring/release-cli)
[![godoc](http://img.shields.io/badge/godoc-reference-blue.svg?style=flat)](https://godoc.org/github.com/exaring/release-cli) 

Release cli is a useful command-line tool for semantic version tags. A semantic version has three parts: major, minor, and patch. For example, 
for v0.1.2, the major version is 0, the minor version is 1, and the patch version is 2. It's necessary for example of the `Go Modules`. For more information read the following article. https://blog.golang.org/using-go-modules

## Installation 

Install using the "go get" command:

`go get github.com/exaring/release-cli/cmd/release`

### Prebuilt binaries
Clone the repo and then run the makefile

`make install`

## Use
```bash
> release -h
NAME:
   release-cli (release tool) - create semantic version tags

USAGE:
   release [global options] command [command options] [arguments...]

VERSION:
   v2.0.0-RC1

DESCRIPTION:
   Release is a useful command line tool for semantic version tags

COMMANDS:
     help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --major                increase major version part. [$RELEASE_MAJOR]
   --minor                increase minor version part. [$RELEASE_MINOR]
   --patch                increase patch version part. This is the default increased part. [$RELEASE_PATCH]
   --pre                  increase release candidate version part. [$RELEASE_PRE]
   -d, --dry              do not change anything. just print the result. [$DRY_RUN]
   -f, --force            ignore untracked & uncommitted changes. [$FORCE]
   -l value, --log value  specifics the log level of the output [$LOG_LEVEL]
   --help, -h             show help
   --version, -v          print the version
```

## Example
```bash
# release the next patch release (default)
> release
Retrieving old version from git.
Latest git version is '1.0.0'.
Releasing version 1.0.1.
Tagging.
Pushing tag.
Release 1.0.1 successful.

# release the next minor version
> release -minor
Latest git version is '1.0.1'.
Releasing version 1.1.0.
Tagging.
Pushing tag.
Release 1.1.0 successful.

# release a major pre-release
> release -major -pre
Retrieving old version from git.
Latest git version is '1.1.0'.
Releasing version 2.0.0-RC1.
Tagging.
Pushing tag.
Release 2.0.0-RC1 successful.
```
