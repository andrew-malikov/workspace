package up

import (
	"context"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	cfg "github.com/andrew-malikov/workspace/config"
	"github.com/andrew-malikov/workspace/projects"
)

func TestResolveProjectDefaultsToCurrentDirectory(t *testing.T) {
	dir := filepath.Join(string(os.PathSeparator), "tmp", "orders")
	config := &cfg.Config{
		Projects: map[string]projects.Project{
			"orders": {Alias: "orders", Dir: dir},
		},
	}

	project, err := resolveProject(config, filepath.Join(dir, "src"), "")
	if err != nil {
		t.Fatalf("resolve project: %v", err)
	}

	if project.Alias != "orders" {
		t.Fatalf("unexpected project alias: %s", project.Alias)
	}
}

func TestResolveProjectMatchesAlias(t *testing.T) {
	config := &cfg.Config{
		Projects: map[string]projects.Project{
			"orders": {Alias: "orders", Dir: filepath.Join(string(os.PathSeparator), "tmp", "orders")},
		},
	}

	project, err := resolveProject(config, "/tmp/other", "orders")
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
	config := &cfg.Config{
		Projects: map[string]projects.Project{
			"orders":  {Alias: "orders", Dir: targetDir, Compose: &compose},
			"billing": {Alias: "billing", Dir: otherDir, Compose: &compose},
		},
	}
	calls := []string{}

	result, err := upProject(context.Background(), config, targetDir, "", false, func(ctx context.Context, project projects.Project) (bool, error) {
		return project.Alias == "billing", nil
	}, func(ctx context.Context, project projects.Project) error {
		calls = append(calls, "up:"+project.Alias)
		return nil
	}, func(ctx context.Context, project projects.Project) error {
		calls = append(calls, "down:"+project.Alias)
		return nil
	})
	if err != nil {
		t.Fatalf("up project: %v", err)
	}

	if result.Alias != "orders" {
		t.Fatalf("unexpected target: %s", result.Alias)
	}
	if !slices.Equal(result.Stopped, []string{"billing"}) {
		t.Fatalf("unexpected stopped projects: %v", result.Stopped)
	}
	if !slices.Equal(calls, []string{"down:billing", "up:orders"}) {
		t.Fatalf("unexpected calls: %v", calls)
	}
}

func TestUpProjectAlongsideKeepsOtherRunningCompose(t *testing.T) {
	targetDir := trackedProjectDir(t)
	otherDir := trackedProjectDir(t)
	compose := "docker-compose.yaml"
	config := &cfg.Config{
		Projects: map[string]projects.Project{
			"orders":  {Alias: "orders", Dir: targetDir, Compose: &compose},
			"billing": {Alias: "billing", Dir: otherDir, Compose: &compose},
		},
	}
	calls := []string{}

	result, err := upProject(context.Background(), config, targetDir, "", true, func(ctx context.Context, project projects.Project) (bool, error) {
		t.Fatal("should not check running projects with --alongside")
		return false, nil
	}, func(ctx context.Context, project projects.Project) error {
		calls = append(calls, "up:"+project.Alias)
		return nil
	}, func(ctx context.Context, project projects.Project) error {
		calls = append(calls, "down:"+project.Alias)
		return nil
	})
	if err != nil {
		t.Fatalf("up project: %v", err)
	}

	if len(result.Stopped) != 0 {
		t.Fatalf("unexpected stopped projects: %v", result.Stopped)
	}
	if !slices.Equal(calls, []string{"up:orders"}) {
		t.Fatalf("unexpected calls: %v", calls)
	}
}

func TestUpProjectFailsWhenComposeMissing(t *testing.T) {
	dir := t.TempDir()
	config := &cfg.Config{
		Projects: map[string]projects.Project{
			"orders": {Alias: "orders", Dir: dir},
		},
	}

	_, err := upProject(context.Background(), config, dir, "", false, nil, nil, nil)
	if err == nil {
		t.Fatal("expected compose error")
	}
	if !strings.Contains(err.Error(), "no docker compose configured for project orders") {
		t.Fatalf("unexpected error: %v", err)
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
