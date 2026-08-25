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
	"sync"
	"testing"
	"time"

	"github.com/andrew-malikov/workspace/projects"
	"github.com/charmbracelet/x/ansi"
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
		Stdout: "live stdout\n",
		Stderr: "live stderr\n",
	}
	commandRunner.BeforeReturn = func(Command) {
		if stdout.String() != "live stdout\n" {
			t.Fatalf("stdout was not visible before process completion: %q", stdout.String())
		}
		if !strings.Contains(stderr.String(), "live stderr\n") {
			t.Fatalf("stderr was not visible before process completion: %q", stderr.String())
		}
	}
	testRunner := makeTestRunner(commandRunner, &stdout, &stderr, newPlainPresentation, staticSummary(summary, nil))

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
	wantArgs := TestArgs(projects.UnitTestKind, "FullyQualifiedName~UnitTests", resultsDir)
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
	if strings.Contains(stderr.String(), ansi.SetModeAltScreenSaveCursor) ||
		strings.Contains(stderr.String(), ansi.ResetModeAltScreenSaveCursor) {
		t.Fatalf("plain output contained terminal controls: %q", stderr.String())
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
	commandRunner := &SpyCommandRunner{Err: failed, Stderr: "failure detail\n"}
	var diagnostics bytes.Buffer
	runner := makeTestRunner(commandRunner, io.Discard, &diagnostics, newPlainPresentation, staticSummary(summary, nil))

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
	missing := errors.New("no TRX result files found")
	runner := makeTestRunner(&SpyCommandRunner{}, io.Discard, &diagnostics, newPlainPresentation, staticSummary(TestSummary{}, missing))

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
	runner := makeTestRunner(commandRunner, io.Discard, io.Discard, newPlainPresentation, staticSummary(TestSummary{}, nil))

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
	runner := makeTestRunner(commandRunner, io.Discard, &rejectingWriter{err: rejected}, newPlainPresentation, staticSummary(TestSummary{}, nil))

	err = runner.Run(t.Context(), t.TempDir(), run, projects.UnitTestKind, projects.TestTarget{})

	if !errors.Is(err, rejected) {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(commandRunner.Executions) != 0 {
		t.Fatalf("command ran after heading failure: %+v", commandRunner.Executions)
	}
}

func TestTestRunnerCollapsesTransientOutputBeforeReporting(t *testing.T) {
	dir := t.TempDir()
	run, err := NewTestRun(dir)
	if err != nil {
		t.Fatalf("create test run: %v", err)
	}
	summary := TestSummary{Total: 2, Passed: 2}
	commandRunner := &SpyCommandRunner{
		Stdout: "live stdout\n",
		Stderr: "live stderr\n",
	}
	var terminal bytes.Buffer
	runner := makeTestRunner(commandRunner, &terminal, &terminal, newAlternateScreenPresentation, staticSummary(summary, nil))

	err = runner.Run(t.Context(), dir, run, projects.UnitTestKind, projects.TestTarget{})

	if err != nil {
		t.Fatalf("run tests: %v", err)
	}
	output := terminal.String()
	enter := strings.Index(output, ansi.SetModeAltScreenSaveCursor)
	heading := strings.Index(output, "----- unit -----")
	raw := strings.Index(output, "live stdout\nlive stderr\n")
	exit := strings.Index(output, ansi.ResetModeAltScreenSaveCursor)
	summaryIndex := strings.Index(output, "unit summary:")
	if !(enter == 0 && enter < heading && heading < raw && raw < exit && exit < summaryIndex) {
		t.Fatalf("unexpected transient lifecycle order: %q", output)
	}

	logged, readErr := os.ReadFile(filepath.Join(run.Root, "unit", "output.log"))
	if readErr != nil {
		t.Fatalf("read output log: %v", readErr)
	}
	if string(logged) != "live stdout\nlive stderr\n" {
		t.Fatalf("unexpected transient log: %q", logged)
	}
}

func TestTestRunnerDoesNotExecuteWhenPresentationEntryFails(t *testing.T) {
	run, err := NewTestRun(t.TempDir())
	if err != nil {
		t.Fatalf("create test run: %v", err)
	}
	rejected := errors.New("presentation rejected")
	commandRunner := &SpyCommandRunner{}
	runner := makeTestRunner(commandRunner, io.Discard, &rejectingWriter{err: rejected}, newAlternateScreenPresentation, staticSummary(TestSummary{}, nil))

	err = runner.Run(t.Context(), t.TempDir(), run, projects.UnitTestKind, projects.TestTarget{})

	if !errors.Is(err, rejected) {
		t.Fatalf("unexpected entry error: %v", err)
	}
	if len(commandRunner.Executions) != 0 {
		t.Fatalf("command ran after entry failure: %+v", commandRunner.Executions)
	}
}

func TestTestRunnerReportsAfterRestorationFailure(t *testing.T) {
	run, err := NewTestRun(t.TempDir())
	if err != nil {
		t.Fatalf("create test run: %v", err)
	}
	restoreErr := errors.New("restore failed")
	var terminal bytes.Buffer
	presentation := &recordingPresentation{output: &terminal, endErr: restoreErr}
	commandRunner := &SpyCommandRunner{Stdout: "raw output\n"}
	runner := makeTestRunner(
		commandRunner,
		&terminal,
		&terminal,
		func(io.Writer) categoryPresentation { return presentation },
		staticSummary(TestSummary{Total: 1, Passed: 1}, nil),
	)

	err = runner.Run(t.Context(), t.TempDir(), run, projects.UnitTestKind, projects.TestTarget{})

	if !errors.Is(err, restoreErr) {
		t.Fatalf("restoration error was not retained: %v", err)
	}
	output := terminal.String()
	if strings.Index(output, "END\n") > strings.Index(output, "unit summary:") {
		t.Fatalf("summary preceded restoration: %q", output)
	}
	if !strings.Contains(output, "Log: ") {
		t.Fatalf("report was suppressed after restoration failure: %q", output)
	}
}

func TestTestRunnerRestoresCanceledCategoryAndRetainsLog(t *testing.T) {
	run, err := NewTestRun(t.TempDir())
	if err != nil {
		t.Fatalf("create test run: %v", err)
	}
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	var terminal bytes.Buffer
	runner := makeTestRunner(
		cancelingCommandRunner{},
		&terminal,
		&terminal,
		newAlternateScreenPresentation,
		staticSummary(TestSummary{}, errors.New("canceled")),
	)

	err = runner.Run(ctx, t.TempDir(), run, projects.ComponentTestKind, projects.TestTarget{})

	if !errors.Is(err, context.Canceled) {
		t.Fatalf("cancellation was not retained: %v", err)
	}
	output := terminal.String()
	if strings.Index(output, ansi.ResetModeAltScreenSaveCursor) > strings.Index(output, "component summary unavailable:") {
		t.Fatalf("canceled category reported before restoration: %q", output)
	}
	logged, readErr := os.ReadFile(filepath.Join(run.Root, "component", "output.log"))
	if readErr != nil {
		t.Fatalf("read canceled output log: %v", readErr)
	}
	if string(logged) != "partial output\n" {
		t.Fatalf("unexpected canceled output log: %q", logged)
	}
}

func TestTestRunnerUsesSeparateTransientLifecyclePerCategory(t *testing.T) {
	run, err := NewTestRun(t.TempDir())
	if err != nil {
		t.Fatalf("create test run: %v", err)
	}
	var terminal bytes.Buffer
	commandRunner := &SpyCommandRunner{}
	runner := makeTestRunner(commandRunner, &terminal, &terminal, newAlternateScreenPresentation, staticSummary(TestSummary{Total: 1, Passed: 1}, nil))

	for _, kind := range []projects.TestKind{projects.UnitTestKind, projects.IntegrationTestKind} {
		if err := runner.Run(t.Context(), t.TempDir(), run, kind, projects.TestTarget{}); err != nil {
			t.Fatalf("run %s tests: %v", kind, err)
		}
	}

	output := terminal.String()
	if strings.Count(output, ansi.SetModeAltScreenSaveCursor) != 2 ||
		strings.Count(output, ansi.ResetModeAltScreenSaveCursor) != 2 {
		t.Fatalf("categories did not receive separate lifecycles: %q", output)
	}
	firstEnter := strings.Index(output, ansi.SetModeAltScreenSaveCursor)
	secondEnterOffset := strings.Index(output[firstEnter+len(ansi.SetModeAltScreenSaveCursor):], ansi.SetModeAltScreenSaveCursor)
	secondEnter := firstEnter + len(ansi.SetModeAltScreenSaveCursor) + secondEnterOffset
	if firstEnter < 0 || secondEnterOffset < 0 || strings.Index(output, "unit summary:") > secondEnter {
		t.Fatalf("unit report did not precede the next category: %q", output)
	}
}

func staticSummary(summary TestSummary, err error) summaryLoader {
	return func(string) (TestSummary, error) {
		return summary, err
	}
}

type recordingPresentation struct {
	output *bytes.Buffer
	endErr error
	ended  bool
}

func (presentation *recordingPresentation) Begin() error {
	_, _ = io.WriteString(presentation.output, "BEGIN\n")
	return nil
}

func (presentation *recordingPresentation) End() error {
	if presentation.ended {
		return nil
	}
	presentation.ended = true
	_, _ = io.WriteString(presentation.output, "END\n")
	return presentation.endErr
}

type cancelingCommandRunner struct{}

func (cancelingCommandRunner) Exec(ctx context.Context, command Command) error {
	_, _ = io.WriteString(command.Output, "partial output\n")
	<-ctx.Done()
	return ctx.Err()
}

type SpyCommandRunner struct {
	Executions   []Command
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
	if runner.BeforeReturn != nil {
		runner.BeforeReturn(command)
	}
	return runner.Err
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

func TestStdCommandRunnerTerminatesChildOnCancellation(t *testing.T) {
	t.Setenv("WS_DOTNET_HELPER", "wait")
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	ready := make(chan struct{})
	stderr := &readyWriter{ready: ready}
	runner := StdCommandRunner{Input: strings.NewReader("")}
	done := make(chan error, 1)

	go func() {
		done <- runner.Exec(ctx, Command{
			Dir:        t.TempDir(),
			Executable: os.Args[0],
			Args:       []string{"-test.run=TestStdCommandRunnerHelper"},
			Output:     io.Discard,
			Error:      stderr,
		})
	}()

	select {
	case <-ready:
		cancel()
	case <-time.After(2 * time.Second):
		t.Fatal("child did not start")
	}

	select {
	case err := <-done:
		if err == nil {
			t.Fatal("canceled child returned success")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("canceled child did not terminate")
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
	if mode == "wait" {
		time.Sleep(time.Minute)
	}
	if mode == "failure" {
		os.Exit(3)
	}
	os.Exit(0)
}

type readyWriter struct {
	ready chan struct{}
	once  sync.Once
}

func (writer *readyWriter) Write(data []byte) (int, error) {
	writer.once.Do(func() { close(writer.ready) })
	return len(data), nil
}

type rejectingWriter struct {
	err error
}

func (writer *rejectingWriter) Write([]byte) (int, error) {
	return 0, writer.err
}
