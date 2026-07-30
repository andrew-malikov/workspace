package dotnet

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/andrew-malikov/workspace/projects"
)

func TestNewTestRunCreatesUniqueProjectLocalRoots(t *testing.T) {
	projectDir := t.TempDir()

	first, err := NewTestRun(projectDir)
	if err != nil {
		t.Fatalf("create first test run: %v", err)
	}
	second, err := NewTestRun(projectDir)
	if err != nil {
		t.Fatalf("create second test run: %v", err)
	}

	logsRoot := filepath.Join(projectDir, ".logs", "tests") + string(filepath.Separator)
	if !strings.HasPrefix(first.Root, logsRoot) {
		t.Fatalf("first root is not project-local: %q", first.Root)
	}
	if first.Root == second.Root {
		t.Fatalf("test run roots must be unique: %q", first.Root)
	}
}

func TestTestRunCreatesCategoryArtifactsUnderSharedRoot(t *testing.T) {
	run, err := NewTestRun(t.TempDir())
	if err != nil {
		t.Fatalf("create test run: %v", err)
	}

	unit, err := run.createArtifacts(projects.UnitTestKind)
	if err != nil {
		t.Fatalf("create unit artifacts: %v", err)
	}
	component, err := run.createArtifacts(projects.ComponentTestKind)
	if err != nil {
		t.Fatalf("create component artifacts: %v", err)
	}

	if filepath.Dir(filepath.Dir(unit.logPath)) != run.Root {
		t.Fatalf("unit log is not under shared root: %q", unit.logPath)
	}
	if filepath.Dir(filepath.Dir(component.logPath)) != run.Root {
		t.Fatalf("component log is not under shared root: %q", component.logPath)
	}
}
