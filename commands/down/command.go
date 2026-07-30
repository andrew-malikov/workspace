package down

import (
	"context"
	"os"
	"text/template"

	"github.com/andrew-malikov/workspace/console"
	"github.com/andrew-malikov/workspace/containers"
	"github.com/andrew-malikov/workspace/view"
	"github.com/andrew-malikov/workspace/workspaces"

	"github.com/urfave/cli/v3"
)

func NewCommand(terminal console.Console, renderer *view.Renderer) *cli.Command {
	return &cli.Command{
		Name:  "down",
		Usage: "stop project docker compose",
		Arguments: []cli.Argument{
			&cli.StringArg{
				Name:      "alias",
				UsageText: "tracked project alias",
			},
		},
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "blank",
				Aliases: []string{"b"},
				Usage:   "cleanup target docker compose volumes while stopping",
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			workspace, err := workspaces.LoadWorkspace()
			if err != nil {
				return err
			}

			dir, err := os.Getwd()
			if err != nil {
				return err
			}

			blank := cmd.Bool("blank")
			compose := containers.NewDockerCompose(terminal.Input, terminal.Output, terminal.Error)
			project, err := workspace.DownProject(ctx, dir, cmd.StringArg("alias"), blank, compose)
			if err != nil {
				return err
			}

			return renderer.Render(RESULT_TEMPLATE, view.Args{
				"Alias": project.Alias,
				"Blank": blank,
			})
		},
	}
}

var RESULT_TEMPLATE = template.Must(template.New("down_result").Funcs(view.TemplateFuncs).Parse(
	`Project *{{.Alias | literal}}* docker compose is down.
{{if .Blank}}Target docker compose volumes were cleaned up.
{{end}}`,
))
