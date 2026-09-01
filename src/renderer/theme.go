package renderer

import (
	"github.com/bababa/Niigata_Real_IaC/src/core/kinds"
)

// DefaultTheme returns the Niigata default light theme. It assigns a distinct
// accent color and a universal glyph to every core entity kind, and maps the
// core relation types to line colors and styles.
func DefaultTheme() *Theme {
	return &Theme{
		ID:   "default",
		Name: "Niigata Default (Light)",
		Colors: &ColorPalette{
			Primary:    "#2563eb",
			Secondary:  "#7c3aed",
			Background: "#ffffff",
			Surface:    "#f8fafc",
			Text:       "#0f172a",
			Border:     "#cbd5e1",
			Success:    "#16a34a",
			Warning:    "#f59e0b",
			Error:      "#dc2626",
			Info:       "#0891b2",
			KindColors: map[string]string{
				string(kinds.Area):          "#2563eb",
				string(kinds.Terrain):       "#ea580c",
				string(kinds.Ground):        "#a16207",
				string(kinds.WaterBody):     "#0891b2",
				string(kinds.Forest):        "#16a34a",
				string(kinds.Species):       "#0d9488",
				string(kinds.Population):    "#7c3aed",
				string(kinds.TourismSpot):   "#e11d48",
				string(kinds.HotSpring):     "#c2410c",
				string(kinds.CulturalAsset): "#ca8a04",
				string(kinds.Event):         "#db2777",
			},
		},
		Typography: &Typography{
			FontFamily:  "system-ui, -apple-system, 'Segoe UI', Roboto, sans-serif",
			FontSize:    12,
			HeadingSize: 18,
			CodeFont:    "ui-monospace, SFMono-Regular, Menlo, monospace",
		},
		Lines: &LineStyles{
			Default:    &LineStyle{Color: "#94a3b8", Width: 1.5, Style: "solid"},
			Ownership:  &LineStyle{Color: "#94a3b8", Width: 1.5, Style: "solid"},
			Connection: &LineStyle{Color: "#64748b", Width: 1.5, Style: "dashed"},
			Dependency: &LineStyle{Color: "#f59e0b", Width: 1.5, Style: "dashed"},
		},
		KindIcons: map[string]string{
			string(kinds.Area):          "🏝",
			string(kinds.Terrain):       "⛰",
			string(kinds.Ground):        "🌋",
			string(kinds.WaterBody):     "💧",
			string(kinds.Forest):        "🌲",
			string(kinds.Species):       "🐾",
			string(kinds.Population):    "📊",
			string(kinds.TourismSpot):   "🏛",
			string(kinds.HotSpring):     "♨",
			string(kinds.CulturalAsset): "🏺",
			string(kinds.Event):         "🎉",
		},
		RelColors: map[string]string{
			"belongs_to": "#94a3b8",
			"located_in": "#64748b",
			"inhabits":   "#0d9488",
			"flows_into": "#0891b2",
			"near":       "#7c3aed",
			"depends_on": "#f59e0b",
		},
		RelStyles: map[string]string{
			"belongs_to": "solid",
			"located_in": "solid",
			"inhabits":   "dashed",
			"flows_into": "dashed",
			"near":       "dotted",
			"depends_on": "dashed",
		},
	}
}

// DarkTheme returns a dark variant of the Niigata theme for use on dark
// terminals and modern dashboards. Kind colors are brightened for contrast, and
// the background/surface/text colors are inverted.
func DarkTheme() *Theme {
	t := DefaultTheme()
	t.ID = "dark"
	t.Name = "Niigata Dark"
	t.Colors.Background = "#0f172a"
	t.Colors.Surface = "#1e293b"
	t.Colors.Text = "#f1f5f9"
	t.Colors.Border = "#334155"
	t.Colors.KindColors = map[string]string{
		string(kinds.Area):          "#60a5fa",
		string(kinds.Terrain):       "#fb923c",
		string(kinds.Ground):        "#d97706",
		string(kinds.WaterBody):     "#22d3ee",
		string(kinds.Forest):        "#4ade80",
		string(kinds.Species):       "#2dd4bf",
		string(kinds.Population):    "#a78bfa",
		string(kinds.TourismSpot):   "#fb7185",
		string(kinds.HotSpring):     "#f97316",
		string(kinds.CulturalAsset): "#facc15",
		string(kinds.Event):         "#f472b6",
	}
	t.Lines.Default = &LineStyle{Color: "#475569", Width: 1.5, Style: "solid"}
	t.Lines.Ownership = &LineStyle{Color: "#64748b", Width: 1.5, Style: "solid"}
	t.Lines.Connection = &LineStyle{Color: "#94a3b8", Width: 1.5, Style: "dashed"}
	t.Lines.Dependency = &LineStyle{Color: "#fbbf24", Width: 1.5, Style: "dashed"}
	return t
}

