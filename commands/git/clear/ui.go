package clear

import (
	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/andrew-malikov/workspace/projects"
)

type focusedPanel int

const (
	focusActive focusedPanel = iota
	focusStale
)

var (
	focusedPanelStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("62"))

	blurredPanelStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("240"))
)

type ui struct {
	active list.Model
	stale  list.Model
	focus  focusedPanel
}

func newUi(branches []projects.StaleBranch) ui {
	activeItems := make([]list.Item, 0)
	staleItems := make([]list.Item, 0)

	for _, branch := range branches {
		item := makeBranchItem(branch)
		if branch.IsStale {
			staleItems = append(staleItems, item)
		} else {
			activeItems = append(activeItems, item)
		}
	}

	activeList := list.New(activeItems, list.NewDefaultDelegate(), 0, 0)
	activeList.Title = "Active Branches"

	staleList := list.New(staleItems, list.NewDefaultDelegate(), 0, 0)
	staleList.Title = "Stale Branches"

	return ui{
		active: activeList,
		stale:  staleList,
		focus:  focusActive,
	}
}

func (m ui) Init() tea.Cmd {
	return nil
}

func (m ui) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "tab", "shift+tab":
			if m.focus == focusActive {
				m.focus = focusStale
			} else {
				m.focus = focusActive
			}
			return m, nil
		}

	case tea.WindowSizeMsg:
		fw, fh := focusedPanelStyle.GetFrameSize()
		panelW := msg.Width/2 - fw
		panelH := msg.Height - fh
		m.active.SetSize(panelW, panelH)
		m.stale.SetSize(panelW, panelH)
		return m, nil
	}

	var cmd tea.Cmd
	if m.focus == focusActive {
		m.active, cmd = m.active.Update(msg)
	} else {
		m.stale, cmd = m.stale.Update(msg)
	}
	return m, cmd
}

func (m ui) View() tea.View {
	leftStyle := blurredPanelStyle
	rightStyle := blurredPanelStyle
	if m.focus == focusActive {
		leftStyle = focusedPanelStyle
	} else {
		rightStyle = focusedPanelStyle
	}

	left := leftStyle.Render(m.active.View())
	right := rightStyle.Render(m.stale.View())

	v := tea.NewView(lipgloss.JoinHorizontal(lipgloss.Top, left, right))
	v.AltScreen = true
	return v
}
