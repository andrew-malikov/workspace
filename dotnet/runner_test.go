package dotnet

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

	"github.com/andrew-malikov/workspace/projects"
)

func TestTestRunnerStreamsAndLogsFilteredDotnetTest(t *testing.T) {
	dir := t.TempDir()
	run, err := NewTestRun(dir)
	if err != nil {
		t.Fatalf("create test run: %v", err)
	}
	summary := TestSummary{Total: 4, Passed: 2, Failed: 1, Skipped: 1}
	var stdout, stderr bytes.Buffer
	commandRunner := &SpyCommandRunner{
		Summary: &summary,
		Stdout:  "live stdout\n",
		Stderr:  "live stderr\n",
	}
	commandRunner.BeforeReturn = func(Command) {
		if stdout.String() != "live stdout\n" {
			t.Fatalf("stdout was not visible before process completion: %q", stdout.String())
		}
		if !strings.Contains(stderr.String(), "live stderr\n") {
			t.Fatalf("stderr was not visible before process completion: %q", stderr.String())
		}
	}
	testRunner := NewTestRunner(commandRunner, &stdout, &stderr)

	err = testRunner.Run(t.Context(), dir, run, projects.UnitTestKind, projects.TestTarget{
		Project: "tests/Orders.UnitTests/Orders.UnitTests.csproj",
		Filter:  "FullyQualifiedName~UnitTests",
	})

	if err != nil {
		t.Fatalf("run tests: %v", err)
	}
	if len(commandRunner.Executions) != 1 {
		t.Fatalf("unexpected execution count: %d", len(commandRunner.Executions))
	}
	execution := commandRunner.Executions[0]
	resultsDir := filepath.Join(run.Root, "unit", "results")
	wantArgs := []string{
		"test",
		"--logger", "trx;LogFilePrefix=unit",
		"--results-directory", resultsDir,
		"--filter", "FullyQualifiedName~UnitTests",
	}
	if strings.Join(execution.Args, "\x00") != strings.Join(wantArgs, "\x00") {
		t.Fatalf("unexpected args: got %q want %q", execution.Args, wantArgs)
	}
	if execution.Dir != dir || execution.Executable != "dotnet" {
		t.Fatalf("unexpected command: %+v", execution)
	}

	logPath := filepath.Join(run.Root, "unit", "output.log")
	logged, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("read output log: %v", err)
	}
	if string(logged) != "live stdout\nlive stderr\n" {
		t.Fatalf("unexpected log: %q", logged)
	}
	if !strings.Contains(stderr.String(), "unit summary: total 4, passed 2, failed 1, skipped 1\n") {
		t.Fatalf("missing summary: %q", stderr.String())
	}
	if !strings.Contains(stderr.String(), "Log: "+logPath+"\n") {
		t.Fatalf("missing absolute log path: %q", stderr.String())
	}
}

func TestTestRunnerReportsBeforeReturningCommandFailure(t *testing.T) {
	dir := t.TempDir()
	run, err := NewTestRun(dir)
	if err != nil {
		t.Fatalf("create test run: %v", err)
	}
	failed := errors.New("command failed")
	summary := TestSummary{Total: 2, Passed: 1, Failed: 1}
	commandRunner := &SpyCommandRunner{Summary: &summary, Err: failed, Stderr: "failure detail\n"}
	var diagnostics bytes.Buffer
	runner := NewTestRunner(commandRunner, io.Discard, &diagnostics)

	err = runner.Run(t.Context(), dir, run, projects.UnitTestKind, projects.TestTarget{})

	if !errors.Is(err, failed) {
		t.Fatalf("command failure was not preserved: %v", err)
	}
	logPath := filepath.Join(run.Root, "unit", "output.log")
	if !strings.Contains(diagnostics.String(), "unit summary: total 2, passed 1, failed 1, skipped 0\nLog: "+logPath+"\n") {
		t.Fatalf("report was not printed before failure: %q", diagnostics.String())
	}
	logged, readErr := os.ReadFile(logPath)
	if readErr != nil {
		t.Fatalf("read failed output log: %v", readErr)
	}
	if string(logged) != "failure detail\n" {
		t.Fatalf("unexpected failed output log: %q", logged)
	}
}

