package mcp

import (
	"context"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"

	"IACForge/src/core"
	_ "IACForge/src/extension/builtin/aws"
)

// mockClientSession implements the subset of ClientSession needed to drive the
// session-context path of getOrCreateSession.
type mockClientSession struct {
	id string
}

func (m *mockClientSession) Initialize()       {}
func (m *mockClientSession) Initialized() bool { return true }
func (m *mockClientSession) NotificationChannel() chan<- mcp.JSONRPCNotification {
	return make(chan mcp.JSONRPCNotification, 1)
}
func (m *mockClientSession) SessionID() string { return m.id }

// TestGetOrCreateSessionWithContext verifies that a request carrying a client
// session resolves to that session's data.
func TestGetOrCreateSessionWithContext(t *testing.T) {
	sm := NewSessionManager()
	s := NewMCPServer(sm)

	ctx := context.Background()
	ctx = s.WithContext(ctx, &mockClientSession{id: "ctx-session"})

	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      "graph_summary",
			Arguments: map[string]interface{}{},
		},
	}
	res, err := s.GetTool("graph_summary").Handler(ctx, req)
	if err != nil {
		t.Fatalf("handler error: %v", err)
	}
	if res.IsError {
		t.Fatalf("expected success, got: %+v", res)
	}
	if sm.GetOrCreate("default") == sm.GetOrCreate("ctx-session") {
		t.Error("expected different session data for context session vs default")
	}
}

// TestEntityToolErrorPaths covers add/get/update/remove_entity failure branches.
func TestEntityToolErrorPaths(t *testing.T) {
	sm := NewSessionManager()
	s := NewMCPServer(sm)

	// add_entity with invalid JSON payloads.
	for _, arg := range []map[string]interface{}{
		{"id": "e1", "kind": "aws.ec2", "name": "E", "labels_json": "{bad"},
		{"id": "e1", "kind": "aws.ec2", "name": "E", "properties_json": "{bad"},
	} {
		res := callTool(t, s, "add_entity", arg)
		if !res.IsError {
			t.Errorf("expected error result for add_entity with %v", arg)
		}
	}

	// get_entity for a missing entity.
	res := callTool(t, s, "get_entity", map[string]interface{}{"id": "missing"})
	if !res.IsError {
		t.Error("expected error result for get_entity of missing entity")
	}

	// update_entity for a missing entity.
	res = callTool(t, s, "update_entity", map[string]interface{}{"id": "missing"})
	if !res.IsError {
		t.Error("expected error result for update_entity of missing entity")
	}

	// remove_entity for a missing entity.
	res = callTool(t, s, "remove_entity", map[string]interface{}{"id": "missing"})
	if !res.IsError {
		t.Error("expected error result for remove_entity of missing entity")
	}
}

// TestRelationToolErrorPaths covers add/get/update/remove_relation failure branches.
func TestRelationToolErrorPaths(t *testing.T) {
	sm := NewSessionManager()
	s := NewMCPServer(sm)

	// add_relation without a valid participant shape.
	res := callTool(t, s, "add_relation", map[string]interface{}{
		"id": "r1", "type": "connects",
	})
	if !res.IsError {
		t.Error("expected error result for add_relation without participants")
	}

	// add_relation with invalid JSON payloads.
	res = callTool(t, s, "add_relation", map[string]interface{}{
		"id": "r2", "type": "connects", "source": "a", "target": "b",
		"labels_json": "{bad",
	})
	if !res.IsError {
		t.Error("expected error result for add_relation with invalid labels_json")
	}

	res = callTool(t, s, "add_relation", map[string]interface{}{
		"id": "r3", "type": "connects", "source": "a", "target": "b",
		"properties_json": "{bad",
	})
	if !res.IsError {
		t.Error("expected error result for add_relation with invalid properties_json")
	}

	// get_relation for a missing relation.
	res = callTool(t, s, "get_relation", map[string]interface{}{"id": "missing"})
	if !res.IsError {
		t.Error("expected error result for get_relation of missing relation")
	}

	// update_relation for a missing relation.
	res = callTool(t, s, "update_relation", map[string]interface{}{"id": "missing"})
	if !res.IsError {
		t.Error("expected error result for update_relation of missing relation")
	}

	// remove_relation for a missing relation.
	res = callTool(t, s, "remove_relation", map[string]interface{}{"id": "missing"})
	if !res.IsError {
		t.Error("expected error result for remove_relation of missing relation")
	}
}

// TestQueryToolErrorPaths covers query_entities failure branches.
func TestQueryToolErrorPaths(t *testing.T) {
	sm := NewSessionManager()
	s := NewMCPServer(sm)

	res := callTool(t, s, "query_entities", map[string]interface{}{"where_json": "{bad"})
	if !res.IsError {
		t.Error("expected error result for invalid where_json")
	}

	// Limit and offset are accepted on an empty graph.
	res = callTool(t, s, "query_entities", map[string]interface{}{
		"kind": "aws.ec2", "limit": 10, "offset": 0,
	})
	if res.IsError {
		t.Fatalf("expected success with limit/offset, got: %+v", res)
	}
}

