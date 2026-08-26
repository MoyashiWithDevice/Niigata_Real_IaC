package renderer

import (
	"strings"
	"testing"

	"IACForge/src/core"
	"IACForge/src/view"
)

func TestNewSVGRenderer(t *testing.T) {
	r := NewSVGRenderer()
	if r.ID() != "svg" {
		t.Errorf("expected ID 'svg', got '%s'", r.ID())
	}
	if r.Name() != "SVG Renderer" {
		t.Errorf("expected Name 'SVG Renderer', got '%s'", r.Name())
	}
	if r.Format() != "svg" {
		t.Errorf("expected Format 'svg', got '%s'", r.Format())
	}
}

func TestNewMermaidRenderer(t *testing.T) {
	r := NewMermaidRenderer()
	if r.ID() != "mermaid" {
		t.Errorf("expected ID 'mermaid', got '%s'", r.ID())
	}
	if r.Format() != "mmd" {
		t.Errorf("expected Format 'mmd', got '%s'", r.Format())
	}
}

func TestNewMarkdownRenderer(t *testing.T) {
	r := NewMarkdownRenderer()
	if r.ID() != "markdown" {
		t.Errorf("expected ID 'markdown', got '%s'", r.ID())
	}
	if r.Format() != "md" {
		t.Errorf("expected Format 'md', got '%s'", r.Format())
	}
}

func TestNewJSONRenderer(t *testing.T) {
	r := NewJSONRenderer()
	if r.ID() != "json" {
		t.Errorf("expected ID 'json', got '%s'", r.ID())
	}
	if r.Format() != "json" {
		t.Errorf("expected Format 'json', got '%s'", r.Format())
	}
}

func TestNewArtifact(t *testing.T) {
	artifact := NewArtifact("art-1", "svg", "view-1", "svg", "<svg></svg>")
	if artifact.ID != "art-1" {
		t.Errorf("expected ID 'art-1', got '%s'", artifact.ID)
	}
	if artifact.RendererID != "svg" {
		t.Errorf("expected RendererID 'svg', got '%s'", artifact.RendererID)
	}
	if artifact.Timestamp == "" {
		t.Error("expected non-empty timestamp")
	}
}

func TestNewRenderOptions(t *testing.T) {
	opts := NewRenderOptions()
	if opts.Width != 800 {
		t.Errorf("expected Width 800, got %f", opts.Width)
	}
	if opts.Height != 600 {
		t.Errorf("expected Height 600, got %f", opts.Height)
	}
}

func TestSVGRendererRender(t *testing.T) {
	g := core.NewGraph()
	e1 := core.NewEntity("srv-1", "server", "Server 1")
	if err := g.AddEntity(e1); err != nil {
		t.Fatalf("failed to add: %v", err)
	}

	e2 := core.NewEntity("srv-2", "server", "Server 2")
	if err := g.AddEntity(e2); err != nil {
		t.Fatalf("failed to add: %v", err)
	}

	v := view.NewView("test-view", "Test View")
	engine := view.NewEngine(g)
	result, err := engine.Apply(v)
	if err != nil {
		t.Fatalf("failed to apply view: %v", err)
	}

	renderer := NewSVGRenderer()
	artifact, err := renderer.Render(result, nil)
	if err != nil {
		t.Fatalf("failed to render: %v", err)
	}

	if artifact.Format != "svg" {
		t.Errorf("expected format 'svg', got '%s'", artifact.Format)
	}
	if len(artifact.Content) == 0 {
		t.Error("expected non-empty content")
	}
}

