package track

import (
	"context"
	"errors"
	"strings"
	"text/template"
	cfg "ws/config"
	"ws/projects"
	"ws/view"

	"github.com/urfave/cli/v3"
)

type input struct {
	alias      string
	dir        string
	compose    string
	migrations string
}

func newInput(cmd *cli.Command) input {
	return input{
		alias:      cmd.StringArg("alias"),
		dir:        cmd.StringArg("dir"),
		compose:    cmd.String("compose"),
		migrations: cmd.String("migrations"),
	}
}

func (input input) Validate() error {
	if strings.TrimSpace(input.alias) == "" {
		return errors.New("alias argument is required")
	}
	return nil
}

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:    "track",
		Aliases: []string{"add", "tr"},
		Usage:   "start tracking directory as a project",
		Arguments: []cli.Argument{
			&cli.StringArg{
				Name:      "alias",
				UsageText: "a shorthand to access the project",
			},
			&cli.StringArg{
				Name:      "dir",
				Value:     "",
				UsageText: "tracking directory",
			},
		},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "compose",
				Aliases: []string{"c"},
				Value:   "docker-compose.yaml",
				Usage:   "filepath to docker-compose.yaml",
			},
			&cli.StringFlag{
				Name:    "migrations",
				Aliases: []string{"m"},
				Usage:   "directory with db migrations",
				Value:   "migrations",
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			input := newInput(cmd)
			err := input.Validate()
			if err != nil {
				return err
			}

			draft := projects.DraftProject{
				Alias:      input.alias,
				Dir:        &input.dir,
				Compose:    &input.compose,
				Migrations: &input.migrations,
			}

			project, err := draft.ToProject()
			if err != nil {
				return err
			}

			config, err := cfg.LoadConfig()
			if err != nil {
				return err
			}

			result, err := config.AddProject(*project)
			if err != nil {
				return err
			}

			err = cfg.SaveConfig(*config)
			if err != nil {
				return err
			}

			return view.Render(RESULT_TEMPLATE, view.Args{
				"Alias":         project.Alias,
				"Dir":           project.Dir,
				"HasCompose":    result.DoesComposeExist,
				"HasMigrations": result.DoMigrationsExist,
			})
		},
	}
}

var RESULT_TEMPLATE = template.Must(template.New("track_result").Parse(
	`Project *{{.Alias}}* is tracked under **{{.Dir}}**

* [{{if .HasCompose}}x{{else}} {{end}}] compose
* [{{if .HasMigrations}}x{{else}} {{end}}] migrations`,
))