// TestParseConditionsErrors covers parseConditions failure branches.
func TestParseConditionsErrors(t *testing.T) {
	cases := []string{
		`not a list`,
		`[{"operator":"eq","value":"x"}]`,  // missing field
		`[{"field":"status","value":"x"}]`, // missing operator
	}
	for _, in := range cases {
		if _, err := parseConditions(in); err == nil {
			t.Errorf("expected error for parseConditions(%q)", in)
		}
	}
}

// TestRenderGraphFormats exercises the non-markdown renderers, grouping, and errors.
func TestRenderGraphFormats(t *testing.T) {
	sm := NewSessionManager()
	s := NewMCPServer(sm)
	_ = seedAWSGraph(t, sm, s)

	// Mermaid output.
	res := callTool(t, s, "render_graph", map[string]interface{}{"format": "mermaid"})
	if res.IsError {
		t.Fatalf("mermaid render failed: %+v", res)
	}
	if !strings.Contains(toolResultText(t, res), "graph") {
		t.Errorf("expected mermaid graph output, got: %s", toolResultText(t, res))
	}

	// SVG output.
	res = callTool(t, s, "render_graph", map[string]interface{}{"format": "svg"})
	if res.IsError {
		t.Fatalf("svg render failed: %+v", res)
	}
	if !strings.Contains(toolResultText(t, res), "<svg") {
		t.Errorf("expected svg output, got: %s", toolResultText(t, res))
	}

	// JSON output.
	res = callTool(t, s, "render_graph", map[string]interface{}{"format": "json"})
	if res.IsError {
		t.Fatalf("json render failed: %+v", res)
	}
	if !strings.Contains(toolResultText(t, res), `"entities"`) {
		t.Errorf("expected json output, got: %s", toolResultText(t, res))
	}

	// Unknown format.
	res = callTool(t, s, "render_graph", map[string]interface{}{"format": "docx"})
	if !res.IsError {
		t.Error("expected error result for unknown render format")
	}

	// Kind filter + group_by.
	res = callTool(t, s, "render_graph", map[string]interface{}{
		"kinds": []interface{}{"aws.ec2", "aws.vpc"}, "group_by": "status",
	})
	if res.IsError {
		t.Fatalf("render with kinds/group_by failed: %+v", res)
	}

	// Valid view_yaml.
	res = callTool(t, s, "render_graph", map[string]interface{}{
		"view_yaml": "id: custom\n",
	})
	if res.IsError {
		t.Fatalf("render with view_yaml failed: %+v", res)
	}

	// view_yaml without an id.
	res = callTool(t, s, "render_graph", map[string]interface{}{
		"view_yaml": "visibility: []\n",
	})
	if !res.IsError {
		t.Error("expected error result for view_yaml without id")
	}

	// Malformed view_yaml.
	res = callTool(t, s, "render_graph", map[string]interface{}{
		"view_yaml": "::not::yaml::",
	})
	if !res.IsError {
		t.Error("expected error result for malformed view_yaml")
	}
}

// TestWhoReferencesPropertyValues exercises the valueReferences branches.
func TestWhoReferencesPropertyValues(t *testing.T) {
	sm := NewSessionManager()
	s := NewMCPServer(sm)
	sd := sm.GetOrCreate("default")

	// Build a small graph with various property reference shapes.
	org := core.NewEntity("org-01", "aws.organization", "Org")
	if err := sd.Graph.AddEntity(org); err != nil {
		t.Fatalf("AddEntity org-01: %v", err)
	}

	ec2 := core.NewEntity("ec2-01", "aws.ec2", "Web")
	ec2.SetOwner("org-01")
	ec2.Properties = map[string]interface{}{
		"subnet": "@subnet-01", // string reference form
	}
	if err := sd.Graph.AddEntity(ec2); err != nil {
		t.Fatalf("AddEntity ec2-01: %v", err)
	}

	// Nested slice reference.
	lb := core.NewEntity("lb-01", "aws.load_balancer", "LB")
	lb.SetOwner("org-01")
	lb.Properties = map[string]interface{}{
		"listeners": []interface{}{"@listener-01", "plain"},
	}
	if err := sd.Graph.AddEntity(lb); err != nil {
		t.Fatalf("AddEntity lb-01: %v", err)
	}

	res := callTool(t, s, "who_references", map[string]interface{}{"id": "subnet-01"})
	if res.IsError {
		t.Fatalf("who_references failed: %+v", res)
	}
	text := toolResultText(t, res)
	if !strings.Contains(text, `"property": "subnet"`) {
		t.Errorf("expected subnet property reference in output:\n%s", text)
	}

	res = callTool(t, s, "who_references", map[string]interface{}{"id": "listener-01"})
	if res.IsError {
		t.Fatalf("who_references (slice) failed: %+v", res)
	}
	text = toolResultText(t, res)
	if !strings.Contains(text, `"property": "listeners"`) {
		t.Errorf("expected listeners property reference in output:\n%s", text)
	}
}
