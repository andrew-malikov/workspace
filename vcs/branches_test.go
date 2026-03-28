package vcs

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/go-git/go-git/v6"
	"github.com/go-git/go-git/v6/plumbing"
	"github.com/go-git/go-git/v6/plumbing/object"
)

func TestBranchOwnershipMatchesIncludeRules(t *testing.T) {
	options, err := NewBranchOwnershipOptions(5,
		BranchOwnershipFilterInput{
			AuthorEmails:    []string{"me@example.com"},
			AuthorNames:     []string{"Me"},
			MessagePatterns: []string{"\\[mine\\]"},
		},
		BranchOwnershipFilterInput{},
	)
	if err != nil {
		t.Fatalf("new options: %v", err)
	}

	commits := []*object.Commit{
		makeCommit("Other", "other@example.com", "refactor"),
		makeCommit("Me", "other@example.com", "refactor"),
	}

	if !options.MatchCommits(commits, User{}) {
		t.Fatal("expected include rule to match commit author name")
	}

	commits = []*object.Commit{makeCommit("Other", "me@example.com", "refactor")}
	if !options.MatchCommits(commits, User{}) {
		t.Fatal("expected include rule to match commit author email")
	}

	commits = []*object.Commit{makeCommit("Other", "other@example.com", "ship [mine]")}
	if !options.MatchCommits(commits, User{}) {
		t.Fatal("expected include rule to match commit message")
	}
}

func TestBranchOwnershipExcludeWins(t *testing.T) {
	options, err := NewBranchOwnershipOptions(5,
		BranchOwnershipFilterInput{AuthorEmails: []string{"me@example.com"}},
		BranchOwnershipFilterInput{MessagePatterns: []string{"^chore"}},
	)
	if err != nil {
		t.Fatalf("new options: %v", err)
	}

	commits := []*object.Commit{
		makeCommit("Me", "me@example.com", "feat: useful"),
		makeCommit("Release Bot", "bot@example.com", "chore: release"),
	}

	if options.MatchCommits(commits, User{}) {
		t.Fatal("expected exclude rule to win over include")
	}
}

func TestBranchOwnershipFallsBackToCurrentUserWhenNoIncludeRules(t *testing.T) {
	options, err := NewBranchOwnershipOptions(5, BranchOwnershipFilterInput{}, BranchOwnershipFilterInput{})
	if err != nil {
		t.Fatalf("new options: %v", err)
	}

	user := User{Name: "Me", Email: "me@example.com"}
	commits := []*object.Commit{makeCommit("Me", "me@example.com", "feat: useful")}

	if !options.MatchCommits(commits, user) {
		t.Fatal("expected fallback current user match")
	}
}

func TestBranchOwnershipIncludeDisablesFallback(t *testing.T) {
	options, err := NewBranchOwnershipOptions(5,
		BranchOwnershipFilterInput{MessagePatterns: []string{"^feat\\(mine\\):"}},
		BranchOwnershipFilterInput{},
	)
	if err != nil {
		t.Fatalf("new options: %v", err)
	}

	user := User{Name: "Me", Email: "me@example.com"}
	commits := []*object.Commit{makeCommit("Me", "me@example.com", "feat: useful")}

	if options.MatchCommits(commits, user) {
		t.Fatal("expected explicit include rules to disable fallback current-user matching")
	}
}

func TestBranchOwnershipMatchingPreviewsUseLatestMatchingCommits(t *testing.T) {
	options, err := NewBranchOwnershipOptions(6,
		BranchOwnershipFilterInput{MessagePatterns: []string{"^feat:"}},
		BranchOwnershipFilterInput{},
	)
	if err != nil {
		t.Fatalf("new options: %v", err)
	}

	result := options.MatchCommitsWithPreviews([]*object.Commit{
		makeCommit("Me", "me@example.com", "feat: four\n\nbody"),
		makeCommit("Me", "me@example.com", "fix: skipped"),
		makeCommit("Me", "me@example.com", "feat: three"),
		makeCommit("Me", "me@example.com", "feat: two"),
		makeCommit("Me", "me@example.com", "feat: one"),
	}, User{Name: "Me", Email: "me@example.com"})

	if !result.Matched {
		t.Fatal("expected branch to match include rules")
	}

	got := previewsToStrings(result.Previews)
	want := []string{"you: feat: four", "you: feat: three", "you: feat: two"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected previews: got %v want %v", got, want)
	}
}