func TestTestRunnerReportsUnavailableSummaryAndLogPath(t *testing.T) {
	dir := t.TempDir()
	run, err := NewTestRun(dir)
	if err != nil {
		t.Fatalf("create test run: %v", err)
	}
	var diagnostics bytes.Buffer
	runner := NewTestRunner(&SpyCommandRunner{}, io.Discard, &diagnostics)

	err = runner.Run(t.Context(), dir, run, projects.ComponentTestKind, projects.TestTarget{})

	if err == nil {
		t.Fatal("expected missing result error")
	}
	logPath := filepath.Join(run.Root, "component", "output.log")
	if !strings.Contains(diagnostics.String(), "component summary unavailable:") {
		t.Fatalf("missing unavailable summary: %q", diagnostics.String())
	}
	if !strings.Contains(diagnostics.String(), "Log: "+logPath+"\n") {
		t.Fatalf("missing log path: %q", diagnostics.String())
	}
}

func TestTestRunnerDoesNotExecuteWhenLogSetupFails(t *testing.T) {
	rootFile := filepath.Join(t.TempDir(), "not-a-directory")
	if err := os.WriteFile(rootFile, []byte("file"), 0o600); err != nil {
		t.Fatalf("create root file: %v", err)
	}
	commandRunner := &SpyCommandRunner{}
	runner := NewTestRunner(commandRunner, io.Discard, io.Discard)

	err := runner.Run(t.Context(), t.TempDir(), TestRun{Root: rootFile}, projects.UnitTestKind, projects.TestTarget{})

	if err == nil {
		t.Fatal("expected log setup error")
	}
	if len(commandRunner.Executions) != 0 {
		t.Fatalf("command ran before log setup: %+v", commandRunner.Executions)
	}
}

func TestTestRunnerPropagatesDiagnosticWriteFailure(t *testing.T) {
	run, err := NewTestRun(t.TempDir())
	if err != nil {
		t.Fatalf("create test run: %v", err)
	}
	rejected := errors.New("diagnostic write rejected")
	commandRunner := &SpyCommandRunner{}
	runner := NewTestRunner(commandRunner, io.Discard, &rejectingWriter{err: rejected})

	err = runner.Run(t.Context(), t.TempDir(), run, projects.UnitTestKind, projects.TestTarget{})

	if !errors.Is(err, rejected) {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(commandRunner.Executions) != 0 {
		t.Fatalf("command ran after heading failure: %+v", commandRunner.Executions)
	}
}

type SpyCommandRunner struct {
	Executions   []Command
	Summary      *TestSummary
	Stdout       string
	Stderr       string
	Err          error
	BeforeReturn func(Command)
}

func (runner *SpyCommandRunner) Exec(_ context.Context, command Command) error {
	runner.Executions = append(runner.Executions, command)
	if runner.Stdout != "" {
		_, _ = io.WriteString(command.Output, runner.Stdout)
	}
	if runner.Stderr != "" {
		_, _ = io.WriteString(command.Error, runner.Stderr)
	}
	if runner.Summary != nil {
		resultsDir := argumentValue(command.Args, "--results-directory")
		writeTRXFile(resultsDir, *runner.Summary)
	}
	if runner.BeforeReturn != nil {
		runner.BeforeReturn(command)
	}
	return runner.Err
}

func argumentValue(args []string, name string) string {
	for index := 0; index+1 < len(args); index++ {
		if args[index] == name {
			return args[index+1]
		}
	}
	return ""
}

func writeTRXFile(resultsDir string, summary TestSummary) {
	if resultsDir == "" {
		return
	}
	_ = os.MkdirAll(resultsDir, 0o755)
	contents := fmt.Sprintf(
		`<TestRun><ResultSummary><Counters total="%d" passed="%d" failed="%d" notExecuted="%d"/></ResultSummary></TestRun>`,
		summary.Total,
		summary.Passed,
		summary.Failed,
		summary.Skipped,
	)
	_ = os.WriteFile(filepath.Join(resultsDir, "result.trx"), []byte(contents), 0o600)
}

func TestStdCommandRunnerPassesConfiguredStreams(t *testing.T) {
	t.Setenv("WS_DOTNET_HELPER", "success")
	input := strings.NewReader("child input")
	var stdout, stderr bytes.Buffer
	runner := StdCommandRunner{Input: input}

	err := runner.Exec(t.Context(), Command{
		Dir:        t.TempDir(),
		Executable: os.Args[0],
		Args:       []string{"-test.run=TestStdCommandRunnerHelper"},
		Output:     &stdout,
		Error:      &stderr,
	})

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
	runner := StdCommandRunner{Input: strings.NewReader("")}

	err := runner.Exec(t.Context(), Command{
		Dir:        t.TempDir(),
		Executable: os.Args[0],
		Args:       []string{"-test.run=TestStdCommandRunnerHelper"},
		Output:     io.Discard,
		Error:      io.Discard,
	})

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
