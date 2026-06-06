package containers

import (
	"context"
	"os/exec"
	"strings"
)

type Compose interface {
	HasRunning(ctx context.Context, dir string, compose string) (bool, error)
	Up(ctx context.Context, dir string, compose string) error
	Down(ctx context.Context, dir string, compose string) error
	Cleanup(ctx context.Context, dir string, compose string) error
}

type DockerCompose struct{}

func (docker DockerCompose) HasRunning(ctx context.Context, dir string, compose string) (bool, error) {
	out, err := docker.run(ctx, dir, compose, "ps", "--status", "running", "--format", "json")
	if err != nil {
		return false, err
	}

	return HasRunning(out), nil
}

func (docker DockerCompose) Up(ctx context.Context, dir string, compose string) error {
	_, err := docker.run(ctx, dir, compose, "up", "-d")
	return err
}

func (docker DockerCompose) Down(ctx context.Context, dir string, compose string) error {
	_, err := docker.run(ctx, dir, compose, "down")
	return err
}

func (docker DockerCompose) Cleanup(ctx context.Context, dir string, compose string) error {
	_, err := docker.run(ctx, dir, compose, "down", "-v")
	return err
}

func (docker DockerCompose) run(ctx context.Context, dir string, compose string, args ...string) ([]byte, error) {
	commandArgs := append([]string{"compose", "-f", compose}, args...)
	command := exec.CommandContext(ctx, "docker", commandArgs...)
	command.Dir = dir
	return command.Output()
}

func HasRunning(out []byte) bool {
	trimmed := strings.TrimSpace(string(out))
	return trimmed != "" && trimmed != "[]"
}