func TestBranchOwnershipMatchingPreviewsFallbackToCurrentUser(t *testing.T) {
	options, err := NewBranchOwnershipOptions(5, BranchOwnershipFilterInput{}, BranchOwnershipFilterInput{})
	if err != nil {
		t.Fatalf("new options: %v", err)
	}

	result := options.MatchCommitsWithPreviews([]*object.Commit{
		makeCommit("Me", "me@example.com", "feat: latest mine"),
		makeCommit("Other", "other@example.com", "feat: not mine"),
		makeCommit("Me", "me@example.com", "feat: older mine"),
	}, User{Name: "Me", Email: "me@example.com"})

	if !result.Matched {
		t.Fatal("expected fallback current-user match")
	}

	got := previewsToStrings(result.Previews)
	want := []string{"you: feat: latest mine", "you: feat: older mine"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected previews: got %v want %v", got, want)
	}
}

func TestBranchOwnershipMatchingPreviewsExcludedBranchReturnsNone(t *testing.T) {
	options, err := NewBranchOwnershipOptions(5,
		BranchOwnershipFilterInput{AuthorEmails: []string{"me@example.com"}},
		BranchOwnershipFilterInput{MessagePatterns: []string{"^chore:"}},
	)
	if err != nil {
		t.Fatalf("new options: %v", err)
	}

	result := options.MatchCommitsWithPreviews([]*object.Commit{
		makeCommit("Me", "me@example.com", "feat: mine"),
		makeCommit("Bot", "bot@example.com", "chore: release"),
	}, User{Name: "Me", Email: "me@example.com"})

	if result.Matched {
		t.Fatal("expected exclude rule to suppress branch match")
	}

	if len(result.Previews) != 0 {
		t.Fatalf("expected no previews when branch is excluded, got %v", previewsToStrings(result.Previews))
	}
}

func TestListBranchCommitsHonorsLookbackAndIncludesHead(t *testing.T) {
	dir := t.TempDir()
	repo, hashes := createCommitHistory(t, dir, []commitSpec{
		{name: "Dev", email: "dev@example.com", message: "one"},
		{name: "Dev", email: "dev@example.com", message: "two"},
		{name: "Dev", email: "dev@example.com", message: "three"},
		{name: "Dev", email: "dev@example.com", message: "four"},
	})

	projectGit := &ProjectGit{dir: dir, repository: repo}
	commits, err := projectGit.listBranchCommits(hashes[len(hashes)-1], 3)
	if err != nil {
		t.Fatalf("list branch commits: %v", err)
	}

	got := make([]string, 0, len(commits))
	for _, commit := range commits {
		got = append(got, commit.Message)
	}

	want := []string{"four", "three", "two"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected commits: got %v want %v", got, want)
	}
}

func makeCommit(name, email, message string) *object.Commit {
	return &object.Commit{
		Author:  object.Signature{Name: name, Email: email},
		Message: message,
	}
}

func previewsToStrings(previews []BranchCommitPreview) []string {
	result := make([]string, 0, len(previews))
	for _, preview := range previews {
		result = append(result, preview.Author+": "+preview.Subject)
	}
	return result
}

type commitSpec struct {
	name    string
	email   string
	message string
}

func createCommitHistory(t *testing.T, dir string, specs []commitSpec) (*git.Repository, []plumbing.Hash) {
	t.Helper()

	repo, err := git.PlainInit(dir, false)
	if err != nil {
		t.Fatalf("init repo: %v", err)
	}

	worktree, err := repo.Worktree()
	if err != nil {
		t.Fatalf("worktree: %v", err)
	}

	hashes := make([]plumbing.Hash, 0, len(specs))
	for index, spec := range specs {
		filePath := filepath.Join(dir, "history.txt")
		content := []byte(spec.message + "\n")
		if err := os.WriteFile(filePath, content, 0o644); err != nil {
			t.Fatalf("write file: %v", err)
		}

		if _, err := worktree.Add("history.txt"); err != nil {
			t.Fatalf("add file: %v", err)
		}

		hash, err := worktree.Commit(spec.message, &git.CommitOptions{
			Author: &object.Signature{
				Name:  spec.name,
				Email: spec.email,
				When:  time.Unix(int64(index+1), 0),
			},
		})
		if err != nil {
			t.Fatalf("commit: %v", err)
		}

		hashes = append(hashes, hash)
	}

	return repo, hashes
}
