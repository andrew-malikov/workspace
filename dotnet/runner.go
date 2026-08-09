package dotnet

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"

	"github.com/andrew-malikov/workspace/projects"
)

type Command struct {
	Dir        string
	Executable string
	Args       []string
	Output     io.Writer
	Error      io.Writer
}

type CommandRunner interface {
	Exec(ctx context.Context, command Command) error
}

type StdCommandRunner struct {
	Input io.Reader
}

func (runner StdCommandRunner) Exec(ctx context.Context, execution Command) error {
	command := exec.CommandContext(ctx, execution.Executable, execution.Args...)
	command.Dir = execution.Dir
	command.Stdout = execution.Output
	command.Stderr = execution.Error
	command.Stdin = runner.Input

	if err := command.Run(); err != nil {
		return fmt.Errorf("executable %s failed: %w", execution.Executable, err)
	}

	return nil
}

type TestRunner struct {
	commandRunner   CommandRunner
	output          io.Writer
	errorOutput     io.Writer
	newPresentation presentationFactory
}

func NewTestRunner(commandRunner CommandRunner, output, errorOutput io.Writer, transient bool) TestRunner {
	factory := newPlainPresentation
	if transient {
		factory = newAlternateScreenPresentation
	}
	return TestRunner{
		commandRunner:   commandRunner,
		output:          output,
		errorOutput:     errorOutput,
		newPresentation: factory,
	}
}

func (runner TestRunner) Run(ctx context.Context, dir string, run TestRun, kind projects.TestKind, target projects.TestTarget) (runErr error) {
	artifacts, err := run.createArtifacts(kind)
	if err != nil {
		return err
	}

	logFile, err := os.OpenFile(artifacts.logPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return fmt.Errorf("open %s test log: %w", kind, err)
	}

	args := []string{
		"test",
		"--logger", "trx;LogFilePrefix=" + string(kind),
		"--results-directory", artifacts.resultsDir,
	}
	if strings.TrimSpace(target.Filter) != "" {
		args = append(args, "--filter", target.Filter)
	}

	presentation := runner.newPresentation(runner.errorOutput)
	if err := presentation.Begin(); err != nil {
		return errors.Join(err, logFile.Close())
	}
	defer func() {
		runErr = errors.Join(runErr, presentation.End())
	}()

	if _, err := fmt.Fprintf(runner.errorOutput, "----- %s -----\ndotnet %s\n", kind, formatCommand(args)); err != nil {
		return errors.Join(err, logFile.Close())
	}

	logOutput := &synchronizedWriter{writer: logFile}
	commandErr := runner.commandRunner.Exec(ctx, Command{
		Dir:        dir,
		Executable: "dotnet",
		Args:       args,
		Output:     teeWriter{terminal: runner.output, log: logOutput},
		Error:      teeWriter{terminal: runner.errorOutput, log: logOutput},
	})
	closeErr := logFile.Close()

	summary, summaryErr := loadTestSummary(artifacts.resultsDir)
	presentationErr := presentation.End()
	reportErr := runner.report(kind, summary, summaryErr, artifacts.logPath)
	return errors.Join(commandErr, closeErr, summaryErr, presentationErr, reportErr)
}

func (runner TestRunner) report(kind projects.TestKind, summary TestSummary, summaryErr error, logPath string) error {
	var reportErr error
	if summaryErr != nil {
		_, reportErr = fmt.Fprintf(runner.errorOutput, "%s summary unavailable: %v\n", kind, summaryErr)
	} else {
		_, reportErr = fmt.Fprintf(
			runner.errorOutput,
			"%s summary: total %d, passed %d, failed %d, skipped %d\n",
			kind,
			summary.Total,
			summary.Passed,
			summary.Failed,
			summary.Skipped,
		)
	}
	_, pathErr := fmt.Fprintf(runner.errorOutput, "Log: %s\n", logPath)
	return errors.Join(reportErr, pathErr)
}

type synchronizedWriter struct {
	mutex  sync.Mutex
	writer io.Writer
}

func (writer *synchronizedWriter) Write(data []byte) (int, error) {
	writer.mutex.Lock()
	defer writer.mutex.Unlock()
	return writer.writer.Write(data)
}

type teeWriter struct {
	terminal io.Writer
	log      io.Writer
}

func (writer teeWriter) Write(data []byte) (int, error) {
	terminalBytes, terminalErr := writer.terminal.Write(data)
	logBytes, logErr := writer.log.Write(data)
	written := min(terminalBytes, logBytes)
	return written, errors.Join(terminalErr, logErr)
}

func formatCommand(args []string) string {
	formatted := make([]string, 0, len(args))
	for _, arg := range args {
		if strings.ContainsAny(arg, " \t\n\"") {
			formatted = append(formatted, strconv.Quote(arg))
		} else {
			formatted = append(formatted, arg)
		}
	}

	return strings.Join(formatted, " ")
}
