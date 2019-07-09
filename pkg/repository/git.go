package repository

import (
	"context"
	"fmt"
	"strings"

	"gopkg.in/src-d/go-git.v4/config"

	"gopkg.in/src-d/go-git.v4/plumbing"

	"gopkg.in/src-d/go-git.v4"
)

type VersioningController struct {
	client *git.Repository
}

func New(path string) (*VersioningController, error) {
	repo, err := git.PlainOpen(path)
	if err != nil {
		return nil, err
	}
	return &VersioningController{
		client: repo,
	}, nil
}

func (vc *VersioningController) LatestCommitHash() string {
	headRef, err := vc.client.Head()
	if err != nil {
		return ""
	}

	return headRef.Hash().String()
}

func (vc *VersioningController) ExistsTag(version string) (bool, error) {
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

func (vc *VersioningController) Tags() []string {
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

func (vc *VersioningController) IsSafe(ctx context.Context) error {
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

func (vc *VersioningController) HasUncommittedChanges() (bool, error) {
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

func (vc *VersioningController) HasStagedChanges() bool {
	// TODO: unsupported
	return false
}

func (vc *VersioningController) IsBehind(ctx context.Context) (bool, error) {
	if err := vc.client.FetchContext(ctx, &git.FetchOptions{}); err != nil {
		return false, err
	}

	return true, nil
}

func (vc *VersioningController) CreateTag(tag string) error {
	name := plumbing.ReferenceName(fmt.Sprintf("refs/tags/%v", tag))
	reference := plumbing.NewHashReference(name, plumbing.NewHash(vc.LatestCommitHash()))
	return vc.client.Storer.SetReference(reference)
}

func (vc *VersioningController) DeleteTag(tag string) error {
	return vc.client.Storer.RemoveReference(
		plumbing.ReferenceName(fmt.Sprintf("refs/tags/%v", tag)),
	)
}

func (vc *VersioningController) Push(ctx context.Context) error {
	return vc.client.PushContext(ctx, &git.PushOptions{
		RefSpecs: []config.RefSpec{"refs/tags/*:refs/tags/*"},
	})
}
