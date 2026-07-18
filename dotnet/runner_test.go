package dotnet

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/andrew-malikov/workspace/projects"
	"github.com/stretchr/testify/assert"
)

func TestTestRunnerRunsDotnetFilteredTest(t *testing.T) {
	// arrange
	dir := t.TempDir()
	commandRunner := SpyCommandRunner{Executions: make([]CommandExecution, 0)}
	var diagnostics bytes.Buffer
	testRunner := NewTestRunner(&commandRunner, &diagnostics)

	// act
	err := testRunner.Run(t.Context(), dir, projects.UnitTestKind, projects.TestTarget{
		Project: "tests/Orders.UnitTests/Orders.UnitTests.csproj",
		Filter:  "FullyQualifiedName~UnitTests",
	})

	// assert
	assert.Nil(t, err)
	assert.Len(t, commandRunner.Executions, 1)
	executedCommand := commandRunner.Executions[0]
	assert.Equal(
		t,
		CommandExecution{
			Dir:        dir,
			Executable: "dotnet",
			Args:       []string{"test", "--filter", "FullyQualifiedName~UnitTests"},
			Cmd:        executedCommand.Cmd,
		},
		executedCommand)
	assert.Contains(t, executedCommand.Cmd.String(), "dotnet test --filter FullyQualifiedName~UnitTests")
	assert.Equal(t, "----- unit -----\ndotnet test --filter FullyQualifiedName~UnitTests\n", diagnostics.String())
}

type CommandExecution struct {
	Dir        string
	Executable string
	Args       []string
	Cmd        *exec.Cmd
}

type SpyCommandRunner struct {
	Executions []CommandExecution
	Err        error
}

func (runner *SpyCommandRunner) Exec(ctx context.Context, dir string, executable string, args []string) error {
	command := exec.CommandContext(ctx, executable, args...)
	command.Dir = dir
	runner.Executions = append(runner.Executions, CommandExecution{
		Dir:        dir,
		Executable: executable,
		Args:       args,
		Cmd:        command,
	})
	return runner.Err
}

func TestTestRunnerPropagatesDiagnosticWriteFailure(t *testing.T) {
	rejected := errors.New("diagnostic write rejected")
	runner := NewTestRunner(&SpyCommandRunner{}, &rejectingWriter{err: rejected})

	err := runner.Run(t.Context(), t.TempDir(), projects.UnitTestKind, projects.TestTarget{})

	if !errors.Is(err, rejected) {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestTestRunnerPreservesHeadingWhenCommandFails(t *testing.T) {
	failed := errors.New("command failed")
	commandRunner := &SpyCommandRunner{Err: failed}
	var diagnostics bytes.Buffer
	runner := NewTestRunner(commandRunner, &diagnostics)

	err := runner.Run(t.Context(), t.TempDir(), projects.UnitTestKind, projects.TestTarget{})

	if !errors.Is(err, failed) {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(diagnostics.String(), "dotnet test") {
		t.Fatalf("missing command preview: %q", diagnostics.String())
	}
}

func TestStdCommandRunnerPassesConfiguredStreams(t *testing.T) {
	t.Setenv("WS_DOTNET_HELPER", "success")
	input := strings.NewReader("child input")
	var stdout, stderr bytes.Buffer
	runner := StdCommandRunner{Input: input, Output: &stdout, Error: &stderr}

	err := runner.Exec(t.Context(), t.TempDir(), os.Args[0], []string{"-test.run=TestStdCommandRunnerHelper"})

	if err != nil {
		t.Fatalf("run helper: %v", err)
	}
	if stdout.String() != "child input" {
		t.Fatalf("unexpected child stdout: %q", stdout.String())
	}
	if stderr.String() != "child stderr" {
		t.Fatalf("unexpected child stderr: %q", stderr.String())
	}
}

func TestStdCommandRunnerReturnsChildFailure(t *testing.T) {
	t.Setenv("WS_DOTNET_HELPER", "failure")
	runner := StdCommandRunner{Input: strings.NewReader(""), Output: io.Discard, Error: io.Discard}

	err := runner.Exec(t.Context(), t.TempDir(), os.Args[0], []string{"-test.run=TestStdCommandRunnerHelper"})

	if err == nil {
		t.Fatal("expected child failure")
	}
}

func TestStdCommandRunnerHelper(t *testing.T) {
	mode := os.Getenv("WS_DOTNET_HELPER")
	if mode == "" {
		return
	}
	if _, err := io.Copy(os.Stdout, os.Stdin); err != nil {
		fmt.Fprint(os.Stderr, err)
		os.Exit(2)
	}
	fmt.Fprint(os.Stderr, "child stderr")
	if mode == "failure" {
		os.Exit(3)
	}
	os.Exit(0)
}

type rejectingWriter struct {
	err error
}

func (writer *rejectingWriter) Write([]byte) (int, error) {
	return 0, writer.err
}
