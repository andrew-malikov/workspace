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
	SharedTerminal bool
	Color          bool
}

func New(input io.Reader, output, errorOutput io.Writer, inputTerminal, outputTerminal, sharedTerminal, noColor bool) Console {
	return Console{
		Input:          input,
		Output:         output,
		Error:          errorOutput,
		InputTerminal:  inputTerminal,
		OutputTerminal: outputTerminal,
		SharedTerminal: sharedTerminal,
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
		sharesTerminal(os.Stdout, os.Stderr, term.IsTerminal),
		noColor,
	)
}

func sharesTerminal(output, errorOutput *os.File, isTerminal func(int) bool) bool {
	if output == nil || errorOutput == nil ||
		!isTerminal(int(output.Fd())) || !isTerminal(int(errorOutput.Fd())) {
		return false
	}

	outputInfo, outputErr := output.Stat()
	errorInfo, errorErr := errorOutput.Stat()
	return outputErr == nil && errorErr == nil && os.SameFile(outputInfo, errorInfo)
}
