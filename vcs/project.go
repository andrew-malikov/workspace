package vcs

import (
	"github.com/go-git/go-git/v6"
)

type ProjectGit struct {
	dir        string
	repository *git.Repository
}

func NewProjectGit(dir string) (*ProjectGit, error) {
	repository, err := git.PlainOpen(dir)
	if err != nil {
		return nil, err
	}
	return &ProjectGit{
		repository: repository,
		dir:        dir,
	}, nil
}
