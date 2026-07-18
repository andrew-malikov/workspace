package clear

import (
	"bytes"
	"context"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/andrew-malikov/workspace/console"
)

func TestClearRejectsNoninteractiveStreamsBeforeLaunchingUI(t *testing.T) {
	var stdout, stderr bytes.Buffer
	terminal := console.New(strings.NewReader(""), &stdout, &stderr, false, false, false)
	built := false
	launched := false
	command := newCommand(
		terminal,
		func(context.Context) (tea.Model, error) {
			built = true
			return nil, nil
		},
		func(tea.Model, ...tea.ProgramOption) error {
			launched = true
			return nil
		},
	)

	err := command.Run(t.Context(), []string{"ws"})

	if err == nil || !strings.Contains(err.Error(), "requires interactive") {
		t.Fatalf("unexpected error: %v", err)
	}
	if built || launched {
		t.Fatalf("noninteractive clear reached UI: built=%v launched=%v", built, launched)
	}
	if strings.Contains(stdout.String(), "\x1b[") || strings.Contains(stderr.String(), "\x1b[") {
		t.Fatalf("noninteractive clear emitted terminal control sequence: stdout=%q stderr=%q", stdout.String(), stderr.String())
	}
}

func TestClearLaunchesUIWithInteractiveStreams(t *testing.T) {
	var stdout, stderr bytes.Buffer
	terminal := console.New(strings.NewReader(""), &stdout, &stderr, true, true, false)
	built := false
	launched := false
	command := newCommand(
		terminal,
		func(context.Context) (tea.Model, error) {
			built = true
			return nil, nil
		},
		func(_ tea.Model, options ...tea.ProgramOption) error {
			launched = true
			if len(options) != 2 {
				t.Fatalf("unexpected Bubble Tea option count: %d", len(options))
			}
			return nil
		},
	)

	if err := command.Run(t.Context(), []string{"ws"}); err != nil {
		t.Fatalf("run clear: %v", err)
	}
	if !built || !launched {
		t.Fatalf("interactive clear did not launch UI: built=%v launched=%v", built, launched)
	}
}
