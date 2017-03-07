# release cli
A command line tool for a simple release workflow based on git tags: check repo state, bump version, tag and push --tags

## Install
`go get github.com/exaring/release-cli/cmd/release`

Make sure $GOPATH/bin is on your PATH.

## Use
```bash
> release -h
Usage: release [OPTIONS]
OPTIONS:
   -major, -minor, -patch   increase version part. default is -patch
   -build <build-name>      include additional build name (e.g. alpha)
   -version <version>       specify the release version. ignores other version modifiers.
   -h                       print this help.
```

## Example
```bash
# release a defined version
> release -version 1.0.0
Tagging version 1.0.0.
Pushing tag.
Release 1.0.0 successful.

# release the next patch release (default)
> release
Tagging version 1.0.1.
Pushing tag.
Release 1.0.1 successful.

# release the next minor version
> release -minor
Retrieving old version from git.
Latest git version is '1.0.1'.Tagging version 1.1.1.
Pushing tag.
Release 1.1.1 successful.

# release the next version with bumped minor and patch parts
> release -minor -patch
Retrieving old version from git.
Latest git version is '1.1.1'.Tagging version 1.2.2.
Pushing tag.
Release 1.2.2 successful.

# release a new defined version with all parts bumped and an additional build part
> release -major -minor -patch -build=alpha -version 1.0.0
bin/release -minor -patch -build=alpha -version 1.0.0
Tagging version 2.1.1+alpha.
Pushing tag.
Release 2.1.1+alpha successful.
```
