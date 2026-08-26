package mcp

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"

	"IACForge/src/core"
	"IACForge/src/parser"
)

func registerYAMLMCPTools(s *mcpserver.MCPServer, sm *SessionManager) {
	s.AddTool(
		mcp.NewTool("load_yaml",
			mcp.WithDescription("Load a YAML infrastructure model file and build the in-memory graph."),
			mcp.WithString("path", mcp.Required(), mcp.Description("Path to the YAML file")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			sd := getOrCreateSession(ctx, sm)
			path, err := req.RequireString("path")
			if err != nil {
				return toolError(err.Error()), nil
			}

			msg, err := loadGraph(sd, path, false)
			if err != nil {
				return toolError(fmt.Sprintf("failed to parse %s: %v", path, err)), nil
			}
			return toolResult(msg), nil
		},
	)

	s.AddTool(
		mcp.NewTool("load_dir",
			mcp.WithDescription("Recursively load all YAML files from a directory and merge them into the in-memory graph. References may span files: relation participants, owners, and @-prefixed property references resolve against the merged graph."),
			mcp.WithString("path", mcp.Required(), mcp.Description("Path to the directory to load recursively")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			sd := getOrCreateSession(ctx, sm)
			path, err := req.RequireString("path")
			if err != nil {
				return toolError(err.Error()), nil
			}

			msg, err := loadGraph(sd, path, true)
			if err != nil {
				return toolError(fmt.Sprintf("failed to load directory %s: %v", path, err)), nil
			}
			return toolResult(msg), nil
		},
	)

	s.AddTool(
		mcp.NewTool("save_yaml",
			mcp.WithDescription("Save the current graph to a YAML file."),
			mcp.WithString("path", mcp.Required(), mcp.Description("Path to write the YAML file")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			sd := getOrCreateSession(ctx, sm)
			path, err := req.RequireString("path")
			if err != nil {
				return toolError(err.Error()), nil
			}

			ser := parser.NewSerializerWithSchema(sd.Schema)
			if err := ser.SerializeFile(sd.Graph, path); err != nil {
				return toolError(fmt.Sprintf("failed to serialize: %v", err)), nil
			}
			return toolResult(fmt.Sprintf("Saved graph to %s", path)), nil
		},
	)

	s.AddTool(
		mcp.NewTool("parse_yaml_string",
			mcp.WithDescription("Parse a YAML string and build the in-memory graph."),
			mcp.WithString("yaml_content", mcp.Required(), mcp.Description("YAML content string")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			sd := getOrCreateSession(ctx, sm)
			content, err := req.RequireString("yaml_content")
			if err != nil {
				return toolError(err.Error()), nil
			}

			p := parser.NewParserWithSchema(sd.Schema)
			g, err := p.Parse([]byte(content))
			if err != nil {
				return toolError(fmt.Sprintf("failed to parse YAML: %v", err)), nil
			}
			sd.Graph = g
			return toolResult(fmt.Sprintf("Parsed %d entities and %d relations", len(g.Entities()), len(g.Relations()))), nil
		},
	)

	s.AddTool(
		mcp.NewTool("serialize_to_string",
			mcp.WithDescription("Serialize the current graph to a YAML string."),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			sd := getOrCreateSession(ctx, sm)

			ser := parser.NewSerializerWithSchema(sd.Schema)
			data, err := ser.Serialize(sd.Graph)
			if err != nil {
				return toolError(fmt.Sprintf("failed to serialize: %v", err)), nil
			}
			return toolResult(string(data)), nil
		},
	)
}

// loadGraph parses a file or directory into the session graph and returns a summary.
// Cross-file references are resolved against the merged graph; any unresolved
// references are appended to the summary as warnings.
func loadGraph(sd *SessionData, path string, isDir bool) (string, error) {
	p := parser.NewParserWithSchema(sd.Schema)
	var g *core.Graph
	var err error
	if isDir {
		g, err = p.ParseDir(path)
	} else {
		g, err = p.ParseFile(path)
	}
	if err != nil {
		return "", err
	}
	sd.Graph = g

	msg := fmt.Sprintf("Loaded %d entities and %d relations from %s", len(g.Entities()), len(g.Relations()), path)
	if errs := parser.ResolveReferences(g); len(errs) > 0 {
		msg += "\nUnresolved references:"
		for _, e := range errs {
			msg += "\n  - " + e.Error()
		}
	}
	return msg, nil
}

func getOrCreateSession(ctx context.Context, sm *SessionManager) *SessionData {
	session := mcpserver.ClientSessionFromContext(ctx)
	if session != nil {
		return sm.GetOrCreate(session.SessionID())
	}
	return sm.GetOrCreate("default")
}
