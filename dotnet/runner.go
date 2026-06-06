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

func RunTest(ctx context.Context, dir string, kind projects.TestKind, target projects.TestTarget) error {
	args := []string{"test"}
	if strings.TrimSpace(target.Filter) != "" {
		args = append(args, "--filter", target.Filter)
	}

	fmt.Printf("----- %s -----\n", kind)
	fmt.Printf("dotnet %s\n", formatCommand(args))

	command := exec.CommandContext(ctx, "dotnet", args...)
	command.Dir = dir
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	command.Stdin = os.Stdin

	if err := command.Run(); err != nil {
		return fmt.Errorf("dotnet test failed for %s: %w", kind, err)
	}

	return nil
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
