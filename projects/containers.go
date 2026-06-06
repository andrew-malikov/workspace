package projects

import (
	"context"

	"github.com/andrew-malikov/workspace/containers"
)

func (project Project) HasRunningContainers(ctx context.Context, compose containers.Compose) (bool, error) {
	if !project.DoesComposeExist() {
		// todo: surely nil is a mistake, need to take time to make decision
		return false, nil
	}

	return compose.HasRunning(ctx, project.Dir, *project.Compose)
}

func (project Project) UpContainers(ctx context.Context, compose containers.Compose) error {
	if !project.DoesComposeExist() {
		return nil
	}

	return compose.Up(ctx, project.Dir, *project.Compose)
}

func (project Project) DownContainers(ctx context.Context, compose containers.Compose) error {
	if !project.DoesComposeExist() {
		return nil
	}

	return compose.Down(ctx, project.Dir, *project.Compose)
}

func (project Project) CleanupContainers(ctx context.Context, compose containers.Compose) error {
	if !project.DoesComposeExist() {
		return nil
	}

	return compose.Cleanup(ctx, project.Dir, *project.Compose)
}