func TestMermaidRendererRender(t *testing.T) {
	g := core.NewGraph()
	e1 := core.NewEntity("srv-1", "server", "Server 1")
	if err := g.AddEntity(e1); err != nil {
		t.Fatalf("failed to add: %v", err)
	}

	e2 := core.NewEntity("srv-2", "server", "Server 2")
	if err := g.AddEntity(e2); err != nil {
		t.Fatalf("failed to add: %v", err)
	}

	r1 := core.NewDirectedRelation("rel-1", "connects", "srv-1", "srv-2")
	if err := g.AddRelation(r1); err != nil {
		t.Fatalf("failed to add: %v", err)
	}

	v := view.NewView("test-view", "Test View")
	engine := view.NewEngine(g)
	result, err := engine.Apply(v)
	if err != nil {
		t.Fatalf("failed to apply view: %v", err)
	}

	renderer := NewMermaidRenderer()
	artifact, err := renderer.Render(result, nil)
	if err != nil {
		t.Fatalf("failed to render: %v", err)
	}

	if artifact.Format != "mmd" {
		t.Errorf("expected format 'mmd', got '%s'", artifact.Format)
	}
	if len(artifact.Content) == 0 {
		t.Error("expected non-empty content")
	}
}

func TestMermaidRendererSymmetricRelation(t *testing.T) {
	g := core.NewGraph()
	e1 := core.NewEntity("eth0-01", "interface", "eth0")
	if err := g.AddEntity(e1); err != nil {
		t.Fatalf("failed to add: %v", err)
	}

	e2 := core.NewEntity("eth0-02", "interface", "eth0")
	if err := g.AddEntity(e2); err != nil {
		t.Fatalf("failed to add: %v", err)
	}

	r1 := core.NewSymmetricRelation("rel-conn", "connects", []string{"eth0-01", "eth0-02"})
	if err := g.AddRelation(r1); err != nil {
		t.Fatalf("failed to add relation: %v", err)
	}

	v := view.NewView("test-view", "Test View")
	engine := view.NewEngine(g)
	result, err := engine.Apply(v)
	if err != nil {
		t.Fatalf("failed to apply view: %v", err)
	}

	renderer := NewMermaidRenderer()
	artifact, err := renderer.Render(result, nil)
	if err != nil {
		t.Fatalf("failed to render: %v", err)
	}

	content := artifact.Content
	if !strings.Contains(content, "eth0_01 -->|connects| eth0_02") {
		t.Errorf("expected symmetric edge between participant endpoints:\n%s", content)
	}
	if strings.Contains(content, "-->|connects| \n") {
		t.Errorf("found dangling edge without endpoints:\n%s", content)
	}
}

func TestMarkdownRendererRender(t *testing.T) {
	g := core.NewGraph()
	e1 := core.NewEntity("srv-1", "server", "Server 1")
	if err := g.AddEntity(e1); err != nil {
		t.Fatalf("failed to add: %v", err)
	}

	v := view.NewView("test-view", "Test View")
	engine := view.NewEngine(g)
	result, err := engine.Apply(v)
	if err != nil {
		t.Fatalf("failed to apply view: %v", err)
	}

	renderer := NewMarkdownRenderer()
	artifact, err := renderer.Render(result, nil)
	if err != nil {
		t.Fatalf("failed to render: %v", err)
	}

	if artifact.Format != "md" {
		t.Errorf("expected format 'md', got '%s'", artifact.Format)
	}
	if len(artifact.Content) == 0 {
		t.Error("expected non-empty content")
	}
}

func TestJSONRendererRender(t *testing.T) {
	g := core.NewGraph()
	e1 := core.NewEntity("srv-1", "server", "Server 1")
	if err := g.AddEntity(e1); err != nil {
		t.Fatalf("failed to add: %v", err)
	}

	v := view.NewView("test-view", "Test View")
	engine := view.NewEngine(g)
	result, err := engine.Apply(v)
	if err != nil {
		t.Fatalf("failed to apply view: %v", err)
	}

	renderer := NewJSONRenderer()
	artifact, err := renderer.Render(result, nil)
	if err != nil {
		t.Fatalf("failed to render: %v", err)
	}

	if artifact.Format != "json" {
		t.Errorf("expected format 'json', got '%s'", artifact.Format)
	}
	if len(artifact.Content) == 0 {
		t.Error("expected non-empty content")
	}
}

