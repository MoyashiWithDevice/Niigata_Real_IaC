package renderer

import (
	"fmt"
	"sort"
	"strings"

	"IACForge/src/view"
)

// SVGRenderer renders views as SVG diagrams.
type SVGRenderer struct {
	id     string
	name   string
	format string
}

// NewSVGRenderer creates a new SVG renderer.
func NewSVGRenderer() *SVGRenderer {
	return &SVGRenderer{
		id:     "svg",
		name:   "SVG Renderer",
		format: "svg",
	}
}

// ID returns the renderer identifier.
func (r *SVGRenderer) ID() string {
	return r.id
}

// Name returns the renderer name.
func (r *SVGRenderer) Name() string {
	return r.name
}

// Format returns the output format.
func (r *SVGRenderer) Format() string {
	return r.format
}

// Render renders a view as SVG.
func (r *SVGRenderer) Render(v *view.ViewResult, opts *RenderOptions) (*Artifact, error) {
	if opts == nil {
		opts = NewRenderOptions()
	}

	width := opts.Width
	height := opts.Height

	layoutEngine := NewLayoutEngine(opts.Layout)
	layout := layoutEngine.ComputeLayout(v)

	if layout.Width > 0 {
		width = layout.Width
	}
	if layout.Height > 0 {
		height = layout.Height
	}

	var svg strings.Builder

	fmt.Fprintf(&svg, `<svg xmlns="http://www.w3.org/2000/svg" width="%.0f" height="%.0f" viewBox="0 0 %.0f %.0f">`, width, height, width, height)
	svg.WriteString("\n")

	if opts.Theme != nil && opts.Theme.Colors != nil {
		fmt.Fprintf(&svg, `<rect width="100%%" height="100%%" fill="%s"/>`, opts.Theme.Colors.Background)
		svg.WriteString("\n")
	} else {
		svg.WriteString(`<rect width="100%" height="100%" fill="#ffffff"/>`)
		svg.WriteString("\n")
	}

	for _, edge := range layout.Edges {
		r.renderEdge(&svg, edge, opts)
	}

	// Draw containers before their contents so child nodes render on top.
	nodesByDepth := make([]NodePosition, len(layout.Nodes))
	copy(nodesByDepth, layout.Nodes)
	depths := make(map[string]int, len(layout.Nodes))
	for _, node := range layout.Nodes {
		for _, childID := range node.Children {
			depths[childID] = depths[node.ID] + 1
		}
	}
	sort.SliceStable(nodesByDepth, func(i, j int) bool {
		return depths[nodesByDepth[i].ID] < depths[nodesByDepth[j].ID]
	})

	for _, node := range nodesByDepth {
		r.renderNode(&svg, node, v, opts, len(node.Children) > 0)
	}

	svg.WriteString("</svg>")

	artifact := NewArtifact(
		fmt.Sprintf("artifact-%s-%s", r.id, v.ViewID),
		r.id,
		v.ViewID,
		r.format,
		svg.String(),
	)
	artifact.Metadata["title"] = v.Title
	artifact.Metadata["description"] = v.Description

	return artifact, nil
}

// renderEdge renders a single edge.
func (r *SVGRenderer) renderEdge(svg *strings.Builder, edge EdgePosition, opts *RenderOptions) {
	if len(edge.Points) < 2 {
		return
	}

	color := "#6b7280"
	width := 2.0
	if opts.Theme != nil && opts.Theme.Lines != nil && opts.Theme.Lines.Default != nil {
		if opts.Theme.Lines.Default.Color != "" {
			color = opts.Theme.Lines.Default.Color
		}
		if opts.Theme.Lines.Default.Width > 0 {
			width = opts.Theme.Lines.Default.Width
		}
	}

	fmt.Fprintf(svg, `<line x1="%.0f" y1="%.0f" x2="%.0f" y2="%.0f" stroke="%s" stroke-width="%.1f"/>`,
		edge.Points[0].X, edge.Points[0].Y,
		edge.Points[1].X, edge.Points[1].Y,
		color, width)
	svg.WriteString("\n")
}

// renderNode renders a single node. Containers (nodes with children) are
// drawn as large labeled boxes; leaves as compact boxes with centered text.
func (r *SVGRenderer) renderNode(svg *strings.Builder, node NodePosition, v *view.ViewResult, opts *RenderOptions, container bool) {
	fill := "#e5e7eb"
	stroke := "#9ca3af"
	textColor := "#111827"

	if container {
		fill = "#f3f4f6"
	}

	if opts.Theme != nil && opts.Theme.Colors != nil {
		if container && opts.Theme.Colors.Background != "" {
			fill = opts.Theme.Colors.Background
		} else if !container && opts.Theme.Colors.Surface != "" {
			fill = opts.Theme.Colors.Surface
		}
		stroke = opts.Theme.Colors.Border
		textColor = opts.Theme.Colors.Text
	}

	fmt.Fprintf(svg, `<rect x="%.0f" y="%.0f" width="%.0f" height="%.0f" rx="4" fill="%s" stroke="%s"/>`,
		node.Position.X, node.Position.Y,
		node.Width, node.Height,
		fill, stroke)
	svg.WriteString("\n")

	fontSize := 12
	if opts.Theme != nil && opts.Theme.Typography != nil && opts.Theme.Typography.FontSize > 0 {
		fontSize = opts.Theme.Typography.FontSize
	}

	name := node.ID
	for _, entity := range v.VisibleEntities {
		if entity.ID == node.ID {
			name = entity.Name
			break
		}
	}

	var textX, textY float64
	anchor := "middle"
	if container {
		anchor = "start"
		textX = node.Position.X + 8
		textY = node.Position.Y + float64(fontSize)
	} else {
		textX = node.Position.X + node.Width/2
		textY = node.Position.Y + node.Height/2 + float64(fontSize)/3
	}

	fmt.Fprintf(svg, `<text x="%.0f" y="%.0f" text-anchor="%s" font-size="%d" fill="%s">%s</text>`,
		textX, textY, anchor, fontSize, textColor, escapeXML(name))
	svg.WriteString("\n")
}

// escapeXML escapes special XML characters.
func escapeXML(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	s = strings.ReplaceAll(s, "'", "&apos;")
	return s
}
