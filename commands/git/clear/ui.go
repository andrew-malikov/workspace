package clear

import (
	"fmt"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/andrew-malikov/workspace/projects"
	"github.com/andrew-malikov/workspace/vcs"
)

var (
	appStyle         = lipgloss.NewStyle().Padding(1, 2)
	titleStyle       = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("230"))
	subtitleStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("246"))
	statusInfoStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("151"))
	statusErrorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("203"))
	emptyStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("246")).Italic(true)
)

type deleteBranchMsg struct {
	branch projects.StaleBranch
	err    error
}

type uiKeyMap struct {
	Delete        key.Binding
	Filter        key.Binding
	ClearFilter   key.Binding
	CursorUp      key.Binding
	CursorDown    key.Binding
	NextPage      key.Binding
	PrevPage      key.Binding
	Quit          key.Binding
	ShowFullHelp  key.Binding
	CloseFullHelp key.Binding
}

func (keys uiKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{keys.CursorUp, keys.CursorDown, keys.Delete, keys.Filter, keys.Quit, keys.ShowFullHelp}
}

func (keys uiKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{keys.CursorUp, keys.CursorDown, keys.NextPage, keys.PrevPage},
		{keys.Delete, keys.Filter, keys.ClearFilter},
		{keys.Quit, keys.ShowFullHelp, keys.CloseFullHelp},
	}
}

func defaultUiKeyMap() uiKeyMap {
	listKeys := list.DefaultKeyMap()
	keys := uiKeyMap{
		Delete: key.NewBinding(
			key.WithKeys("d"),
			key.WithHelp("d", "delete branch"),
		),
		Filter:        listKeys.Filter,
		ClearFilter:   listKeys.ClearFilter,
		CursorUp:      listKeys.CursorUp,
		CursorDown:    listKeys.CursorDown,
		NextPage:      listKeys.NextPage,
		PrevPage:      listKeys.PrevPage,
		Quit:          listKeys.Quit,
		ShowFullHelp:  key.NewBinding(key.WithKeys("?"), key.WithHelp("?", "more")),
		CloseFullHelp: key.NewBinding(key.WithKeys("?"), key.WithHelp("?", "less")),
	}
	keys.CloseFullHelp.SetEnabled(false)
	return keys
}

type ui struct {
	branches  list.Model
	git       *vcs.ProjectGit
	help      help.Model
	keys      uiKeyMap
	width     int
	height    int
	status    string
	statusErr bool
}

func newUi(branches []projects.StaleBranch, projectGit *vcs.ProjectGit) ui {
	branchList := list.New(toListItems(branches), newBubbleDelegate(), 0, 0)
	branchList.Title = "Your git branches"
	branchList.SetShowStatusBar(false)
	branchList.SetFilteringEnabled(true)
	branchList.SetShowHelp(false)
	branchList.SetShowTitle(false)
	branchList.DisableQuitKeybindings()

	helpModel := help.New()
	helpModel.ShowAll = false

	status := "Press d to delete the selected branch locally and remotely."
	if len(branches) == 0 {
		status = "No branches found for the current git user."
	}

	return ui{
		branches: branchList,
		git:      projectGit,
		help:     helpModel,
		keys:     defaultUiKeyMap(),
		status:   status,
	}
}

func (model ui) Init() tea.Cmd { return nil }

func (model ui) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch typed := msg.(type) {
	case tea.WindowSizeMsg:
		model.width = typed.Width
		model.height = typed.Height
		model.resize()
		return model, nil

	case tea.KeyPressMsg:
		switch typed.String() {
		case "ctrl+c", "q":
			return model, tea.Quit
		case "?":
			model.help.ShowAll = !model.help.ShowAll
			model.keys.ShowFullHelp.SetEnabled(!model.help.ShowAll)
			model.keys.CloseFullHelp.SetEnabled(model.help.ShowAll)
			model.resize()
			return model, nil
		case "d":
			if model.branches.SettingFilter() {
				return model, nil
			}
			item, ok := model.branches.SelectedItem().(branchItem)
			if !ok {
				return model, nil
			}
			return model, func() tea.Msg {
				return deleteBranchMsg{branch: item.branch, err: model.git.DeleteBranch(item.branch.Branch)}
			}
		}

	case deleteBranchMsg:
		if typed.err != nil {
			model.status = typed.err.Error()
			model.statusErr = true
			return model, nil
		}

		items := model.branches.Items()
		for index, raw := range items {
			item, ok := raw.(branchItem)
			if !ok {
				continue
			}
			if item.branch.Name == typed.branch.Name {
				model.branches.RemoveItem(index)
				break
			}
		}

		model.status = deletedBranchMessage(typed.branch.Branch)
		model.statusErr = false
		return model, nil
	}

	var cmd tea.Cmd
	model.branches, cmd = model.branches.Update(msg)
	return model, cmd
}

func (model ui) View() tea.View {
	header := titleStyle.Render("ws git clear")
	subtitle := subtitleStyle.Render("Branches owned by the current git user from branch history")

	statusStyle := statusInfoStyle
	if model.statusErr {
		statusStyle = statusErrorStyle
	}

	body := model.branches.View()
	if len(model.branches.Items()) == 0 && !model.branches.SettingFilter() {
		body = emptyStyle.Render("Nothing to clear.")
	}

	helpModel := model.help
	helpModel.SetWidth(model.width)
	helpBar := helpModel.View(model.keys)

	content := lipgloss.JoinVertical(lipgloss.Left,
		header,
		subtitle,
		statusStyle.Render(model.status),
		"",
		body,
		helpBar,
	)

	v := tea.NewView(appStyle.Render(content))
	v.AltScreen = true
	return v
}

func (model *ui) resize() {
	if model.width <= 0 || model.height <= 0 {
		return
	}

	helpModel := model.help
	helpModel.SetWidth(model.width)
	helpHeight := lipgloss.Height(helpModel.View(model.keys))
	frameWidth, frameHeight := appStyle.GetFrameSize()
	listWidth := model.width - frameWidth
	listHeight := model.height - frameHeight - helpHeight - 5
	if listWidth < 20 {
		listWidth = 20
	}
	if listHeight < 5 {
		listHeight = 5
	}
	model.branches.SetSize(listWidth, listHeight)
}

func deletedBranchMessage(branch vcs.Branch) string {
	parts := make([]string, 0, 2)
	if branch.HasLocal() {
		parts = append(parts, "local")
	}
	if remote, ok := branch.ResolveRemote(); ok {
		parts = append(parts, fmt.Sprintf("remote %s", remote))
	}
	if len(parts) == 0 {
		return fmt.Sprintf("Deleted %s.", branch.Name)
	}
	return fmt.Sprintf("Deleted %s from %s.", branch.Name, joinWithAnd(parts))
}

func joinWithAnd(parts []string) string {
	if len(parts) == 1 {
		return parts[0]
	}
	return parts[0] + " and " + parts[1]
}
