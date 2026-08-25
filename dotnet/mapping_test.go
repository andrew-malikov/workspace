package dotnet

import (
	"errors"
	"slices"
	"strings"
	"testing"

	"github.com/andrew-malikov/workspace/projects"
)

func TestTestArgs(t *testing.T) {
	tests := []struct {
		name       string
		kind       projects.TestKind
		filter     string
		resultsDir string
		want       []string
	}{
		{
			name:       "unfiltered category",
			kind:       projects.UnitTestKind,
			resultsDir: "/tmp/results",
			want: []string{
				"test",
				"--logger", "trx;LogFilePrefix=unit",
				"--results-directory", "/tmp/results",
			},
		},
		{
			name:       "filtered category",
			kind:       projects.UnitTestKind,
			filter:     "FullyQualifiedName~UnitTests",
			resultsDir: "/tmp/results",
			want: []string{
				"test",
				"--logger", "trx;LogFilePrefix=unit",
				"--results-directory", "/tmp/results",
				"--filter", "FullyQualifiedName~UnitTests",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := TestArgs(tt.kind, tt.filter, tt.resultsDir)
			if !slices.Equal(got, tt.want) {
				t.Fatalf("unexpected args: got %q want %q", got, tt.want)
			}
		})
	}
}

func TestCategoryReport(t *testing.T) {
	t.Run("summary present", func(t *testing.T) {
		got := CategoryReport(projects.UnitTestKind, TestSummary{Total: 4, Passed: 2, Failed: 1, Skipped: 1}, nil, "/tmp/unit.log")
		if !strings.Contains(got, "unit summary: total 4, passed 2, failed 1, skipped 1") {
			t.Fatalf("missing summary: %q", got)
		}
		if !strings.Contains(got, "Log: /tmp/unit.log") {
			t.Fatalf("missing log path: %q", got)
		}
	})

	t.Run("summary missing", func(t *testing.T) {
		got := CategoryReport(projects.ComponentTestKind, TestSummary{}, errors.New("no TRX"), "/tmp/component.log")
		if !strings.Contains(got, "component summary unavailable:") {
			t.Fatalf("missing unavailable summary: %q", got)
		}
		if !strings.Contains(got, "Log: /tmp/component.log") {
			t.Fatalf("missing log path: %q", got)
		}
	})
}
