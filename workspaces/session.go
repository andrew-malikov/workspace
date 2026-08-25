package workspaces

import (
	"os"

	"github.com/andrew-malikov/workspace/console"
	"github.com/andrew-malikov/workspace/containers"
)

type Session struct {
	Load    func() (*Workspace, error)
	Save    func(Workspace) error
	Cwd     func() (string, error)
	Compose func(console.Console) containers.Compose
}

func DefaultSession() Session {
	return Session{
		Load: func() (*Workspace, error) {
			path, err := ConfigPath()
			if err != nil {
				return nil, err
			}
			return Load(path)
		},
		Save: func(workspace Workspace) error {
			path, err := ConfigPath()
			if err != nil {
				return err
			}
			return Save(path, workspace)
		},
		Cwd: os.Getwd,
		Compose: func(terminal console.Console) containers.Compose {
			return containers.NewDockerCompose(terminal.Input, terminal.Output, terminal.Error)
		},
	}
}
