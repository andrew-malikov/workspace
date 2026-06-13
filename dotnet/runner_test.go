package dotnet

import (
	"context"
	"os/exec"
	"testing"

	"github.com/andrew-malikov/workspace/projects"
	"github.com/stretchr/testify/assert"
)

func TestTestRunnerRunsDotnetFilteredTest(t *testing.T) {
	// arrange
	dir := t.TempDir()
	commandRunner := SpyCommandRunner{Executions: make([]CommandExecution, 0)}
	testRunner := NewTestRunner(&commandRunner)

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
}

type CommandExecution struct {
	Dir        string
	Executable string
	Args       []string
	Cmd        *exec.Cmd
}

type SpyCommandRunner struct {
	Executions []CommandExecution
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
	return nil
}
