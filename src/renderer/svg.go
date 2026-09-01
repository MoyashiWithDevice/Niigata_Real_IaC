package renderer

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/bababa/Niigata_Real_IaC/src/core"
	"github.com/bababa/Niigata_Real_IaC/src/view"
)

// headerBarHeight is the height of the colored title bar drawn on containers.
const headerBarHeight = 24.0

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
	theme := ResolveTheme(opts)

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

	// Fonts and reusable arrow marker.
	fmt.Fprintf(&svg, `<style>text{font-family:%s;}</style>`, theme.FontFamily())
	svg.WriteString("\n")
	fmt.Fprintf(&svg, `<defs><marker id="arrow" markerWidth="8" markerHeight="8" refX="7" refY="4" orient="auto" markerUnits="userSpaceOnUse"><path d="M0,0 L8,4 L0,8 z"/></marker></defs>`)
	svg.WriteString("\n")

	fmt.Fprintf(&svg, `<rect width="100%%" height="100%%" fill="%s"/>`, theme.BackgroundColor())
	svg.WriteString("\n")

	if isMapLayout(opts.Layout) {
		r.drawMapFrame(&svg, layout, theme)
	}

	relTypeByID := relationTypesByID(v)

	for _, edge := range layout.Edges {
		r.renderEdge(&svg, edge, relTypeByID[edge.ID], theme)
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
		r.renderNode(&svg, node, v, theme, len(node.Children) > 0)
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

// isMapLayout reports whether the render options request the geographic map
// layout ("map" or its "geographic" alias).
func isMapLayout(cfg *LayoutConfig) bool {
	if cfg == nil {
		return false
	}
	return cfg.Type == "map" || cfg.Type == "geographic"
}

// drawMapFrame draws the geographic reference frame of a map layout: a border
// around the projected area and min/max latitude/longitude axis labels.
func (r *SVGRenderer) drawMapFrame(svg *strings.Builder, layout *LayoutResult, theme *Theme) {
	x, okX := parseMetaFloat(layout.Meta, "geo_x")
	y, okY := parseMetaFloat(layout.Meta, "geo_y")
	w, okW := parseMetaFloat(layout.Meta, "geo_w")
	h, okH := parseMetaFloat(layout.Meta, "geo_h")
	if !okX || !okY || !okW || !okH {
		return
	}

	fmt.Fprintf(svg, `<rect x="%.0f" y="%.0f" width="%.0f" height="%.0f" fill="none" stroke="%s" stroke-width="1" stroke-dasharray="4 4"/>`,
		x, y, w, h, theme.Colors.Border)
	svg.WriteString("\n")

	labelColor := "#64748b"
	if theme.Colors != nil && theme.Colors.Secondary != "" {
		labelColor = theme.Colors.Secondary
	}
	fontSize := 10

	minLat, okMinLat := layout.Meta["geo_min_lat"]
	maxLat, okMaxLat := layout.Meta["geo_max_lat"]
	minLon, okMinLon := layout.Meta["geo_min_lon"]
	maxLon, okMaxLon := layout.Meta["geo_max_lon"]

	if okMaxLat {
		fmt.Fprintf(svg, `<text x="%.0f" y="%.0f" text-anchor="start" font-size="%d" fill="%s">N %s°</text>`,
			x+6, y-6, fontSize, labelColor, escapeXML(maxLat))
		svg.WriteString("\n")
	}
	if okMinLat {
		fmt.Fprintf(svg, `<text x="%.0f" y="%.0f" text-anchor="start" font-size="%d" fill="%s">S %s°</text>`,
			x+6, y+h+float64(fontSize)+4, fontSize, labelColor, escapeXML(minLat))
		svg.WriteString("\n")
	}
	if okMaxLon {
		fmt.Fprintf(svg, `<text x="%.0f" y="%.0f" text-anchor="end" font-size="%d" fill="%s">E %s°</text>`,
			x+w-6, y-6, fontSize, labelColor, escapeXML(maxLon))
		svg.WriteString("\n")
	}
	if okMinLon {
		fmt.Fprintf(svg, `<text x="%.0f" y="%.0f" text-anchor="end" font-size="%d" fill="%s">W %s°</text>`,
			x+w-6, y+h+float64(fontSize)+4, fontSize, labelColor, escapeXML(minLon))
		svg.WriteString("\n")
	}

	if rowY, ok := parseMetaFloat(layout.Meta, "fallback_row_y"); ok {
		fmt.Fprintf(svg, `<text x="%.0f" y="%.0f" text-anchor="start" font-size="%d" fill="%s">no coordinates</text>`,
			x, rowY-8, fontSize, labelColor)
		svg.WriteString("\n")
	}
}

// parseMetaFloat reads a float from layout metadata.
func parseMetaFloat(meta map[string]string, key string) (float64, bool) {
	if meta == nil {
		return 0, false
	}
	s, ok := meta[key]
	if !ok {
		return 0, false
	}
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, false
	}
	return f, true
}

// renderEdge renders a single edge. The look is driven by the relation type:
// line color and stroke style come from the theme, and a small arrowhead is
// drawn at the target when the edge is directed.
func (r *SVGRenderer) renderEdge(svg *strings.Builder, edge EdgePosition, relType string, theme *Theme) {
	if len(edge.Points) < 2 {
		return
	}

	p1 := edge.Points[0]
	p2 := edge.Points[1]

	color := theme.LineColor(relType)
	width := theme.LineWidth(relType)
	dash := theme.LineDash(relType)

	attrs := fmt.Sprintf(`x1="%.0f" y1="%.0f" x2="%.0f" y2="%.0f" stroke="%s" stroke-width="%.1f"`,
		p1.X, p1.Y, p2.X, p2.Y, color, width)
	if dash != "" {
		attrs += fmt.Sprintf(` stroke-dasharray="%s"`, dash)
	}
	fmt.Fprintf(svg, `<line %s marker-end="url(#arrow)"/>`, attrs)
	svg.WriteString("\n")
}

