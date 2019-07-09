package main

import (
	"context"
	"os"
	"sort"

	"github.com/am3o/release-cli/pkg/version"

	"github.com/am3o/release-cli/pkg/repository"

	"github.com/sirupsen/logrus"

	"github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()
	app.Name = "release-cli (release tool)"
	app.Version = "experimental-build"

	var (
		flagMajor, flagMinor, flagPatch, flagPre, dryRun, force bool
	)
	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:        "major",
			Destination: &flagMajor,
			Usage:       "increase major version part.",
			EnvVar:      "RELEASE_MAJOR",
		},
		cli.BoolFlag{
			Name:        "minor",
			Destination: &flagMinor,
			Usage:       "increase minor version part.",
			EnvVar:      "RELEASE_MINOR",
		},
		cli.BoolFlag{
			Name:        "patch",
			Destination: &flagPatch,
			Usage:       "increase patch version part. This is the default increased part.",
			EnvVar:      "RELEASE_PATCH",
		},
		cli.BoolFlag{
			Name:        "pre",
			Destination: &flagPre,
			Usage:       "increase release candidate version part.",
			EnvVar:      "RELEASE_PRE",
		},
		cli.BoolFlag{
			Name:        "d, dry",
			Destination: &dryRun,
			Usage:       "do not change anything. just print the result.",
			EnvVar:      "DRY_RUN",
		},
		cli.BoolFlag{
			Name:        "f, force",
			Destination: &force,
			Usage:       "ignore untracked & uncommitted changes.",
			EnvVar:      "FORCE",
		},
	}

	app.Action = run
	if err := app.Run(os.Args); err != nil {
		logrus.WithError(err).Error("Couldn't release a new version")
	}
}

func run(ctx *cli.Context) error {
	var logger logrus.FieldLogger = logrus.StandardLogger()
	var mode string
	if ctx.IsSet("dry") {
		mode = "[dry-run] "
	}

	currentPath, err := os.Getwd()
	if err != nil {
		return err
	}
	logger.WithField("repository", currentPath).Infof("%vRead the directory", mode)

	repo, err := repository.New(currentPath)
	if err != nil {
		return err
	}
	logger.WithField("repository", currentPath).Infof("%vAnalyse the git repository", mode)

	currentTag, err := LatestTag(repo)
	if err != nil {
		return err
	}
	logger.WithFields(logrus.Fields{
		"repository": currentPath,
		"Tag":        currentTag,
	}).Infof("%vDetect latest tag of the repository", mode)

	currentTag.Increase(ctx.IsSet("major"), ctx.IsSet("minor"), ctx.IsSet("patch"), ctx.IsSet("pre"))
	logger.WithFields(logrus.Fields{
		"repository": currentPath,
		"tag":        currentTag,
	}).Infof("%vCreate new releasing version", mode)

	if err := repo.IsSafe(context.Background()); !ctx.IsSet("force") && err != nil {
		return err
	}

	if err := repo.CreateTag(currentTag.String()); err != nil {
		if err := repo.DeleteTag(currentTag.String()); err != nil {
			logger.WithError(err).Errorf("Couldn't remove the creates tag: %v", currentTag)
		}
		return err
	}
	logger.WithFields(logrus.Fields{
		"Repository": currentPath,
		"Version":    currentTag,
	}).Infof("%vTagging the current repository", mode)

	if err := repo.Push(context.Background()); err != nil {
		if err := repo.DeleteTag(currentTag.String()); err != nil {
			logger.WithError(err).Errorf("Couldn't remove the creates tag: %v", currentTag)
		}
		return err
	}
	logger.WithFields(logrus.Fields{
		"Repository": currentPath,
		"Version":    currentTag,
	}).Infof("%vPushing new tag to the origin repository", mode)

	currentTag, err = LatestTag(repo)
	if err != nil {
		return err
	}
	logger.WithFields(logrus.Fields{
		"Repository": currentPath,
		"Version":    currentTag,
	}).Infof("%vRelease new version", mode)

	return nil
}

type Repository interface {
	Tags() []string
}

func LatestTag(vc Repository) (version.Version, error) {
	var lihtweightTags version.Versions
	for _, tag := range vc.Tags() {
		o, err := version.New(tag)
		if err != nil {
			return version.Version{}, err
		}
		lihtweightTags = append(lihtweightTags, o)
	}

	sort.Sort(lihtweightTags)

	return lihtweightTags[len(lihtweightTags)-1], nil
}
