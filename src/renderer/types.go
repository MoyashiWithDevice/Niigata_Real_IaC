package renderer

import (
	"time"

	"github.com/bababa/Niigata_Real_IaC/src/view"
)

// Renderer is the interface that all renderers must implement.
type Renderer interface {
	Render(v *view.ViewResult, opts *RenderOptions) (*Artifact, error)
	ID() string
	Name() string
	Format() string
}

// RenderOptions configures the rendering process.
type RenderOptions struct {
	Width   float64                `yaml:"width,omitempty"`
	Height  float64                `yaml:"height,omitempty"`
	Theme   *Theme                 `yaml:"theme,omitempty"`
	Layout  *LayoutConfig          `yaml:"layout,omitempty"`
	Options map[string]interface{} `yaml:"options,omitempty"`
}

// Artifact represents a rendered output.
type Artifact struct {
	ID         string                 `yaml:"id"`
	RendererID string                 `yaml:"renderer_id"`
	ViewID     string                 `yaml:"view_id"`
	Format     string                 `yaml:"format"`
	Content    string                 `yaml:"content"`
	Metadata   map[string]interface{} `yaml:"metadata,omitempty"`
	Timestamp  string                 `yaml:"timestamp"`
}

// Theme defines presentation characteristics.
type Theme struct {
	ID         string        `yaml:"id"`
	Name       string        `yaml:"name"`
	Colors     *ColorPalette `yaml:"colors,omitempty"`
	Typography *Typography   `yaml:"typography,omitempty"`
	Lines      *LineStyles   `yaml:"lines,omitempty"`
	// KindIcons maps an entity kind (e.g. "species") to a short glyph used by
	// icon-capable renderers. The glyph may be a Unicode/emoji character or a
	// small label. Renderers that cannot show glyphs simply ignore it.
	KindIcons map[string]string `yaml:"kind_icons,omitempty"`
	// RelColors maps a relation type (e.g. "belongs_to") to a stroke color used
	// when drawing the corresponding edges.
	RelColors map[string]string `yaml:"rel_colors,omitempty"`
	// RelStyles maps a relation type to a line style ("solid", "dashed",
	// "dotted"). Unspecified types fall back to the default line style.
	RelStyles map[string]string `yaml:"rel_styles,omitempty"`
}

// ColorPalette defines color definitions.
type ColorPalette struct {
	Primary    string `yaml:"primary,omitempty"`
	Secondary  string `yaml:"secondary,omitempty"`
	Background string `yaml:"background,omitempty"`
	Surface    string `yaml:"surface,omitempty"`
	Text       string `yaml:"text,omitempty"`
	Border     string `yaml:"border,omitempty"`
	Success    string `yaml:"success,omitempty"`
	Warning    string `yaml:"warning,omitempty"`
	Error      string `yaml:"error,omitempty"`
	Info       string `yaml:"info,omitempty"`
	// KindColors maps an entity kind (e.g. "forest") to its accent color used to
	// fill node backgrounds and containers.
	KindColors map[string]string `yaml:"kind_colors,omitempty"`
}

// KindColor returns the accent color for an entity kind, falling back to the
// ColorPalette.Primary color when the kind has no explicit color.
func (cp *ColorPalette) KindColor(kind string) string {
	if cp == nil {
		return ""
	}
	if c, ok := cp.KindColors[kind]; ok && c != "" {
		return c
	}
	return cp.Primary
}

// KindIcon returns the glyph for an entity kind, returning "" when the kind has
// no icon and the theme has no glyphs at all.
func (t *Theme) KindIcon(kind string) string {
	if t == nil || t.KindIcons == nil {
		return ""
	}
	return t.KindIcons[kind]
}

// RelColor returns the stroke color for a relation type, falling back to the
// default line color when no specific color is set.
func (t *Theme) RelColor(relType string, defaultColor string) string {
	if t == nil || t.RelColors == nil {
		return defaultColor
	}
	if c, ok := t.RelColors[relType]; ok && c != "" {
		return c
	}
	return defaultColor
}

// RelStyle returns the line style name for a relation type ("solid", "dashed",
// "dotted"), falling back to "solid".
func (t *Theme) RelStyle(relType string) string {
	if t == nil || t.RelStyles == nil {
		return "solid"
	}
	if s, ok := t.RelStyles[relType]; ok && s != "" {
		return s
	}
	return "solid"
}

// Typography defines font definitions.
type Typography struct {
	FontFamily  string `yaml:"font_family,omitempty"`
	FontSize    int    `yaml:"font_size,omitempty"`
	HeadingSize int    `yaml:"heading_size,omitempty"`
	CodeFont    string `yaml:"code_font,omitempty"`
}

// LineStyles defines line style definitions.
type LineStyles struct {
	Default    *LineStyle `yaml:"default,omitempty"`
	Connection *LineStyle `yaml:"connection,omitempty"`
	Ownership  *LineStyle `yaml:"ownership,omitempty"`
	Dependency *LineStyle `yaml:"dependency,omitempty"`
}

// LineStyle defines a single line style.
type LineStyle struct {
	Color string  `yaml:"color,omitempty"`
	Width float64 `yaml:"width,omitempty"`
	Style string  `yaml:"style,omitempty"`
}

// LayoutConfig configures spatial arrangement.
type LayoutConfig struct {
	Type      string  `yaml:"type,omitempty"`
	Direction string  `yaml:"direction,omitempty"`
	Spacing   float64 `yaml:"spacing,omitempty"`
	Padding   float64 `yaml:"padding,omitempty"`
	Alignment string  `yaml:"alignment,omitempty"`
}

// Position represents a 2D position.
type Position struct {
	X float64
	Y float64
}

// NodePosition represents a positioned node in the layout.
type NodePosition struct {
	ID       string
	Position Position
	Width    float64
	Height   float64
	Children []string
}

// EdgePosition represents a positioned edge in the layout.
type EdgePosition struct {
	ID     string
	Source string
	Target string
	Points []Position
}

// LayoutResult represents the result of a layout computation.
type LayoutResult struct {
	Nodes  []NodePosition
	Edges  []EdgePosition
	Width  float64
	Height float64
	// Meta carries layout-specific auxiliary data (e.g. the geographic frame of
	// a map layout) that renderers may use for decorations such as axis labels.
	Meta map[string]string
}

// NewArtifact creates a new Artifact.
func NewArtifact(id, rendererID, viewID, format, content string) *Artifact {
	return &Artifact{
		ID:         id,
		RendererID: rendererID,
		ViewID:     viewID,
		Format:     format,
		Content:    content,
		Metadata:   make(map[string]interface{}),
		Timestamp:  time.Now().UTC().Format(time.RFC3339),
	}
}

// NewRenderOptions creates new RenderOptions with defaults.
func NewRenderOptions() *RenderOptions {
	return &RenderOptions{
		Width:  800,
		Height: 600,
	}
}
