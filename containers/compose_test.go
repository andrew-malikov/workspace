package containers

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestHasRunning(t *testing.T) {
	tests := []struct {
		name string
		out  []byte
		want bool
	}{
		{
			name: "empty output",
			want: false,
		},
		{
			name: "empty json array",
			out:  []byte("[]\n"),
			want: false,
		},
		{
			name: "json object",
			out:  []byte(`{"Name":"api"}`),
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := HasRunning(tt.out)
			if got != tt.want {
				t.Fatalf("unexpected result: got %t want %t", got, tt.want)
			}
		})
	}
}

func TestDockerComposeQueryIsCapturedAndSilent(t *testing.T) {
	var stdout, stderr bytes.Buffer
	docker := NewDockerCompose(strings.NewReader("unused"), &stdout, &stderr)
	docker.command = composeHelperFactory(false)

	running, err := docker.HasRunning(t.Context(), Target{
		Alias: "api",
		Dir:   t.TempDir(),
		File:  "compose.yml",
	})
	if err != nil {
		t.Fatalf("has running: %v", err)
	}
	if !running {
		t.Fatal("expected running compose")
	}
	if stdout.Len() != 0 || stderr.Len() != 0 {
		t.Fatalf("query leaked output: stdout=%q stderr=%q", stdout.String(), stderr.String())
	}
}

func TestDockerComposeActionsUseConfiguredStreamsAndArguments(t *testing.T) {
	tests := []struct {
		name       string
		operation  string
		preview    string
		helperArgs string
		run        func(DockerCompose, context.Context, Target) error
	}{
		{
			name:       "up",
			operation:  "start",
			preview:    "docker compose -f compose.yml up -d",
			helperArgs: "compose|-f|compose.yml|up|-d",
			run:        func(d DockerCompose, ctx context.Context, target Target) error { return d.Up(ctx, target) },
		},
		{
			name:       "down",
			operation:  "stop",
			preview:    "docker compose -f compose.yml down",
			helperArgs: "compose|-f|compose.yml|down",
			run:        func(d DockerCompose, ctx context.Context, target Target) error { return d.Down(ctx, target) },
		},
		{
			name:       "cleanup",
			operation:  "clean volumes",
			preview:    "docker compose -f compose.yml down -v",
			helperArgs: "compose|-f|compose.yml|down|-v",
			run:        func(d DockerCompose, ctx context.Context, target Target) error { return d.Cleanup(ctx, target) },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			docker := NewDockerCompose(strings.NewReader("request"), &stdout, &stderr)
			docker.command = composeHelperFactory(false)
			target := Target{Alias: "api", Dir: t.TempDir(), File: "compose.yml"}

			if err := tt.run(docker, t.Context(), target); err != nil {
				t.Fatalf("%s: %v", tt.name, err)
			}

			if stdout.String() != "stdout:request" {
				t.Fatalf("unexpected stdout: %q", stdout.String())
			}
			canonicalDir, err := filepath.EvalSymlinks(target.Dir)
			if err != nil {
				t.Fatalf("resolve target directory: %v", err)
			}
			for _, expected := range []string{
				fmt.Sprintf("project %q in %q: %s", target.Alias, target.Dir, tt.operation),
				tt.preview,
				"stderr-dir:" + canonicalDir,
				"stderr-args:" + tt.helperArgs,
			} {
				if !strings.Contains(stderr.String(), expected) {
					t.Fatalf("stderr missing %q: %q", expected, stderr.String())
				}
			}
		})
	}
}

func TestDockerComposeActionMetadataEscapesDynamicValues(t *testing.T) {
	var stdout, stderr bytes.Buffer
	docker := NewDockerCompose(strings.NewReader(""), &stdout, &stderr)
	docker.command = composeHelperFactory(false)
	target := Target{
		Alias: "api\n\x1b",
		Dir:   t.TempDir(),
		File:  "compose file\n\"\x1b.yml",
	}

	if err := docker.Down(t.Context(), target); err != nil {
		t.Fatalf("down: %v", err)
	}

	metadata := strings.SplitN(stderr.String(), "stderr-dir:", 2)[0]
	if strings.ContainsRune(metadata, '\x1b') {
		t.Fatalf("metadata contains raw escape: %q", metadata)
	}
	for _, expected := range []string{
		strconv.QuoteToASCII(target.Alias),
		strconv.QuoteToASCII(target.Dir),
		strconv.QuoteToASCII(target.File),
	} {
		if !strings.Contains(metadata, expected) {
			t.Fatalf("metadata missing escaped value %q: %q", expected, metadata)
		}
	}
}

