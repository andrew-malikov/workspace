package dotnet

import (
	"bytes"
	"errors"
	"testing"

	"github.com/charmbracelet/x/ansi"
)

func TestAlternateScreenPresentationWritesOneLifecycle(t *testing.T) {
	var output bytes.Buffer
	presentation := newAlternateScreenPresentation(&output)

	if err := presentation.Begin(); err != nil {
		t.Fatalf("begin presentation: %v", err)
	}
	if err := presentation.Begin(); err != nil {
		t.Fatalf("repeat begin presentation: %v", err)
	}
	if err := presentation.End(); err != nil {
		t.Fatalf("end presentation: %v", err)
	}
	if err := presentation.End(); err != nil {
		t.Fatalf("repeat end presentation: %v", err)
	}

	want := ansi.SetModeAltScreenSaveCursor + ansi.ResetModeAltScreenSaveCursor
	if output.String() != want {
		t.Fatalf("unexpected control sequence: got %q want %q", output.String(), want)
	}
}

func TestPlainPresentationWritesNoControlSequences(t *testing.T) {
	var output bytes.Buffer
	presentation := newPlainPresentation(&output)

	if err := presentation.Begin(); err != nil {
		t.Fatalf("begin presentation: %v", err)
	}
	if err := presentation.End(); err != nil {
		t.Fatalf("end presentation: %v", err)
	}
	if output.Len() != 0 {
		t.Fatalf("plain presentation wrote %q", output.String())
	}
}

func TestAlternateScreenPresentationRestoresAfterEntryWriteFailure(t *testing.T) {
	rejected := errors.New("control write rejected")
	writer := &rejectingControlWriter{err: rejected}
	presentation := newAlternateScreenPresentation(writer)

	err := presentation.Begin()

	if !errors.Is(err, rejected) {
		t.Fatalf("unexpected begin error: %v", err)
	}
	if writer.writes != 2 {
		t.Fatalf("entry failure made %d writes, want entry and restoration", writer.writes)
	}
	if err := presentation.End(); err != nil {
		t.Fatalf("restoration was not idempotent: %v", err)
	}
	if writer.writes != 2 {
		t.Fatalf("idempotent restoration wrote again: %d", writer.writes)
	}
}

func TestAlternateScreenPresentationReturnsRestorationFailureOnce(t *testing.T) {
	rejected := errors.New("restore rejected")
	writer := &failAfterWriter{allowed: 1, err: rejected}
	presentation := newAlternateScreenPresentation(writer)
	if err := presentation.Begin(); err != nil {
		t.Fatalf("begin presentation: %v", err)
	}

	err := presentation.End()

	if !errors.Is(err, rejected) {
		t.Fatalf("unexpected restoration error: %v", err)
	}
	if err := presentation.End(); err != nil {
		t.Fatalf("restoration failure was retried: %v", err)
	}
	if writer.writes != 2 {
		t.Fatalf("unexpected write count: %d", writer.writes)
	}
}

type rejectingControlWriter struct {
	writes int
	err    error
}

func (writer *rejectingControlWriter) Write([]byte) (int, error) {
	writer.writes++
	return 0, writer.err
}

type failAfterWriter struct {
	writes  int
	allowed int
	err     error
}

func (writer *failAfterWriter) Write(data []byte) (int, error) {
	writer.writes++
	if writer.writes > writer.allowed {
		return 0, writer.err
	}
	return len(data), nil
}
