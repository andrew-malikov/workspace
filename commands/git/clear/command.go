package clear

import (
	"context"
	"fmt"
	"os"

	cfg "github.com/andrew-malikov/workspace/config"
	"github.com/andrew-malikov/workspace/projects"
	"github.com/andrew-malikov/workspace/vcs"
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

type branchItem struct {
	title, desc string
}

func (i branchItem) Title() string       { return i.title }
func (i branchItem) Description() string { return i.desc }
func (i branchItem) FilterValue() string { return i.title }

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
	items := make([]list.Item, 0)
	for _, branch := range branches {
		stale := "[ ]"
		if branch.IsStale {
			stale = "[x]"
		}
		title := fmt.Sprintf("%s %s", stale, branch.Name)
		if branch.Remote != nil {
			title = fmt.Sprintf("%s at %s", title, *branch.Remote)
		} else if branch.Related != nil {
			status := ""
			switch branch.Related.Status {
			case vcs.AheadBranch:
				status = "ahead"
			case vcs.BehindBranch:
				status = "behind"
			case vcs.DivergedBranch:
				status = "diverged"
			case vcs.SyncedBranch:
				status = "synced"
			}
			title = fmt.Sprintf("%s is %s with %s", title, status, *branch.Related.Remote)
		} else {
			title = fmt.Sprintf("%s is local only", title)
		}
		items = append(items, branchItem{
			title: title,
			desc:  fmt.Sprintf("%s at %s", branch.Author, branch.UpdatedAt.Format("2 Jan 2006, 3:04 PM")),
		})
	}
	m := ui{list: list.New(items, list.NewDefaultDelegate(), 0, 0)}
	m.list.Title = "Stale Branches"
	return m
}
