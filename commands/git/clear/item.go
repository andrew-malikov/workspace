package clear

import (
	"fmt"
	"strings"

	"charm.land/bubbles/v2/list"
	"github.com/andrew-malikov/workspace/projects"
)

type branchItem struct {
	branch projects.StaleBranch
	title  string
	desc   string
	meta   string
}

func (item branchItem) Title() string       { return item.title }
func (item branchItem) Description() string { return item.desc }
func (item branchItem) FilterValue() string {
	return strings.Join([]string{item.title, item.meta}, " ")
}

func makeBranchItem(branch projects.StaleBranch) branchItem {
	locations := make([]string, 0, 2)
	if branch.HasLocal() {
		locations = append(locations, "local")
	}
	if remote, ok := branch.ResolveRemote(); ok {
		locations = append(locations, remote)
	}

	desc := fmt.Sprintf("%s updated %s", branch.Author, branch.UpdatedAt.Format("2 Jan 2006, 3:04 PM"))
	meta := strings.Join(locations, "  ")
	if meta == "" {
		meta = "detached"
	}

	return branchItem{
		branch: branch,
		title:  branch.Name,
		desc:   desc,
		meta:   meta,
	}
}

func toListItems(branches []projects.StaleBranch) []list.Item {
	items := make([]list.Item, 0, len(branches))
	for _, branch := range branches {
		items = append(items, makeBranchItem(branch))
	}
	return items
}
