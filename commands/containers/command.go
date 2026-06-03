package containers

import (
	"context"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"text/template"

	cfg "github.com/andrew-malikov/workspace/config"
	"github.com/andrew-malikov/workspace/projects"
	"github.com/andrew-malikov/workspace/view"

	"github.com/urfave/cli/v3"
)

type composeRunner func(ctx context.Context, dir string, compose string) ([]byte, error)

type runningProject struct {
	Alias   string
	Dir     string
	Compose string
}

func NewCommand() *cli.Command {
	return newCommand(runDockerCompose)
}

func newCommand(runner composeRunner) *cli.Command {
	return &cli.Command{
		Name:    "containers",
		Aliases: []string{"ctrs"},
		Usage:   "list projects with running docker compose services",
		Action: func(ctx context.Context, cmd *cli.Command) error {
			config, err := cfg.LoadConfig()
			if err != nil {
				return err
			}

			running, err := listRunningProjects(ctx, config.Projects, runner)
			if err != nil {
				return err
			}

			return view.Render(RESULT_TEMPLATE, view.Args{
				"Projects": running,
			})
		},
	}
}

func listRunningProjects(ctx context.Context, tracked map[string]projects.Project, runner composeRunner) ([]runningProject, error) {
	running := make([]runningProject, 0, len(tracked))
	for _, project := range tracked {
		if !project.DoesComposeExist() {
			continue
		}

		out, err := runner(ctx, project.Dir, *project.Compose)
		if err != nil {
			return nil, err
		}

		if !hasRunningContainers(out) {
			continue
		}

		running = append(running, runningProject{
			Alias:   project.Alias,
			Dir:     project.Dir,
			Compose: filepath.Join(project.Dir, *project.Compose),
		})
	}

	sort.Slice(running, func(i, j int) bool {
		return running[i].Alias < running[j].Alias
	})

	return running, nil
}

func runDockerCompose(ctx context.Context, dir string, compose string) ([]byte, error) {
	command := exec.CommandContext(ctx, "docker", "compose", "-f", compose, "ps", "--status", "running", "--format", "json")
	command.Dir = dir
	return command.Output()
}

func hasRunningContainers(out []byte) bool {
	trimmed := strings.TrimSpace(string(out))
	return trimmed != "" && trimmed != "[]"
}

var RESULT_TEMPLATE = template.Must(template.New("containers_result").Parse(
	`{{if .Projects}}| alias | directory | compose |
| --- | --- | --- |
{{range .Projects}}| {{.Alias}} | {{.Dir}} | {{.Compose}} |
{{end}}{{else}}No running docker compose projects found.
{{end}}`,
))
