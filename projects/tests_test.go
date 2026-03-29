package projects

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestDiscoverTestsMatchesConfiguredKinds(t *testing.T) {
	dir := t.TempDir()
	writeTestProject(t, dir, "tests/Orders.UnitTests/Orders.UnitTests.csproj")
	writeTestProject(t, dir, "tests/Orders.IntegrationTests/Orders.IntegrationTests.csproj")
	writeTestProject(t, dir, "src/Orders/Orders.csproj")

	project := Project{Alias: "orders", Dir: dir}
	matches, err := project.DiscoverTests(TestDiscoveryConfig{
		Unit: TestDiscoveryTarget{
			ProjectPatterns: []string{"(^|/)tests/.+UnitTests\\.csproj$"},
		},
		Integration: TestDiscoveryTarget{
			ProjectPatterns: []string{"(^|/)tests/.+IntegrationTests\\.csproj$"},
		},
		Component: TestDiscoveryTarget{
			ProjectPatterns: []string{"(^|/)tests/.+ComponentTests\\.csproj$"},
		},
	})
	if err != nil {
		t.Fatalf("discover tests: %v", err)
	}

	want := map[TestKind]string{
		UnitTestKind:        "tests/Orders.UnitTests/Orders.UnitTests.csproj",
		IntegrationTestKind: "tests/Orders.IntegrationTests/Orders.IntegrationTests.csproj",
	}
	if !reflect.DeepEqual(matches, want) {
		t.Fatalf("unexpected matches: got %v want %v", matches, want)
	}
}

func TestDiscoverTestsFailsOnMultipleMatches(t *testing.T) {
	dir := t.TempDir()
	writeTestProject(t, dir, "tests/Orders.ComponentTests/Orders.ComponentTests.csproj")
	writeTestProject(t, dir, "tests/Orders.ComponentTests/Orders.Api.ComponentTests.csproj")

	project := Project{Alias: "orders", Dir: dir}
	_, err := project.DiscoverTests(TestDiscoveryConfig{
		Component: TestDiscoveryTarget{
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
