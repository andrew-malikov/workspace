package scaffold

import (
	"context"
	"fmt"
	"strings"
	"text/template"

	"github.com/andrew-malikov/workspace/projects"
	"github.com/andrew-malikov/workspace/view"
	"github.com/andrew-malikov/workspace/workspaces"

	"github.com/urfave/cli/v3"
)


func newTestCommand(renderer *view.Renderer, session workspaces.Session) *cli.Command {
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
			workspace, err := session.Load()
			if err != nil {
				return err
			}

			alsdir := cmd.StringArg("directory_or_alias")
			if strings.TrimSpace(alsdir) == "" {
				dir, err := session.Cwd()
				if err != nil {
					return err
				}
				alsdir = dir
			}

			project, err := resolveProject(workspace, alsdir)
			if err != nil {
				return err
			}

			if err := workspace.ScaffoldProjectTests(project); err != nil {
				return err
			}

			workspace.Projects[project.Alias] = *project
			if err := session.Save(*workspace); err != nil {
				return err
			}


			return renderer.Render(TEST_RESULT_TEMPLATE, view.Args{
				"Alias":               project.Alias,
				"HasUnitTests":        project.Test.Unit.IsConfigured(),
				"HasIntegrationTests": project.Test.Integration.IsConfigured(),
				"HasComponentTests":   project.Test.Component.IsConfigured(),
			})
		},
	}
}

func resolveProject(workspace *workspaces.Workspace, alsdir string) (*projects.Project, error) {
	if project, ok := workspace.Projects[alsdir]; ok {
		return &project, nil
	}


	project := workspace.ResolveProjectByDir(alsdir)
	if project != nil {
		return project, nil
	}

	return nil, fmt.Errorf("project is not tracked: %s", alsdir)
}

var TEST_RESULT_TEMPLATE = template.Must(template.New("scaffold_test_result").Funcs(view.TemplateFuncs).Parse(
	`Project *{{.Alias | literal}}* test config scaffolded

* [{{if .HasUnitTests}}x{{else}} {{end}}] unit tests
* [{{if .HasIntegrationTests}}x{{else}} {{end}}] integration tests
* [{{if .HasComponentTests}}x{{else}} {{end}}] component tests`,
))
