package main

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/andrew-malikov/workspace/console"
	"github.com/andrew-malikov/workspace/projects"
	"github.com/andrew-malikov/workspace/workspaces"
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

func TestRunDownStreamsComposeAndSupportsBlankFlags(t *testing.T) {
	tests := []struct {
		name        string
		flag        string
		wantArgs    string
		wantVolumes bool
	}{
		{name: "default", wantArgs: "compose -f docker-compose.yaml down"},
		{name: "long blank", flag: "--blank", wantArgs: "compose -f docker-compose.yaml down -v", wantVolumes: true},
		{name: "short blank", flag: "-b", wantArgs: "compose -f docker-compose.yaml down -v", wantVolumes: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			useIsolatedWorkspace(t)
			setupComposeWorkspace(t, "orders")
			callsPath := installFakeDocker(t)
			var stdout, stderr bytes.Buffer
			terminal := console.New(bytes.NewReader(nil), &stdout, &stderr, false, false, false)
			args := []string{"ws", "down", "orders"}
			if tt.flag != "" {
				args = append(args, tt.flag)
			}

			status := run(t.Context(), args, terminal)
			if status != 0 {
				t.Fatalf("unexpected status: %d; stderr: %q", status, stderr.String())
			}
			if !strings.Contains(stdout.String(), "docker-stdout:orders") ||
				!strings.Contains(stdout.String(), "docker compose is down.") {
				t.Fatalf("missing streamed or result output: %q", stdout.String())
			}
			hasVolumes := strings.Contains(stdout.String(), "volumes were cleaned up")
			if hasVolumes != tt.wantVolumes {
				t.Fatalf("unexpected volume result: %q", stdout.String())
			}
			for _, expected := range []string{
				`project "orders"`,
				"docker " + tt.wantArgs,
				"docker-stderr:orders",
			} {
				if !strings.Contains(stderr.String(), expected) {
					t.Fatalf("stderr missing %q: %q", expected, stderr.String())
				}
			}
			calls, err := os.ReadFile(callsPath)
			if err != nil {
				t.Fatalf("read docker calls: %v", err)
			}
			if strings.TrimSpace(string(calls)) != "orders|"+tt.wantArgs {
				t.Fatalf("unexpected docker calls: %q", calls)
			}
		})
	}
}

func TestRunDownFailureSuppressesSuccessResult(t *testing.T) {
	useIsolatedWorkspace(t)
	setupComposeWorkspace(t, "orders")
	installFakeDocker(t)
	t.Setenv("WS_DOCKER_FAIL_ALIAS", "orders")
	var stdout, stderr bytes.Buffer
	terminal := console.New(bytes.NewReader(nil), &stdout, &stderr, false, false, false)

	status := run(t.Context(), []string{"ws", "down", "orders"}, terminal)
	if status == 0 {
		t.Fatal("compose failure returned success")
	}
	if !strings.Contains(stdout.String(), "docker-stdout:orders") {
		t.Fatalf("missing partial child stdout: %q", stdout.String())
	}
	if strings.Contains(stdout.String(), "docker compose is down") {
		t.Fatalf("failure rendered success result: %q", stdout.String())
	}
	if !strings.Contains(stderr.String(), "docker-stderr:orders") ||
		!strings.Contains(stderr.String(), "exit status 7") {
		t.Fatalf("missing process failure output: %q", stderr.String())
	}
}

func TestRunUpLabelsMultipleComposeActionsInStableOrder(t *testing.T) {
	useIsolatedWorkspace(t)
	dirs := setupComposeWorkspace(t, "worker", "orders", "api")
	callsPath := installFakeDocker(t)
	t.Setenv("WS_RUNNING_ALIASES", "api,worker")
	oldDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("get working directory: %v", err)
	}
	if err := os.Chdir(dirs["orders"]); err != nil {
		t.Fatalf("change to target: %v", err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(oldDir); err != nil {
			t.Errorf("restore target directory: %v", err)
		}
	})
	var stdout, stderr bytes.Buffer
	terminal := console.New(bytes.NewReader(nil), &stdout, &stderr, false, false, false)

	status := run(t.Context(), []string{"ws", "up"}, terminal)
	if status != 0 {
		t.Fatalf("unexpected status: %d; stderr: %q", status, stderr.String())
	}
	if strings.Contains(stdout.String(), `{"Name"`) {
		t.Fatalf("running probe leaked to stdout: %q", stdout.String())
	}
	if strings.Count(stdout.String(), "docker-stdout:") != 3 ||
		!strings.Contains(stdout.String(), "docker compose is up.") {
		t.Fatalf("unexpected stdout: %q", stdout.String())
	}
	api := strings.Index(stderr.String(), `project "api"`)
	worker := strings.Index(stderr.String(), `project "worker"`)
	orders := strings.Index(stderr.String(), `project "orders"`)
	if api < 0 || worker <= api || orders <= worker {
		t.Fatalf("action headings not in alias/target order: %q", stderr.String())
	}
	calls, err := os.ReadFile(callsPath)
	if err != nil {
		t.Fatalf("read docker calls: %v", err)
	}
	expectedCalls := "api|compose -f docker-compose.yaml down\n" +
		"worker|compose -f docker-compose.yaml down\n" +
		"orders|compose -f docker-compose.yaml up -d\n"
	if string(calls) != expectedCalls {
		t.Fatalf("unexpected docker calls: %q", calls)
	}
}

func setupComposeWorkspace(t *testing.T, aliases ...string) map[string]string {
	t.Helper()
	root, err := os.Getwd()
	if err != nil {
		t.Fatalf("get workspace root: %v", err)
	}
	dirs := make(map[string]string, len(aliases))
	tracked := make(map[string]projects.Project, len(aliases))
	compose := "docker-compose.yaml"
	for _, alias := range aliases {
		dir := filepath.Join(root, alias)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatalf("create project directory: %v", err)
		}
		if err := os.WriteFile(filepath.Join(dir, compose), []byte("services: {}\n"), 0o644); err != nil {
			t.Fatalf("write compose file: %v", err)
		}
		dirs[alias] = dir
		tracked[alias] = projects.Project{Alias: alias, Dir: dir, Compose: &compose}
	}
	if err := workspaces.SaveWorkspace(workspaces.Workspace{Projects: tracked}); err != nil {
		t.Fatalf("save workspace: %v", err)
	}
	return dirs
}

func installFakeDocker(t *testing.T) string {
	t.Helper()
	bin := t.TempDir()
	executable := filepath.Join(bin, "docker")
	script := `#!/bin/sh
project=${PWD##*/}
case " $* " in
  *" ps "*)
    case ",${WS_RUNNING_ALIASES}," in
      *",${project},"*) printf '{"Name":"%s"}' "$project" ;;
      *) printf '[]' ;;
    esac
    exit 0
    ;;
esac
printf 'docker-stdout:%s\n' "$project"
printf 'docker-stderr:%s\n' "$project" >&2
printf '%s|%s\n' "$project" "$*" >> "$WS_DOCKER_CALLS"
if [ "$WS_DOCKER_FAIL_ALIAS" = "$project" ]; then
  exit 7
fi
`
	if err := os.WriteFile(executable, []byte(script), 0o755); err != nil {
		t.Fatalf("write fake docker: %v", err)
	}
	callsPath := filepath.Join(t.TempDir(), "calls")
	t.Setenv("WS_DOCKER_CALLS", callsPath)
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	return callsPath
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
