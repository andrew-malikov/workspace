package untrack

import (
	"context"
	"strings"
	"text/template"

	"github.com/andrew-malikov/workspace/view"
	"github.com/andrew-malikov/workspace/workspaces"

	"github.com/urfave/cli/v3"
)


func NewCommand(renderer *view.Renderer, session workspaces.Session) *cli.Command {
	return &cli.Command{
		Name:    "untrack",
		Aliases: []string{"remove", "untr"},
		Usage:   "stop tracking project",
		Arguments: []cli.Argument{
			&cli.StringArg{
				Name:      "alsdir",
				UsageText: "an alias or directory",
			},
		},
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "all",
				Aliases: []string{"a"},
				Value:   false,
				Usage:   "allogithub.com/andrew-malikov/workspace to remove all the matches projects",
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			alsdir := cmd.StringArg("alsdir")
			if strings.TrimSpace(alsdir) == "" {
				wd, err := session.Cwd()
				if err != nil {
					return err
				}
				alsdir = wd
			}

			workspace, err := session.Load()
			if err != nil {
				return err
			}

			removedProject, err := workspace.RemoveProject(alsdir)
			if err != nil {
				return err
			}

			if err := session.Save(*workspace); err != nil {
				return err
			}


			return renderer.Render(RESULT_TEMPLATE, view.Args{
				"Alias": removedProject.Alias,
				"Dir":   removedProject.Dir,
			})
		},
	}
}

var RESULT_TEMPLATE = template.Must(template.New("untrack_result").Funcs(view.TemplateFuncs).Parse(
	`Project *{{.Alias | literal}}* under **{{.Dir | literal}}** is no longer tracked`,
))
