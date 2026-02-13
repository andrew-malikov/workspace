package view

import (
	"bytes"
	"fmt"
	"text/template"
	flr "ws/failure"

	"github.com/charmbracelet/glamour"
)

type Args = map[string]any

func Render(tmpl *template.Template, data Args) error {
	var buf bytes.Buffer
	tmpl.Execute(&buf, data)

	out, err := glamour.Render(buf.String(), "dark")
	if err != nil {
		return err
	}

	fmt.Print(out)
	return nil
}

var UNHANDLED_FAILURE_TEMPLATE = template.Must(template.New("unhandled_failure").Parse(
	// todo: show up the data along the way as json HA_HA_HA
	`Unhandler failure is found **{{.Type}}**`,
))

func RenderUnhandledFailure(ctx flr.Context) error {
	return Render(UNHANDLED_FAILURE_TEMPLATE, Args{
		"Type": ctx.Type,
	})
}

var DIRECTORY_IS_NOT_TRACKED_YET = template.Must(template.New("directory_is_untracked").Parse(
	`Current directory **{{.Dir}}** isn't tracked yet`,
))

func RenderDirectoryIsNotTrackedYet(dir string) error {
	return Render(DIRECTORY_IS_NOT_TRACKED_YET, Args{
		"Dir": dir,
	})
}
