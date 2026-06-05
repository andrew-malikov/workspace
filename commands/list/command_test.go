package list

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/andrew-malikov/workspace/projects"
)

func TestEnabledSummary(t *testing.T) {
	dir := t.TempDir()
	compose := "docker-compose.yaml"
	migrations := "migrations"

	if err := os.WriteFile(filepath.Join(dir, compose), []byte("services: {}"), 0o644); err != nil {
		t.Fatalf("write compose: %v", err)
	}
	if err := os.Mkdir(filepath.Join(dir, migrations), 0o755); err != nil {
		t.Fatalf("mkdir migrations: %v", err)
	}

	project := projects.Project{
		Alias:      "orders",
		Dir:        dir,
		Compose:    &compose,
		Migrations: &migrations,
		Test: projects.TestConfig{
			Unit:      projects.TestTarget{Project: "tests/Orders.UnitTests.csproj"},
			Component: projects.TestTarget{Project: "tests/Orders.ComponentTests.csproj"},
		},
	}

	got := enabledSummary(project)
	want := "docker, migrations, unit tests, component tests"
	if got != want {
		t.Fatalf("unexpected summary: got %q want %q", got, want)
	}
}

func TestEnabledSummaryNone(t *testing.T) {
	got := enabledSummary(projects.Project{Alias: "orders", Dir: t.TempDir()})
	if got != "none" {
		t.Fatalf("unexpected summary: %q", got)
	}
}

func TestBuildListedProjectsSortsByName(t *testing.T) {
	listed := buildListedProjects(map[string]projects.Project{
		"web": {Alias: "web", Dir: "/repo/web"},
		"api": {Alias: "api", Dir: "/repo/api"},
	})

	got := []string{listed[0].Name, listed[1].Name}
	want := []string{"api", "web"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected order: got %v want %v", got, want)
	}
}
