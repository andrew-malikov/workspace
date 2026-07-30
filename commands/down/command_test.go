package down

import (
	"bytes"
	"strings"
	"testing"

	"github.com/andrew-malikov/workspace/view"
)

func TestResultTemplateReportsBlankCleanupOnlyWhenRequested(t *testing.T) {
	tests := []struct {
		name        string
		blank       bool
		wantCleanup bool
	}{
		{name: "default"},
		{name: "blank", blank: true, wantCleanup: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var output bytes.Buffer
			renderer, err := view.NewRenderer(&output, false)
			if err != nil {
				t.Fatalf("create renderer: %v", err)
			}

			if err := renderer.Render(RESULT_TEMPLATE, view.Args{"Alias": "orders", "Blank": tt.blank}); err != nil {
				t.Fatalf("render result: %v", err)
			}
			if !strings.Contains(output.String(), "docker compose is down.") {
				t.Fatalf("missing down result: %q", output.String())
			}
			hasCleanup := strings.Contains(output.String(), "volumes were cleaned up")
			if hasCleanup != tt.wantCleanup {
				t.Fatalf("unexpected cleanup result: %q", output.String())
			}
		})
	}
}
