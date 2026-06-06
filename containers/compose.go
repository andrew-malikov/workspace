package containers

import (
	"context"
	"os/exec"
	"strings"
)

type Runner func(ctx context.Context, dir string, compose string) ([]byte, error)

func DockerCompose(ctx context.Context, dir string, compose string) ([]byte, error) {
	return RunDockerCompose(ctx, dir, compose, "ps", "--status", "running", "--format", "json")
}

func DockerComposeUp(ctx context.Context, dir string, compose string) error {
	_, err := RunDockerCompose(ctx, dir, compose, "up", "-d")
	return err
}

func DockerComposeDown(ctx context.Context, dir string, compose string) error {
	_, err := RunDockerCompose(ctx, dir, compose, "down")
	return err
}

func RunDockerCompose(ctx context.Context, dir string, compose string, args ...string) ([]byte, error) {
	commandArgs := append([]string{"compose", "-f", compose}, args...)
	command := exec.CommandContext(ctx, "docker", commandArgs...)
	command.Dir = dir
	return command.Output()
}

func HasRunning(out []byte) bool {
	trimmed := strings.TrimSpace(string(out))
	return trimmed != "" && trimmed != "[]"
}
