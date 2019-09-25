package repository

import (
	"context"
	"fmt"
	"os"
	"strings"

	"gopkg.in/src-d/go-git.v4/config"

	"gopkg.in/src-d/go-git.v4/plumbing/object"

	"gopkg.in/src-d/go-git.v4/plumbing"

	"gopkg.in/src-d/go-git.v4"
)

// Git is the version control client for git
type Git struct {
	client *git.Repository
}

// New creates an new instance of the git client
func New(path string) (*Git, error) {
	repo, err := git.PlainOpen(path)
	if err != nil {
		return nil, err
	}
	return &Git{
		client: repo,
	}, nil
}

// LatestCommitHash returns the latest commit hash of the git repo. In case of an error the result is empty.
func (vc *Git) LatestCommitHash() string {
	headRef, err := vc.client.Head()
	if err != nil {
		return ""
	}

	return headRef.Hash().String()
}

// ExistsTag validates the parameter version and returns the existence of the git tag.
func (vc *Git) ExistsTag(version string) (bool, error) {
	t, err := vc.client.Tags()
	if err != nil {
		return false, err
	}

	var existsTag bool
	if err := t.ForEach(func(reference *plumbing.Reference) error {
		if version == reference.Name().String() {
			existsTag = true
		}
		return nil
	}); err != nil {
		return false, err
	}

	return existsTag, nil
}

// Tags lists all existing tags of the git repo.
func (vc *Git) Tags() []string {
	tIter, err := vc.client.Tags()
	if err != nil {
		return nil
	}

	var tags = make([]string, 0)
	if err := tIter.ForEach(func(reference *plumbing.Reference) error {
		tags = append(tags, reference.String())
		return nil
	}); err != nil {
		return nil
	}

	return tags
}

// BranchTags lists all existing tags associated to commits of the given branch.
func (vc *Git) BranchTags(branchName string) []string {

	// get branch reference
	branch, err := vc.client.Branch(branchName)
	if err != nil {
		return nil
	}
	ref, err := vc.client.Reference(branch.Merge, true)
	if err != nil {
		return nil
	}

	// get all commit hashes of the given branch
	logs, err := vc.client.Log(&git.LogOptions{
		From: ref.Hash(),
	})
	if err != nil {
		return nil
	}
	var branchCommits = make(map[plumbing.Hash]bool)
	if err := logs.ForEach(func(commit *object.Commit) error {
		branchCommits[commit.Hash] = true
		return nil
	}); err != nil {
		return nil
	}
	logs.Close()

	// get all tags of the repository with their associated commit hash
	tIter, err := vc.client.Tags()
	if err != nil {
		return nil
	}
	var tagsWithCommits = make(map[string]plumbing.Hash)
	if err := tIter.ForEach(func(ref *plumbing.Reference) error {
		if annotedTag, err := vc.client.TagObject(ref.Hash()); err != plumbing.ErrObjectNotFound {
			if annotedTag.TargetType == plumbing.CommitObject {
				tagsWithCommits[ref.String()] = annotedTag.Target
			}
			return nil
		}
		tagsWithCommits[ref.String()] = ref.Hash()
		return nil
	}); err != nil {
		return nil
	}

	// only return tags whose associated commit hash belongs to the master branch
	var branchTags = make([]string, 0)
	for tag, commit := range tagsWithCommits {
		if _, ok := branchCommits[commit]; ok {
			branchTags = append(branchTags, tag)
		}
	}

	return branchTags

}

// IsSafe validate the state of the git repo and returns an error if the repo is unsafe like include uncommitted files
// or the local branch is behind the origin.
func (vc *Git) IsSafe(ctx context.Context) error {
	if hasUncommittedChanges, err := vc.HasUncommittedChanges(); err != nil {
		return err
	} else if hasUncommittedChanges {
		return fmt.Errorf("your client has uncommitted changes.")
	}

	if vc.HasStagedChanges() {
		return fmt.Errorf("your client has unpushed changes. Please push.")
	}

	if isBehind, err := vc.IsBehind(ctx); err != nil {
		if strings.Contains(err.Error(), "already up-to-date") {
			return nil
		}
		return fmt.Errorf("could not determine remote status: %v.", err)
	} else if !isBehind {
		return fmt.Errorf("your branch is behind the remote. Please pull.")
	}

	return nil
}

// HasUncommittedChanges checks the git repo for uncommitted changes.
func (vc *Git) HasUncommittedChanges() (bool, error) {
	w, err := vc.client.Worktree()
	if err != nil {
		return false, err
	}

	status, err := w.Status()
	if err != nil {
		return false, err
	}

	return !status.IsClean(), nil
}

// HasStagedChanges checks the git repo stash for staged changes
// Note: the current version of the go-git doesn't support this functionality.
func (vc *Git) HasStagedChanges() bool {
	// TODO: unsupported function from the go-git lib
	return false
}

// IsBehind checks the local repo with the origin and validate the state of the git repo
func (vc *Git) IsBehind(ctx context.Context) (bool, error) {
	if err := vc.client.FetchContext(ctx, &git.FetchOptions{}); err != nil {
		return false, err
	}

	return true, nil
}

// CreateTag creates a local git tag.
func (vc *Git) CreateTag(tag string) error {
	name := plumbing.ReferenceName(fmt.Sprintf("refs/tags/%v", tag))
	reference := plumbing.NewHashReference(name, plumbing.NewHash(vc.LatestCommitHash()))
	return vc.client.Storer.SetReference(reference)
}

// DeleteTag deletes a local git tag.
func (vc *Git) DeleteTag(tag string) error {
	return vc.client.Storer.RemoveReference(
		plumbing.ReferenceName(fmt.Sprintf("refs/tags/%v", tag)),
	)
}

// Push pushes the local repo state to the origin.
func (vc *Git) Push(ctx context.Context) error {
	return vc.client.PushContext(ctx, &git.PushOptions{
		RefSpecs: []config.RefSpec{"refs/tags/*:refs/tags/*"},
	})
}

// NoOpRepository is the implementation of an no-operation client.
type NoOpRepository struct{}

// NewNoOp creates an new instance of the No-Operation object
func NewNoOp() *NoOpRepository {
	return &NoOpRepository{}
}

// LatestCommitHash does nothing.
func (noop *NoOpRepository) LatestCommitHash() string {
	return ""
}

// ExistsTag does nothing.
func (noop *NoOpRepository) ExistsTag(version string) (bool, error) {
	return true, nil
}

// Tags does nothing
func (noop *NoOpRepository) Tags() []string {
	currentPath, err := os.Getwd()
	if err != nil {
		return make([]string, 0)
	}

	repository, err := New(currentPath)
	if err != nil {
		return make([]string, 0)
	}

	return repository.Tags()
}

// BranchTags does nothing
func (noop *NoOpRepository) BranchTags(branchName string) []string {
	currentPath, err := os.Getwd()
	if err != nil {
		return make([]string, 0)
	}

	repository, err := New(currentPath)
	if err != nil {
		return make([]string, 0)
	}

	return repository.BranchTags(branchName)
}

// IsSafe does nothing.
func (noop *NoOpRepository) IsSafe(ctx context.Context) error {
	return nil
}

// CreateTag does nothing.
func (noop *NoOpRepository) CreateTag(tag string) error {
	return nil
}

// DeleteTag does nothing.
func (noop *NoOpRepository) DeleteTag(tag string) error {
	return nil
}

// Push does nothing.
func (noop *NoOpRepository) Push(ctx context.Context) error {
	return nil
}
