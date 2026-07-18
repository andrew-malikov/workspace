package view

import (
	"bytes"
	"errors"
	"strings"
	"testing"
	"text/template"

	"github.com/andrew-malikov/workspace/console"
)

func TestRendererSelectsRichOrPlainOutput(t *testing.T) {
	tmpl := template.Must(template.New("result").Parse("Successful result"))

	var richOutput bytes.Buffer
	rich, err := NewRenderer(&richOutput, true)
	if err != nil {
		t.Fatalf("create rich renderer: %v", err)
	}
	if err := rich.Render(tmpl, nil); err != nil {
		t.Fatalf("render rich output: %v", err)
	}
	if !strings.Contains(richOutput.String(), "\x1b[") {
		t.Fatalf("rich output contains no ANSI sequence: %q", richOutput.String())
	}

	var plainOutput bytes.Buffer
	plain, err := NewRenderer(&plainOutput, false)
	if err != nil {
		t.Fatalf("create plain renderer: %v", err)
	}
	if err := plain.Render(tmpl, nil); err != nil {
		t.Fatalf("render plain output: %v", err)
	}
	if strings.Contains(plainOutput.String(), "\x1b[") {
		t.Fatalf("plain output contains ANSI sequence: %q", plainOutput.String())
	}
	if strings.TrimSpace(plainOutput.String()) != "Successful result" {
		t.Fatalf("unexpected plain output: %q", plainOutput.String())
	}
	if len(plainOutput.String()) > 100 {
		t.Fatalf("plain output is padded to renderer width: %d bytes", len(plainOutput.String()))
	}
}

func TestRendererHonorsNoColorOnTerminal(t *testing.T) {
	var output, errorOutput bytes.Buffer
	terminal := console.New(nil, &output, &errorOutput, true, true, true)
	renderer, err := NewRenderer(terminal.Output, terminal.Color)
	if err != nil {
		t.Fatalf("create renderer: %v", err)
	}
	if err := renderer.Render(template.Must(template.New("result").Parse("result")), nil); err != nil {
		t.Fatalf("render output: %v", err)
	}
	if strings.Contains(output.String(), "\x1b[") {
		t.Fatalf("NO_COLOR output contains ANSI sequence: %q", output.String())
	}
}

func TestRendererWritesDynamicTableValuesLiterally(t *testing.T) {
	tmpl := template.Must(template.New("table").Funcs(TemplateFuncs).Parse(
		"| value |\n| --- |\n| {{.Value | tableCell}} |",
	))
	var output bytes.Buffer
	renderer, err := NewRenderer(&output, false)
	if err != nil {
		t.Fatalf("create renderer: %v", err)
	}

	if err := renderer.Render(tmpl, Args{"Value": "a|*b*`c`\nnext"}); err != nil {
		t.Fatalf("render table: %v", err)
	}

	got := output.String()
	for _, literal := range []string{"a|*b*`c`", `\nnext`} {
		if !strings.Contains(got, literal) {
			t.Fatalf("output does not preserve %q: %q", literal, got)
		}
	}
}

func TestRendererPropagatesWriterErrorOnce(t *testing.T) {
	rejected := errors.New("write rejected")
	writer := &rejectingWriter{err: rejected}
	renderer, err := NewRenderer(writer, false)
	if err != nil {
		t.Fatalf("create renderer: %v", err)
	}

	err = renderer.Render(template.Must(template.New("result").Parse("result")), nil)
	if !errors.Is(err, rejected) {
		t.Fatalf("unexpected render error: %v", err)
	}
	if writer.attempts != 1 {
		t.Fatalf("unexpected write attempts: %d", writer.attempts)
	}
}

type rejectingWriter struct {
	err      error
	attempts int
}

func (writer *rejectingWriter) Write([]byte) (int, error) {
	writer.attempts++
	return 0, writer.err
}
