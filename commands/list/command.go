package list

import (
	"context"
	"sort"
	"strings"
	"text/template"

	"github.com/andrew-malikov/workspace/projects"
	"github.com/andrew-malikov/workspace/view"
	"github.com/andrew-malikov/workspace/workspaces"

	"github.com/urfave/cli/v3"
)

type listedProject struct {
	Name      string
	Dir       string
	Enabled   string
	sortOrder string
}

func NewCommand(renderer *view.Renderer) *cli.Command {
	return &cli.Command{
		Name:    "list",
		Aliases: []string{"ls"},
		Usage:   "list tracked projects",
		Action: func(ctx context.Context, cmd *cli.Command) error {
			workspace, err := workspaces.LoadWorkspace()
			if err != nil {
				return err
			}

			return renderer.Render(RESULT_TEMPLATE, view.Args{
				"Projects": buildListedProjects(workspace.Projects),
			})
		},
	}
}

func buildListedProjects(projectsByAlias map[string]projects.Project) []listedProject {
	listed := make([]listedProject, 0, len(projectsByAlias))
	for _, project := range projectsByAlias {
		listed = append(listed, listedProject{
			Name:      project.Alias,
			Dir:       project.Dir,
			Enabled:   enabledSummary(project),
			sortOrder: project.Alias,
		})
	}

	sort.Slice(listed, func(i, j int) bool {
		return listed[i].sortOrder < listed[j].sortOrder
	})

	return listed
}

func enabledSummary(project projects.Project) string {
	enabled := make([]string, 0, 5)
	if project.DoesComposeExist() {
		enabled = append(enabled, "docker")
	}
	if project.DoMigrationsExist() {
		enabled = append(enabled, "migrations")
	}
	if project.Test.Unit.IsConfigured() {
		enabled = append(enabled, "unit tests")
	}
	if project.Test.Integration.IsConfigured() {
		enabled = append(enabled, "integration tests")
	}
	if project.Test.Component.IsConfigured() {
		enabled = append(enabled, "component tests")
	}
	if len(enabled) == 0 {
		return "none"
	}
	return strings.Join(enabled, ", ")
}

var RESULT_TEMPLATE = template.Must(template.New("list_result").Funcs(view.TemplateFuncs).Parse(
	`{{if .Projects}}| name | directory | enabled |
| --- | --- | --- |
{{range .Projects}}| {{.Name | tableCell}} | {{.Dir | tableCell}} | {{.Enabled | tableCell}} |
{{end}}{{else}}No tracked projects found.
{{end}}`,
))
