package mcp

import (
	"context"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

func TestRenderGraphMarkdown(t *testing.T) {
	sm := NewSessionManager()
	_ = seedTestGraph(t, sm)
	s := NewMCPServer(sm)

	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      "render_graph",
			Arguments: map[string]interface{}{"format": "markdown"},
		},
	}
	res, err := s.GetTool("render_graph").Handler(context.Background(), req)
	if err != nil {
		t.Fatalf("handler error: %v", err)
	}
	if res.IsError {
		t.Fatalf("expected success, got: %+v", res)
	}
	text := toolResultText(t, res)
	for _, want := range []string{"# Infrastructure View", "## Entities", "srv-01", "Server 01", "## Relations", "rel-hosts"} {
		if !strings.Contains(text, want) {
			t.Errorf("expected %q in markdown render:\n%s", want, text)
		}
	}
}

func TestRenderGraphMermaid(t *testing.T) {
	sm := NewSessionManager()
	_ = seedTestGraph(t, sm)
	s := NewMCPServer(sm)

	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      "render_graph",
			Arguments: map[string]interface{}{"format": "mermaid"},
		},
	}
	res, err := s.GetTool("render_graph").Handler(context.Background(), req)
	if err != nil {
		t.Fatalf("handler error: %v", err)
	}
	if res.IsError {
		t.Fatalf("expected success, got: %+v", res)
	}
	text := toolResultText(t, res)
	for _, want := range []string{"graph TB", `srv_01["Server 01"]`, "srv_01 -->|hosts| vm_01"} {
		if !strings.Contains(text, want) {
			t.Errorf("expected %q in mermaid render:\n%s", want, text)
		}
	}
}

func TestRenderGraphKindFilter(t *testing.T) {
	sm := NewSessionManager()
	_ = seedTestGraph(t, sm)
	s := NewMCPServer(sm)

	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "render_graph",
			Arguments: map[string]interface{}{
				"format": "markdown",
				"kinds":  []string{"server"},
			},
		},
	}
	res, err := s.GetTool("render_graph").Handler(context.Background(), req)
	if err != nil {
		t.Fatalf("handler error: %v", err)
	}
	if res.IsError {
		t.Fatalf("expected success, got: %+v", res)
	}
	text := toolResultText(t, res)
	if !strings.Contains(text, "- **Server 01** (`srv-01`, server)") {
		t.Errorf("expected server entity in filtered render:\n%s", text)
	}
	if strings.Contains(text, "(`vm-01`") {
		t.Errorf("did not expect vm entity in server-only render:\n%s", text)
	}
}

func TestRenderGraphViewYAML(t *testing.T) {
	sm := NewSessionManager()
	_ = seedTestGraph(t, sm)
	s := NewMCPServer(sm)

	viewYAML := "id: custom-view\nname: Custom View\nvisibility:\n  - target: entities\n    kind: server\n    action: show\n"
	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "render_graph",
			Arguments: map[string]interface{}{
				"format":    "markdown",
				"view_yaml": viewYAML,
			},
		},
	}
	res, err := s.GetTool("render_graph").Handler(context.Background(), req)
	if err != nil {
		t.Fatalf("handler error: %v", err)
	}
	if res.IsError {
		t.Fatalf("expected success, got: %+v", res)
	}
	text := toolResultText(t, res)
	if !strings.Contains(text, "# Custom View") {
		t.Errorf("expected custom view title in render:\n%s", text)
	}
	if !strings.Contains(text, "srv-01") {
		t.Errorf("expected server in custom view render:\n%s", text)
	}
}
