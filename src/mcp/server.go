package mcp

import (
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	_ "IACForge/src/extension/builtin/aws"
)

// NewMCPServer creates and configures a new MCP server with all IACForge tools.
func NewMCPServer(sm *SessionManager) *server.MCPServer {
	s := server.NewMCPServer(
		"IACForge",
		"0.1.0",
		server.WithToolCapabilities(false),
	)

	registerYAMLMCPTools(s, sm)
	registerEntityMCPTools(s, sm)
	registerRelationMCPTools(s, sm)
	registerGraphMCPTools(s, sm)
	registerSchemaMCPTools(s, sm)
	registerQueryMCPTools(s, sm)
	registerRenderMCPTools(s, sm)
	registerExtensionMCPTools(s, sm)

	return s
}

// ServeStdio starts the MCP server using stdio transport.
func ServeStdio(s *server.MCPServer) error {
	return server.ServeStdio(s)
}

// NewSSEServer creates an SSE server bound to the given address.
func NewSSEServer(s *server.MCPServer, addr string) error {
	sse := server.NewSSEServer(s,
		server.WithBaseURL("http://"+addr),
	)
	return sse.Start(addr)
}

func toolResult(text string) *mcp.CallToolResult {
	return mcp.NewToolResultText(text)
}

func toolError(msg string) *mcp.CallToolResult {
	return mcp.NewToolResultError(msg)
}
