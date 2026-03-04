package projects

import (
	"os"
	"path/filepath"
	"time"

	"github.com/andrew-malikov/workspace/vcs"
)

type Project struct {
	Alias      string
	Dir        string
	Compose    *string
	Migrations *string
}

func (project Project) DoesComposeExist() bool {
	if project.Compose == nil {
		return false
	}

	info, err := os.Stat(filepath.Join(project.Dir, *project.Compose))
	if err != nil {
		return false
	}

	if info.IsDir() {
		return false
	}

	return true
}

func (project Project) DoMigrationsExist() bool {
	if project.Migrations == nil {
		return false
	}

	info, err := os.Stat(filepath.Join(project.Dir, *project.Migrations))
	if err != nil {
		return false
	}

	if !info.IsDir() {
		return false
	}

	return true
}

func (project *Project) ResetCompose() {
	project.Compose = nil
}

func (project *Project) ResetMigrations() {
	project.Migrations = nil
}

type StaleBranch struct {
	IsStale bool
	vcs.Branch
}

var STALE_INTERVAL = time.Hour * 24 * 14

// todo: define list options instead of bunch of args
func (project Project) ListStaleBranches(fetch bool) ([]StaleBranch, error) {
	git, err := vcs.NewProjectGit(project.Dir)
	if err != nil {
		return nil, err
	}

	if err := git.Fetch(); err != nil {
		return nil, err
	}

	branches, err := git.ListBranches()
	if err != nil {
		return nil, err
	}

	staleBranches := make([]StaleBranch, len(branches))
	for index := range branches {
		isStale := false
		if inactive := time.Now().Sub(branches[index].UpdatedAt); inactive > STALE_INTERVAL {
			isStale = true
		}
		staleBranches[index] = StaleBranch{
			IsStale: isStale,
			Branch:  branches[index],
		}
	}

	return staleBranches, nil
}
