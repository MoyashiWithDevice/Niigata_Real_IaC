package renderer

import (
	"time"

	"IACForge/src/view"
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
