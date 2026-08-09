package test

import (
	"context"
	"errors"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/andrew-malikov/workspace/dotnet"
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

func TestRunCategoriesSharesRunAndPreservesOrder(t *testing.T) {
	project := &projects.Project{
		Alias: "orders",
		Dir:   t.TempDir(),
		Test: projects.TestConfig{
			Unit:        projects.TestTarget{Project: "unit.csproj"},
			Integration: projects.TestTarget{Project: "integration.csproj"},
			Component:   projects.TestTarget{Project: "component.csproj"},
		},
	}
	kinds := []projects.TestKind{
		projects.UnitTestKind,
		projects.IntegrationTestKind,
		projects.ComponentTestKind,
	}
	runner := &recordingCategoryRunner{}
	factoryCalls := 0
	sharedRun := dotnet.TestRun{Root: t.TempDir()}

	err := runCategories(
		t.Context(),
		project,
		kinds,
		runner,
		func(projectDir string) (dotnet.TestRun, error) {
			factoryCalls++
			if projectDir != project.Dir {
				t.Fatalf("unexpected project directory: %q", projectDir)
			}
			return sharedRun, nil
		},
	)

	if err != nil {
		t.Fatalf("run categories: %v", err)
	}
	if factoryCalls != 1 {
		t.Fatalf("run factory called %d times", factoryCalls)
	}
	if !reflect.DeepEqual(runner.kinds, kinds) {
		t.Fatalf("unexpected category order: got %v want %v", runner.kinds, kinds)
	}
	for _, run := range runner.runs {
		if run.Root != sharedRun.Root {
			t.Fatalf("categories did not share a run root: %q", run.Root)
		}
	}
}

func TestRunCategoriesStopsAfterReportedFailure(t *testing.T) {
	project := &projects.Project{
		Alias: "orders",
		Dir:   t.TempDir(),
		Test: projects.TestConfig{
			Unit:        projects.TestTarget{Project: "unit.csproj"},
			Integration: projects.TestTarget{Project: "integration.csproj"},
			Component:   projects.TestTarget{Project: "component.csproj"},
		},
	}
	failed := errors.New("integration failed")
	runner := &recordingCategoryRunner{
		failKind: projects.IntegrationTestKind,
		err:      failed,
	}

	err := runCategories(
		t.Context(),
		project,
		projects.AllTestKinds,
		runner,
		func(string) (dotnet.TestRun, error) {
			return dotnet.TestRun{Root: t.TempDir()}, nil
		},
	)

	if !errors.Is(err, failed) {
		t.Fatalf("failure was not preserved: %v", err)
	}
	wantKinds := []projects.TestKind{projects.UnitTestKind, projects.IntegrationTestKind}
	if !reflect.DeepEqual(runner.kinds, wantKinds) {
		t.Fatalf("runner did not stop after failure: got %v want %v", runner.kinds, wantKinds)
	}
	if !reflect.DeepEqual(runner.reported, wantKinds) {
		t.Fatalf("failed category returned before reporting: got %v want %v", runner.reported, wantKinds)
	}
}

func TestInterruptContextCancelsOnInterrupt(t *testing.T) {
	ctx, stop := interruptContext(t.Context())
	defer stop()

	process, err := os.FindProcess(os.Getpid())
	if err != nil {
		t.Fatalf("find test process: %v", err)
	}
	if err := process.Signal(os.Interrupt); err != nil {
		t.Fatalf("send interrupt: %v", err)
	}

	select {
	case <-ctx.Done():
		if !errors.Is(ctx.Err(), context.Canceled) {
			t.Fatalf("unexpected interrupt error: %v", ctx.Err())
		}
	case <-time.After(time.Second):
		t.Fatal("interrupt context was not canceled")
	}
}

type recordingCategoryRunner struct {
	kinds    []projects.TestKind
	runs     []dotnet.TestRun
	reported []projects.TestKind
	failKind projects.TestKind
	err      error
}

func (runner *recordingCategoryRunner) Run(
	_ context.Context,
	_ string,
	run dotnet.TestRun,
	kind projects.TestKind,
	_ projects.TestTarget,
) error {
	runner.kinds = append(runner.kinds, kind)
	runner.runs = append(runner.runs, run)
	runner.reported = append(runner.reported, kind)
	if kind == runner.failKind {
		return runner.err
	}
	return nil
}
