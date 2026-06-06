package projects

import (
	"context"

	"github.com/andrew-malikov/workspace/containers"
)

func (project Project) HasRunningContainers(ctx context.Context) (bool, error) {
	if !project.DoesComposeExist() {
		return false, nil
	}

	out, err := containers.DockerCompose(ctx, project.Dir, *project.Compose)
	if err != nil {
		return false, err
	}

	return containers.HasRunning(out), nil
}

func (project Project) UpContainers(ctx context.Context) error {
	if !project.DoesComposeExist() {
		return nil
	}

	return containers.DockerComposeUp(ctx, project.Dir, *project.Compose)
}

func (project Project) DownContainers(ctx context.Context) error {
	if !project.DoesComposeExist() {
		return nil
	}

	return containers.DockerComposeDown(ctx, project.Dir, *project.Compose)
}
