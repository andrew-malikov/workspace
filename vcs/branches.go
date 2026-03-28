package vcs

import (
	"errors"
	"fmt"
	"regexp"
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
	MatchingCommits    []BranchCommitPreview
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
	MatchingCommits    []BranchCommitPreview
	Remote             *string
	Status             BranchStatus
}

type BranchCommitPreview struct {
	Author  string
	Subject string
	At      time.Time
}

type BranchOwnershipFilterInput struct {
	AuthorEmails    []string
	AuthorNames     []string
	MessagePatterns []string
}

type BranchOwnershipFilter struct {
	authorEmails    map[string]struct{}
	authorNames     map[string]struct{}
	messagePatterns []*regexp.Regexp
}

type BranchOwnershipOptions struct {
	LookbackCommits int
	Include         BranchOwnershipFilter
	Exclude         BranchOwnershipFilter
}

func NewBranchOwnershipOptions(lookbackCommits int, include BranchOwnershipFilterInput, exclude BranchOwnershipFilterInput) (BranchOwnershipOptions, error) {
	compiledInclude, err := compileBranchOwnershipFilter(include)
	if err != nil {
		return BranchOwnershipOptions{}, err
	}

	compiledExclude, err := compileBranchOwnershipFilter(exclude)
	if err != nil {
		return BranchOwnershipOptions{}, err
	}

	if lookbackCommits <= 0 {
		lookbackCommits = 3
	}

	return BranchOwnershipOptions{
		LookbackCommits: lookbackCommits,
		Include:         compiledInclude,
		Exclude:         compiledExclude,
	}, nil
}

func compileBranchOwnershipFilter(input BranchOwnershipFilterInput) (BranchOwnershipFilter, error) {
	filter := BranchOwnershipFilter{
		authorEmails:    make(map[string]struct{}, len(input.AuthorEmails)),
		authorNames:     make(map[string]struct{}, len(input.AuthorNames)),
		messagePatterns: make([]*regexp.Regexp, 0, len(input.MessagePatterns)),
	}

	for _, email := range input.AuthorEmails {
		if email == "" {
			continue
		}
		filter.authorEmails[email] = struct{}{}
	}

	for _, name := range input.AuthorNames {
		if name == "" {
			continue
		}
		filter.authorNames[name] = struct{}{}
	}

	for _, pattern := range input.MessagePatterns {
		if pattern == "" {
			continue
		}
		compiled, err := regexp.Compile(pattern)
		if err != nil {
			return BranchOwnershipFilter{}, fmt.Errorf("invalid ownership message regex %q: %w", pattern, err)
		}
		filter.messagePatterns = append(filter.messagePatterns, compiled)
	}

	return filter, nil
}

func (filter BranchOwnershipFilter) HasRules() bool {
	return len(filter.authorEmails) > 0 || len(filter.authorNames) > 0 || len(filter.messagePatterns) > 0
}

func (filter BranchOwnershipFilter) MatchCommit(commit *object.Commit) bool {
	if len(filter.authorEmails) > 0 {
		if _, ok := filter.authorEmails[commit.Author.Email]; ok {
			return true
		}
	}

	if len(filter.authorNames) > 0 {
		if _, ok := filter.authorNames[commit.Author.Name]; ok {
			return true
		}
	}

	for _, pattern := range filter.messagePatterns {
		if pattern.MatchString(commit.Message) {
			return true
		}
	}

	return false
}

func (options BranchOwnershipOptions) HasIncludeRules() bool {
	return options.Include.HasRules()
}

func (options BranchOwnershipOptions) HasExcludeRules() bool {
	return options.Exclude.HasRules()
}

func (projectGit *ProjectGit) Fetch() error {
	err := projectGit.repository.Fetch(&git.FetchOptions{RemoteName: "origin", Prune: true})
	if err != nil && err != git.NoErrAlreadyUpToDate {
		return err
	}
	return nil
}

