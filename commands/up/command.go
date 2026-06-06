package up

import (
	"context"
	"fmt"
	"os"
	"strings"
	"text/template"

	cfg "github.com/andrew-malikov/workspace/config"
	"github.com/andrew-malikov/workspace/projects"
	"github.com/andrew-malikov/workspace/view"

	"github.com/urfave/cli/v3"
)

type composeRunner func(context.Context, projects.Project) error
type statusChecker func(context.Context, projects.Project) (bool, error)

type upResult struct {
	Alias     string
	Alongside bool
	Stopped   []string
}

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:  "up",
		Usage: "start project docker compose",
		Arguments: []cli.Argument{
			&cli.StringArg{
				Name:      "alias",
				UsageText: "tracked project alias",
			},
		},
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  "alongside",
				Usage: "keep other running docker compose projects active",
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			config, err := cfg.LoadConfig()
			if err != nil {
				return err
			}

			dir, err := os.Getwd()
			if err != nil {
				return err
			}

			result, err := upProject(ctx, config, dir, cmd.StringArg("alias"), cmd.Bool("alongside"), hasRunningContainers, runProjectUp, runProjectDown)
			if err != nil {
				return err
			}

			return view.Render(RESULT_TEMPLATE, view.Args{
				"Alias":     result.Alias,
				"Alongside": result.Alongside,
				"Stopped":   strings.Join(result.Stopped, ", "),
			})
		},
	}
}

func upProject(ctx context.Context, config *cfg.Config, cwd string, alias string, alongside bool, hasRunning statusChecker, up composeRunner, down composeRunner) (*upResult, error) {
	target, err := resolveProject(config, cwd, alias)
	if err != nil {
		return nil, err
	}

	if !target.DoesComposeExist() {
		return nil, fmt.Errorf("no docker compose configured for project %s", target.Alias)
	}

	stopped := []string{}
	if !alongside {
		for _, project := range config.Projects {
			if project.Alias == target.Alias {
				continue
			}

			hasRunning, err := hasRunning(ctx, project)
			if err != nil {
				return nil, err
			}

			if !hasRunning {
				continue
			}

			if err := down(ctx, project); err != nil {
				return nil, err
			}
			stopped = append(stopped, project.Alias)
		}
	}

	if err := up(ctx, *target); err != nil {
		return nil, err
	}

	return &upResult{Alias: target.Alias, Alongside: alongside, Stopped: stopped}, nil
}

func resolveProject(config *cfg.Config, cwd string, alias string) (*projects.Project, error) {
	if strings.TrimSpace(alias) == "" {
		project := config.ResolveProjectByDir(cwd)
		if project != nil {
			return project, nil
		}
		return nil, fmt.Errorf("project is not tracked: %s", cwd)
	}

	if project, ok := config.Projects[alias]; ok {
		return &project, nil
	}

	return nil, fmt.Errorf("project is not tracked: %s", alias)
}

func hasRunningContainers(ctx context.Context, project projects.Project) (bool, error) {
	return project.HasRunningContainers(ctx)
}

func runProjectUp(ctx context.Context, project projects.Project) error {
	return project.UpContainers(ctx)
}

func runProjectDown(ctx context.Context, project projects.Project) error {
	return project.DownContainers(ctx)
}

var RESULT_TEMPLATE = template.Must(template.New("up_result").Parse(
	`Project *{{.Alias}}* docker compose is up.
{{if .Alongside}}Other running docker compose projects left active.
{{else}}{{if .Stopped}}Stopped running docker compose projects: {{.Stopped}}.
{{else}}No other running docker compose projects found.
{{end}}{{end}}`,
))