func TestDockerComposeActionStopsWhenMetadataWriteFails(t *testing.T) {
	for _, rejectAt := range []int{1, 2} {
		t.Run(fmt.Sprintf("write %d", rejectAt), func(t *testing.T) {
			writer := &rejectingComposeWriter{rejectAt: rejectAt}
			docker := NewDockerCompose(strings.NewReader(""), io.Discard, writer)
			started := 0
			docker.command = func(ctx context.Context, name string, args ...string) *exec.Cmd {
				started++
				return composeHelperFactory(false)(ctx, name, args...)
			}

			err := docker.Down(t.Context(), Target{Alias: "api", Dir: t.TempDir(), File: "compose.yml"})
			if !errors.Is(err, errComposeWriteRejected) {
				t.Fatalf("unexpected error: %v", err)
			}
			if started != 0 {
				t.Fatalf("process factory called %d times", started)
			}
		})
	}
}

func TestDockerComposeActionPreservesProcessFailureAndOutput(t *testing.T) {
	var stdout, stderr bytes.Buffer
	docker := NewDockerCompose(strings.NewReader("request"), &stdout, &stderr)
	docker.command = composeHelperFactory(true)
	target := Target{Alias: "api", Dir: t.TempDir(), File: "compose.yml"}

	err := docker.Up(t.Context(), target)
	if err == nil {
		t.Fatal("expected process failure")
	}
	if stdout.String() != "stdout:request" {
		t.Fatalf("unexpected stdout: %q", stdout.String())
	}
	if !strings.Contains(stderr.String(), "stderr-dir:") {
		t.Fatalf("missing streamed stderr: %q", stderr.String())
	}
}

var errComposeWriteRejected = errors.New("write rejected")

type rejectingComposeWriter struct {
	writes   int
	rejectAt int
}

func (writer *rejectingComposeWriter) Write(data []byte) (int, error) {
	writer.writes++
	if writer.writes == writer.rejectAt {
		return 0, errComposeWriteRejected
	}
	return len(data), nil
}

func composeHelperFactory(fail bool) commandFactory {
	return func(ctx context.Context, name string, args ...string) *exec.Cmd {
		helperArgs := append([]string{"-test.run=TestDockerComposeHelperProcess", "--", name}, args...)
		command := exec.CommandContext(ctx, os.Args[0], helperArgs...)
		command.Env = append(os.Environ(),
			"WS_COMPOSE_HELPER=1",
			fmt.Sprintf("WS_COMPOSE_HELPER_FAIL=%t", fail),
		)
		return command
	}
}

func TestDockerComposeHelperProcess(t *testing.T) {
	if os.Getenv("WS_COMPOSE_HELPER") != "1" {
		return
	}

	separator := 0
	for i, arg := range os.Args {
		if arg == "--" {
			separator = i + 1
			break
		}
	}
	args := os.Args[separator+1:]
	if len(args) > 0 && args[len(args)-1] == "json" {
		_, _ = fmt.Fprint(os.Stdout, `{"Name":"api"}`)
		os.Exit(0)
	}

	input, _ := io.ReadAll(os.Stdin)
	dir, _ := os.Getwd()
	_, _ = fmt.Fprintf(os.Stdout, "stdout:%s", input)
	_, _ = fmt.Fprintf(os.Stderr, "stderr-dir:%s\nstderr-args:%s", dir, strings.Join(args, "|"))
	if os.Getenv("WS_COMPOSE_HELPER_FAIL") == "true" {
		os.Exit(1)
	}
	os.Exit(0)
}
