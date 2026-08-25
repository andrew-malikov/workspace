package projects

import (
	"os"
	"path/filepath"
	"testing"
)

func TestIsComposeConfiguredDoesNotTouchDisk(t *testing.T) {
	compose := "docker-compose.yaml"
	project := Project{Alias: "orders", Compose: &compose}

	if !project.IsComposeConfigured() {
		t.Fatal("expected configured compose filename")
	}
	if project.DoesComposeExist() {
		t.Fatal("configured filename must not imply a present file")
	}
}

func TestIsComposeConfiguredRejectsEmptyFilename(t *testing.T) {
	empty := "  "
	if (Project{Compose: &empty}).IsComposeConfigured() {
		t.Fatal("empty filename is not configured")
	}
	if (Project{}).IsComposeConfigured() {
		t.Fatal("missing pointer is not configured")
	}
}

func TestDoesComposeExistRequiresFile(t *testing.T) {
	dir := t.TempDir()
	compose := "docker-compose.yaml"
	project := Project{Alias: "orders", Dir: dir, Compose: &compose}
	if project.DoesComposeExist() {
		t.Fatal("missing file is not present")
	}
	if err := os.WriteFile(filepath.Join(dir, compose), []byte("services: {}\n"), 0o644); err != nil {
		t.Fatalf("write compose: %v", err)
	}
	if !project.DoesComposeExist() {
		t.Fatal("expected present compose file")
	}
}
