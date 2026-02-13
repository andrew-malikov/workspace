package clear

import (
	"context"
	"os"
	"text/template"

	cfg "github.com/andrew-malikov/workspace/config"
	"github.com/andrew-malikov/workspace/view"

	"github.com/urfave/cli/v3"
)

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:    "clear",
		Aliases: []string{"clr", "c"},
		Usage:   "clear hanging branches",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "dry-run",
				Aliases: []string{"d"},
				Value:   false,
				Usage:   "only prints out the plan",
			},
			&cli.BoolFlag{
				Name:    "team",
				Aliases: []string{"t"},
				Value:   false,
				Usage:   "include team members",
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			config, err := cfg.LoadConfig()
			if err != nil {
				return err
			}

			dir, err := os.Getwd()
			if err != nil {
				return err
			}

			project := config.ResolveProjectByDir(dir)

			if project == nil {
				return view.RenderDirectoryIsNotTrackedYet(dir)
			}

			branches, err := project.ListStaleBranches()
			if err != nil {
				return err
			}

			return view.Render(RESULT_TEMPLATE, view.Args{
				"Branches": branches,
			})
		},
	}
}

var RESULT_TEMPLATE = template.Must(template.New("clear_result").Parse(
	`| [ ] | branch           | author | updated at          | status |
| --- | ---------------- | ------ | ------------------- | ------ |
{{range .Branches}}| [{{if .IsStale}}x{{else}} {{end}}] | {{.Name}} | {{.Author}} | {{.UpdatedAt.Format "2 Jan 2006, 3:04 PM"}} | {{.Status}} |{{printf "\n"}}{{end}}`,
))
