package workspaces

import (
	"context"
	"fmt"
	"strings"

	"github.com/andrew-malikov/workspace/containers"
	"github.com/andrew-malikov/workspace/projects"
)

type UpProjectResult struct {
	Alias     string
	Alongside bool
	Stopped   []string
}

func (workspace Workspace) ResolveProject(cwd string, alias string) (*projects.Project, error) {
	if strings.TrimSpace(alias) == "" {
		project := workspace.ResolveProjectByDir(cwd)
		if project != nil {
			return project, nil
		}
		return nil, fmt.Errorf("project is not tracked: %s", cwd)
	}

	if project, ok := workspace.Projects[alias]; ok {
		return &project, nil
	}

	return nil, fmt.Errorf("project is not tracked: %s", alias)
}

func (workspace Workspace) UpProject(ctx context.Context, cwd string, alias string, alongside bool, compose containers.Compose) (*UpProjectResult, error) {
	target, err := workspace.ResolveProject(cwd, alias)
	if err != nil {
		return nil, err
	}

	if !target.DoesComposeExist() {
		return nil, fmt.Errorf("no docker compose configured for project %s", target.Alias)
	}

	stopped := []string{}
	if !alongside {
		for _, project := range workspace.Projects {
			if project.Alias == target.Alias {
				continue
			}

			hasRunning, err := project.HasRunningContainers(ctx, compose)
			if err != nil {
				return nil, err
			}

			if !hasRunning {
				continue
			}

			if err := project.DownContainers(ctx, compose); err != nil {
				return nil, err
			}
			stopped = append(stopped, project.Alias)
		}
	}

	if err := target.UpContainers(ctx, compose); err != nil {
		return nil, err
	}

	return &UpProjectResult{Alias: target.Alias, Alongside: alongside, Stopped: stopped}, nil
}

func (workspace Workspace) DownProject(ctx context.Context, cwd string, alias string, compose containers.Compose) (*projects.Project, error) {
	project, err := workspace.ResolveProject(cwd, alias)
	if err != nil {
		return nil, err
	}

	if !project.DoesComposeExist() {
		return nil, fmt.Errorf("no docker compose configured for project %s", project.Alias)
	}

	if err := project.DownContainers(ctx, compose); err != nil {
		return nil, err
	}

	return project, nil
}
