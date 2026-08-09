package dotnet

import (
	"errors"
	"fmt"
	"io"

	"github.com/charmbracelet/x/ansi"
)

type categoryPresentation interface {
	Begin() error
	End() error
}

type presentationFactory func(io.Writer) categoryPresentation

type plainPresentation struct{}

func newPlainPresentation(io.Writer) categoryPresentation {
	return plainPresentation{}
}

func (plainPresentation) Begin() error { return nil }
func (plainPresentation) End() error   { return nil }

type alternateScreenPresentation struct {
	writer io.Writer
	active bool
}

func newAlternateScreenPresentation(writer io.Writer) categoryPresentation {
	return &alternateScreenPresentation{writer: writer}
}

func (presentation *alternateScreenPresentation) Begin() error {
	if presentation.active {
		return nil
	}

	presentation.active = true
	if err := writeControl(presentation.writer, ansi.SetModeAltScreenSaveCursor); err != nil {
		return errors.Join(
			fmt.Errorf("enter alternate screen: %w", err),
			presentation.End(),
		)
	}
	return nil
}

func (presentation *alternateScreenPresentation) End() error {
	if !presentation.active {
		return nil
	}

	presentation.active = false
	if err := writeControl(presentation.writer, ansi.ResetModeAltScreenSaveCursor); err != nil {
		return fmt.Errorf("restore normal screen: %w", err)
	}
	return nil
}

func writeControl(writer io.Writer, sequence string) error {
	written, err := io.WriteString(writer, sequence)
	if err != nil {
		return err
	}
	if written != len(sequence) {
		return io.ErrShortWrite
	}
	return nil
}
