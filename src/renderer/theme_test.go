package renderer

import (
	"testing"

	"github.com/bababa/Niigata_Real_IaC/src/core/kinds"
)

func TestDefaultThemeCoversCoreKinds(t *testing.T) {
	theme := DefaultTheme()
	coreKinds := []string{
		string(kinds.Area),
		string(kinds.Terrain),
		string(kinds.Ground),
		string(kinds.WaterBody),
		string(kinds.Forest),
		string(kinds.Species),
		string(kinds.Population),
		string(kinds.TourismSpot),
		string(kinds.HotSpring),
		string(kinds.CulturalAsset),
		string(kinds.Event),
	}
	for _, kind := range coreKinds {
		if theme.Colors.KindColors[kind] == "" {
			t.Errorf("expected a color for kind %q in the default theme", kind)
		}
		if theme.KindIcons[kind] == "" {
			t.Errorf("expected an icon for kind %q in the default theme", kind)
		}
	}
}

func TestResolveTheme(t *testing.T) {
	if got := ResolveTheme(nil); got == nil || got.ID != "default" {
		t.Errorf("ResolveTheme(nil) = %+v, expected default theme", got)
	}

	custom := &Theme{ID: "custom", Name: "Custom"}
	opts := &RenderOptions{Theme: custom}
	if got := ResolveTheme(opts); got != custom {
		t.Errorf("ResolveTheme with opts should return the provided theme")
	}
}

func TestLookupTheme(t *testing.T) {
	if got := LookupTheme("dark"); got == nil || got.ID != "dark" {
		t.Errorf("LookupTheme('dark') = %+v, expected dark theme", got)
	}
	if got := LookupTheme("default"); got == nil || got.ID != "default" {
		t.Errorf("LookupTheme('default') = %+v, expected default theme", got)
	}
	if got := LookupTheme("missing"); got != nil {
		t.Errorf("LookupTheme('missing') = %+v, expected nil", got)
	}
}

func TestKindColorFallback(t *testing.T) {
	cp := &ColorPalette{Primary: "#123456"}
	if got := cp.KindColor("nonexistent"); got != "#123456" {
		t.Errorf("expected fallback to primary, got %q", got)
	}
	cp.KindColors = map[string]string{"species": "#abcdef"}
	if got := cp.KindColor("species"); got != "#abcdef" {
		t.Errorf("expected explicit kind color, got %q", got)
	}
}

func TestRelStyleAndDash(t *testing.T) {
	theme := DefaultTheme()
	if theme.LineDash("inhabits") != "6 4" {
		t.Errorf("expected dashed dasharray for inhabits, got %q", theme.LineDash("inhabits"))
	}
	if theme.LineDash("near") != "2 3" {
		t.Errorf("expected dotted dasharray for near, got %q", theme.LineDash("near"))
	}
	if theme.LineDash("belongs_to") != "" {
		t.Errorf("expected solid (empty) dasharray for belongs_to, got %q", theme.LineDash("belongs_to"))
	}
	if theme.LineDash("unknown-type") != "" {
		t.Errorf("expected solid fallback for unknown type, got %q", theme.LineDash("unknown-type"))
	}
}

func TestLabelWithIcon(t *testing.T) {
	theme := DefaultTheme()
	if got := labelWithIcon(theme, "area", "Sado Island"); got != "🏝 Sado Island" {
		t.Errorf("expected icon before name, got %q", got)
	}
	if got := labelWithIcon(theme, "server", "Server 1"); got != "Server 1" {
		t.Errorf("expected no stray space for unknown kind, got %q", got)
	}
}

func TestDarkThemeBrightensColors(t *testing.T) {
	dark := DarkTheme()
	if dark.Colors.Background == DefaultTheme().Colors.Background {
		t.Errorf("expected dark theme to change the background color")
	}
	if dark.AccentColor(string(kinds.Area)) == DefaultTheme().AccentColor(string(kinds.Area)) {
		t.Errorf("expected dark theme to brighten the area accent color")
	}
}
