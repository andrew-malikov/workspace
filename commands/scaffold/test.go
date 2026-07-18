package scaffold

import (
	"context"
	"fmt"
	"os"
	"strings"
	"text/template"

	"github.com/andrew-malikov/workspace/projects"
	"github.com/andrew-malikov/workspace/view"
	"github.com/andrew-malikov/workspace/workspaces"

	"github.com/urfave/cli/v3"
)

func newTestCommand(renderer *view.Renderer) *cli.Command {
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
			workspace, err := workspaces.LoadWorkspace()
			if err != nil {
				return err
			}

			project, err := resolveProject(workspace, cmd.StringArg("directory_or_alias"))
			if err != nil {
				return err
			}

			if err := workspace.ScaffoldProjectTests(project); err != nil {
				return err
			}

			workspace.Projects[project.Alias] = *project
			if err := workspaces.SaveWorkspace(*workspace); err != nil {
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
	if strings.TrimSpace(alsdir) == "" {
		dir, err := os.Getwd()
		if err != nil {
			return nil, err
		}
		return resolveProject(workspace, dir)
	}

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
