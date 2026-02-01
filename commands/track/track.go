package track

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strings"
	"text/template"
	cfg "ws/config"
	"ws/projects"

	"github.com/charmbracelet/glamour"
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

			return output(*project, *result)
		},
	}
}

var TEMPLATE = template.Must(template.New("track_result").Parse(
	`Project *{{.Alias}}* is tracked under **{{.Dir}}**

* [{{if .HasCompose}}x{{else}} {{end}}] compose
* [{{if .HasMigrations}}x{{else}} {{end}}] migrations`,
))

func output(project projects.Project, result cfg.AddProjectResult) error {
	return render(TEMPLATE, map[string]any{
		"Alias":         project.Alias,
		"Dir":           project.Dir,
		"HasCompose":    result.DoesComposeExist,
		"HasMigrations": result.DoMigrationsExist,
	})
}

// todo: move out into a view package
func render(tmpl *template.Template, data map[string]any) error {
	var buf bytes.Buffer
	tmpl.Execute(&buf, data)

	out, err := glamour.Render(buf.String(), "dark")
	if err != nil {
		return err
	}

	fmt.Print(out)
	return nil
}
