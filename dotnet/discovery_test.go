package dotnet

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/andrew-malikov/workspace/projects"
)

func TestDiscoverTestsMatchesConfiguredKinds(t *testing.T) {
	dir := t.TempDir()
	writeTestProject(t, dir, "tests/Orders.UnitTests/Orders.UnitTests.csproj")
	writeTestProject(t, dir, "tests/Orders.IntegrationTests/Orders.IntegrationTests.csproj")
	writeTestProject(t, dir, "src/Orders/Orders.csproj")

	got, err := DiscoverTests(dir, projects.TestDiscoveryConfig{
		Unit: projects.TestDiscoveryTarget{
			ProjectPatterns: []string{"(^|/)tests/.+UnitTests\\.csproj$"},
		},
		Integration: projects.TestDiscoveryTarget{
			ProjectPatterns: []string{"(^|/)tests/.+IntegrationTests\\.csproj$"},
		},
		Component: projects.TestDiscoveryTarget{
			ProjectPatterns: []string{"(^|/)tests/.+ComponentTests\\.csproj$"},
		},
	})
	if err != nil {
		t.Fatalf("discover tests: %v", err)
	}

	want := projects.TestConfig{
		Unit:        projects.TestTarget{Project: "tests/Orders.UnitTests/Orders.UnitTests.csproj"},
		Integration: projects.TestTarget{Project: "tests/Orders.IntegrationTests/Orders.IntegrationTests.csproj"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected matches: got %+v want %+v", got, want)
	}
}

func TestDiscoverTestsCopiesFilterToResolvedTarget(t *testing.T) {
	dir := t.TempDir()
	writeTestProject(t, dir, "tests/Orders.UnitTests/Orders.UnitTests.csproj")

	got, err := DiscoverTests(dir, projects.TestDiscoveryConfig{
		Unit: projects.TestDiscoveryTarget{
			ProjectPatterns: []string{"(^|/)tests/.+UnitTests\\.csproj$"},
			Filter:          "Category=Unit",
		},
	})
	if err != nil {
		t.Fatalf("discover tests: %v", err)
	}

	if got.Unit.Filter != "Category=Unit" {
		t.Fatalf("unexpected filter: %s", got.Unit.Filter)
	}
}

func TestDiscoverTestsFailsOnMultipleMatches(t *testing.T) {
	dir := t.TempDir()
	writeTestProject(t, dir, "tests/Orders.ComponentTests/Orders.ComponentTests.csproj")
	writeTestProject(t, dir, "tests/Orders.ComponentTests/Orders.Api.ComponentTests.csproj")

	_, err := DiscoverTests(dir, projects.TestDiscoveryConfig{
		Component: projects.TestDiscoveryTarget{
			ProjectPatterns: []string{"(^|/)tests/.+ComponentTests\\.csproj$"},
		},
	})
	if err == nil {
		t.Fatal("expected multiple matches error")
	}

	if !strings.Contains(err.Error(), "multiple component test projects") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func writeTestProject(t *testing.T, root string, rel string) {
	t.Helper()

	path := filepath.Join(root, rel)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir project dir: %v", err)
	}

	if err := os.WriteFile(path, []byte("<Project />"), 0o644); err != nil {
		t.Fatalf("write project: %v", err)
	}
}
