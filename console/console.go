package console

import (
	"io"
	"os"

	"golang.org/x/term"
)

type Console struct {
	Input          io.Reader
	Output         io.Writer
	Error          io.Writer
	InputTerminal  bool
	OutputTerminal bool
	Color          bool
}

func New(input io.Reader, output, errorOutput io.Writer, inputTerminal, outputTerminal, noColor bool) Console {
	return Console{
		Input:          input,
		Output:         output,
		Error:          errorOutput,
		InputTerminal:  inputTerminal,
		OutputTerminal: outputTerminal,
		Color:          outputTerminal && !noColor,
	}
}

func OS() Console {
	_, noColor := os.LookupEnv("NO_COLOR")
	return New(
		os.Stdin,
		os.Stdout,
		os.Stderr,
		term.IsTerminal(int(os.Stdin.Fd())),
		term.IsTerminal(int(os.Stdout.Fd())),
		noColor,
	)
}
