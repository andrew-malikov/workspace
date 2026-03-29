package cfg

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/andrew-malikov/workspace/projects"
)

func TestLoadConfigNormalizesOwnershipLookback(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_CONFIG_HOME", ".config")

	configPath := filepath.Join(home, ".config", "ws", "config.toml")
	if err := os.MkdirAll(filepath.Dir(configPath), 0o755); err != nil {
		t.Fatalf("mkdir config dir: %v", err)
	}

	content := []byte("[projects]\n")
	if err := os.WriteFile(configPath, content, 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	config, err := LoadConfig()
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	if config.Git.Clear.Ownership.LookbackCommits != 3 {
		t.Fatalf("expected default lookback 3, got %d", config.Git.Clear.Ownership.LookbackCommits)
	}
}

func TestLoadConfigRejectsInvalidOwnershipRegex(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_CONFIG_HOME", ".config")

	configPath := filepath.Join(home, ".config", "ws", "config.toml")
	if err := os.MkdirAll(filepath.Dir(configPath), 0o755); err != nil {
		t.Fatalf("mkdir config dir: %v", err)
	}

	content := []byte(strings.Join([]string{
		"[git.clear.ownership.include]",
		"message_patterns = ['(']",
	}, "\n"))
	if err := os.WriteFile(configPath, content, 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	_, err := LoadConfig()
	if err == nil {
		t.Fatal("expected invalid regex error")
	}

	if !strings.Contains(err.Error(), "git.clear.ownership.include.message_patterns") {
		t.Fatalf("expected field name in error, got %v", err)
	}
}

func TestLoadConfigRejectsInvalidTestRegex(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_CONFIG_HOME", ".config")

	configPath := filepath.Join(home, ".config", "ws", "config.toml")
	if err := os.MkdirAll(filepath.Dir(configPath), 0o755); err != nil {
		t.Fatalf("mkdir config dir: %v", err)
	}

	content := []byte(strings.Join([]string{
		"[test.unit]",
		"project_patterns = ['(']",
	}, "\n"))
	if err := os.WriteFile(configPath, content, 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	_, err := LoadConfig()
	if err == nil {
		t.Fatal("expected invalid regex error")
	}

	if !strings.Contains(err.Error(), "test.unit.project_patterns") {
		t.Fatalf("expected field name in error, got %v", err)
	}
}

func TestResolveProjectByDirMatchesNestedDirectory(t *testing.T) {
	config := Config{
		Projects: map[string]projects.Project{
			"orders": {
				Alias: "orders",
				Dir:   filepath.Join(string(os.PathSeparator), "tmp", "orders"),
			},
		},
	}

	project := config.ResolveProjectByDir(filepath.Join(string(os.PathSeparator), "tmp", "orders", "src"))
	if project == nil {
		t.Fatal("expected project to resolve")
	}

	if project.Alias != "orders" {
		t.Fatalf("unexpected project alias: %s", project.Alias)
	}

	project = config.ResolveProjectByDir(filepath.Join(string(os.PathSeparator), "tmp", "orders-api"))
	if project != nil {
		t.Fatalf("expected no match for sibling directory, got %s", project.Alias)
	}
}

func TestAddProjectCopiesDiscoveredTests(t *testing.T) {
	dir := t.TempDir()
	projectPath := filepath.Join(dir, "tests", "Orders.UnitTests", "Orders.UnitTests.csproj")
	if err := os.MkdirAll(filepath.Dir(projectPath), 0o755); err != nil {
		t.Fatalf("mkdir project dir: %v", err)
	}
	if err := os.WriteFile(projectPath, []byte("<Project />"), 0o644); err != nil {
		t.Fatalf("write csproj: %v", err)
	}

	config := Config{
		Projects: map[string]projects.Project{},
		Test: projects.TestDiscoveryConfig{
			Unit: projects.TestDiscoveryTarget{
				ProjectPatterns: []string{"(^|/)tests/.+UnitTests\\.csproj$"},
				Filter:          "Category=Unit",
			},
		},
	}

	result, err := config.AddProject(projects.Project{Alias: "orders", Dir: dir})
	if err != nil {
		t.Fatalf("add project: %v", err)
	}

	if !result.HasUnitTests {
		t.Fatal("expected unit tests to be detected")
	}

	stored := config.Projects["orders"]
	if stored.Test.Unit.Project != "tests/Orders.UnitTests/Orders.UnitTests.csproj" {
		t.Fatalf("unexpected project path: %s", stored.Test.Unit.Project)
	}

	if stored.Test.Unit.Filter != "Category=Unit" {
		t.Fatalf("unexpected filter: %s", stored.Test.Unit.Filter)
	}
}
