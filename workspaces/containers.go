package workspaces

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/andrew-malikov/workspace/containers"
	"github.com/andrew-malikov/workspace/projects"
)

type UpProjectResult struct {
	Alias     string
	Alongside bool
	Blank     bool
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

func (workspace Workspace) UpProject(ctx context.Context, cwd string, alias string, alongside bool, blank bool, compose containers.Compose) (*UpProjectResult, error) {
	target, err := workspace.ResolveProject(cwd, alias)
	if err != nil {
		return nil, err
	}

	if !target.DoesComposeExist() {
		return nil, fmt.Errorf("no docker compose configured for project %s", target.Alias)
	}

	var others []OtherProject
	if !alongside {
		others, err = workspace.probeOthers(ctx, target.Alias, compose)
		if err != nil {
			return nil, err
		}
	}

	plan := PlanUp(target.Alias, others, alongside, blank)
	if err := workspace.execute(ctx, compose, plan); err != nil {
		return nil, err
	}

	return &UpProjectResult{
		Alias:     target.Alias,
		Alongside: alongside,
		Blank:     blank,
		Stopped:   stoppedAliases(plan),
	}, nil
}

func (workspace Workspace) DownProject(ctx context.Context, cwd string, alias string, blank bool, compose containers.Compose) (*projects.Project, error) {
	project, err := workspace.ResolveProject(cwd, alias)
	if err != nil {
		return nil, err
	}

	if !project.DoesComposeExist() {
		return nil, fmt.Errorf("no docker compose configured for project %s", project.Alias)
	}

	if err := workspace.execute(ctx, compose, PlanDown(project.Alias, blank)); err != nil {
		return nil, err
	}

	return project, nil
}

func (workspace Workspace) probeOthers(ctx context.Context, targetAlias string, compose containers.Compose) ([]OtherProject, error) {
	aliases := make([]string, 0, len(workspace.Projects))
	for alias := range workspace.Projects {
		if alias == targetAlias {
			continue
		}
		aliases = append(aliases, alias)
	}
	sort.Strings(aliases)

	others := make([]OtherProject, 0, len(aliases))
	for _, alias := range aliases {
		project := workspace.Projects[alias]
		configured := project.IsComposeConfigured()
		running := false
		if configured {
			var err error
			running, err = project.HasRunningContainers(ctx, compose)
			if err != nil {
				return nil, err
			}
		}
		others = append(others, OtherProject{Alias: alias, Configured: configured, Running: running})
	}
	return others, nil
}

func (workspace Workspace) execute(ctx context.Context, compose containers.Compose, plan []Action) error {
	for _, action := range plan {
		project, ok := workspace.Projects[action.Alias]
		if !ok {
			return fmt.Errorf("planned project is not tracked: %s", action.Alias)
		}
		target := containers.Target{Alias: project.Alias, Dir: project.Dir, File: *project.Compose}
		var err error
		switch action.Kind {
		case ActionStop:
			err = compose.Down(ctx, target)
		case ActionCleanup:
			err = compose.Cleanup(ctx, target)
		case ActionStart:
			err = compose.Up(ctx, target)
		default:
			return fmt.Errorf("unknown compose action %q", action.Kind)
		}
		if err != nil {
			return err
		}
	}
	return nil
}
