package view

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"text/template"
	"unicode"

	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/glamour/styles"
)

type Args = map[string]any

type Renderer struct {
	writer   io.Writer
	rich     bool
	renderer *glamour.TermRenderer
}

func NewRenderer(writer io.Writer, rich bool) (*Renderer, error) {
	result := &Renderer{writer: writer, rich: rich}
	if !rich {
		return result, nil
	}

	renderer, err := newTermRenderer("dark", 80)
	if err != nil {
		return nil, err
	}
	result.renderer = renderer
	return result, nil
}

func newTermRenderer(style string, wordWrap int) (*glamour.TermRenderer, error) {
	return glamour.NewTermRenderer(
		glamour.WithStandardStyle(style),
		glamour.WithWordWrap(wordWrap),
		glamour.WithPreservedNewLines(),
	)
}

func newPlainRenderer(wordWrap int) (*glamour.TermRenderer, error) {
	style := styles.NoTTYStyleConfig
	noMargin := uint(0)
	style.Document.Margin = &noMargin
	style.CodeBlock.Margin = &noMargin
	return glamour.NewTermRenderer(
		glamour.WithStyles(style),
		glamour.WithWordWrap(wordWrap),
		glamour.WithPreservedNewLines(),
	)
}

func (renderer *Renderer) Render(tmpl *template.Template, data Args) error {
	var markdown bytes.Buffer
	if err := tmpl.Execute(&markdown, data); err != nil {
		return err
	}

	termRenderer := renderer.renderer
	if !renderer.rich {
		var err error
		termRenderer, err = newPlainRenderer(markdownWidth(markdown.String()))
		if err != nil {
			return err
		}
	}

	out, err := termRenderer.Render(markdown.String())
	if err != nil {
		return err
	}
	if !renderer.rich {
		lines := strings.Split(out, "\n")
		for index := range lines {
			lines[index] = strings.TrimRight(lines[index], " \t")
		}
		out = strings.Join(lines, "\n")
	}
	_, err = io.WriteString(renderer.writer, out)
	return err
}

func markdownWidth(markdown string) int {
	width := 1
	for line := range strings.SplitSeq(markdown, "\n") {
		if len(line) > width {
			width = len(line)
		}
	}
	return width
}

var TemplateFuncs = template.FuncMap{
	"literal":   MarkdownLiteral,
	"tableCell": MarkdownTableCell,
}

func SafeText(value string) string {
	var escaped strings.Builder
	escaped.Grow(len(value))
	for _, r := range value {
		if unicode.IsControl(r) && r != '\n' && r != '\t' {
			fmt.Fprintf(&escaped, `\x%02x`, r)
			continue
		}
		escaped.WriteRune(r)
	}
	return escaped.String()
}

func MarkdownLiteral(value string) string {
	value = SafeText(value)
	var escaped strings.Builder
	escaped.Grow(len(value))
	for _, r := range value {
		if strings.ContainsRune(`\`+"`*_{}[]<>()#+-.!|", r) {
			escaped.WriteByte('\\')
		}
		escaped.WriteRune(r)
	}
	return escaped.String()
}

func MarkdownTableCell(value string) string {
	return strings.ReplaceAll(MarkdownLiteral(value), "\n", `\n`)
}
