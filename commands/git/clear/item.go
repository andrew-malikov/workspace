package clear

import (
	"fmt"

	"github.com/andrew-malikov/workspace/projects"
	"github.com/andrew-malikov/workspace/vcs"
)

type branchItem struct {
	title, desc string
}

func (i branchItem) Title() string       { return i.title }
func (i branchItem) Description() string { return i.desc }
func (i branchItem) FilterValue() string { return i.title }

func makeBranchItem(branch projects.StaleBranch) branchItem {
	title := branch.Name

	if branch.Remote != nil {
		title = fmt.Sprintf("%s at %s", title, *branch.Remote)
	} else if branch.Related != nil {
		status := ""
		switch branch.Related.Status {
		case vcs.AheadBranch:
			status = "ahead of"
		case vcs.BehindBranch:
			status = "behind"
		case vcs.DivergedBranch:
			status = "diverged from"
		case vcs.SyncedBranch:
			status = "synced with"
		}
		remote := ""
		if branch.Related.Remote != nil {
			remote = *branch.Related.Remote + "/"
		}
		title = fmt.Sprintf("%s %s %s%s", title, status, remote, branch.Related.Name)
	} else {
		title = fmt.Sprintf("%s local only", title)
	}

	desc := fmt.Sprintf("%s %s", branch.Author, branch.UpdatedAt.Format("2 Jan 2006, 3:04 PM"))
	return branchItem{title: title, desc: desc}
}
