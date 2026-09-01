package renderer

import (
	"fmt"

	"github.com/bababa/Niigata_Real_IaC/src/view"
)

// RenderFormat renders a view result to a string in the requested format.
// Supported formats are "mermaid", "svg", "json", "markdown", and "md".
func RenderFormat(v *view.ViewResult, format string) (string, error) {
	return RenderFormatWithOptions(v, format, NewRenderOptions())
}

// RenderFormatWithOptions renders a view result to a string in the requested
// format, honoring the provided render options (theme, size, etc.). This lets
// callers supply a theme from the CLI or an MCP tool.
func RenderFormatWithOptions(v *view.ViewResult, format string, opts *RenderOptions) (string, error) {
	switch format {
	case "mermaid":
		artifact, err := NewMermaidRenderer().Render(v, opts)
		if err != nil {
			return "", err
		}
		return artifact.Content, nil
	case "svg":
		artifact, err := NewSVGRenderer().Render(v, opts)
		if err != nil {
			return "", err
		}
		return artifact.Content, nil
	case "json":
		artifact, err := NewJSONRenderer().Render(v, opts)
		if err != nil {
			return "", err
		}
		return artifact.Content, nil
	case "markdown", "md":
		artifact, err := NewMarkdownRenderer().Render(v, opts)
		if err != nil {
			return "", err
		}
		return artifact.Content, nil
	default:
		return "", fmt.Errorf("unknown render format: %s (supported: mermaid, svg, json, markdown)", format)
	}
}
