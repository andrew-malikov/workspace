package vcs

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/go-git/go-git/v6"
	"github.com/go-git/go-git/v6/config"
	"github.com/go-git/go-git/v6/plumbing"
	"github.com/go-git/go-git/v6/plumbing/object"
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
	Name               string
	Hash               plumbing.Hash
	Ref                plumbing.ReferenceName
	UpdatedAt          time.Time
	Author             string
	OwnedByCurrentUser bool
	Remote             *string
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

func (branch Branch) HasLocal() bool {
	return branch.Remote == nil
}

func (branch Branch) HasRemote() bool {
	if branch.Remote != nil {
		return true
	}
	return branch.Related != nil && branch.Related.Remote != nil
}

func (branch Branch) ResolveRemote() (string, bool) {
	if branch.Remote != nil {
		return *branch.Remote, true
	}
	if branch.Related != nil && branch.Related.Remote != nil {
		return *branch.Related.Remote, true
	}
	return "", false
}

type RelatedBranch struct {
	Name               string
	Hash               plumbing.Hash
	Ref                plumbing.ReferenceName
	UpdatedAt          time.Time
	Author             string
	OwnedByCurrentUser bool
	Remote             *string
	Status             BranchStatus
}

func (projectGit *ProjectGit) Fetch() error {
	err := projectGit.repository.Fetch(&git.FetchOptions{RemoteName: "origin", Prune: true})
	if err != nil && err != git.NoErrAlreadyUpToDate {
		return err
	}
	return nil
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
		remoteKey := localName
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

			commit, err := projectGit.repository.CommitObject(remoteRef.hash)
			if err != nil {
				return nil, err
			}

			author := commit.Author.Name
			if user.Match(commit.Author) {
				author = "you"
			}

			ownedByCurrentUser, err := projectGit.isBranchOwnedByUser(remoteRef.hash, user)
			if err != nil {
				return nil, err
			}

			related = &RelatedBranch{
				Name:               remoteRef.name,
				Hash:               remoteRef.hash,
				Ref:                remoteRef.refName,
				UpdatedAt:          commit.Author.When,
				Author:             author,
				OwnedByCurrentUser: ownedByCurrentUser,
				Remote:             &remoteRef.remote,
				Status:             status,
			}
			seenRemoteKeys[remoteKey] = struct{}{}
		}

		commit, err := projectGit.repository.CommitObject(localRef.hash)
		if err != nil {
			return nil, err
		}

		author := commit.Author.Name
		if user.Match(commit.Author) {
			author = "you"
		}

		ownedByCurrentUser, err := projectGit.isBranchOwnedByUser(localRef.hash, user)
		if err != nil {
			return nil, err
		}

		result = append(result, Branch{
			Name:               localRef.name,
			Hash:               localRef.hash,
			Ref:                localRef.refName,
			Author:             author,
			OwnedByCurrentUser: ownedByCurrentUser,
			UpdatedAt:          commit.Author.When,
			Related:            related,
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
		if user.Match(commit.Author) {
			author = "you"
		}

		ownedByCurrentUser, err := projectGit.isBranchOwnedByUser(remoteRef.hash, user)
		if err != nil {
			return nil, err
		}

		result = append(result, Branch{
			Name:               remoteRef.name,
			Hash:               remoteRef.hash,
			Ref:                remoteRef.refName,
			Author:             author,
			OwnedByCurrentUser: ownedByCurrentUser,
			UpdatedAt:          commit.Author.When,
			Remote:             &remoteRef.remote,
		})
	}

	return result, nil
}

func (projectGit *ProjectGit) isBranchOwnedByUser(branchHash plumbing.Hash, user User) (bool, error) {
	commit, err := projectGit.repository.CommitObject(branchHash)
	if err != nil {
		return false, err
	}

	isOwnedByUser := false
	depth := 0
	err = commit.Parents().ForEach(func(commit *object.Commit) error {
		if depth > 2 || isOwnedByUser {
			return nil
		}

		isOwnedByUser = user.Match(commit.Author)

		depth++

		return nil
	})

	return isOwnedByUser, err
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
	remotes, err := projectGit.repository.Remotes()
	if err != nil {
		return nil, err
	}

	refs, err := projectGit.repository.References()
	if err != nil {
		return nil, err
	}

	branches := make([]branchRef, 0)
	err = refs.ForEach(func(ref *plumbing.Reference) error {
		if !ref.Name().IsRemote() {
			return nil
		}

		short := ref.Name().Short() // e.g. origin/main
		remote, branch, err := splitRemoteBranchShort(short, remotes)
		if err != nil {
			return nil
		}

		if *branch == plumbing.HEAD.Short() {
			return nil
		}

		branches = append(branches, branchRef{
			name:     *branch,
			refName:  ref.Name(),
			hash:     ref.Hash(),
			isRemote: true,
			remote:   remote.Config().Name,
		})
		return nil
	})
	if err != nil {
		return nil, err
	}
	return branches, nil
}

func splitRemoteBranchShort(short string, remotes []*git.Remote) (remote *git.Remote, branch *string, err error) {
	// remote names can be tricky to define e.g. "my-origin/here"
	// is a valid remote name and you can't parse it out of "refs/remotes/my-origin/here/my-branch"
	// without knowing what part the remote name actually takes between the branch and the remote
	for _, remote := range remotes {
		if branch, cutout := strings.CutPrefix(short, remote.Config().Name+"/"); cutout {
			return remote, &branch, nil
		}
	}
	return nil, nil, fmt.Errorf("found no remotes in ref %s", short)
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

func (user User) Match(author object.Signature) bool {
	if user.Email != "" && author.Email != "" {
		return user.Email == author.Email
	}
	if user.Name != "" && author.Name != "" {
		return user.Name == author.Name
	}
	return false
}

func (projectGit *ProjectGit) DeleteBranch(branch Branch) error {
	if branch.HasLocal() {
		if err := projectGit.deleteLocalBranch(branch); err != nil {
			return err
		}
	}

	if remote, ok := branch.ResolveRemote(); ok {
		if err := projectGit.deleteRemoteBranch(remote, branch.Name); err != nil {
			return err
		}
	}

	return nil
}

func (projectGit *ProjectGit) deleteLocalBranch(branch Branch) error {
	head, err := projectGit.repository.Head()
	if err != nil {
		return err
	}

	if head.Name().IsBranch() && head.Name().Short() == branch.Name {
		return fmt.Errorf("cannot delete the checked out branch %s", branch.Name)
	}

	if err := projectGit.repository.Storer.RemoveReference(branch.Ref); err != nil {
		return err
	}

	if err := projectGit.repository.DeleteBranch(branch.Name); err != nil && err != git.ErrBranchNotFound {
		return err
	}

	return nil
}

func (projectGit *ProjectGit) deleteRemoteBranch(remote string, name string) error {
	err := projectGit.repository.Push(&git.PushOptions{
		RemoteName: remote,
		RefSpecs: []config.RefSpec{
			config.RefSpec(fmt.Sprintf(":refs/heads/%s", name)),
		},
	})
	if err != nil && !errors.Is(err, git.NoErrAlreadyUpToDate) {
		return err
	}
	return nil
}
