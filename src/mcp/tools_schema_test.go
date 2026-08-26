package mcp

import (
	"context"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

func TestListEntityKinds(t *testing.T) {
	sm := NewSessionManager()
	s := NewMCPServer(sm)

	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      "list_entity_kinds",
			Arguments: map[string]interface{}{},
		},
	}
	res, err := s.GetTool("list_entity_kinds").Handler(context.Background(), req)
	if err != nil {
		t.Fatalf("handler error: %v", err)
	}
	if res.IsError {
		t.Fatalf("expected success, got: %+v", res)
	}

	text := toolResultText(t, res)
	for _, want := range []string{`"kind": "server"`, `"kind": "network"`, `"nest_keys"`} {
		if !strings.Contains(text, want) {
			t.Errorf("expected %q in list_entity_kinds output:\n%s", want, text)
		}
	}
}

func TestGetEntityKind(t *testing.T) {
	sm := NewSessionManager()
	s := NewMCPServer(sm)

	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      "get_entity_kind",
			Arguments: map[string]interface{}{"kind": "server"},
		},
	}
	res, err := s.GetTool("get_entity_kind").Handler(context.Background(), req)
	if err != nil {
		t.Fatalf("handler error: %v", err)
	}
	if res.IsError {
		t.Fatalf("expected success, got: %+v", res)
	}

	text := toolResultText(t, res)
	for _, want := range []string{`"kind": "server"`, `"properties"`, `"nesting_defs"`, `"networks"`} {
		if !strings.Contains(text, want) {
			t.Errorf("expected %q in get_entity_kind output:\n%s", want, text)
		}
	}

	// Unknown kind returns an error result.
	bad := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      "get_entity_kind",
			Arguments: map[string]interface{}{"kind": "does-not-exist"},
		},
	}
	badRes, err := s.GetTool("get_entity_kind").Handler(context.Background(), bad)
	if err != nil {
		t.Fatalf("handler error: %v", err)
	}
	if !badRes.IsError {
		t.Error("expected error result for unknown kind")
	}
}

func TestListRelationTypes(t *testing.T) {
	sm := NewSessionManager()
	s := NewMCPServer(sm)

	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      "list_relation_types",
			Arguments: map[string]interface{}{},
		},
	}
	res, err := s.GetTool("list_relation_types").Handler(context.Background(), req)
	if err != nil {
		t.Fatalf("handler error: %v", err)
	}
	if res.IsError {
		t.Fatalf("expected success, got: %+v", res)
	}

	text := toolResultText(t, res)
	for _, want := range []string{`"type": "connects"`, `"direction": "symmetric"`, `"type": "hosts"`} {
		if !strings.Contains(text, want) {
			t.Errorf("expected %q in list_relation_types output:\n%s", want, text)
		}
	}
}

func TestGetSchema(t *testing.T) {
	sm := NewSessionManager()
	s := NewMCPServer(sm)

	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      "get_schema",
			Arguments: map[string]interface{}{},
		},
	}
	res, err := s.GetTool("get_schema").Handler(context.Background(), req)
	if err != nil {
		t.Fatalf("handler error: %v", err)
	}
	if res.IsError {
		t.Fatalf("expected success, got: %+v", res)
	}

	text := toolResultText(t, res)
	for _, want := range []string{"entity_kinds:", "relation_types:", "server:", "connects:"} {
		if !strings.Contains(text, want) {
			t.Errorf("expected %q in get_schema YAML output:\n%s", want, text)
		}
	}
}
