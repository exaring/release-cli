package repository

import (
	"context"
	"fmt"
	"os"
	"strings"

	"gopkg.in/src-d/go-git.v4/config"

	"gopkg.in/src-d/go-git.v4/plumbing"

	"gopkg.in/src-d/go-git.v4"
)

type Git struct {
	client *git.Repository
}

func New(path string) (*Git, error) {
	repo, err := git.PlainOpen(path)
	if err != nil {
		return nil, err
	}
	return &Git{
		client: repo,
	}, nil
}

func (vc *Git) LatestCommitHash() string {
	headRef, err := vc.client.Head()
	if err != nil {
		return ""
	}

	return headRef.Hash().String()
}

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

func (vc *Git) IsSafe(ctx context.Context) error {
	if hasUncomittedChanges, err := vc.HasUncommittedChanges(); err != nil {
		return err
	} else if hasUncomittedChanges {
		return fmt.Errorf("your client has uncommited changes.")
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

func (vc *Git) HasStagedChanges() bool {
	// TODO: unsupported function from the go-git lib
	return false
}

func (vc *Git) IsBehind(ctx context.Context) (bool, error) {
	if err := vc.client.FetchContext(ctx, &git.FetchOptions{}); err != nil {
		return false, err
	}

	return true, nil
}

func (vc *Git) CreateTag(tag string) error {
	name := plumbing.ReferenceName(fmt.Sprintf("refs/tags/%v", tag))
	reference := plumbing.NewHashReference(name, plumbing.NewHash(vc.LatestCommitHash()))
	return vc.client.Storer.SetReference(reference)
}

func (vc *Git) DeleteTag(tag string) error {
	return vc.client.Storer.RemoveReference(
		plumbing.ReferenceName(fmt.Sprintf("refs/tags/%v", tag)),
	)
}

func (vc *Git) Push(ctx context.Context) error {
	return vc.client.PushContext(ctx, &git.PushOptions{
		RefSpecs: []config.RefSpec{"refs/tags/*:refs/tags/*"},
	})
}

type NoOpRepository struct {
}

func NewNoOp() *NoOpRepository {
	return &NoOpRepository{}
}

func (noop *NoOpRepository) LatestCommitHash() string {
	return ""
}

func (noop *NoOpRepository) ExistsTag(version string) (bool, error) {
	return true, nil
}

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

func (noop *NoOpRepository) IsSafe(ctx context.Context) error {
	return nil
}

func (noop *NoOpRepository) CreateTag(tag string) error {
	return nil
}

func (noop *NoOpRepository) DeleteTag(tag string) error {
	return nil
}

func (noop *NoOpRepository) Push(ctx context.Context) error {
	return nil
}
