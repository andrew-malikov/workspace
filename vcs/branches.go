package vcs

import (
	"time"

	"github.com/go-git/go-git/v6"
	"github.com/go-git/go-git/v6/config"
	"github.com/go-git/go-git/v6/plumbing"
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

const MASTER_BRANCH = "main"

type Branch struct {
	Name      string
	UpdatedAt time.Time
	Author    string
	Status    string
}

func (projectGit *ProjectGit) ListBranches() ([]Branch, error) {
	user := projectGit.ResolveUser()

	refs, err := projectGit.repository.Branches()
	if err != nil {
		return nil, err
	}

	result := make([]Branch, 0)
	err = refs.ForEach(func(ref *plumbing.Reference) error {
		status := "local"
		if branch, _ := projectGit.repository.Branch(ref.Name().Short()); branch != nil && branch.Remote != "" {
			status = "synced"
		}

		commit, err := projectGit.repository.CommitObject(ref.Hash())
		if err != nil {
			return err
		}

		author := commit.Author.Name
		if author == user.Name {
			author = "you"
		}

		result = append(result, Branch{
			Name:      ref.Name().Short(),
			Author:    author,
			Status:    status,
			UpdatedAt: commit.Author.When,
		})

		return nil
	})

	if err != nil {
		return nil, err
	}

	return result, nil
}

type User struct {
	Name  string
	Email string
}

func (user User) hasMissingData() bool {
	return user.Email == "" || user.Name == ""
}

func (user *User) Apply(scopedCfg *config.Config) {
	if user.Email == "" && scopedCfg.User.Email != "" {
		user.Email = scopedCfg.User.Email
	}
	if user.Name == "" && scopedCfg.User.Name != "" {
		user.Name = scopedCfg.User.Name
	}
}

func (projectGit *ProjectGit) ResolveUser() User {
	user := &User{
		Name:  "",
		Email: "",
	}

	localCfg, _ := projectGit.repository.Config()
	if localCfg != nil {
		user.Apply(localCfg)
	}

	if !user.hasMissingData() {
		return *user
	}

	globalCfg, _ := config.LoadConfig(config.GlobalScope)
	if globalCfg != nil {
		user.Apply(globalCfg)
	}

	if !user.hasMissingData() {
		return *user
	}

	systemCfg, _ := config.LoadConfig(config.SystemScope)
	if systemCfg != nil {
		user.Apply(systemCfg)
	}

	return *user
}
