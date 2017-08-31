// Package git encapsulates git functionality required by the timecard utility.
package git

////////////////////////////////////////////////////////////////////////////////

import (
	"errors"

	"gopkg.in/src-d/go-git.v4"
	// "gopkg.in/src-d/go-git.v4/plumbing/object"
)

////////////////////////////////////////////////////////////////////////////////

var (
	ErrNotGitRepo = errors.New("not a valid git repo")
)

////////////////////////////////////////////////////////////////////////////////

type Git struct {
	cwd  string
	repo *git.Repository
}

func New(dp string) (*Git, error) {
	r, err := git.PlainOpen(dp)
	if err != nil {
		return nil, err
	}

	return &Git{
		cwd:  dp,
		repo: r,
	}, nil
}

////////////////////////////////////////////////////////////////////////////////

// Returns the current commit has for the git repo.
func (g *Git) GetCurrentHash() (string, error) {
	head, err := g.repo.Head()
	if err != nil {
		return "", err
	}
	return head.Hash().String(), nil
}

////////////////////////////////////////////////////////////////////////////////
