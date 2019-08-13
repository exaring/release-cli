package main

import (
	"context"
	"fmt"
	"os"
	"sort"

	"github.com/exaring/release-cli/pkg/repository"
	"github.com/exaring/release-cli/pkg/version"

	"github.com/sirupsen/logrus"

	"github.com/urfave/cli"
)

var Version = "v0.0.0"

func main() {
	app := cli.NewApp()
	app.Name = "release-cli (release tool)"
	app.Usage = "create semantic version tags"
	app.Description = "Release is a useful command line tool for semantic version tags"
	app.Version = Version

	var (
		flagMajor, flagMinor, flagPatch, flagPre, dryRun, force bool
		flagLog                                                 string
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
		cli.StringFlag{
			Name:        "l, log",
			Destination: &flagLog,
			Usage:       "specifics the log level of the output",
			EnvVar:      "LOG_LEVEL",
		},
	}

	var exitCode int
	defer os.Exit(exitCode)

	app.Action = run
	if err := app.Run(os.Args); err != nil {
		logrus.WithError(err).Error("Couldn't release a new version")
	}
}

type Repository interface {
	LatestCommitHash() string
	ExistsTag(version string) (bool, error)
	Tags() []string
	IsSafe(ctx context.Context) error
	CreateTag(tag string) error
	DeleteTag(tag string) error
	Push(ctx context.Context) error
}

func run(ctx *cli.Context) error {
	switch ctx.String("log") {
	case "debug":
		logrus.SetLevel(logrus.DebugLevel)
	case "error":
		logrus.SetLevel(logrus.ErrorLevel)
	default:
		logrus.SetLevel(logrus.InfoLevel)
	}
	logger := logrus.StandardLogger()
	dryModus := ctx.IsSet("dry")

	currentPath, err := os.Getwd()
	if err != nil {
		return err
	}
	logger.WithFields(logrus.Fields{
		"Repository": currentPath,
		"Dry":        dryModus,
	}).Debug("Read the directory")

	var repo Repository
	repo, err = repository.New(currentPath)
	if err != nil {
		return err
	}

	if ctx.IsSet("dry") {
		repo = repository.NewNoOp()
	}

	logger.WithFields(logrus.Fields{
		"Repository": currentPath,
		"Dry":        dryModus,
	}).Debug("Analyse the git repository")
	currentTag, err := LatestTag(repo)
	if err != nil {
		return err
	}
	logger.WithFields(logrus.Fields{
		"Repository": currentPath,
		"Tag":        currentTag,
		"repository": currentPath,
		"Dry":        dryModus,
	}).Debug("Detect latest tag of the repository")

	currentTag.Increase(
		ctx.IsSet("major"),
		ctx.IsSet("minor"),
		ctx.IsSet("patch"),
		ctx.IsSet("pre"))
	logger.WithFields(logrus.Fields{
		"Repository": currentPath,
		"Tag":        currentTag,
		"Dry":        dryModus,
	}).Info("Create new releasing version")

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
		"Dry":        dryModus,
	}).Debug("Tagging the current repository")

	if err := repo.Push(context.Background()); err != nil {
		if err := repo.DeleteTag(currentTag.String()); err != nil {
			logger.WithError(err).Errorf("Couldn't remove the creates tag: %v", currentTag)
		}
		return err
	}
	logger.WithFields(logrus.Fields{
		"Repository": currentPath,
		"Version":    currentTag,
		"Dry":        dryModus,
	}).Debug("Pushing new tag to the origin repository")

	currentTag, err = LatestTag(repo)
	if err != nil {
		return err
	}
	logger.WithFields(logrus.Fields{
		"Repository": currentPath,
		"Version":    currentTag,
		"Dry":        dryModus,
	}).Info("Release new version")

	return nil
}

func LatestTag(vc Repository) (version.Version, error) {
	var lightweightTags version.Versions
	for _, tag := range vc.Tags() {
		o, err := version.New(tag)
		if err != nil {
			return version.Version{}, err
		}
		lightweightTags = append(lightweightTags, o)
	}

	sort.Sort(lightweightTags)

	if len(lightweightTags) > 0 {
		return lightweightTags[len(lightweightTags)-1], nil
	}

	return version.Version{}, fmt.Errorf("the version list is empty")
}
