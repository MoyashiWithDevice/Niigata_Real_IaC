package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"

	"github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"

	"IACForge/src/extension"
)

func registerExtensionMCPTools(s *mcpserver.MCPServer, sm *SessionManager) {
	s.AddTool(
		mcp.NewTool("load_extension_dir",
			mcp.WithDescription("Load all Go plugin extensions (.so) from a directory and register them into the session schema."),
			mcp.WithString("path", mcp.Required(), mcp.Description("Path to the directory containing .so extension plugins")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			sd := getOrCreateSession(ctx, sm)
			path, err := req.RequireString("path")
			if err != nil {
				return toolError(err.Error()), nil
			}

			if err := sd.Extensions.LoadFromDir(path); err != nil {
				return toolError(fmt.Sprintf("failed to load extensions from %s: %v", path, err)), nil
			}

			loaded := sd.Extensions.LoadOrder()
			return toolResult(fmt.Sprintf("Loaded %d extension(s) from %s: %v", len(loaded), path, loaded)), nil
		},
	)

	s.AddTool(
		mcp.NewTool("list_extensions",
			mcp.WithDescription("List all registered extensions with their metadata."),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			sd := getOrCreateSession(ctx, sm)

			type extInfo struct {
				ID             string   `json:"id"`
				Name           string   `json:"name"`
				Version        string   `json:"version"`
				Namespace      string   `json:"namespace"`
				Description    string   `json:"description,omitempty"`
				ExtensionPoint []string `json:"extension_points"`
			}

			exts := sd.Extensions.Extensions()
			infos := make([]extInfo, 0, len(exts))
			for _, e := range exts {
				infos = append(infos, extInfo{
					ID:             e.Manifest.ID,
					Name:           e.Manifest.Name,
					Version:        e.Manifest.Version,
					Namespace:      e.Manifest.Namespace,
					Description:    e.Manifest.Description,
					ExtensionPoint: e.Manifest.ExtensionPoints,
				})
			}

			data, err := json.MarshalIndent(infos, "", "  ")
			if err != nil {
				return toolError(fmt.Sprintf("failed to marshal response: %v", err)), nil
			}
			return toolResult(string(data)), nil
		},
	)

	s.AddTool(
		mcp.NewTool("list_extension_kinds",
			mcp.WithDescription("List all entity kinds contributed by extensions, grouped by extension."),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			sd := getOrCreateSession(ctx, sm)

			epAny, ok := sd.Extensions.GetExtensionPoint(extension.ExtensionPointEntityKinds)
			if !ok {
				return toolResult("{}"), nil
			}
			ep, ok := epAny.(*extension.EntityKindsExtensionPoint)
			if !ok {
				return toolResult("{}"), nil
			}

			byExt := make(map[string][]string)
			for kind, extID := range ep.AllExtendedKinds() {
				byExt[extID] = append(byExt[extID], string(kind))
			}

			extIDs := make([]string, 0, len(byExt))
			for extID := range byExt {
				extIDs = append(extIDs, extID)
			}
			sort.Strings(extIDs)
			for _, extID := range extIDs {
				sort.Strings(byExt[extID])
			}

			data, err := json.MarshalIndent(byExt, "", "  ")
			if err != nil {
				return toolError(fmt.Sprintf("failed to marshal response: %v", err)), nil
			}
			return toolResult(string(data)), nil
		},
	)
}