func TestLayoutEngineHierarchical(t *testing.T) {
	g := core.NewGraph()
	e1 := core.NewEntity("region-1", "region", "Region 1")
	if err := g.AddEntity(e1); err != nil {
		t.Fatalf("failed to add: %v", err)
	}

	e2 := core.NewEntity("rack-1", "rack", "Rack 1")
	e2.SetOwner("region-1")
	if err := g.AddEntity(e2); err != nil {
		t.Fatalf("failed to add: %v", err)
	}

	v := view.NewView("test-view", "Test View")
	engine := view.NewEngine(g)
	result, err := engine.Apply(v)
	if err != nil {
		t.Fatalf("failed to apply view: %v", err)
	}

	layoutEngine := NewLayoutEngine(&LayoutConfig{
		Type:      "hierarchical",
		Direction: "top-down",
		Spacing:   50,
		Padding:   20,
	})

	layout := layoutEngine.ComputeLayout(result)
	if len(layout.Nodes) != 2 {
		t.Errorf("expected 2 nodes, got %d", len(layout.Nodes))
	}
	if layout.Width <= 0 {
		t.Errorf("expected positive width, got %f", layout.Width)
	}
	if layout.Height <= 0 {
		t.Errorf("expected positive height, got %f", layout.Height)
	}
}

func TestLayoutEngineForceDirected(t *testing.T) {
	g := core.NewGraph()
	e1 := core.NewEntity("srv-1", "server", "Server 1")
	if err := g.AddEntity(e1); err != nil {
		t.Fatalf("failed to add: %v", err)
	}

	e2 := core.NewEntity("srv-2", "server", "Server 2")
	if err := g.AddEntity(e2); err != nil {
		t.Fatalf("failed to add: %v", err)
	}

	v := view.NewView("test-view", "Test View")
	engine := view.NewEngine(g)
	result, err := engine.Apply(v)
	if err != nil {
		t.Fatalf("failed to apply view: %v", err)
	}

	layoutEngine := NewLayoutEngine(&LayoutConfig{
		Type:    "force-directed",
		Spacing: 100,
		Padding: 20,
	})

	layout := layoutEngine.ComputeLayout(result)
	if len(layout.Nodes) != 2 {
		t.Errorf("expected 2 nodes, got %d", len(layout.Nodes))
	}
}

func TestSVGRendererRenderWithTheme(t *testing.T) {
	g := core.NewGraph()
	e1 := core.NewEntity("srv-1", "server", "Server 1")
	if err := g.AddEntity(e1); err != nil {
		t.Fatalf("failed to add: %v", err)
	}

	v := view.NewView("test-view", "Test View")
	engine := view.NewEngine(g)
	result, err := engine.Apply(v)
	if err != nil {
		t.Fatalf("failed to apply view: %v", err)
	}

	theme := &Theme{
		ID:   "dark",
		Name: "Dark Theme",
		Colors: &ColorPalette{
			Background: "#ffffff",
		},
	}
	opts := &RenderOptions{
		Width:  1024,
		Height: 768,
		Theme:  theme,
	}

	renderer := NewSVGRenderer()
	artifact, err := renderer.Render(result, opts)
	if err != nil {
		t.Fatalf("failed to render: %v", err)
	}

	if artifact.Format != "svg" {
		t.Errorf("expected format 'svg', got '%s'", artifact.Format)
	}
}

func TestSanitizeID(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"srv-1", "srv_1"},
		{"my.server", "my_server"},
		{"path/to/file", "path_to_file"},
	}

	for _, test := range tests {
		result := sanitizeID(test.input)
		if result != test.expected {
			t.Errorf("sanitizeID(%s) = %s, expected %s", test.input, result, test.expected)
		}
	}
}

