package vcs

import (
	"github.com/go-git/go-git/v6"
	"github.com/go-git/go-git/v6/plumbing"
	"github.com/go-git/go-git/v6/plumbing/object"
	"github.com/go-git/go-git/v6/plumbing/storer"
)

type commitRelation int

const (
	commitEqual commitRelation = iota
	commitAhead
	commitBehind
	commitDiverged
)

func (projectGit *ProjectGit) compareCommits(localHash, remoteHash plumbing.Hash) (commitRelation, error) {
	if localHash == remoteHash {
		return commitEqual, nil
	}

	localReachable, err := isReachableFrom(projectGit.repository, localHash, remoteHash)
	if err != nil {
		return commitDiverged, err
	}
	if localReachable {
		return commitBehind, nil
	}

	remoteReachable, err := isReachableFrom(projectGit.repository, remoteHash, localHash)
	if err != nil {
		return commitDiverged, err
	}
	if remoteReachable {
		return commitAhead, nil
	}

	return commitDiverged, nil
}

func isReachableFrom(repo *git.Repository, target, from plumbing.Hash) (bool, error) {
	if target == from {
		return true, nil
	}

	start, err := repo.CommitObject(from)
	if err != nil {
		return false, err
	}

	seen := map[plumbing.Hash]struct{}{start.Hash: {}}
	stack := []*object.Commit{start}

	for len(stack) > 0 {
		n := len(stack) - 1
		commit := stack[n]
		stack = stack[:n]

		if commit.Hash == target {
			return true, nil
		}

		parents := commit.Parents()
		err := parents.ForEach(func(p *object.Commit) error {
			if _, ok := seen[p.Hash]; ok {
				return nil
			}
			seen[p.Hash] = struct{}{}
			stack = append(stack, p)
			return nil
		})
		if err != nil && err != storer.ErrStop {
			return false, err
		}
	}

	return false, nil
}
