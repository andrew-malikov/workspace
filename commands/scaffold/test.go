package scaffold

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

func newTestCommand() *cli.Command {
	return &cli.Command{
		Name:  "test",
		Usage: "copy global test options into a tracked project",
		Arguments: []cli.Argument{
			&cli.StringArg{
				Name:      "directory_or_alias",
				UsageText: "tracked directory or alias",
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			config, err := cfg.LoadConfig()
			if err != nil {
				return err
			}

			project, err := resolveProject(config, cmd.StringArg("directory_or_alias"))
			if err != nil {
				return err
			}

			if err := config.ScaffoldProjectTests(project); err != nil {
				return err
			}

			config.Projects[project.Alias] = *project
			if err := cfg.SaveConfig(*config); err != nil {
				return err
			}

			return view.Render(TEST_RESULT_TEMPLATE, view.Args{
				"Alias":               project.Alias,
				"HasUnitTests":        project.Test.Unit.IsConfigured(),
				"HasIntegrationTests": project.Test.Integration.IsConfigured(),
				"HasComponentTests":   project.Test.Component.IsConfigured(),
			})
		},
	}
}

func resolveProject(config *cfg.Config, alsdir string) (*projects.Project, error) {
	if strings.TrimSpace(alsdir) == "" {
		dir, err := os.Getwd()
		if err != nil {
			return nil, err
		}
		return resolveProject(config, dir)
	}

	if project, ok := config.Projects[alsdir]; ok {
		return &project, nil
	}

	project := config.ResolveProjectByDir(alsdir)
	if project != nil {
		return project, nil
	}

	return nil, fmt.Errorf("project is not tracked: %s", alsdir)
}

var TEST_RESULT_TEMPLATE = template.Must(template.New("scaffold_test_result").Parse(
	`Project *{{.Alias}}* test config scaffolded

* [{{if .HasUnitTests}}x{{else}} {{end}}] unit tests
* [{{if .HasIntegrationTests}}x{{else}} {{end}}] integration tests
* [{{if .HasComponentTests}}x{{else}} {{end}}] component tests`,
))