func TestEscapeXML(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"hello", "hello"},
		{"<script>", "&lt;script&gt;"},
		{`"quoted"`, "&quot;quoted&quot;"},
	}

	for _, test := range tests {
		result := escapeXML(test.input)
		if result != test.expected {
			t.Errorf("escapeXML(%s) = %s, expected %s", test.input, result, test.expected)
		}
	}
}

func TestEscapeMermaid(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"hello", "hello"},
		{`"quoted"`, "'quoted'"},
		{"line\nbreak", "line break"},
	}

	for _, test := range tests {
		result := escapeMermaid(test.input)
		if result != test.expected {
			t.Errorf("escapeMermaid(%s) = %s, expected %s", test.input, result, test.expected)
		}
	}
}

func TestRenderFormat(t *testing.T) {
	g := core.NewGraph()
	if err := g.AddEntity(core.NewEntity("srv-1", "server", "Server 1")); err != nil {
		t.Fatalf("failed to add: %v", err)
	}

	v := view.NewView("test-view", "Test View")
	engine := view.NewEngine(g)
	result, err := engine.Apply(v)
	if err != nil {
		t.Fatalf("failed to apply view: %v", err)
	}

	mermaid, err := RenderFormat(result, "mermaid")
	if err != nil {
		t.Fatalf("mermaid render failed: %v", err)
	}
	if !strings.HasPrefix(mermaid, "graph ") {
		t.Errorf("expected mermaid graph, got:\n%s", mermaid)
	}

	for _, format := range []string{"markdown", "md"} {
		md, err := RenderFormat(result, format)
		if err != nil {
			t.Fatalf("%s render failed: %v", format, err)
		}
		if !strings.Contains(md, "srv-1") {
			t.Errorf("expected srv-1 in %s output:\n%s", format, md)
		}
	}
}

func TestRenderFormatUnknown(t *testing.T) {
	v := view.NewView("test-view", "Test View")
	engine := view.NewEngine(core.NewGraph())
	result, err := engine.Apply(v)
	if err != nil {
		t.Fatalf("failed to apply view: %v", err)
	}

	if _, err := RenderFormat(result, "xml"); err == nil {
		t.Error("expected error for unknown format")
	} else if !strings.Contains(err.Error(), "unknown render format") {
		t.Errorf("expected unknown render format error, got: %v", err)
	}
}

func TestMermaidRendererSkipsDanglingEdge(t *testing.T) {
	e1 := core.NewEntity("srv-1", "server", "Server 1")
	e2 := core.NewEntity("srv-2", "server", "Server 2")
	rel := core.NewDirectedRelation("rel-1", "connects", "srv-1", "missing")

	vr := &view.ViewResult{
		ViewID:           "dangling",
		Title:            "Dangling",
		VisibleEntities:  []*core.Entity{e1, e2},
		VisibleRelations: []*core.Relation{rel},
		Annotations:      make(map[string]map[string]interface{}),
	}

	artifact, err := NewMermaidRenderer().Render(vr, nil)
	if err != nil {
		t.Fatalf("failed to render: %v", err)
	}

	content := artifact.Content
	if strings.Contains(content, "missing") {
		t.Errorf("found dangling edge referencing missing node:\n%s", content)
	}
}

func ownershipViewResult() *view.ViewResult {
	region := core.NewEntity("region-1", "region", "Region 1")
	rack := core.NewEntity("rack-1", "rack", "Rack 1")
	rack.SetOwner("region-1")
	server := core.NewEntity("srv-1", "server", "Server 1")
	server.SetOwner("rack-1")

	return &view.ViewResult{
		ViewID:          "hierarchy",
		Title:           "Hierarchy",
		VisibleEntities: []*core.Entity{region, rack, server},
		Annotations:     make(map[string]map[string]interface{}),
	}
}

