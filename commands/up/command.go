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
			workspace, err := workspaces.LoadWorkspace()
			if err != nil {
				return err
			}

			dir, err := os.Getwd()
			if err != nil {
				return err
			}

			result, err := workspace.UpProject(ctx, dir, cmd.StringArg("alias"), cmd.Bool("alongside"), containers.DockerCompose{})
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

var RESULT_TEMPLATE = template.Must(template.New("up_result").Parse(
	`Project *{{.Alias}}* docker compose is up.
{{if .Alongside}}Other running docker compose projects left active.
{{else}}{{if .Stopped}}Stopped running docker compose projects: {{.Stopped}}.
{{else}}No other running docker compose projects found.
{{end}}{{end}}`,
))
