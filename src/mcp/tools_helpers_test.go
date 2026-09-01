package mcp

import (
	"context"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/bababa/Niigata_Real_IaC/src/core"
)

// callTool invokes the named tool on the server with the given arguments and
// returns the result, failing the test when the handler errors.
func callTool(t *testing.T, s *server.MCPServer, name string, args map[string]interface{}) *mcp.CallToolResult {
	t.Helper()
	tool := s.GetTool(name)
	if tool == nil {
		t.Fatalf("tool %q is not registered", name)
	}
	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      name,
			Arguments: args,
		},
	}
	res, err := tool.Handler(context.Background(), req)
	if err != nil {
		t.Fatalf("%s handler error: %v", name, err)
	}
	return res
}

// seedNiigataGraph populates the default session graph with a small valid
// Niigata-domain model for render/query coverage tests.
func seedNiigataGraph(t *testing.T, sm *SessionManager, s *server.MCPServer) *SessionData {
	t.Helper()
	sd := sm.GetOrCreate("default")

	area := core.NewEntity("area-sado", "area", "Sado Island")
	sd.Graph.AddEntity(area)

	forest := core.NewEntity("forest-01", "forest", "Beech Forest")
	forest.SetOwner("area-sado")
	sd.Graph.AddEntity(forest)

	spot := core.NewEntity("spot-01", "tourism_spot", "Viewpoint")
	spot.SetOwner("area-sado")
	sd.Graph.AddEntity(spot)

	rel := core.NewDirectedRelation("rel-near-01", "near", "spot-01", "forest-01")
	sd.Graph.AddRelation(rel)

	return sd
}
