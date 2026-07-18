package up

import (
	"context"
	"os"
	"strings"
	"text/template"

	"github.com/andrew-malikov/workspace/containers"
	"github.com/andrew-malikov/workspace/view"
	"github.com/andrew-malikov/workspace/workspaces"

	"github.com/urfave/cli/v3"
)

func NewCommand(renderer *view.Renderer) *cli.Command {
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
			&cli.BoolFlag{
				Name:    "blank",
				Aliases: []string{"b"},
				Usage:   "cleanup target docker compose volumes before starting",
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

			result, err := workspace.UpProject(ctx, dir, cmd.StringArg("alias"), cmd.Bool("alongside"), cmd.Bool("blank"), containers.DockerCompose{})
			if err != nil {
				return err
			}

			return renderer.Render(RESULT_TEMPLATE, view.Args{
				"Alias":     result.Alias,
				"Alongside": result.Alongside,
				"Blank":     result.Blank,
				"Stopped":   strings.Join(result.Stopped, ", "),
			})
		},
	}
}

var RESULT_TEMPLATE = template.Must(template.New("up_result").Funcs(view.TemplateFuncs).Parse(
	`Project *{{.Alias | literal}}* docker compose is up.
{{if .Blank}}Target docker compose volumes were cleaned up before start.
{{end}}
{{if .Alongside}}Other running docker compose projects left active.
{{else}}{{if .Stopped}}Stopped running docker compose projects: {{.Stopped | literal}}.
{{else}}No other running docker compose projects found.
{{end}}{{end}}`,
))
