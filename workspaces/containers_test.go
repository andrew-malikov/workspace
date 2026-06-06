package workspaces

import (
	"context"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/andrew-malikov/workspace/projects"
)

type composeCall struct {
	Action  string
	Dir     string
	Compose string
}

type composeSpy struct {
	running map[string]bool
	calls   []composeCall
}

func (spy *composeSpy) HasRunning(ctx context.Context, dir string, compose string) (bool, error) {
	spy.calls = append(spy.calls, composeCall{Action: "ps", Dir: dir, Compose: compose})
	return spy.running[dir], nil
}

func (spy *composeSpy) Up(ctx context.Context, dir string, compose string) error {
	spy.calls = append(spy.calls, composeCall{Action: "up", Dir: dir, Compose: compose})
	return nil
}

func (spy *composeSpy) Down(ctx context.Context, dir string, compose string) error {
	spy.calls = append(spy.calls, composeCall{Action: "down", Dir: dir, Compose: compose})
	return nil
}

func TestResolveProjectDefaultsToCurrentDirectory(t *testing.T) {
	dir := filepath.Join(string(os.PathSeparator), "tmp", "orders")
	workspace := Workspace{
		Projects: map[string]projects.Project{
			"orders": {Alias: "orders", Dir: dir},
		},
	}

	project, err := workspace.ResolveProject(filepath.Join(dir, "src"), "")
	if err != nil {
		t.Fatalf("resolve project: %v", err)
	}

	if project.Alias != "orders" {
		t.Fatalf("unexpected project alias: %s", project.Alias)
	}
}

func TestResolveProjectMatchesAlias(t *testing.T) {
	workspace := Workspace{
		Projects: map[string]projects.Project{
			"orders": {Alias: "orders", Dir: filepath.Join(string(os.PathSeparator), "tmp", "orders")},
		},
	}

	project, err := workspace.ResolveProject("/tmp/other", "orders")
	if err != nil {
		t.Fatalf("resolve project: %v", err)
	}

	if project.Alias != "orders" {
		t.Fatalf("unexpected project alias: %s", project.Alias)
	}
}

func TestUpProjectStopsOtherRunningComposeByDefault(t *testing.T) {
	targetDir := trackedProjectDir(t)
	otherDir := trackedProjectDir(t)
	compose := "docker-compose.yaml"
	workspace := Workspace{
		Projects: map[string]projects.Project{
			"orders":  {Alias: "orders", Dir: targetDir, Compose: &compose},
			"billing": {Alias: "billing", Dir: otherDir, Compose: &compose},
		},
	}
	composeSpy := &composeSpy{running: map[string]bool{otherDir: true}}

	result, err := workspace.UpProject(context.Background(), targetDir, "", false, composeSpy)
	if err != nil {
		t.Fatalf("up project: %v", err)
	}

	if result.Alias != "orders" {
		t.Fatalf("unexpected target: %s", result.Alias)
	}
	if !slices.Equal(result.Stopped, []string{"billing"}) {
		t.Fatalf("unexpected stopped projects: %v", result.Stopped)
	}
	expectedCalls := []composeCall{
		{Action: "ps", Dir: otherDir, Compose: compose},
		{Action: "down", Dir: otherDir, Compose: compose},
		{Action: "up", Dir: targetDir, Compose: compose},
	}
	if !slices.Equal(composeSpy.calls, expectedCalls) {
		t.Fatalf("unexpected calls: %v", composeSpy.calls)
	}
}

func TestUpProjectAlongsideKeepsOtherRunningCompose(t *testing.T) {
	targetDir := trackedProjectDir(t)
	otherDir := trackedProjectDir(t)
	compose := "docker-compose.yaml"
	workspace := Workspace{
		Projects: map[string]projects.Project{
			"orders":  {Alias: "orders", Dir: targetDir, Compose: &compose},
			"billing": {Alias: "billing", Dir: otherDir, Compose: &compose},
		},
	}
	composeSpy := &composeSpy{running: map[string]bool{otherDir: true}}

	result, err := workspace.UpProject(context.Background(), targetDir, "", true, composeSpy)
	if err != nil {
		t.Fatalf("up project: %v", err)
	}

	if len(result.Stopped) != 0 {
		t.Fatalf("unexpected stopped projects: %v", result.Stopped)
	}
	expectedCalls := []composeCall{{Action: "up", Dir: targetDir, Compose: compose}}
	if !slices.Equal(composeSpy.calls, expectedCalls) {
		t.Fatalf("unexpected calls: %v", composeSpy.calls)
	}
}

func TestUpProjectFailsWhenComposeMissing(t *testing.T) {
	dir := t.TempDir()
	workspace := Workspace{
		Projects: map[string]projects.Project{
			"orders": {Alias: "orders", Dir: dir},
		},
	}

	composeSpy := &composeSpy{}
	_, err := workspace.UpProject(context.Background(), dir, "", false, composeSpy)
	if err == nil {
		t.Fatal("expected compose error")
	}
	if !strings.Contains(err.Error(), "no docker compose configured for project orders") {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(composeSpy.calls) != 0 {
		t.Fatalf("unexpected calls: %v", composeSpy.calls)
	}
}

func trackedProjectDir(t *testing.T) string {
	t.Helper()

	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "docker-compose.yaml"), []byte("services: {}"), 0o644); err != nil {
		t.Fatalf("write compose: %v", err)
	}
	return dir
}
