package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/andrew-malikov/workspace/console"
	"github.com/andrew-malikov/workspace/containers"
	"github.com/andrew-malikov/workspace/projects"
	"github.com/andrew-malikov/workspace/workspaces"
)

func TestRunUsesStdoutForSuccessfulResult(t *testing.T) {
	var stdout, stderr bytes.Buffer
	terminal := console.New(bytes.NewReader(nil), &stdout, &stderr, false, false, false, false)

	status := run(t.Context(), []string{"ws", "list"}, terminal, memorySession(t, workspaces.Workspace{}, t.TempDir(), &routingCompose{}))

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
	var stdout, stderr bytes.Buffer
	terminal := console.New(bytes.NewReader(nil), &stdout, &stderr, false, false, false, false)

	status := run(t.Context(), []string{"ws", "test"}, terminal, memorySession(t, workspaces.Workspace{}, t.TempDir(), &routingCompose{}))

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
	var stdout, stderr bytes.Buffer
	terminal := console.New(bytes.NewReader(nil), &stdout, &stderr, false, false, false, false)

	status := run(t.Context(), []string{"ws", "clear"}, terminal, memorySession(t, workspaces.Workspace{}, t.TempDir(), &routingCompose{}))

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
	var stdout, stderr bytes.Buffer
	terminal := console.New(bytes.NewReader(nil), &stdout, &stderr, false, false, false, false)

	status := run(t.Context(), []string{"ws", "unknown"}, terminal, memorySession(t, workspaces.Workspace{}, t.TempDir(), &routingCompose{}))

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
	stdout := &applicationRejectingWriter{err: errors.New("stdout rejected")}
	var stderr bytes.Buffer
	terminal := console.New(bytes.NewReader(nil), stdout, &stderr, false, false, false, false)

	status := run(t.Context(), []string{"ws", "list"}, terminal, memorySession(t, workspaces.Workspace{}, t.TempDir(), &routingCompose{}))

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

func TestRunDownStreamsComposeAndSupportsBlankFlags(t *testing.T) {
	tests := []struct {
		name        string
		flag        string
		wantAction  string
		wantVolumes bool
	}{
		{name: "default", wantAction: "down"},
		{name: "long blank", flag: "--blank", wantAction: "cleanup", wantVolumes: true},
		{name: "short blank", flag: "-b", wantAction: "cleanup", wantVolumes: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			workspace, cwd := composeWorkspace(t, "orders")
			compose := &routingCompose{}
			var stdout, stderr bytes.Buffer
			terminal := console.New(bytes.NewReader(nil), &stdout, &stderr, false, false, false, false)
			args := []string{"ws", "down", "orders"}
			if tt.flag != "" {
				args = append(args, tt.flag)
			}

			status := run(t.Context(), args, terminal, memorySession(t, workspace, cwd, compose))
			if status != 0 {
				t.Fatalf("unexpected status: %d; stderr: %q", status, stderr.String())
			}
			if !strings.Contains(stdout.String(), "compose-stdout:orders") ||
				!strings.Contains(stdout.String(), "docker compose is down.") {
				t.Fatalf("missing streamed or result output: %q", stdout.String())
			}
			hasVolumes := strings.Contains(stdout.String(), "volumes were cleaned up")
			if hasVolumes != tt.wantVolumes {
				t.Fatalf("unexpected volume result: %q", stdout.String())
			}
			if !strings.Contains(stderr.String(), "compose-stderr:orders") {
				t.Fatalf("stderr missing streamed compose output: %q", stderr.String())
			}
			if got := strings.Join(compose.calls, "\n"); got != "orders|"+tt.wantAction {
				t.Fatalf("unexpected compose calls: %q", got)
			}
		})
	}
}

func TestRunDownFailureSuppressesSuccessResult(t *testing.T) {
	workspace, cwd := composeWorkspace(t, "orders")
	compose := &routingCompose{failAlias: "orders"}
	var stdout, stderr bytes.Buffer
	terminal := console.New(bytes.NewReader(nil), &stdout, &stderr, false, false, false, false)

	status := run(t.Context(), []string{"ws", "down", "orders"}, terminal, memorySession(t, workspace, cwd, compose))
	if status == 0 {
		t.Fatal("compose failure returned success")
	}
	if !strings.Contains(stdout.String(), "compose-stdout:orders") {
		t.Fatalf("missing partial child stdout: %q", stdout.String())
	}
	if strings.Contains(stdout.String(), "docker compose is down") {
		t.Fatalf("failure rendered success result: %q", stdout.String())
	}
	if !strings.Contains(stderr.String(), "compose-stderr:orders") ||
		!strings.Contains(stderr.String(), "compose failed") {
		t.Fatalf("missing process failure output: %q", stderr.String())
	}
}

