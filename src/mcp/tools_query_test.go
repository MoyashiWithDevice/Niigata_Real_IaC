package mcp

import (
	"context"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"

	"IACForge/src/core"
)

// seedTestGraph populates the default session with a small infrastructure graph
// and returns the session data. Shared by query and render tool tests.
func seedTestGraph(t *testing.T, sm *SessionManager) *SessionData {
	t.Helper()
	sd := sm.GetOrCreate("default")

	g := sd.Graph
	addEntity := func(id, kind, name string, owner string) {
		t.Helper()
		e := core.NewEntity(id, core.EntityKind(kind), name)
		if owner != "" {
			e.SetOwner(owner)
		}
		if err := g.AddEntity(e); err != nil {
			t.Fatalf("failed to add entity %s: %v", id, err)
		}
	}

	addEntity("srv-01", "server", "Server 01", "")
	addEntity("srv-02", "server", "Server 02", "")
	addEntity("vm-01", "vm", "VM 01", "srv-01")
	addEntity("net-mgmt", "network", "Management Network", "srv-01")
	addEntity("eth0", "interface", "eth0", "net-mgmt")

	vlan := core.NewEntity("vlan-100", "vlan", "VLAN 100")
	vlan.SetProperty("associated_network", core.NewReferenceValue("@net-mgmt"))
	if err := g.AddEntity(vlan); err != nil {
		t.Fatalf("failed to add vlan-100: %v", err)
	}

	rel := core.NewDirectedRelation("rel-hosts", core.RelationType("hosts"), "srv-01", "vm-01")
	if err := g.AddRelation(rel); err != nil {
		t.Fatalf("failed to add relation: %v", err)
	}

	if err := g.BuildOwnershipPaths(); err != nil {
		t.Fatalf("failed to build ownership paths: %v", err)
	}
	return sd
}

func TestQueryEntities(t *testing.T) {
	sm := NewSessionManager()
	_ = seedTestGraph(t, sm)
	s := NewMCPServer(sm)

	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      "query_entities",
			Arguments: map[string]interface{}{"kind": "server"},
		},
	}
	res, err := s.GetTool("query_entities").Handler(context.Background(), req)
	if err != nil {
		t.Fatalf("handler error: %v", err)
	}
	if res.IsError {
		t.Fatalf("expected success, got: %+v", res)
	}
	text := toolResultText(t, res)
	if !strings.Contains(text, `"id": "srv-01"`) || !strings.Contains(text, `"id": "srv-02"`) {
		t.Errorf("expected both servers in query_entities output:\n%s", text)
	}
}

func TestQueryEntitiesWithWhere(t *testing.T) {
	sm := NewSessionManager()
	sd := seedTestGraph(t, sm)

	srv1, _ := sd.Graph.GetEntity("srv-01")
	srv1.SetStatus(core.StatusActive)
	srv2, _ := sd.Graph.GetEntity("srv-02")
	srv2.SetStatus(core.StatusOffline)

	s := NewMCPServer(sm)
	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "query_entities",
			Arguments: map[string]interface{}{
				"kind":       "server",
				"where_json": `[{"field":"status","operator":"eq","value":"active"}]`,
			},
		},
	}
	res, err := s.GetTool("query_entities").Handler(context.Background(), req)
	if err != nil {
		t.Fatalf("handler error: %v", err)
	}
	if res.IsError {
		t.Fatalf("expected success, got: %+v", res)
	}
	text := toolResultText(t, res)
	if !strings.Contains(text, `"id": "srv-01"`) {
		t.Errorf("expected active server in output:\n%s", text)
	}
	if strings.Contains(text, `"id": "srv-02"`) {
		t.Errorf("did not expect offline server in output:\n%s", text)
	}
}

func TestQueryRelatedDescendants(t *testing.T) {
	sm := NewSessionManager()
	_ = seedTestGraph(t, sm)
	s := NewMCPServer(sm)

	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "query_related",
			Arguments: map[string]interface{}{
				"from":      "srv-01",
				"operation": "descendants",
			},
		},
	}
	res, err := s.GetTool("query_related").Handler(context.Background(), req)
	if err != nil {
		t.Fatalf("handler error: %v", err)
	}
	if res.IsError {
		t.Fatalf("expected success, got: %+v", res)
	}
	text := toolResultText(t, res)
	for _, want := range []string{`"id": "vm-01"`, `"id": "net-mgmt"`, `"id": "eth0"`} {
		if !strings.Contains(text, want) {
			t.Errorf("expected %q in descendants output:\n%s", want, text)
		}
	}
}

