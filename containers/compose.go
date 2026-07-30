package containers

import (
	"context"
	"fmt"
	"io"
	"os/exec"
	"strconv"
	"strings"
	"unicode"
)

type Target struct {
	Alias string
	Dir   string
	File  string
}

type Compose interface {
	HasRunning(ctx context.Context, target Target) (bool, error)
	Up(ctx context.Context, target Target) error
	Down(ctx context.Context, target Target) error
	Cleanup(ctx context.Context, target Target) error
}

type commandFactory func(ctx context.Context, name string, args ...string) *exec.Cmd

type DockerCompose struct {
	input       io.Reader
	output      io.Writer
	errorOutput io.Writer
	command     commandFactory
}

func NewDockerCompose(input io.Reader, output, errorOutput io.Writer) DockerCompose {
	return DockerCompose{
		input:       input,
		output:      output,
		errorOutput: errorOutput,
		command:     exec.CommandContext,
	}
}

func (docker DockerCompose) HasRunning(ctx context.Context, target Target) (bool, error) {
	out, err := docker.query(ctx, target, "ps", "--status", "running", "--format", "json")
	if err != nil {
		return false, err
	}

	return HasRunning(out), nil
}

func (docker DockerCompose) Up(ctx context.Context, target Target) error {
	return docker.action(ctx, target, "start", "up", "-d")
}

func (docker DockerCompose) Down(ctx context.Context, target Target) error {
	return docker.action(ctx, target, "stop", "down")
}

func (docker DockerCompose) Cleanup(ctx context.Context, target Target) error {
	return docker.action(ctx, target, "clean volumes", "down", "-v")
}

func (docker DockerCompose) query(ctx context.Context, target Target, args ...string) ([]byte, error) {
	commandArgs := append([]string{"compose", "-f", target.File}, args...)
	command := docker.command(ctx, "docker", commandArgs...)
	command.Dir = target.Dir
	return command.Output()
}

func (docker DockerCompose) action(ctx context.Context, target Target, operation string, args ...string) error {
	commandArgs := append([]string{"compose", "-f", target.File}, args...)
	if _, err := fmt.Fprintf(
		docker.errorOutput,
		"----- project %s in %s: %s -----\n",
		strconv.QuoteToASCII(target.Alias),
		strconv.QuoteToASCII(target.Dir),
		operation,
	); err != nil {
		return fmt.Errorf("write docker compose action heading: %w", err)
	}
	if _, err := fmt.Fprintf(docker.errorOutput, "docker %s\n", formatCommand(commandArgs)); err != nil {
		return fmt.Errorf("write docker compose command preview: %w", err)
	}

	command := docker.command(ctx, "docker", commandArgs...)
	command.Dir = target.Dir
	command.Stdin = docker.input
	command.Stdout = docker.output
	command.Stderr = docker.errorOutput
	return command.Run()
}

func formatCommand(args []string) string {
	formatted := make([]string, len(args))
	for i, arg := range args {
		formatted[i] = formatCommandArg(arg)
	}
	return strings.Join(formatted, " ")
}

func formatCommandArg(arg string) string {
	if arg != "" && strings.IndexFunc(arg, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r) && !strings.ContainsRune("-._/:", r)
	}) == -1 {
		return arg
	}
	return strconv.QuoteToASCII(arg)
}

func HasRunning(out []byte) bool {
	trimmed := strings.TrimSpace(string(out))
	return trimmed != "" && trimmed != "[]"
}