// renderNode renders a single node. Containers (nodes with children) are drawn
// as large labeled boxes tinted by their kind; leaves as compact boxes filled
// with the kind accent color, white text, and an optional subtitle.
func (r *SVGRenderer) renderNode(svg *strings.Builder, node NodePosition, v *view.ViewResult, theme *Theme, container bool) {
	entity := entityByID(v, node.ID)
	kind := ""
	if entity != nil {
		kind = string(entity.Kind)
	}
	accent := theme.AccentColor(kind)

	fill := accent
	stroke := accent
	textColor := "#ffffff"

	if container {
		// Containers get a soft tinted surface with a colored border and title.
		fill = theme.ContainerFill()
		stroke = accent
		textColor = theme.TextColor()
	}

	fmt.Fprintf(svg, `<rect x="%.0f" y="%.0f" width="%.0f" height="%.0f" rx="8" fill="%s" stroke="%s" stroke-width="1.5"/>`,
		node.Position.X, node.Position.Y,
		node.Width, node.Height,
		fill, stroke)
	svg.WriteString("\n")

	name := node.ID
	if entity != nil {
		name = entity.Name
	}

	fontSize := theme.FontSize()
	if fontSize <= 0 {
		fontSize = theme.Typography.FontSize
	}

	if container {
		// Title bar: colored accent strip + kind label + entity name.
		fmt.Fprintf(svg, `<rect x="%.0f" y="%.0f" width="%.0f" height="%.0f" rx="8" fill="%s"/>`,
			node.Position.X, node.Position.Y, node.Width, headerBarHeight, accent)
		svg.WriteString("\n")
		title := escapeXML(labelWithIcon(theme, kind, name))
		fmt.Fprintf(svg, `<text x="%.0f" y="%.0f" text-anchor="start" font-size="%d" font-weight="bold" fill="#ffffff">%s</text>`,
			node.Position.X+8, node.Position.Y+float64(fontSize), fontSize, title)
		svg.WriteString("\n")
		return
	}

	// Leaf node: icon + name centered; optional subtitle underneath.
	subtitle := entitySubtitle(entity)
	centerX := node.Position.X + node.Width/2
	title := escapeXML(labelWithIcon(theme, kind, name))

	if subtitle == "" {
		fmt.Fprintf(svg, `<text x="%.0f" y="%.0f" text-anchor="middle" font-size="%d" font-weight="bold" fill="%s">%s</text>`,
			centerX, node.Position.Y+node.Height/2+float64(fontSize)/3, fontSize, textColor, title)
		svg.WriteString("\n")
	} else {
		fmt.Fprintf(svg, `<text x="%.0f" y="%.0f" text-anchor="middle" font-size="%d" font-weight="bold" fill="%s">%s</text>`,
			centerX, node.Position.Y+16, fontSize, textColor, title)
		svg.WriteString("\n")
		fmt.Fprintf(svg, `<text x="%.0f" y="%.0f" text-anchor="middle" font-size="10" fill="rgba(255,255,255,0.85)">%s</text>`,
			centerX, node.Position.Y+34, escapeXML(subtitle))
		svg.WriteString("\n")
	}
}

// relationTypesByID builds a map from relation ID to its type for edge styling.
func relationTypesByID(v *view.ViewResult) map[string]string {
	m := make(map[string]string, len(v.VisibleRelations))
	for _, rel := range v.VisibleRelations {
		m[rel.ID] = string(rel.Type)
	}
	return m
}

// entityByID finds the visible entity with the given ID, or nil.
func entityByID(v *view.ViewResult, id string) *core.Entity {
	for _, e := range v.VisibleEntities {
		if e.ID == id {
			return e
		}
	}
	return nil
}

// entitySubtitle picks a human-readable secondary value for a leaf node, so
// diagrams convey a meaningful fact (population count, elevation, etc.) instead
// of only the name. Returns "" when no suitable value exists.
func entitySubtitle(e *core.Entity) string {
	if e == nil {
		return ""
	}
	switch e.Kind {
	case "population":
		return propertyString(e, "count")
	case "water_body":
		return propertyString(e, "water_type")
	case "terrain":
		return propertyString(e, "elevation_m", "m")
	case "ground":
		return propertyString(e, "soil_type")
	case "forest":
		return propertyString(e, "forest_type")
	case "hot_spring":
		return propertyString(e, "spring_quality")
	case "cultural_asset":
		return propertyString(e, "asset_type")
	case "event":
		return propertyString(e, "season")
	default:
		return ""
	}
}

// propertyString returns a JSON-style scalar property as a string, optionally
// with a unit suffix. Only scalars are rendered.
func propertyString(e *core.Entity, key string, unit ...string) string {
	val, ok := e.GetProperty(key)
	if !ok || val == nil {
		return ""
	}
	s := fmt.Sprintf("%v", val)
	if len(unit) > 0 && unit[0] != "" {
		// Strip a trailing decimal when it is insignificant.
		s = strings.TrimSuffix(strings.TrimSuffix(s, "0"), ".")
		s += unit[0]
	}
	return s
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