func (projectGit *ProjectGit) ListBranches(ownership BranchOwnershipOptions) ([]Branch, error) {
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

			ownershipResult, err := projectGit.resolveBranchOwnership(remoteRef.hash, user, ownership)
			if err != nil {
				return nil, err
			}

			related = &RelatedBranch{
				Name:               remoteRef.name,
				Hash:               remoteRef.hash,
				Ref:                remoteRef.refName,
				UpdatedAt:          commit.Author.When,
				Author:             author,
				OwnedByCurrentUser: ownershipResult.Matched,
				MatchingCommits:    ownershipResult.Previews,
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

		ownershipResult, err := projectGit.resolveBranchOwnership(localRef.hash, user, ownership)
		if err != nil {
			return nil, err
		}

		result = append(result, Branch{
			Name:               localRef.name,
			Hash:               localRef.hash,
			Ref:                localRef.refName,
			Author:             author,
			OwnedByCurrentUser: ownershipResult.Matched,
			MatchingCommits:    ownershipResult.Previews,
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

		ownershipResult, err := projectGit.resolveBranchOwnership(remoteRef.hash, user, ownership)
		if err != nil {
			return nil, err
		}

		result = append(result, Branch{
			Name:               remoteRef.name,
			Hash:               remoteRef.hash,
			Ref:                remoteRef.refName,
			Author:             author,
			OwnedByCurrentUser: ownershipResult.Matched,
			MatchingCommits:    ownershipResult.Previews,
			UpdatedAt:          commit.Author.When,
			Remote:             &remoteRef.remote,
		})
	}

	return result, nil
}

type branchOwnershipResult struct {
	Matched  bool
	Previews []BranchCommitPreview
}

func (projectGit *ProjectGit) isBranchOwnedByUser(branchHash plumbing.Hash, user User, ownership BranchOwnershipOptions) (bool, error) {
	result, err := projectGit.resolveBranchOwnership(branchHash, user, ownership)
	if err != nil {
		return false, err
	}

	return result.Matched, nil
}

func (projectGit *ProjectGit) resolveBranchOwnership(branchHash plumbing.Hash, user User, ownership BranchOwnershipOptions) (branchOwnershipResult, error) {
	commits, err := projectGit.listBranchCommits(branchHash, ownership.LookbackCommits)
	if err != nil {
		return branchOwnershipResult{}, err
	}

	return ownership.MatchCommitsWithPreviews(commits, user), nil
}

func (options BranchOwnershipOptions) MatchCommits(commits []*object.Commit, user User) bool {
	return options.MatchCommitsWithPreviews(commits, user).Matched
}

func (options BranchOwnershipOptions) MatchCommitsWithPreviews(commits []*object.Commit, user User) branchOwnershipResult {
	included := false
	useFallbackUser := !options.HasIncludeRules()
	previews := make([]BranchCommitPreview, 0, 3)

	for _, commit := range commits {
		if options.HasExcludeRules() && options.Exclude.MatchCommit(commit) {
			return branchOwnershipResult{}
		}

		if options.HasIncludeRules() {
			if options.Include.MatchCommit(commit) {
				included = true
				previews = appendBranchCommitPreview(previews, commit, user)
			}
			continue
		}

		if useFallbackUser && user.Match(commit.Author) {
			included = true
			previews = appendBranchCommitPreview(previews, commit, user)
		}
	}

	return branchOwnershipResult{
		Matched:  included,
		Previews: previews,
	}
}

func appendBranchCommitPreview(previews []BranchCommitPreview, commit *object.Commit, user User) []BranchCommitPreview {
	if len(previews) >= 3 {
		return previews
	}

	author := commit.Author.Name
	if user.Match(commit.Author) {
		author = "you"
	}

	subject := strings.TrimSpace(commit.Message)
	if index := strings.IndexByte(subject, '\n'); index >= 0 {
		subject = subject[:index]
	}
	if subject == "" {
		subject = "(no commit message)"
	}

	return append(previews, BranchCommitPreview{
		Author:  author,
		Subject: subject,
		At:      commit.Author.When,
	})
}

func (projectGit *ProjectGit) listBranchCommits(branchHash plumbing.Hash, limit int) ([]*object.Commit, error) {
	if limit <= 0 {
		limit = 3
	}

	start, err := projectGit.repository.CommitObject(branchHash)
	if err != nil {
		return nil, err
	}

	commits := make([]*object.Commit, 0, limit)
	seen := map[plumbing.Hash]struct{}{}
	stack := []*object.Commit{start}

	for len(stack) > 0 && len(commits) < limit {
		n := len(stack) - 1
		commit := stack[n]
		stack = stack[:n]

		if _, ok := seen[commit.Hash]; ok {
			continue
		}
		seen[commit.Hash] = struct{}{}
		commits = append(commits, commit)

		parents := make([]*object.Commit, 0, commit.NumParents())
		err := commit.Parents().ForEach(func(parent *object.Commit) error {
			parents = append(parents, parent)
			return nil
		})
		if err != nil {
			return nil, err
		}

		for index := len(parents) - 1; index >= 0; index-- {
			if _, ok := seen[parents[index].Hash]; ok {
				continue
			}
			stack = append(stack, parents[index])
		}
	}

	return commits, nil
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
