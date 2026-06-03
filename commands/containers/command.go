package containers

import (
	"context"
	"path/filepath"
	"sort"
	"text/template"

	cfg "github.com/andrew-malikov/workspace/config"
	"github.com/andrew-malikov/workspace/view"

	"github.com/urfave/cli/v3"
)

type runningProject struct {
	Alias   string
	Dir     string
	Compose string
}

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:    "containers",
		Aliases: []string{"ctrs"},
		Usage:   "list projects with running docker compose services",
		Action: func(ctx context.Context, cmd *cli.Command) error {
			config, err := cfg.LoadConfig()
			if err != nil {
				return err
			}

			running := make([]runningProject, 0, len(config.Projects))
			for _, project := range config.Projects {
				hasRunning, err := project.HasRunningContainers(ctx)
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

			return view.Render(RESULT_TEMPLATE, view.Args{
				"Projects": running,
			})
		},
	}
}

var RESULT_TEMPLATE = template.Must(template.New("containers_result").Parse(
	`{{if .Projects}}| alias | directory | compose |
| --- | --- | --- |
{{range .Projects}}| {{.Alias}} | {{.Dir}} | {{.Compose}} |
{{end}}{{else}}No running docker compose projects found.
{{end}}`,
))
