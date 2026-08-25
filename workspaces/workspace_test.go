package workspaces

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/andrew-malikov/workspace/projects"
)

func TestLoadReadsProvidedPathWithoutHome(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.toml")
	if err := os.WriteFile(path, []byte("[projects]\n"), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	workspace, err := Load(path)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if workspace.Git.Clear.Ownership.LookbackCommits != 3 {
		t.Fatalf("expected default lookback 3, got %d", workspace.Git.Clear.Ownership.LookbackCommits)
	}
}

func TestSaveWritesProvidedPathWithoutHome(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", "config.toml")
	workspace := Workspace{Projects: map[string]projects.Project{
		"orders": {Alias: "orders", Dir: "/tmp/orders"},
	}}

	if err := Save(path, workspace); err != nil {
		t.Fatalf("save workspace: %v", err)
	}
	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("reload workspace: %v", err)
	}
	if loaded.Projects["orders"].Dir != "/tmp/orders" {
		t.Fatalf("unexpected saved project: %+v", loaded.Projects["orders"])
	}
}

func TestLoadRejectsInvalidOwnershipRegex(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.toml")
	content := []byte(strings.Join([]string{
		"[git.clear.ownership.include]",
		"message_patterns = ['(']",
	}, "\n"))
	if err := os.WriteFile(path, content, 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	_, err := Load(path)
	if err == nil {
		t.Fatal("expected invalid regex error")
	}
	if !strings.Contains(err.Error(), "git.clear.ownership.include.message_patterns") {
		t.Fatalf("expected field name in error, got %v", err)
	}
}

func TestLoadRejectsInvalidTestRegex(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.toml")
	content := []byte(strings.Join([]string{
		"[test.unit]",
		"project_patterns = ['(']",
	}, "\n"))
	if err := os.WriteFile(path, content, 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	_, err := Load(path)
	if err == nil {
		t.Fatal("expected invalid regex error")
	}
	if !strings.Contains(err.Error(), "test.unit.project_patterns") {
		t.Fatalf("expected field name in error, got %v", err)
	}
}

func TestConfigPathUsesHomeAndXDGRules(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_CONFIG_HOME", "")

	path, err := ConfigPath()
	if err != nil {
		t.Fatalf("config path: %v", err)
	}
	want := filepath.Join(home, DEFAULT_UNIX_CONFIG_DIR, "ws", "config.toml")
	if path != want {
		t.Fatalf("unexpected default path: got %q want %q", path, want)
	}

	t.Setenv("XDG_CONFIG_HOME", ".config")
	path, err = ConfigPath()
	if err != nil {
		t.Fatalf("config path with xdg: %v", err)
	}
	want = filepath.Join(home, ".config", "ws", "config.toml")
	if path != want {
		t.Fatalf("unexpected xdg path: got %q want %q", path, want)
	}
}


func TestResolveProjectByDirMatchesNestedDirectory(t *testing.T) {
	workspace := Workspace{
		Projects: map[string]projects.Project{
			"orders": {
				Alias: "orders",
				Dir:   filepath.Join(string(os.PathSeparator), "tmp", "orders"),
			},
		},
	}

	project := workspace.ResolveProjectByDir(filepath.Join(string(os.PathSeparator), "tmp", "orders", "src"))
	if project == nil {
		t.Fatal("expected project to resolve")
	}

	if project.Alias != "orders" {
		t.Fatalf("unexpected project alias: %s", project.Alias)
	}

	project = workspace.ResolveProjectByDir(filepath.Join(string(os.PathSeparator), "tmp", "orders-api"))
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

	workspace := Workspace{
		Projects: map[string]projects.Project{},
		Test: projects.TestDiscoveryConfig{
			Unit: projects.TestDiscoveryTarget{
				ProjectPatterns: []string{"(^|/)tests/.+UnitTests\\.csproj$"},
				Filter:          "Category=Unit",
			},
		},
	}

	result, err := workspace.AddProject(projects.Project{Alias: "orders", Dir: dir})
	if err != nil {
		t.Fatalf("add project: %v", err)
	}

	if !result.HasUnitTests {
		t.Fatal("expected unit tests to be detected")
	}

	stored := workspace.Projects["orders"]
	if stored.Test.Unit.Project != "tests/Orders.UnitTests/Orders.UnitTests.csproj" {
		t.Fatalf("unexpected project path: %s", stored.Test.Unit.Project)
	}

	if stored.Test.Unit.Filter != "Category=Unit" {
		t.Fatalf("unexpected filter: %s", stored.Test.Unit.Filter)
	}
}

func TestScaffoldProjectTestsCopiesDiscoveredTestOptions(t *testing.T) {
	dir := t.TempDir()
	projectPath := filepath.Join(dir, "tests", "Orders.ComponentTests", "Orders.ComponentTests.csproj")
	if err := os.MkdirAll(filepath.Dir(projectPath), 0o755); err != nil {
		t.Fatalf("mkdir project dir: %v", err)
	}
	if err := os.WriteFile(projectPath, []byte("<Project />"), 0o644); err != nil {
		t.Fatalf("write csproj: %v", err)
	}

	workspace := Workspace{
		Test: projects.TestDiscoveryConfig{
			Component: projects.TestDiscoveryTarget{
				ProjectPatterns: []string{"(^|/)tests/.+ComponentTests\\.csproj$"},
				Filter:          "Category=Component",
			},
		},
	}
	project := projects.Project{
		Alias: "orders",
		Dir:   dir,
		Test: projects.TestConfig{
			Component: projects.TestTarget{Project: "old.csproj", Filter: "Old"},
		},
	}

	if err := workspace.ScaffoldProjectTests(&project); err != nil {
		t.Fatalf("scaffold tests: %v", err)
	}

	if project.Test.Component.Project != "tests/Orders.ComponentTests/Orders.ComponentTests.csproj" {
		t.Fatalf("unexpected project path: %s", project.Test.Component.Project)
	}

	if project.Test.Component.Filter != "Category=Component" {
		t.Fatalf("unexpected filter: %s", project.Test.Component.Filter)
	}
}
