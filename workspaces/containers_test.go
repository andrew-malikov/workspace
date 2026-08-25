package workspaces

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/andrew-malikov/workspace/containers"
	"github.com/andrew-malikov/workspace/projects"
)

type composeCall struct {
	Action string
	Target containers.Target
}

type composeSpy struct {
	running map[string]bool
	calls   []composeCall
	fail    map[string]error
}

func (spy *composeSpy) HasRunning(ctx context.Context, target containers.Target) (bool, error) {
	spy.calls = append(spy.calls, composeCall{Action: "ps", Target: target})
	if err := spy.failure("ps", target); err != nil {
		return false, err
	}
	return spy.running[target.Dir], nil
}

func (spy *composeSpy) Up(ctx context.Context, target containers.Target) error {
	spy.calls = append(spy.calls, composeCall{Action: "up", Target: target})
	return spy.failure("up", target)
}

func (spy *composeSpy) Down(ctx context.Context, target containers.Target) error {
	spy.calls = append(spy.calls, composeCall{Action: "down", Target: target})
	return spy.failure("down", target)
}

func (spy *composeSpy) Cleanup(ctx context.Context, target containers.Target) error {
	spy.calls = append(spy.calls, composeCall{Action: "cleanup", Target: target})
	return spy.failure("cleanup", target)
}

func (spy *composeSpy) failure(action string, target containers.Target) error {
	return spy.fail[action+":"+target.Alias]
}

func TestResolveProjectDefaultsToCurrentDirectory(t *testing.T) {
	dir := filepath.Join(string(os.PathSeparator), "tmp", "orders")
	workspace := Workspace{
		Projects: map[string]projects.Project{
			"orders": {Alias: "orders", Dir: dir},
		},
	}

	project, err := workspace.ResolveProject(filepath.Join(dir, "src"), "")
	if err != nil {
		t.Fatalf("resolve project: %v", err)
	}

	if project.Alias != "orders" {
		t.Fatalf("unexpected project alias: %s", project.Alias)
	}
}

func TestResolveProjectMatchesAlias(t *testing.T) {
	workspace := Workspace{
		Projects: map[string]projects.Project{
			"orders": {Alias: "orders", Dir: filepath.Join(string(os.PathSeparator), "tmp", "orders")},
		},
	}

	project, err := workspace.ResolveProject("/tmp/other", "orders")
	if err != nil {
		t.Fatalf("resolve project: %v", err)
	}

	if project.Alias != "orders" {
		t.Fatalf("unexpected project alias: %s", project.Alias)
	}
}

func TestUpProjectExecutesPlannedMutatingActions(t *testing.T) {
	targetDir := trackedProjectDir(t)
	apiDir := trackedProjectDir(t)
	billingDir := trackedProjectDir(t)
	compose := "docker-compose.yaml"
	workspace := Workspace{Projects: map[string]projects.Project{
		"orders":  {Alias: "orders", Dir: targetDir, Compose: &compose},
		"api":     {Alias: "api", Dir: apiDir, Compose: &compose},
		"billing": {Alias: "billing", Dir: billingDir, Compose: &compose},
	}}
	spy := &composeSpy{running: map[string]bool{apiDir: true, billingDir: true}}

	result, err := workspace.UpProject(t.Context(), targetDir, "", false, false, spy)
	if err != nil {
		t.Fatalf("up project: %v", err)
	}
	if result.Alias != "orders" || !slices.Equal(result.Stopped, []string{"api", "billing"}) {
		t.Fatalf("unexpected result: %+v", result)
	}

	expected := []composeCall{
		{Action: "ps", Target: containers.Target{Alias: "api", Dir: apiDir, File: compose}},
		{Action: "ps", Target: containers.Target{Alias: "billing", Dir: billingDir, File: compose}},
		{Action: "down", Target: containers.Target{Alias: "api", Dir: apiDir, File: compose}},
		{Action: "down", Target: containers.Target{Alias: "billing", Dir: billingDir, File: compose}},
		{Action: "up", Target: containers.Target{Alias: "orders", Dir: targetDir, File: compose}},
	}
	if !slices.Equal(spy.calls, expected) {
		t.Fatalf("unexpected calls: %v", spy.calls)
	}
}

