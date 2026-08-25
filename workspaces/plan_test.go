package workspaces

import (
	"slices"
	"testing"
)

func TestPlanUp(t *testing.T) {
	target := "orders"
	runningOthers := []OtherProject{
		{Alias: "billing", Configured: true, Running: true},
		{Alias: "api", Configured: true, Running: true},
		{Alias: "worker", Configured: true, Running: false},
	}

	tests := []struct {
		name       string
		others     []OtherProject
		alongside  bool
		blank      bool
		want       []Action
	}{
		{
			name:   "exclusive up stops other running projects",
			others: runningOthers,
			want: []Action{
				{Kind: ActionStop, Alias: "api"},
				{Kind: ActionStop, Alias: "billing"},
				{Kind: ActionStart, Alias: "orders"},
			},
		},
		{
			name:      "alongside up ignores other projects",
			others:    runningOthers,
			alongside: true,
			want:      []Action{{Kind: ActionStart, Alias: "orders"}},
		},
		{
			name:   "blank exclusive up cleans the target after stops",
			others: []OtherProject{{Alias: "api", Configured: true, Running: true}},
			blank:  true,
			want: []Action{
				{Kind: ActionStop, Alias: "api"},
				{Kind: ActionCleanup, Alias: "orders"},
				{Kind: ActionStart, Alias: "orders"},
			},
		},
		{
			name:      "blank alongside up skips other projects",
			others:    runningOthers,
			alongside: true,
			blank:     true,
			want: []Action{
				{Kind: ActionCleanup, Alias: "orders"},
				{Kind: ActionStart, Alias: "orders"},
			},
		},
		{
			name: "unconfigured others are omitted",
			others: []OtherProject{
				{Alias: "docs", Configured: false, Running: true},
				{Alias: "api", Configured: true, Running: true},
			},
			want: []Action{
				{Kind: ActionStop, Alias: "api"},
				{Kind: ActionStart, Alias: "orders"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := PlanUp(target, tt.others, tt.alongside, tt.blank)
			if !slices.Equal(got, tt.want) {
				t.Fatalf("unexpected plan: got %#v want %#v", got, tt.want)
			}
		})
	}
}

func TestPlanDown(t *testing.T) {
	tests := []struct {
		name  string
		blank bool
		want  []Action
	}{
		{name: "default down", want: []Action{{Kind: ActionStop, Alias: "orders"}}},
		{name: "blank down", blank: true, want: []Action{{Kind: ActionCleanup, Alias: "orders"}}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := PlanDown("orders", tt.blank)
			if !slices.Equal(got, tt.want) {
				t.Fatalf("unexpected plan: got %#v want %#v", got, tt.want)
			}
		})
	}
}
