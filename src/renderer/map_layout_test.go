package renderer

import (
	"strings"
	"testing"

	"github.com/bababa/Niigata_Real_IaC/src/core"
	"github.com/bababa/Niigata_Real_IaC/src/view"
)

// mapViewResult builds a view result with two located entities (differing in
// latitude and longitude) and one entity without coordinates.
func mapViewResult() *view.ViewResult {
	north := core.NewEntity("spot-north", "tourism_spot", "North Spot")
	north.SetProperty("latitude", 38.20)
	north.SetProperty("longitude", 138.30)

	south := core.NewEntity("spot-south", "tourism_spot", "South Spot")
	south.SetProperty("latitude", 37.90)
	south.SetProperty("longitude", 138.50)

	loose := core.NewEntity("species-toki", "species", "Crested Ibis")

	return &view.ViewResult{
		ViewID:          "map",
		Title:           "Map",
		VisibleEntities: []*core.Entity{north, south, loose},
		Annotations:     make(map[string]map[string]interface{}),
	}
}

func TestComputeMapLayoutProjection(t *testing.T) {
	layout := NewLayoutEngine(&LayoutConfig{Type: "map"}).ComputeLayout(mapViewResult())

	if len(layout.Nodes) != 3 {
		t.Fatalf("expected 3 nodes, got %d", len(layout.Nodes))
	}

	var north, south NodePosition
	for _, node := range layout.Nodes {
		switch node.ID {
		case "spot-north":
			north = node
		case "spot-south":
			south = node
		}
	}

	// North up: higher latitude must map to a smaller y.
	if north.Position.Y >= south.Position.Y {
		t.Errorf("expected north spot above south spot: north y=%f south y=%f",
			north.Position.Y, south.Position.Y)
	}
	// East right: higher longitude must map to a larger x.
	if north.Position.X >= south.Position.X {
		t.Errorf("expected eastward spot to the right: north x=%f south x=%f",
			north.Position.X, south.Position.X)
	}
	// All geo nodes stay inside the map frame.
	frameX, _ := parseMetaFloat(layout.Meta, "geo_x")
	frameY, _ := parseMetaFloat(layout.Meta, "geo_y")
	frameW, _ := parseMetaFloat(layout.Meta, "geo_w")
	frameH, _ := parseMetaFloat(layout.Meta, "geo_h")
	for _, node := range layout.Nodes {
		if node.ID == "species-toki" {
			continue
		}
		if node.Position.X < frameX || node.Position.Y < frameY ||
			node.Position.X+node.Width > frameX+frameW ||
			node.Position.Y+node.Height > frameY+frameH {
			t.Errorf("node %s escapes the map frame: %+v", node.ID, node.Position)
		}
	}
}

func TestComputeMapLayoutFallbackRow(t *testing.T) {
	layout := NewLayoutEngine(&LayoutConfig{Type: "map"}).ComputeLayout(mapViewResult())

	rowY, ok := parseMetaFloat(layout.Meta, "fallback_row_y")
	if !ok {
		t.Fatal("expected fallback_row_y metadata")
	}
	frameBottom, _ := parseMetaFloat(layout.Meta, "geo_y")
	frameH, _ := parseMetaFloat(layout.Meta, "geo_h")
	frameBottom += frameH

	var loose *NodePosition
	for i := range layout.Nodes {
		if layout.Nodes[i].ID == "species-toki" {
			loose = &layout.Nodes[i]
		}
	}
	if loose == nil {
		t.Fatal("expected the uncoordinated entity to be placed")
	}
	if loose.Position.Y < frameBottom {
		t.Errorf("expected uncoordinated entity below the map frame: y=%f frame bottom=%f",
			loose.Position.Y, frameBottom)
	}
	if loose.Position.Y != rowY {
		t.Errorf("expected uncoordinated entity on the fallback row: y=%f row y=%f",
			loose.Position.Y, rowY)
	}
}

func TestComputeMapLayoutMeta(t *testing.T) {
	layout := NewLayoutEngine(&LayoutConfig{Type: "map"}).ComputeLayout(mapViewResult())

	for _, key := range []string{"geo_min_lat", "geo_max_lat", "geo_min_lon", "geo_max_lon"} {
		if layout.Meta[key] == "" {
			t.Errorf("expected %s in map layout metadata", key)
		}
	}
	if layout.Meta["geo_min_lat"] != "37.90" || layout.Meta["geo_max_lat"] != "38.20" {
		t.Errorf("unexpected latitude bounds: %v", layout.Meta)
	}
}

func TestSVGRendererMapLayout(t *testing.T) {
	artifact, err := NewSVGRenderer().Render(mapViewResult(), &RenderOptions{
		Layout: &LayoutConfig{Type: "map"},
	})
	if err != nil {
		t.Fatalf("failed to render: %v", err)
	}

	content := artifact.Content
	for _, want := range []string{
		`stroke-dasharray="4 4"`, // map frame
		"N 38.20°",
		"S 37.90°",
		"E 138.50°",
		"W 138.30°",
		"no coordinates",
		"🐾 Crested Ibis",
	} {
		if !strings.Contains(content, want) {
			t.Errorf("expected %q in map SVG:\n%s", want, content)
		}
	}
}

func TestGeoCoordinateParsing(t *testing.T) {
	tests := []struct {
		name  string
		value interface{}
		want  float64
		ok    bool
	}{
		{"float", 38.04, 38.04, true},
		{"int", 38, 38, true},
		{"numeric string", "37.99", 37.99, true},
		{"non-numeric string", "north", 0, false},
		{"nil", nil, 0, false},
	}
	for _, tt := range tests {
		e := core.NewEntity("e", "tourism_spot", "E")
		e.SetProperty("latitude", tt.value)
		e.SetProperty("longitude", 138.0)
		got, ok := geoCoordinate(e, "latitude")
		if ok != tt.ok {
			t.Errorf("%s: expected ok=%v, got %v", tt.name, tt.ok, ok)
		}
		if ok && got != tt.want {
			t.Errorf("%s: expected %v, got %v", tt.name, tt.want, got)
		}
	}

	// Both coordinates must be present for an entity to count as located.
	partial := core.NewEntity("p", "tourism_spot", "P")
	partial.SetProperty("latitude", 38.0)
	if _, _, ok := entityGeoCoordinates(partial); ok {
		t.Error("expected entity with only latitude to be unlocated")
	}
}
