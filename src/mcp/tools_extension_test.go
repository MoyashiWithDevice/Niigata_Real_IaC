package mcp

import (
	"context"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

func TestExtensionToolsRegistered(t *testing.T) {
	sm := NewSessionManager()
	s := NewMCPServer(sm)

	for _, name := range []string{"load_extension_dir", "list_extensions", "list_extension_kinds"} {
		if s.GetTool(name) == nil {
			t.Errorf("expected %s tool to be registered", name)
		}
	}
}

func TestListExtensionsEmptyWithoutBuiltin(t *testing.T) {
	sm := NewSessionManager()
	s := NewMCPServer(sm)

	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      "list_extensions",
			Arguments: map[string]interface{}{},
		},
	}
	res, err := s.GetTool("list_extensions").Handler(context.Background(), req)
	if err != nil {
		t.Fatalf("handler error: %v", err)
	}
	if res.IsError {
		t.Fatalf("expected success, got: %+v", res)
	}
	text := toolResultText(t, res)
	if strings.Contains(text, "aws") {
		t.Errorf("expected no builtin extensions in list, got: %s", text)
	}
}

func TestListExtensionKindsCoreOnly(t *testing.T) {
	sm := NewSessionManager()
	s := NewMCPServer(sm)

	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      "list_extension_kinds",
			Arguments: map[string]interface{}{},
		},
	}
	res, err := s.GetTool("list_extension_kinds").Handler(context.Background(), req)
	if err != nil {
		t.Fatalf("handler error: %v", err)
	}
	if res.IsError {
		t.Fatalf("expected success, got: %+v", res)
	}
	text := toolResultText(t, res)
	if strings.Contains(text, "aws") {
		t.Errorf("expected no extension kinds, got: %s", text)
	}
}

func TestLoadExtensionDirEmptySucceeds(t *testing.T) {
	sm := NewSessionManager()
	s := NewMCPServer(sm)

	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      "load_extension_dir",
			Arguments: map[string]interface{}{"path": t.TempDir()},
		},
	}
	res, err := s.GetTool("load_extension_dir").Handler(context.Background(), req)
	if err != nil {
		t.Fatalf("handler error: %v", err)
	}
	if res.IsError {
		t.Fatalf("expected success, got: %+v", res)
	}
}

func TestLoadExtensionDirNonexistent(t *testing.T) {
	sm := NewSessionManager()
	s := NewMCPServer(sm)

	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      "load_extension_dir",
			Arguments: map[string]interface{}{"path": "/nonexistent/path"},
		},
	}
	res, err := s.GetTool("load_extension_dir").Handler(context.Background(), req)
	if err != nil {
		t.Fatalf("handler error: %v", err)
	}
	if !res.IsError {
		t.Error("expected error result for nonexistent directory")
	}
}
