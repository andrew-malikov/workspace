package dotnet

import (
	"context"
	"fmt"
	"io"
	"os/exec"
	"strconv"
	"strings"

	"github.com/andrew-malikov/workspace/projects"
)

type CommandRunner interface {
	Exec(ctx context.Context, dir string, executable string, args []string) error
}

type StdCommandRunner struct {
	Input  io.Reader
	Output io.Writer
	Error  io.Writer
}

func (runner StdCommandRunner) Exec(ctx context.Context, dir string, executable string, args []string) error {
	command := exec.CommandContext(ctx, executable, args...)
	command.Dir = dir
	command.Stdout = runner.Output
	command.Stderr = runner.Error
	command.Stdin = runner.Input

	if err := command.Run(); err != nil {
		return fmt.Errorf("executable %s failed: %w", executable, err)
	}

	return nil
}

type TestRunner struct {
	commandRunner CommandRunner
	diagnostics   io.Writer
}

func NewTestRunner(commandRunner CommandRunner, diagnostics io.Writer) TestRunner {
	return TestRunner{commandRunner: commandRunner, diagnostics: diagnostics}
}

func (runner TestRunner) Run(ctx context.Context, dir string, kind projects.TestKind, target projects.TestTarget) error {
	args := []string{"test"}
	if strings.TrimSpace(target.Filter) != "" {
		args = append(args, "--filter", target.Filter)
	}

	if _, err := fmt.Fprintf(runner.diagnostics, "----- %s -----\ndotnet %s\n", kind, formatCommand(args)); err != nil {
		return err
	}
	return runner.commandRunner.Exec(ctx, dir, "dotnet", args)
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
