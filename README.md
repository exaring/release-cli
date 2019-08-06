# release cli
[![Go Report Card](https://goreportcard.com/badge/github.com/Am3o/release-cli)](https://goreportcard.com/report/github.com/Am3o/release-cli)
[![Build Status](https://travis-ci.org/Am3o/release-cli.svg?branch=master)](https://travis-ci.org/Am3o/release-cli)
[![godoc](http://img.shields.io/badge/godoc-reference-blue.svg?style=flat)](https://godoc.org/github.com/Am3o/release-cli) 

A command line tool for a simple release workflow based on git tags: check repo state, bump version, tag and push --tags

## Install
`go get github.com/exaring/release-cli/cmd/release`

Make sure $GOPATH/bin is on your PATH.

## Use
```bash
> release -h
Usage: release [OPTIONS]
OPTIONS:
   -major, -minor, -patch, -pre   increase version part. default is -patch.
                                  only -pre may be combined with others.
   -version <version>             specify the release version. ignores other version modifiers.
   -pre-version <pre-release>     specify the pre-release version. implies -pre. default is 'RC' (when only -pre is set).
   -dry                           do not change anything. just print the result.
   -f                             ignore untracked & uncommitted changes.
   -h                             print this help.
```

## Example
```bash
# release a defined version
> release -version 1.0.0
Releasing version 1.0.0.
Tagging.
Pushing tag.
Release 1.0.0 successful.

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

# release a specific pre-release version
> release -pre-version debug1
Retrieving old version from git.
Latest git version is '2.0.0-RC1'.
Releasing version 2.0.0-debug1.
Tagging.
Pushing tag.
Release 2.0.0-debug1 successful.
```
