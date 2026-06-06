package untrack

import (
	"context"
	"os"
	"strings"
	"text/template"

	flr "github.com/andrew-malikov/workspace/failure"
	"github.com/andrew-malikov/workspace/view"
	"github.com/andrew-malikov/workspace/workspaces"

	"github.com/urfave/cli/v3"
)

func NewCommand() *cli.Command {
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
				if wd, err := os.Getwd(); err != nil {
					return renderFailure(*flr.OfError(err))
				} else {
					alsdir = wd
				}
			}

			workspace, err := workspaces.LoadWorkspace()
			if err != nil {
				return err
			}

			removedProject, failure := workspace.RemoveProject(alsdir)
			if failure != nil {
				return renderFailure(*failure)
			}

			err = workspaces.SaveWorkspace(*workspace)
			if err != nil {
				return renderFailure(*flr.OfError(err))
			}

			return view.Render(RESULT_TEMPLATE, view.Args{
				"Alias": removedProject.Alias,
				"Dir":   removedProject.Dir,
			})
		},
	}
}

func renderFailure(failure flr.Failure) error {
	if failure.Context == nil {
		return failure.Error
	}
	if ctx, matched := flr.Is[workspaces.NoProjectFound]("NO_PROJECT_FOUND", failure); matched {
		return view.Render(FAILURE_PROJECT_NOT_FOUND, view.Args{
			"Alsdir": ctx.Alsdir,
		})
	}
	return view.RenderUnhandledFailure(*failure.Context)
}

var FAILURE_PROJECT_NOT_FOUND = template.Must(template.New("untrack_project_not_found").Parse(
	`No project is found by **{{.Alsdir}}**`,
))

var RESULT_TEMPLATE = template.Must(template.New("untrack_result").Parse(
	`Project *{{.Alias}}* under **{{.Dir}}** is no longer tracked`,
))
