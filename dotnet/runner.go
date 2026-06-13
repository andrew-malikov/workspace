package dotnet

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/andrew-malikov/workspace/projects"
)

type CommandRunner interface {
	Exec(ctx context.Context, dir string, executable string, args []string) error
}

type StdCommandRunner struct{}

func (runner StdCommandRunner) Exec(ctx context.Context, dir string, executable string, args []string) error {
	command := exec.CommandContext(ctx, executable, args...)
	command.Dir = dir
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	command.Stdin = os.Stdin

	if err := command.Run(); err != nil {
		return fmt.Errorf("executable %s failed: %w", executable, err)
	}

	return nil
}

type TestRunner struct {
	commandRunner CommandRunner
}

func NewTestRunner(commandRunner CommandRunner) TestRunner {
	return TestRunner{commandRunner: commandRunner}
}

func (runner TestRunner) Run(ctx context.Context, dir string, kind projects.TestKind, target projects.TestTarget) error {
	args := []string{"test"}
	if strings.TrimSpace(target.Filter) != "" {
		args = append(args, "--filter", target.Filter)
	}

	fmt.Printf("----- %s -----\n", kind)
	fmt.Printf("dotnet %s\n", formatCommand(args))

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
