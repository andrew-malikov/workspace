package vcs

import (
	"time"

	"github.com/go-git/go-git/v6"
	"github.com/go-git/go-git/v6/config"
	"github.com/go-git/go-git/v6/plumbing"
)

type BranchStatus = int

const (
	SyncedBranch BranchStatus = iota
	AheadBranch
	BehindBranch
	DivergedBranch
)

// Types of branches
//
//  1. local only
//
//     has no Remote and Related
//
//  2. remote only
//
//     has Remote and no Related
//
//  3. local with remote
//
//     has Related and no Remote and Related has Remote
type Branch struct {
	Name      string
	UpdatedAt time.Time
	Author    string
	Remote    *string
	// todo: sure there may be many remotes
	// 		 it looks like branches are grouped by some kind of ref
	// 	     so that even different refs can be accostomed via different upstreams
	// 		 which gives a net of whatever that is grouped by whatever but that's not
	// 		 the problem to deal with here, in most scenarios there's one remote only,
	// 		 and most of the branches either behind, ahead or synced, the diverged one
	// 		 is a broken one, people usually try to resolve, but hanging branches alike
	// 		 might still exist, because people are still messy mess that digress
	Related *RelatedBranch
}

type RelatedBranch struct {
	Name      string
	UpdatedAt time.Time
	Author    string
	Remote    *string
	Status    BranchStatus
}

func (projectGit *ProjectGit) Fetch() error {
	return projectGit.repository.Fetch(&git.FetchOptions{Prune: true})
}

func (projectGit *ProjectGit) ListBranches() ([]Branch, error) {
	user := projectGit.ResolveUser()

	local, err := projectGit.getLocalBranches()
	if err != nil {
		return nil, err
	}
	localByName := make(map[string]branchRef)
	for _, branch := range local {
		localByName[branch.name] = branch
	}

	remote, err := projectGit.getRemoteBranches()
	if err != nil {
		return nil, err
	}
	remoteByName := make(map[string]branchRef)
	for _, branch := range remote {
		remoteByName[branch.name] = branch
	}

	result := make([]Branch, 0, len(localByName)+len(remoteByName))
	seenRemoteKeys := make(map[string]struct{})

	for localName, localRef := range localByName {
		remoteKey := "origin/" + localName
		remoteRef, hasRemote := remoteByName[remoteKey]

		var related *RelatedBranch
		if hasRemote {
			relation, relErr := projectGit.compareCommits(localRef.hash, remoteRef.hash)
			if relErr != nil {
				return nil, relErr
			}
			var status BranchStatus
			switch relation {
			case commitEqual:
				status = SyncedBranch
			case commitAhead:
				status = AheadBranch
			case commitBehind:
				status = BehindBranch
			case commitDiverged:
				status = DivergedBranch
			}
			related = &RelatedBranch{
				Name:      remoteRef.name,
				UpdatedAt: time.Now(),
				Author:    "",
				Remote:    &remoteRef.remote,
				Status:    status,
			}
			seenRemoteKeys[remoteKey] = struct{}{}
		}

		commit, err := projectGit.repository.CommitObject(localRef.hash)
		if err != nil {
			return nil, err
		}

		author := commit.Author.Name
		if author == user.Name {
			author = "you"
		}

		result = append(result, Branch{
			Name:      localRef.name,
			Author:    author,
			UpdatedAt: commit.Author.When,
			Related:   related,
		})
	}

	for key, remoteRef := range remoteByName {
		if _, seen := seenRemoteKeys[key]; seen {
			continue
		}

		commit, err := projectGit.repository.CommitObject(remoteRef.hash)
		if err != nil {
			return nil, err
		}

		author := commit.Author.Name
		if author == user.Name {
			author = "you"
		}

		result = append(result, Branch{
			Name:      remoteRef.name,
			Author:    author,
			UpdatedAt: commit.Author.When,
			Remote:    &remoteRef.remote,
		})
	}

	return result, nil
}

type branchRef struct {
	name     string
	refName  plumbing.ReferenceName
	hash     plumbing.Hash
	isRemote bool
	remote   string
}

func (projectGit *ProjectGit) getLocalBranches() ([]branchRef, error) {
	localIter, err := projectGit.repository.Branches()
	if err != nil {
		return nil, err
	}

	branches := make([]branchRef, 0)
	err = localIter.ForEach(func(ref *plumbing.Reference) error {
		name := ref.Name().Short()
		branches = append(branches, branchRef{
			name:     name,
			refName:  ref.Name(),
			hash:     ref.Hash(),
			isRemote: false,
		})
		return nil
	})
	if err != nil {
		return nil, err
	}
	return branches, nil
}

func (projectGit *ProjectGit) getRemoteBranches() ([]branchRef, error) {
	refs, err := projectGit.repository.References()
	if err != nil {
		return nil, err
	}

	branches := make([]branchRef, 0)
	err = refs.ForEach(func(ref *plumbing.Reference) error {
		if !ref.Name().IsRemote() || !ref.Name().IsBranch() {
			return nil
		}

		short := ref.Name().Short() // e.g. origin/main
		remote, branchName := splitRemoteBranchShort(short)
		if remote == "" || branchName == "" {
			return nil
		}

		key := remote + "/" + branchName
		branches = append(branches, branchRef{
			name:     key,
			refName:  ref.Name(),
			hash:     ref.Hash(),
			isRemote: true,
			remote:   remote,
		})
		return nil
	})
	if err != nil {
		return nil, err
	}
	return branches, nil
}

func splitRemoteBranchShort(short string) (remote string, branch string) {
	for i := 0; i < len(short); i++ {
		if short[i] == '/' {
			if i == 0 || i == len(short)-1 {
				return "", ""
			}
			return short[:i], short[i+1:]
		}
	}
	return "", ""
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