func TestQueryEntitiesMermaid(t *testing.T) {
	sm := NewSessionManager()
	_ = seedTestGraph(t, sm)
	s := NewMCPServer(sm)

	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "query_entities",
			Arguments: map[string]interface{}{
				"kind":   "server",
				"format": "mermaid",
			},
		},
	}
	res, err := s.GetTool("query_entities").Handler(context.Background(), req)
	if err != nil {
		t.Fatalf("handler error: %v", err)
	}
	if res.IsError {
		t.Fatalf("expected success, got: %+v", res)
	}
	text := toolResultText(t, res)

	for _, want := range []string{"graph TB", "srv_01", "srv_02"} {
		if !strings.Contains(text, want) {
			t.Errorf("expected %q in mermaid output:\n%s", want, text)
		}
	}
	if strings.Contains(text, "vm_01") {
		t.Errorf("did not expect out-of-result entity vm_01 in mermaid output:\n%s", text)
	}
}

func TestQueryRelatedMarkdown(t *testing.T) {
	sm := NewSessionManager()
	_ = seedTestGraph(t, sm)
	s := NewMCPServer(sm)

	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "query_related",
			Arguments: map[string]interface{}{
				"from":      "srv-01",
				"operation": "descendants",
				"format":    "markdown",
			},
		},
	}
	res, err := s.GetTool("query_related").Handler(context.Background(), req)
	if err != nil {
		t.Fatalf("handler error: %v", err)
	}
	if res.IsError {
		t.Fatalf("expected success, got: %+v", res)
	}
	text := toolResultText(t, res)

	for _, want := range []string{"## Entities", "vm-01", "net-mgmt", "eth0"} {
		if !strings.Contains(text, want) {
			t.Errorf("expected %q in markdown output:\n%s", want, text)
		}
	}
}

func TestQueryEntitiesUnknownFormat(t *testing.T) {
	sm := NewSessionManager()
	_ = seedTestGraph(t, sm)
	s := NewMCPServer(sm)

	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "query_entities",
			Arguments: map[string]interface{}{
				"kind":   "server",
				"format": "xml",
			},
		},
	}
	res, err := s.GetTool("query_entities").Handler(context.Background(), req)
	if err != nil {
		t.Fatalf("handler error: %v", err)
	}
	if !res.IsError {
		t.Fatal("expected error result for unknown format")
	}
	if !strings.Contains(toolResultText(t, res), "unknown format") {
		t.Errorf("expected unknown format message, got: %s", toolResultText(t, res))
	}
}

func TestResolvePath(t *testing.T) {
	sm := NewSessionManager()
	_ = seedTestGraph(t, sm)
	s := NewMCPServer(sm)

	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      "resolve_path",
			Arguments: map[string]interface{}{"ref": "srv-01/net-mgmt/eth0"},
		},
	}
	res, err := s.GetTool("resolve_path").Handler(context.Background(), req)
	if err != nil {
		t.Fatalf("handler error: %v", err)
	}
	if res.IsError {
		t.Fatalf("expected success, got: %+v", res)
	}
	text := toolResultText(t, res)
	if !strings.Contains(text, `"id": "eth0"`) {
		t.Errorf("expected resolved entity eth0:\n%s", text)
	}

	// Unresolvable reference returns an error result.
	bad := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      "resolve_path",
			Arguments: map[string]interface{}{"ref": "does-not/exist"},
		},
	}
	badRes, err := s.GetTool("resolve_path").Handler(context.Background(), bad)
	if err != nil {
		t.Fatalf("handler error: %v", err)
	}
	if !badRes.IsError {
		t.Error("expected error result for unresolvable reference")
	}
}

func TestWhoReferences(t *testing.T) {
	sm := NewSessionManager()
	_ = seedTestGraph(t, sm)
	s := NewMCPServer(sm)

	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      "who_references",
			Arguments: map[string]interface{}{"id": "net-mgmt"},
		},
	}
	res, err := s.GetTool("who_references").Handler(context.Background(), req)
	if err != nil {
		t.Fatalf("handler error: %v", err)
	}
	if res.IsError {
		t.Fatalf("expected success, got: %+v", res)
	}
	text := toolResultText(t, res)
	if !strings.Contains(text, `"id": "eth0"`) {
		t.Errorf("expected child eth0 in who_references output:\n%s", text)
	}
	if !strings.Contains(text, `"property": "associated_network"`) {
		t.Errorf("expected vlan property reference in who_references output:\n%s", text)
	}

	// A server referenced by a relation should list the relation.
	relReq := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      "who_references",
			Arguments: map[string]interface{}{"id": "vm-01"},
		},
	}
	relRes, err := s.GetTool("who_references").Handler(context.Background(), relReq)
	if err != nil {
		t.Fatalf("handler error: %v", err)
	}
	if relRes.IsError {
		t.Fatalf("expected success, got: %+v", relRes)
	}
	relText := toolResultText(t, relRes)
	if !strings.Contains(relText, `"id": "rel-hosts"`) {
		t.Errorf("expected relation rel-hosts in who_references output:\n%s", relText)
	}
}
