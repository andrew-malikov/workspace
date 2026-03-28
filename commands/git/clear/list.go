package clear

import (
	"fmt"
	"io"
	"strings"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

type bubbleItem interface {
	list.Item
	Title() string
	Description() string
	Meta() string
	Commits() []string
}

func (item branchItem) Meta() string      { return item.meta }
func (item branchItem) Commits() []string { return item.commits }

type bubbleDelegate struct {
	styles bubbleStyles
	width  int
	height int
}

type bubbleStyles struct {
	normalBubble   lipgloss.Style
	selectedBubble lipgloss.Style
	normalTitle    lipgloss.Style
	selectedTitle  lipgloss.Style
	normalDesc     lipgloss.Style
	selectedDesc   lipgloss.Style
	normalMeta     lipgloss.Style
	selectedMeta   lipgloss.Style
	match          lipgloss.Style
	normalCommits  lipgloss.Style
}

func newBubbleStyles() bubbleStyles {
	return bubbleStyles{
		normalBubble: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("240")).
			Padding(0, 1).
			MarginBottom(1),
		selectedBubble: lipgloss.NewStyle().
			Border(lipgloss.ThickBorder()).
			BorderForeground(lipgloss.Color("81")).
			Padding(0, 1).
			MarginBottom(1),
		normalTitle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("230")).
			Bold(true),
		selectedTitle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("230")).
			Bold(true),
		normalDesc: lipgloss.NewStyle().
			Foreground(lipgloss.Color("249")),
		selectedDesc: lipgloss.NewStyle().
			Foreground(lipgloss.Color("255")),
		normalMeta: lipgloss.NewStyle().
			Foreground(lipgloss.Color("151")),
		selectedMeta: lipgloss.NewStyle().
			Foreground(lipgloss.Color("121")).
			Bold(true),
		match:         lipgloss.NewStyle().Underline(true),
		normalCommits: lipgloss.NewStyle().Foreground(lipgloss.Color("249")),
	}
}

func newBubbleDelegate() bubbleDelegate {
	return bubbleDelegate{
		styles: newBubbleStyles(),
		height: 6,
		width:  0,
	}
}

func (delegate bubbleDelegate) Height() int {
	return delegate.height
}

func (delegate bubbleDelegate) Spacing() int {
	return 0
}

func (delegate bubbleDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd {
	return nil
}

func (delegate bubbleDelegate) Render(writer io.Writer, model list.Model, index int, listItem list.Item) {
	item, ok := listItem.(bubbleItem)
	if !ok {
		return
	}

	style := delegate.styles.normalBubble
	titleStyle := delegate.styles.normalTitle
	descStyle := delegate.styles.normalDesc
	metaStyle := delegate.styles.normalMeta
	commitsStyle := delegate.styles.normalCommits

	if index == model.Index() && model.FilterState() != list.Filtering {
		style = delegate.styles.selectedBubble
		titleStyle = delegate.styles.selectedTitle
		descStyle = delegate.styles.selectedDesc
		metaStyle = delegate.styles.selectedMeta
	}

	textWidth := model.Width() - style.GetHorizontalFrameSize()
	if textWidth < 10 {
		textWidth = 10
	}

	title := ansi.Truncate(item.Title(), textWidth, "...")
	desc := ansi.Truncate(item.Description(), textWidth, "...")
	meta := ansi.Truncate(item.Meta(), textWidth, "...")
	commitLines := make([]string, 0, len(item.Commits()))
	for _, commit := range item.Commits() {
		commitLines = append(commitLines, commitsStyle.Render(ansi.Truncate(commit, textWidth, "...")))
	}

	if state := model.FilterState(); state == list.Filtering || state == list.FilterApplied {
		matched := model.MatchesForItem(index)
		if len(matched) > 0 {
			plain := titleStyle.Inline(true)
			title = lipgloss.StyleRunes(title, matched, plain.Inherit(delegate.styles.match), plain)
		}
	}

	contentLines := []string{
		titleStyle.Render(title),
		descStyle.Render(desc),
		metaStyle.Render(meta),
	}
	contentLines = append(contentLines, commitLines...)
	content := strings.Join(contentLines, "\n")

	fmt.Fprint(writer, style.Width(model.Width()).Render(content))
}

func (delegate bubbleDelegate) ShortHelp() []key.Binding { return nil }

func (delegate bubbleDelegate) FullHelp() [][]key.Binding { return nil }
