package mcp

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"

	"IACForge/src/core"
)

func TestSessionManager(t *testing.T) {
	sm := NewSessionManager()

	sd1 := sm.GetOrCreate("session-1")
	sd2 := sm.GetOrCreate("session-1")
	if sd1 != sd2 {
		t.Fatal("expected same session data for same ID")
	}

	sd3 := sm.GetOrCreate("session-2")
	if sd1 == sd3 {
		t.Fatal("expected different session data for different IDs")
	}

	sm.Remove("session-1")
	sd4 := sm.GetOrCreate("session-1")
	if sd1 == sd4 {
		t.Fatal("expected new session data after removal")
	}
}

func TestSessionManagerGraph(t *testing.T) {
	sm := NewSessionManager()
	sd := sm.GetOrCreate("test")

	if sd.Graph == nil {
		t.Fatal("expected non-nil graph")
	}

	e := core.NewEntity("srv-01", "server", "Server 01")
	if err := sd.Graph.AddEntity(e); err != nil {
		t.Fatalf("failed to add entity: %v", err)
	}

	entities := sd.Graph.Entities()
	if len(entities) != 1 {
		t.Fatalf("expected 1 entity, got %d", len(entities))
	}
	if entities[0].ID != "srv-01" {
		t.Errorf("expected entity ID srv-01, got %s", entities[0].ID)
	}
}

func TestNewMCPServer(t *testing.T) {
	sm := NewSessionManager()
	s := NewMCPServer(sm)
	if s == nil {
		t.Fatal("expected non-nil MCP server")
	}
}

func TestLoadDirToolRegistered(t *testing.T) {
	sm := NewSessionManager()
	s := NewMCPServer(sm)

	ts := s.GetTool("load_dir")
	if ts == nil {
		t.Fatal("expected load_dir tool to be registered")
	}
	if ts.Handler == nil {
		t.Fatal("expected load_dir tool to have a handler")
	}
}

func TestLoadDirToolCrossFileReferences(t *testing.T) {
	sm := NewSessionManager()
	s := NewMCPServer(sm)

	dir := t.TempDir()
	files := map[string]string{
		"fileA.yaml": `
objects:
  - id: region-ap-northeast-1
    kind: region
    name: Tokyo Datacenter 1

  - id: srv-proxmox-01
    kind: server
    name: Proxmox Node 01
    attributes:
      owner: region-ap-northeast-1
    spec:
      networks:
        - id: net-mgmt
          name: Management
          interfaces:
            - id: eth0
              name: eth0

  - id: net-storage
    kind: network
    name: Storage Network
    spec:
      cidr: 192.168.10.0/24
`,
		"fileB.yaml": `
objects:
  - id: vlan-100
    kind: vlan
    name: VLAN 100
    attributes:
      owner: region-ap-northeast-1
    spec:
      vlan_id: 100
      associated_network: "@net-storage"

  - id: rel-connects
    type: connects
    participants:
      - srv-proxmox-01/net-mgmt/eth0
      - sw-core-01/port1
`,
		"nested/fileC.yaml": `
objects:
  - id: sw-core-01
    kind: switch
    name: Core Switch 01
    attributes:
      owner: region-ap-northeast-1
    spec:
      interfaces:
        - id: port1
          name: port1
`,
	}
	for relPath, content := range files {
		full := filepath.Join(dir, relPath)
		if err := os.MkdirAll(filepath.Dir(full), 0755); err != nil {
			t.Fatalf("failed to create dir: %v", err)
		}
		if err := os.WriteFile(full, []byte(content), 0644); err != nil {
			t.Fatalf("failed to write %s: %v", relPath, err)
		}
	}

	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      "load_dir",
			Arguments: map[string]interface{}{"path": dir},
		},
	}

	res, err := s.GetTool("load_dir").Handler(context.Background(), req)
	if err != nil {
		t.Fatalf("handler returned error: %v", err)
	}
	if res.IsError {
		t.Fatalf("expected success, got error result: %+v", res)
	}

	text := toolResultText(t, res)
	if !strings.Contains(text, "Loaded") {
		t.Errorf("expected summary in result text, got: %s", text)
	}
	if strings.Contains(text, "Unresolved references") {
		t.Errorf("cross-file references should resolve, got warnings: %s", text)
	}

	sd := sm.GetOrCreate("default")
	if _, ok := sd.Graph.GetEntity("srv-proxmox-01"); !ok {
		t.Error("expected entity from fileA in session graph")
	}
	if _, ok := sd.Graph.GetEntity("sw-core-01"); !ok {
		t.Error("expected entity from nested file in session graph")
	}
	if _, ok := sd.Graph.GetRelation("rel-connects"); !ok {
		t.Error("expected cross-file relation in session graph")
	}
}

func toolResultText(t *testing.T, res *mcp.CallToolResult) string {
	t.Helper()
	if len(res.Content) == 0 {
		t.Fatal("expected content in result")
	}
	tc, ok := mcp.AsTextContent(res.Content[0])
	if !ok {
		t.Fatalf("expected text content, got %T", res.Content[0])
	}
	return tc.Text
}
