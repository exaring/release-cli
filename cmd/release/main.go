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

// Version is the version of the cli application
var Version = "v2.0.1"

func main() {
	app := cli.NewApp()
	app.Name = "release-cli (release tool)"
	app.Usage = "create semantic version tags"
	app.Description = "Release is a useful command line tool for semantic version tags"
	app.Version = Version

	var (
		flagMajor, flagMinor, flagPatch, flagPre, dryRun, force bool
		flagBranch, flagLog                                     string
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
			Name:        "b, branch",
			Destination: &flagBranch,
			Usage:       "only track tags related to the given branch when creating new version tags.",
			EnvVar:      "ONLY_BRANCH",
		},
		cli.StringFlag{
			Name:        "l, log",
			Destination: &flagLog,
			Usage:       "specifics the log level of the output",
			EnvVar:      "LOG_LEVEL",
		},
	}

	app.Action = run
	if err := app.Run(os.Args); err != nil {
		logrus.WithError(err).Error("Couldn't release a new version")
		os.Exit(1)
	}
}

// Repository is a abstraction of the version control system client
type Repository interface {
	// LatestCommitHash returns the latest commit hash of the repository. In case of an error the result is empty.
	LatestCommitHash() string
	// ExistsTag validates the parameter version and returns the existence of the repository tag.
	ExistsTag(version string) (bool, error)
	// Tags lists all existing tags of the repository.
	Tags() []string
	// BranchTags lists all existing tags related to commits of the given branch.
	BranchTags(branchName string) []string
	// IsSafe validate the state of the repository and returns an error if the repository is unsafe like include uncommitted files
	// or the local branch is behind the origin.
	IsSafe(ctx context.Context) error
	// CreateTag creates a local version control system tag.
	CreateTag(tag string) error
	// DeleteTag deletes a local version control system  tag.
	DeleteTag(tag string) error
	// Push pushes the local repo state to the origin.
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
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	logger.Debug("Read the directory")

	var repo Repository
	repo, err = repository.New(currentPath)
	if err != nil {
		return fmt.Errorf("failed to open the git repository metadata directory: %w", err)
	}

	if ctx.IsSet("dry") {
		repo = repository.NewNoOp()
	}

	logger.Debug("Analyse the git repository")

	var currentTag version.Version
	if ctx.IsSet("branch") {
		currentTag, err = LatestBranchTag(repo, ctx.String("branch"))
		if err != nil {
			return fmt.Errorf("failed to fetch the tag on the given branch: %w", err)
		}
	} else {
		currentTag, err = LatestTag(repo)
		if err != nil {
			return fmt.Errorf("failed to fetch the tag in the repository: %w", err)
		}
	}
	logger.WithFields(logrus.Fields{
		"Tag": currentTag,
	}).Debug("Detect latest tag of the repository")

	currentTag.Increase(
		ctx.IsSet("major"),
		ctx.IsSet("minor"),
		ctx.IsSet("patch"),
		ctx.IsSet("pre"))
	logger.WithFields(logrus.Fields{
		"Tag": currentTag,
	}).Info("Create new releasing version")

	if err := repo.IsSafe(context.Background()); !ctx.IsSet("force") && err != nil {
		return fmt.Errorf("repository is in unsafe state and force is not set: %w", err)
	}

	if err := repo.CreateTag(currentTag.String()); err != nil {
		if deleteErr := repo.DeleteTag(currentTag.String()); deleteErr != nil {
			logger.WithError(deleteErr).Errorf("Couldn't remove the creates tag: %v", currentTag)
		}
		return fmt.Errorf("failed to create tag: %w", err)
	}
	logger.WithFields(logrus.Fields{
		"Version": currentTag,
	}).Debug("Tagging the current repository")

	if err := repo.Push(context.Background()); err != nil {
		if deleteErr := repo.DeleteTag(currentTag.String()); deleteErr != nil {
			logger.WithError(deleteErr).Errorf("Couldn't remove the creates tag: %v", currentTag)
		}
		return fmt.Errorf("failed to push tag: %w", err)
	}
	logger.WithFields(logrus.Fields{
		"Version": currentTag,
	}).Debug("Pushing new tag to the origin repository")

	if ctx.IsSet("branch") {
		currentTag, err = LatestBranchTag(repo, ctx.String("branch"))
		if err != nil {
			return fmt.Errorf("failed to fetch the tag on the given branch: %w", err)
		}
	} else {
		currentTag, err = LatestTag(repo)
		if err != nil {
			return fmt.Errorf("failed to fetch the tag in the repository: %w", err)
		}
	}

	if dryModus {
		logger.Info("Don't publish the new releases, because of the dry-run mode")
		return nil
	}

	logger.WithFields(logrus.Fields{
		"Version": currentTag,
	}).Info("Release new version")

	return nil
}

// LatestTag returns the latest tag of the repository.
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

// LatestBranchTag returns the latest tag of the given branch.
func LatestBranchTag(vc Repository, branchName string) (version.Version, error) {
	var branchTags version.Versions
	for _, tag := range vc.BranchTags(branchName) {
		o, err := version.New(tag)
		if err != nil {
			return version.Version{}, err
		}
		branchTags = append(branchTags, o)
	}

	sort.Sort(branchTags)

	if len(branchTags) > 0 {
		return branchTags[len(branchTags)-1], nil
	}

	return version.Version{}, fmt.Errorf("the master branch version list is empty")

}