func TestUpProjectStopsAfterFirstActionFailure(t *testing.T) {
	targetDir := trackedProjectDir(t)
	apiDir := trackedProjectDir(t)
	billingDir := trackedProjectDir(t)
	compose := "docker-compose.yaml"
	actionErr := errors.New("down failed")
	workspace := Workspace{Projects: map[string]projects.Project{
		"orders":  {Alias: "orders", Dir: targetDir, Compose: &compose},
		"api":     {Alias: "api", Dir: apiDir, Compose: &compose},
		"billing": {Alias: "billing", Dir: billingDir, Compose: &compose},
	}}
	spy := &composeSpy{
		running: map[string]bool{apiDir: true, billingDir: true},
		fail:    map[string]error{"down:api": actionErr},
	}

	_, err := workspace.UpProject(t.Context(), targetDir, "", false, false, spy)
	if !errors.Is(err, actionErr) {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := []composeCall{
		{Action: "ps", Target: containers.Target{Alias: "api", Dir: apiDir, File: compose}},
		{Action: "ps", Target: containers.Target{Alias: "billing", Dir: billingDir, File: compose}},
		{Action: "down", Target: containers.Target{Alias: "api", Dir: apiDir, File: compose}},
	}
	if !slices.Equal(spy.calls, expected) {
		t.Fatalf("unexpected calls after failure: %v", spy.calls)
	}
}

func TestUpProjectFailsWhenComposeFileMissingBeforePlan(t *testing.T) {
	dir := t.TempDir()
	compose := "docker-compose.yaml"
	workspace := Workspace{Projects: map[string]projects.Project{
		"orders": {Alias: "orders", Dir: dir, Compose: &compose},
	}}
	spy := &composeSpy{}

	_, err := workspace.UpProject(t.Context(), dir, "", false, false, spy)
	if err == nil || !strings.Contains(err.Error(), "no docker compose configured for project orders") {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(spy.calls) != 0 {
		t.Fatalf("plan ran before missing-file gate: %v", spy.calls)
	}
}

func TestDownProjectExecutesPlannedAction(t *testing.T) {
	dir := trackedProjectDir(t)
	compose := "docker-compose.yaml"
	workspace := Workspace{Projects: map[string]projects.Project{
		"orders": {Alias: "orders", Dir: dir, Compose: &compose},
	}}
	spy := &composeSpy{}

	project, err := workspace.DownProject(t.Context(), t.TempDir(), "orders", false, spy)
	if err != nil {
		t.Fatalf("down project: %v", err)
	}
	if project.Alias != "orders" {
		t.Fatalf("unexpected project: %s", project.Alias)
	}
	expected := []composeCall{{
		Action: "down",
		Target: containers.Target{Alias: "orders", Dir: dir, File: compose},
	}}
	if !slices.Equal(spy.calls, expected) {
		t.Fatalf("unexpected calls: %v", spy.calls)
	}
}

func TestDownProjectRejectsMissingTargetBeforeComposeAction(t *testing.T) {
	spy := &composeSpy{}
	workspace := Workspace{Projects: map[string]projects.Project{}}

	_, err := workspace.DownProject(t.Context(), t.TempDir(), "missing", false, spy)
	if err == nil || !strings.Contains(err.Error(), "project is not tracked: missing") {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(spy.calls) != 0 {
		t.Fatalf("unexpected calls: %v", spy.calls)
	}
}

func TestDownProjectRejectsMissingComposeBeforeAction(t *testing.T) {
	dir := t.TempDir()
	compose := "docker-compose.yaml"
	spy := &composeSpy{}
	workspace := Workspace{Projects: map[string]projects.Project{
		"orders": {Alias: "orders", Dir: dir, Compose: &compose},
	}}

	_, err := workspace.DownProject(t.Context(), dir, "", false, spy)
	if err == nil || !strings.Contains(err.Error(), "no docker compose configured for project orders") {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(spy.calls) != 0 {
		t.Fatalf("unexpected calls: %v", spy.calls)
	}
}

func TestDownProjectPreservesComposeFailure(t *testing.T) {
	dir := trackedProjectDir(t)
	compose := "docker-compose.yaml"
	actionErr := errors.New("cleanup failed")
	spy := &composeSpy{fail: map[string]error{"cleanup:orders": actionErr}}
	workspace := Workspace{Projects: map[string]projects.Project{
		"orders": {Alias: "orders", Dir: dir, Compose: &compose},
	}}

	_, err := workspace.DownProject(t.Context(), dir, "", true, spy)
	if !errors.Is(err, actionErr) {
		t.Fatalf("unexpected error: %v", err)
	}
}

func trackedProjectDir(t *testing.T) string {
	t.Helper()

	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "docker-compose.yaml"), []byte("services: {}"), 0o644); err != nil {
		t.Fatalf("write compose: %v", err)
	}
	return dir
}
