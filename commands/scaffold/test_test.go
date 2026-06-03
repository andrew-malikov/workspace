package scaffold

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	cfg "github.com/andrew-malikov/workspace/config"
	"github.com/andrew-malikov/workspace/projects"
)

func TestResolveProjectMatchesAlias(t *testing.T) {
	config := &cfg.Config{
		Projects: map[string]projects.Project{
			"orders": {Alias: "orders", Dir: filepath.Join(string(os.PathSeparator), "tmp", "orders")},
		},
	}

	project, err := resolveProject(config, "orders")
	if err != nil {
		t.Fatalf("resolve project: %v", err)
	}

	if project.Alias != "orders" {
		t.Fatalf("unexpected project alias: %s", project.Alias)
	}
}

func TestResolveProjectMatchesDirectory(t *testing.T) {
	dir := filepath.Join(string(os.PathSeparator), "tmp", "orders")
	config := &cfg.Config{
		Projects: map[string]projects.Project{
			"orders": {Alias: "orders", Dir: dir},
		},
	}

	project, err := resolveProject(config, filepath.Join(dir, "src"))
	if err != nil {
		t.Fatalf("resolve project: %v", err)
	}

	if project.Alias != "orders" {
		t.Fatalf("unexpected project alias: %s", project.Alias)
	}
}

func TestResolveProjectFailsForUntrackedInput(t *testing.T) {
	config := &cfg.Config{Projects: map[string]projects.Project{}}

	_, err := resolveProject(config, "missing")
	if err == nil {
		t.Fatal("expected untracked project error")
	}

	if !strings.Contains(err.Error(), "project is not tracked: missing") {
		t.Fatalf("unexpected error: %v", err)
	}
}
