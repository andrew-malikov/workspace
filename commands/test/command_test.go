package test

import (
	"reflect"
	"testing"

	"github.com/andrew-malikov/workspace/projects"
)

func TestResolveRequestedKinds(t *testing.T) {
	tests := []struct {
		name        string
		unit        bool
		integration bool
		component   bool
		want        []projects.TestKind
	}{
		{
			name: "none",
			want: []projects.TestKind{},
		},
		{
			name:      "unit and component",
			unit:      true,
			component: true,
			want:      []projects.TestKind{projects.UnitTestKind, projects.ComponentTestKind},
		},
		{
			name:        "all",
			unit:        true,
			integration: true,
			component:   true,
			want:        []projects.TestKind{projects.UnitTestKind, projects.IntegrationTestKind, projects.ComponentTestKind},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := resolveRequestedKinds(tt.unit, tt.integration, tt.component)
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("unexpected kinds: got %v want %v", got, tt.want)
			}
		})
	}
}