func TestRunUpUsesInjectedWorkspaceAndCompose(t *testing.T) {
	workspace, dirs := composeWorkspaceDirs(t, "worker", "orders", "api")
	compose := &routingCompose{running: map[string]bool{"api": true, "worker": true}}
	var stdout, stderr bytes.Buffer
	terminal := console.New(bytes.NewReader(nil), &stdout, &stderr, false, false, false, false)

	status := run(t.Context(), []string{"ws", "up"}, terminal, memorySession(t, workspace, dirs["orders"], compose))
	if status != 0 {
		t.Fatalf("unexpected status: %d; stderr: %q", status, stderr.String())
	}
	if strings.Count(stdout.String(), "compose-stdout:") != 3 ||
		!strings.Contains(stdout.String(), "docker compose is up.") {
		t.Fatalf("unexpected stdout: %q", stdout.String())
	}
	expectedCalls := "api|down\nworker|down\norders|up"
	if got := strings.Join(compose.calls, "\n"); got != expectedCalls {
		t.Fatalf("unexpected compose calls: %q", got)
	}
}

func memorySession(t *testing.T, workspace workspaces.Workspace, cwd string, compose *routingCompose) workspaces.Session {
	t.Helper()
	current := workspace
	if current.Projects == nil {
		current.Projects = map[string]projects.Project{}
	}
	return workspaces.Session{
		Load: func() (*workspaces.Workspace, error) {
			copy := current
			return &copy, nil
		},
		Save: func(next workspaces.Workspace) error {
			current = next
			return nil
		},
		Cwd: func() (string, error) {
			return cwd, nil
		},
		Compose: func(terminal console.Console) containers.Compose {
			compose.output = terminal.Output
			compose.errorOutput = terminal.Error
			return compose
		},
	}
}

func composeWorkspace(t *testing.T, aliases ...string) (workspaces.Workspace, string) {
	t.Helper()
	workspace, dirs := composeWorkspaceDirs(t, aliases...)
	return workspace, dirs[aliases[0]]
}

func composeWorkspaceDirs(t *testing.T, aliases ...string) (workspaces.Workspace, map[string]string) {
	t.Helper()
	dirs := make(map[string]string, len(aliases))
	tracked := make(map[string]projects.Project, len(aliases))
	compose := "docker-compose.yaml"
	for _, alias := range aliases {
		dir := t.TempDir()
		if err := os.WriteFile(filepath.Join(dir, compose), []byte("services: {}\n"), 0o644); err != nil {
			t.Fatalf("write compose file: %v", err)
		}
		dirs[alias] = dir
		tracked[alias] = projects.Project{Alias: alias, Dir: dir, Compose: &compose}
	}
	return workspaces.Workspace{Projects: tracked}, dirs
}

type routingCompose struct {
	output      io.Writer
	errorOutput io.Writer
	running     map[string]bool
	calls       []string
	failAlias   string
}

func (compose *routingCompose) HasRunning(_ context.Context, target containers.Target) (bool, error) {
	return compose.running[target.Alias], nil
}

func (compose *routingCompose) Up(_ context.Context, target containers.Target) error {
	return compose.record("up", target)
}

func (compose *routingCompose) Down(_ context.Context, target containers.Target) error {
	return compose.record("down", target)
}

func (compose *routingCompose) Cleanup(_ context.Context, target containers.Target) error {
	return compose.record("cleanup", target)
}

func (compose *routingCompose) record(action string, target containers.Target) error {
	compose.calls = append(compose.calls, target.Alias+"|"+action)
	if _, err := fmt.Fprintf(compose.output, "compose-stdout:%s\n", target.Alias); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(compose.errorOutput, "compose-stderr:%s\n", target.Alias); err != nil {
		return err
	}
	if compose.failAlias == target.Alias {
		return errors.New("compose failed")
	}
	return nil
}

type applicationRejectingWriter struct {
	err      error
	attempts int
}

func (writer *applicationRejectingWriter) Write([]byte) (int, error) {
	writer.attempts++
	return 0, writer.err
}