// LookupTheme returns a built-in theme by ID ("default" or "dark"), or nil when
// the ID is unknown.
func LookupTheme(id string) *Theme {
	switch id {
	case "", "default":
		return DefaultTheme()
	case "dark":
		return DarkTheme()
	default:
		return nil
	}
}

// ResolveTheme returns opts.Theme when set, otherwise the built-in default
// theme. Renderers call this so a nil theme yields the polished default look
// instead of an unstyled white canvas.
func ResolveTheme(opts *RenderOptions) *Theme {
	if opts != nil && opts.Theme != nil {
		return opts.Theme
	}
	return DefaultTheme()
}

// FontFamily returns the theme font stack, falling back to a safe default.
func (t *Theme) FontFamily() string {
	if t == nil || t.Typography == nil || t.Typography.FontFamily == "" {
		return "system-ui, -apple-system, 'Segoe UI', Roboto, sans-serif"
	}
	return t.Typography.FontFamily
}

// FontSize returns the base font size, falling back to 12.
func (t *Theme) FontSize() int {
	if t == nil || t.Typography == nil || t.Typography.FontSize <= 0 {
		return 12
	}
	return t.Typography.FontSize
}

// BackgroundColor returns the canvas background color.
func (t *Theme) BackgroundColor() string {
	if t == nil || t.Colors == nil {
		return "#ffffff"
	}
	if t.Colors.Background != "" {
		return t.Colors.Background
	}
	return "#ffffff"
}

// TextColor returns the default text color.
func (t *Theme) TextColor() string {
	if t == nil || t.Colors == nil {
		return "#111827"
	}
	if t.Colors.Text != "" {
		return t.Colors.Text
	}
	return "#111827"
}

// ContainerFill returns the fill used for grouped container nodes: a soft
// surface tint that keeps children readable.
func (t *Theme) ContainerFill() string {
	if t == nil || t.Colors == nil {
		return "#f3f4f6"
	}
	if t.Colors.Surface != "" {
		return t.Colors.Surface
	}
	return "#f3f4f6"
}

// AccentColor returns the accent color for an entity kind, falling back to the
// palette primary.
func (t *Theme) AccentColor(kind string) string {
	if t == nil || t.Colors == nil {
		return "#2563eb"
	}
	if c := t.Colors.KindColor(kind); c != "" {
		return c
	}
	if t.Colors.Primary != "" {
		return t.Colors.Primary
	}
	return "#2563eb"
}

// LineColor returns the stroke color for a relation type, resolving theme and
// per-type overrides with the palette default line color as the fallback.
func (t *Theme) LineColor(relType string) string {
	fallback := "#64748b"
	if t != nil && t.Lines != nil && t.Lines.Default != nil && t.Lines.Default.Color != "" {
		fallback = t.Lines.Default.Color
	}
	if t == nil {
		return fallback
	}
	return t.RelColor(relType, fallback)
}

// LineWidth returns the stroke width for a relation type.
func (t *Theme) LineWidth(relType string) float64 {
	if t != nil && t.Lines != nil && t.Lines.Default != nil && t.Lines.Default.Width > 0 {
		return t.Lines.Default.Width
	}
	return 1.5
}

// LineDash returns the SVG stroke-dasharray for a relation type ("", "6 4", or
// "2 3"), following the theme line style.
func (t *Theme) LineDash(relType string) string {
	style := ""
	if t != nil {
		style = t.RelStyle(relType)
	}
	switch style {
	case "dashed":
		return "6 4"
	case "dotted":
		return "2 3"
	default:
		return ""
	}
}

// labelWithIcon joins a kind glyph and a text label without leaving a stray
// leading or duplicated space when the kind has no icon.
func labelWithIcon(theme *Theme, kind, name string) string {
	icon := theme.KindIcon(kind)
	if icon == "" {
		return name
	}
	return icon + " " + name
}
