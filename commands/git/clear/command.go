package clear

import (
	"context"
	"fmt"
	"os"

	cfg "github.com/andrew-malikov/workspace/config"
	"github.com/andrew-malikov/workspace/projects"
	"github.com/andrew-malikov/workspace/view"

	"github.com/urfave/cli/v3"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:    "clear",
		Aliases: []string{"clr", "c"},
		Usage:   "clear hanging branches",
		// todo: probably outdated by tui
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

			// todo: use the wd as a draft project instead of exiting
			if project == nil {
				return view.RenderDirectoryIsNotTrackedYet(dir)
			}

			branches, err := project.ListStaleBranches(false)
			if err != nil {
				return err
			}

			if _, err := tea.NewProgram(newUi(branches), tea.WithAltScreen()).Run(); err != nil {
				return err
			}

			return nil
		},
	}
}

var docStyle = lipgloss.NewStyle().Margin(1, 2)

type branch struct {
	title, desc string
}

func (i branch) Title() string       { return i.title }
func (i branch) Description() string { return i.desc }
func (i branch) FilterValue() string { return i.title }

type ui struct {
	list list.Model
}

func (model ui) Init() tea.Cmd {
	return nil
}

func (m ui) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v)
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m ui) View() string {
	return docStyle.Render(m.list.View())
}

func newUi(branches []projects.StaleBranch) ui {
	items := make([]list.Item, len(branches))
	for i := range branches {
		stale := "[ ]"
		if branches[i].IsStale {
			stale = "[x]"
		}
		items[i] = branch{
			title: fmt.Sprintf("%s %s is %s", stale, branches[i].Name, branches[i].Status),
			desc:  fmt.Sprintf("%s at %s", branches[i].Author, branches[i].UpdatedAt.Format("2 Jan 2006, 3:04 PM")),
		}
	}
	m := ui{list: list.New(items, list.NewDefaultDelegate(), 0, 0)}
	m.list.Title = "Stale Branches"
	return m
}
