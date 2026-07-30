package dotnet

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/andrew-malikov/workspace/projects"
)

// TestRun identifies the artifact root shared by every category in one ws test invocation.
type TestRun struct {
	Root string
}

type testArtifacts struct {
	logPath    string
	resultsDir string
}

// NewTestRun creates a unique project-local artifact root.
func NewTestRun(projectDir string) (TestRun, error) {
	absoluteProjectDir, err := filepath.Abs(projectDir)
	if err != nil {
		return TestRun{}, fmt.Errorf("resolve project directory: %w", err)
	}

	testsDir := filepath.Join(absoluteProjectDir, ".logs", "tests")
	if err := os.MkdirAll(testsDir, 0o755); err != nil {
		return TestRun{}, fmt.Errorf("create test logs directory: %w", err)
	}

	prefix := time.Now().UTC().Format("20060102T150405.000000000Z") + "-"
	root, err := os.MkdirTemp(testsDir, prefix)
	if err != nil {
		return TestRun{}, fmt.Errorf("create test run directory: %w", err)
	}
	return TestRun{Root: root}, nil
}

func (run TestRun) createArtifacts(kind projects.TestKind) (testArtifacts, error) {
	categoryDir := filepath.Join(run.Root, string(kind))
	resultsDir := filepath.Join(categoryDir, "results")
	if err := os.MkdirAll(resultsDir, 0o755); err != nil {
		return testArtifacts{}, fmt.Errorf("create %s test artifact directory: %w", kind, err)
	}
	return testArtifacts{
		logPath:    filepath.Join(categoryDir, "output.log"),
		resultsDir: resultsDir,
	}, nil
}
