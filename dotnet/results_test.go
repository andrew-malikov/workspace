package dotnet

import (
	"os"
	"path/filepath"
	"strconv"
	"testing"
)

func TestLoadTestSummaryAggregatesTRXFiles(t *testing.T) {
	resultsDir := t.TempDir()
	writeTRX(t, filepath.Join(resultsDir, "first.trx"), 4, 2, 1, 1)
	writeTRX(t, filepath.Join(resultsDir, "nested", "second.TRX"), 3, 3, 0, 0)

	summary, err := loadTestSummary(resultsDir)

	if err != nil {
		t.Fatalf("load summary: %v", err)
	}
	want := (TestSummary{Total: 7, Passed: 5, Failed: 1, Skipped: 1})
	if summary != want {
		t.Fatalf("unexpected summary: got %+v want %+v", summary, want)
	}
}

func TestLoadTestSummarySingleFile(t *testing.T) {
	resultsDir := t.TempDir()
	writeTRX(t, filepath.Join(resultsDir, "result.trx"), 5, 4, 0, 1)

	summary, err := loadTestSummary(resultsDir)

	if err != nil {
		t.Fatalf("load summary: %v", err)
	}
	want := (TestSummary{Total: 5, Passed: 4, Skipped: 1})
	if summary != want {
		t.Fatalf("unexpected summary: got %+v want %+v", summary, want)
	}
}

func TestLoadTestSummaryRejectsUnavailableResults(t *testing.T) {
	tests := []struct {
		name    string
		prepare func(*testing.T, string)
	}{
		{
			name: "missing file",
			prepare: func(t *testing.T, dir string) {
				t.Helper()
			},
		},
		{
			name: "malformed file",
			prepare: func(t *testing.T, dir string) {
				t.Helper()
				writeFile(t, filepath.Join(dir, "result.trx"), "<TestRun>")
			},
		},
		{
			name: "missing counters",
			prepare: func(t *testing.T, dir string) {
				t.Helper()
				writeFile(t, filepath.Join(dir, "result.trx"), `<TestRun><ResultSummary outcome="Completed"/></TestRun>`)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resultsDir := t.TempDir()
			tt.prepare(t, resultsDir)

			if _, err := loadTestSummary(resultsDir); err == nil {
				t.Fatal("expected summary error")
			}
		})
	}
}

func writeTRX(t *testing.T, path string, total, passed, failed, skipped int) {
	t.Helper()
	contents := `<TestRun xmlns="http://microsoft.com/schemas/VisualStudio/TeamTest/2010"><ResultSummary outcome="Completed"><Counters total="` + strconv.Itoa(total) + `" passed="` + strconv.Itoa(passed) + `" failed="` + strconv.Itoa(failed) + `" notExecuted="` + strconv.Itoa(skipped) + `"/></ResultSummary></TestRun>`
	writeFile(t, path, contents)
}

func writeFile(t *testing.T, path, contents string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("create fixture directory: %v", err)
	}
	if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
}
