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
	Test       TestConfig `toml:"test"`
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
func (project Project) ListStaleBranches(fetch bool, ownership vcs.BranchOwnershipOptions) ([]StaleBranch, error) {
	git, err := vcs.NewProjectGit(project.Dir)
	if err != nil {
		return nil, err
	}

	if fetch {
		if err := git.Fetch(); err != nil {
			return nil, err
		}
	}

	branches, err := git.ListBranches(ownership)
	if err != nil {
		return nil, err
	}

	staleBranches := make([]StaleBranch, 0, len(branches))
	for _, branch := range branches {
		if branch.OwnedByCurrentUser || (branch.Related != nil && branch.Related.OwnedByCurrentUser) {
			isStale := false
			if inactive := time.Now().Sub(branch.UpdatedAt); inactive > STALE_INTERVAL {
				isStale = true
			}

			staleBranches = append(staleBranches, StaleBranch{
				IsStale: isStale,
				Branch:  branch,
			})
		}

	}

	return staleBranches, nil
}