func TestMarkdownRendererOwnershipHierarchy(t *testing.T) {
	vr := ownershipViewResult()

	artifact, err := NewMarkdownRenderer().Render(vr, nil)
	if err != nil {
		t.Fatalf("failed to render: %v", err)
	}

	content := artifact.Content
	if !strings.Contains(content, "- **Region 1** (`region-1`, region)\n") {
		t.Errorf("expected region as root bullet:\n%s", content)
	}
	if !strings.Contains(content, "\n  - **Rack 1** (`rack-1`, rack)\n") {
		t.Errorf("expected rack nested inside region:\n%s", content)
	}
	if !strings.Contains(content, "\n    - **Server 1** (`srv-1`, server)\n") {
		t.Errorf("expected server nested inside rack:\n%s", content)
	}
}

func TestMermaidRendererOwnershipSubgraphs(t *testing.T) {
	vr := ownershipViewResult()

	artifact, err := NewMermaidRenderer().Render(vr, nil)
	if err != nil {
		t.Fatalf("failed to render: %v", err)
	}

	content := artifact.Content
	regionIdx := strings.Index(content, `subgraph region_1["Region 1"]`)
	rackIdx := strings.Index(content, `subgraph rack_1["Rack 1"]`)
	serverIdx := strings.Index(content, `srv_1["Server 1"]`)
	endIdx := strings.Index(content, "end")
	if regionIdx < 0 || rackIdx < 0 || serverIdx < 0 || endIdx < 0 {
		t.Fatalf("expected nested subgraphs and node declarations:\n%s", content)
	}
	if !(regionIdx < rackIdx && rackIdx < serverIdx && serverIdx < endIdx) {
		t.Errorf("expected parents declared before children:\n%s", content)
	}
}

func TestSVGRendererContainment(t *testing.T) {
	vr := ownershipViewResult()

	layout := NewLayoutEngine(nil).ComputeLayout(vr)
	if len(layout.Nodes) != 3 {
		t.Fatalf("expected 3 nodes, got %d", len(layout.Nodes))
	}

	var region, rack NodePosition
	for _, node := range layout.Nodes {
		switch node.ID {
		case "region-1":
			region = node
		case "rack-1":
			rack = node
		case "srv-1":
			node.Children = nil
		}
	}
	if len(region.Children) == 0 {
		t.Error("expected region to contain children")
	}
	if rack.Position.X <= region.Position.X || rack.Position.Y <= region.Position.Y {
		t.Errorf("expected rack inside region box: region=%+v rack=%+v", region.Position, rack.Position)
	}
	if right := rack.Position.X + rack.Width; right > region.Position.X+region.Width {
		t.Errorf("rack exceeds region width: %f > %f", right, region.Position.X+region.Width)
	}

	artifact, err := NewSVGRenderer().Render(vr, nil)
	if err != nil {
		t.Fatalf("failed to render: %v", err)
	}

	content := artifact.Content
	regionRect := strings.Index(content, `fill="#f3f4f6"`)
	if regionRect < 0 {
		t.Fatalf("expected region container rect at origin:\n%s", content)
	}
	regionText := strings.Index(content, ">Region 1</text>")
	rackText := strings.Index(content, ">Rack 1</text>")
	serverText := strings.Index(content, ">Server 1</text>")
	if !(regionText < rackText && rackText < serverText) {
		t.Errorf("expected containers drawn before children:\n%s", content)
	}
}

func TestJSONRendererHierarchy(t *testing.T) {
	vr := ownershipViewResult()

	artifact, err := NewJSONRenderer().Render(vr, nil)
	if err != nil {
		t.Fatalf("failed to render: %v", err)
	}

	content := artifact.Content
	if !strings.Contains(content, `"hierarchy"`) {
		t.Fatalf("expected hierarchy field:\n%s", content)
	}
	regionIdx := strings.Index(content, `"id": "region-1"`)
	serverIdx := strings.Index(content, `"children"`)
	if regionIdx < 0 || serverIdx < 0 || regionIdx > serverIdx {
		t.Errorf("expected nested children under root:\n%s", content)
	}
}
