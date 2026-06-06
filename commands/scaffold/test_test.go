package scaffold

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/andrew-malikov/workspace/projects"
	"github.com/andrew-malikov/workspace/workspaces"
)

func TestResolveProjectMatchesAlias(t *testing.T) {
	workspace := &workspaces.Workspace{
		Projects: map[string]projects.Project{
			"orders": {Alias: "orders", Dir: filepath.Join(string(os.PathSeparator), "tmp", "orders")},
		},
	}

	project, err := resolveProject(workspace, "orders")
	if err != nil {
		t.Fatalf("resolve project: %v", err)
	}

	if project.Alias != "orders" {
		t.Fatalf("unexpected project alias: %s", project.Alias)
	}
}

func TestResolveProjectMatchesDirectory(t *testing.T) {
	dir := filepath.Join(string(os.PathSeparator), "tmp", "orders")
	workspace := &workspaces.Workspace{
		Projects: map[string]projects.Project{
			"orders": {Alias: "orders", Dir: dir},
		},
	}

	project, err := resolveProject(workspace, filepath.Join(dir, "src"))
	if err != nil {
		t.Fatalf("resolve project: %v", err)
	}

	if project.Alias != "orders" {
		t.Fatalf("unexpected project alias: %s", project.Alias)
	}
}

func TestResolveProjectFailsForUntrackedInput(t *testing.T) {
	workspace := &workspaces.Workspace{Projects: map[string]projects.Project{}}

	_, err := resolveProject(workspace, "missing")
	if err == nil {
		t.Fatal("expected untracked project error")
	}

	if !strings.Contains(err.Error(), "project is not tracked: missing") {
		t.Fatalf("unexpected error: %v", err)
	}
}
