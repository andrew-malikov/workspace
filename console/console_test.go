package console

import (
	"bytes"
	"os"
	"testing"
)

func TestNewPreservesConsoleCapabilities(t *testing.T) {
	input := bytes.NewReader(nil)
	var output, errorOutput bytes.Buffer

	terminal := New(input, &output, &errorOutput, true, true, true, false)

	if terminal.Input != input || terminal.Output != &output || terminal.Error != &errorOutput {
		t.Fatal("console did not preserve configured streams")
	}
	if !terminal.InputTerminal || !terminal.OutputTerminal || !terminal.SharedTerminal || !terminal.Color {
		t.Fatalf("unexpected console capabilities: %+v", terminal)
	}

	noColor := New(input, &output, &errorOutput, true, true, false, true)
	if noColor.Color {
		t.Fatal("NO_COLOR did not disable rich output")
	}
}

func TestSharesTerminalRequiresTwoTerminalDescriptorsForSameDevice(t *testing.T) {
	shared, err := os.CreateTemp(t.TempDir(), "shared-terminal")
	if err != nil {
		t.Fatalf("create shared descriptor: %v", err)
	}
	defer shared.Close()

	other, err := os.CreateTemp(t.TempDir(), "other-terminal")
	if err != nil {
		t.Fatalf("create distinct descriptor: %v", err)
	}
	defer other.Close()

	terminal := func(int) bool { return true }
	if !sharesTerminal(shared, shared, terminal) {
		t.Fatal("same terminal device was not detected")
	}
	if sharesTerminal(shared, other, terminal) {
		t.Fatal("distinct terminal devices were reported as shared")
	}
	if sharesTerminal(shared, shared, func(int) bool { return false }) {
		t.Fatal("redirected descriptors were reported as shared")
	}
	if sharesTerminal(nil, shared, terminal) {
		t.Fatal("unknown output descriptor was reported as shared")
	}
}

func TestSharesTerminalDefaultsToFalseWhenDescriptorCannotBeInspected(t *testing.T) {
	closed, err := os.CreateTemp(t.TempDir(), "closed-terminal")
	if err != nil {
		t.Fatalf("create descriptor: %v", err)
	}
	if err := closed.Close(); err != nil {
		t.Fatalf("close descriptor: %v", err)
	}

	if sharesTerminal(closed, closed, func(int) bool { return true }) {
		t.Fatal("uninspectable descriptors were reported as shared")
	}
}
