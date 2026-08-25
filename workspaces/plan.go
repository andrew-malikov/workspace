package workspaces

import "sort"


type ActionKind string

const (
	ActionStop    ActionKind = "stop"
	ActionCleanup ActionKind = "cleanup"
	ActionStart   ActionKind = "start"
)

type Action struct {
	Kind  ActionKind
	Alias string
}

type OtherProject struct {
	Alias      string
	Configured bool
	Running    bool
}

func PlanUp(targetAlias string, others []OtherProject, alongside bool, blank bool) []Action {
	plan := make([]Action, 0, len(others)+2)
	if !alongside {
		ordered := append([]OtherProject(nil), others...)
		sort.Slice(ordered, func(i, j int) bool {
			return ordered[i].Alias < ordered[j].Alias
		})
		for _, other := range ordered {
			if !other.Configured || !other.Running {
				continue
			}
			plan = append(plan, Action{Kind: ActionStop, Alias: other.Alias})
		}
	}
	if blank {
		plan = append(plan, Action{Kind: ActionCleanup, Alias: targetAlias})
	}
	return append(plan, Action{Kind: ActionStart, Alias: targetAlias})
}

func PlanDown(targetAlias string, blank bool) []Action {
	if blank {
		return []Action{{Kind: ActionCleanup, Alias: targetAlias}}
	}
	return []Action{{Kind: ActionStop, Alias: targetAlias}}
}

func stoppedAliases(plan []Action) []string {
	stopped := make([]string, 0)
	for _, action := range plan {
		if action.Kind == ActionStop {
			stopped = append(stopped, action.Alias)
		}
	}
	return stopped
}

