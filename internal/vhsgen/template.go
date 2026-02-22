package vhsgen

import (
	"embed"
	"fmt"
	"strings"
	"text/template"
)

//go:embed templates/base.tape.tmpl
var templateFS embed.FS

// TapeData holds the data for rendering a VHS tape template.
type TapeData struct {
	FeatureName      string
	ScenarioName     string
	GIFPath          string
	ASCIIPath        string
	ConfigSourcePath string
	SetupCommands    string
	DemoCommands     string
}

// RenderTape renders a VHS tape template with the provided data.
//
// Expected: TapeData with feature name, scenario name, paths, and commands.
// Returns: Rendered tape content as string, or error if template rendering fails.
// Side effects: None.
func RenderTape(data TapeData) (string, error) {
	tmplContent, err := templateFS.ReadFile("templates/base.tape.tmpl")
	if err != nil {
		return "", fmt.Errorf("reading template: %w", err)
	}

	tmpl, err := template.New("base.tape").Parse(string(tmplContent))
	if err != nil {
		return "", fmt.Errorf("parsing template: %w", err)
	}

	var buf strings.Builder
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("executing template: %w", err)
	}

	return buf.String(), nil
}
