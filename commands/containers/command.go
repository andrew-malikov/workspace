package containers

import (
	"context"
	"path/filepath"
	"sort"
	"text/template"

	"github.com/andrew-malikov/workspace/console"
	"github.com/andrew-malikov/workspace/view"
	"github.com/andrew-malikov/workspace/workspaces"

	"github.com/urfave/cli/v3"
)


type runningProject struct {
	Alias   string
	Dir     string
	Compose string
}

func NewCommand(terminal console.Console, renderer *view.Renderer, session workspaces.Session) *cli.Command {
	return &cli.Command{
		Name:    "containers",
		Aliases: []string{"ctrs"},
		Usage:   "list projects with running docker compose services",
		Action: func(ctx context.Context, cmd *cli.Command) error {
			workspace, err := session.Load()
			if err != nil {
				return err
			}

			compose := session.Compose(terminal)

			running := make([]runningProject, 0, len(workspace.Projects))
			for _, project := range workspace.Projects {
				hasRunning, err := project.HasRunningContainers(ctx, compose)
				if err != nil {
					return err
				}

				if hasRunning {
					running = append(running, runningProject{
						Alias:   project.Alias,
						Dir:     project.Dir,
						Compose: filepath.Join(project.Dir, *project.Compose),
					})
				}
			}

			sort.Slice(running, func(i, j int) bool {
				return running[i].Alias < running[j].Alias
			})

			return renderer.Render(RESULT_TEMPLATE, view.Args{
				"Projects": running,
			})
		},
	}
}

var RESULT_TEMPLATE = template.Must(template.New("containers_result").Funcs(view.TemplateFuncs).Parse(
	`{{if .Projects}}| alias | directory | compose |
| --- | --- | --- |
{{range .Projects}}| {{.Alias | tableCell}} | {{.Dir | tableCell}} | {{.Compose | tableCell}} |
{{end}}{{else}}No running docker compose projects found.
{{end}}`,
))
