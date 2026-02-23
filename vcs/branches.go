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

func (projectGit *ProjectGit) Fetch() error {
	return projectGit.repository.Fetch(&git.FetchOptions{Prune: true})
}

func (projectGit *ProjectGit) ListBranches() ([]Branch, error) {
	user := projectGit.ResolveUser()

	// types of branches
	// 1. local only
	// 2. remote only
	// 3. local is synced to remote
	// 4. local is latest, remote is out of sync
	// 5. remote is latest, local is out of sync
	// 6. local is diverged from remote

	branches, err := projectGit.repository.Branches()
	if err != nil {
		return nil, err
	}

	// todo: surely list all the remotes and get through all of them
	//       though this way shoots out 99.9% cases
	remote, err := projectGit.repository.Remote("origin")
	if err != nil {
		return nil, err
	}

	originRefs, err := remote.List(&git.ListOptions{
		PeelingOption: git.AppendPeeled,
	})
	if err != nil {
		return nil, err
	}

	result := make([]Branch, 0)
	err = branches.ForEach(func(ref *plumbing.Reference) error {
		status := "local"
		for _, originRef := range originRefs {
			if !originRef.Name().IsBranch() {
				continue
			}
			// todo: check the actual upstream for the branch instead of guessing through equality
			if originRef.Name().Short() == ref.Name().Short() {
				status = "synced"
				break
			}
		}

		commit, err := projectGit.repository.CommitObject(ref.Hash())
		if err != nil {
			return err
		}

		author := commit.Author.Name
		if author == user.Name {
			author = "you"
		}

		// todo: keeping the remote name is 100% important
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
