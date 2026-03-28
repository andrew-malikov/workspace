package cfg

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
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
