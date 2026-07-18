package main

import (
	"bytes"
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/andrew-malikov/workspace/console"
)

func TestRunUsesStdoutForSuccessfulResult(t *testing.T) {
	useIsolatedWorkspace(t)
	var stdout, stderr bytes.Buffer
	terminal := console.New(bytes.NewReader(nil), &stdout, &stderr, false, false, false)

	status := run(t.Context(), []string{"ws", "list"}, terminal)

	if status != 0 {
		t.Fatalf("unexpected status: %d; stderr: %q", status, stderr.String())
	}
	if !strings.Contains(stdout.String(), "No tracked projects found") {
		t.Fatalf("missing successful result: %q", stdout.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("success wrote stderr: %q", stderr.String())
	}
}

func TestRunUsesStderrAndNonzeroForDomainFailure(t *testing.T) {
	useIsolatedWorkspace(t)
	var stdout, stderr bytes.Buffer
	terminal := console.New(bytes.NewReader(nil), &stdout, &stderr, false, false, false)

	status := run(t.Context(), []string{"ws", "test"}, terminal)

	if status == 0 {
		t.Fatal("domain failure returned success")
	}
	if !strings.Contains(stderr.String(), "isn't tracked yet") {
		t.Fatalf("missing domain diagnostic: %q", stderr.String())
	}
	if stdout.Len() != 0 {
		t.Fatalf("domain failure wrote stdout: %q", stdout.String())
	}
}

func TestRunRejectsNoninteractiveClearWithoutControlSequences(t *testing.T) {
	useIsolatedWorkspace(t)
	var stdout, stderr bytes.Buffer
	terminal := console.New(bytes.NewReader(nil), &stdout, &stderr, false, false, false)

	status := run(t.Context(), []string{"ws", "clear"}, terminal)

	if status == 0 {
		t.Fatal("noninteractive clear returned success")
	}
	if !strings.Contains(stderr.String(), "requires interactive") {
		t.Fatalf("missing clear diagnostic: %q", stderr.String())
	}
	if strings.Contains(stdout.String(), "\x1b[") || strings.Contains(stderr.String(), "\x1b[") {
		t.Fatalf("clear emitted terminal control sequence: stdout=%q stderr=%q", stdout.String(), stderr.String())
	}
}

func TestRunUsesStderrAndNonzeroForCommandFailure(t *testing.T) {
	useIsolatedWorkspace(t)
	var stdout, stderr bytes.Buffer
	terminal := console.New(bytes.NewReader(nil), &stdout, &stderr, false, false, false)

	status := run(t.Context(), []string{"ws", "unknown"}, terminal)

	if status == 0 {
		t.Fatal("command failure returned success")
	}
	if stderr.Len() == 0 {
		t.Fatal("command failure wrote no diagnostic")
	}
	if stdout.Len() != 0 {
		t.Fatalf("command failure wrote stdout: %q", stdout.String())
	}
}

func TestRunPropagatesRenderFailure(t *testing.T) {
	useIsolatedWorkspace(t)
	stdout := &applicationRejectingWriter{err: errors.New("stdout rejected")}
	var stderr bytes.Buffer
	terminal := console.New(bytes.NewReader(nil), stdout, &stderr, false, false, false)

	status := run(t.Context(), []string{"ws", "list"}, terminal)

	if status == 0 {
		t.Fatal("render failure returned success")
	}
	if !strings.Contains(stderr.String(), "stdout rejected") {
		t.Fatalf("missing render diagnostic: %q", stderr.String())
	}
	if stdout.attempts != 1 {
		t.Fatalf("unexpected stdout attempts: %d", stdout.attempts)
	}
}

func useIsolatedWorkspace(t *testing.T) {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_CONFIG_HOME", ".config")
	oldDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("get working directory: %v", err)
	}
	if err := os.Chdir(home); err != nil {
		t.Fatalf("change working directory: %v", err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(oldDir); err != nil {
			t.Errorf("restore working directory: %v", err)
		}
	})
}

type applicationRejectingWriter struct {
	err      error
	attempts int
}

func (writer *applicationRejectingWriter) Write([]byte) (int, error) {
	writer.attempts++
	return 0, writer.err
}
