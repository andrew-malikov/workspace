package track

import (
	"context"
	"errors"
	"fmt"
	"strings"
	cfg "ws/config"
	"ws/projects"

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
		Name:  "track",
		Usage: "start tracking directory",
		Arguments: []cli.Argument{
			&cli.StringArg{
				Name:      "alias",
				UsageText: "a shorthand to access the project",
			},
			&cli.StringArg{
				Name:      "dir",
				Value:     ".",
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

			// todo: define a draft project to convert into a valid project
			// so the path validation happens before the config loading

			config, err := cfg.LoadConfig()
			if err != nil {
				return err
			}

			project := projects.Project{
				Alias:      input.alias,
				Dir:        input.dir,
				Compose:    &input.compose,
				Migrations: &input.migrations,
			}

			result, err := config.AddProject(project)
			if err != nil {
				return err
			}

			err = cfg.SaveConfig(*config)
			if err != nil {
				return err
			}

			fmt.Printf("Project %s has been successfully added to the tracker at %s", project.Alias, project.Dir)
			fmt.Printf("Compose found at %s : %v", *project.Compose, result.DoesComposeExist)
			fmt.Printf("Migrations found at %s : %v", *project.Migrations, result.DoesComposeExist)

			return nil
		},
	}
}
