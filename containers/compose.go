package containers

import (
	"context"
	"os/exec"
	"strings"
)

type Runner func(ctx context.Context, dir string, compose string) ([]byte, error)

func DockerCompose(ctx context.Context, dir string, compose string) ([]byte, error) {
	command := exec.CommandContext(ctx, "docker", "compose", "-f", compose, "ps", "--status", "running", "--format", "json")
	command.Dir = dir
	return command.Output()
}

func HasRunning(out []byte) bool {
	trimmed := strings.TrimSpace(string(out))
	return trimmed != "" && trimmed != "[]"
}
