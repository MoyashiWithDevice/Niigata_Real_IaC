package mcp

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"

	"IACForge/src/core"
	"IACForge/src/query"
)

func registerEntityMCPTools(s *mcpserver.MCPServer, sm *SessionManager) {
	s.AddTool(
		mcp.NewTool("add_entity",
			mcp.WithDescription("Add an entity to the graph."),
			mcp.WithString("id", mcp.Required(), mcp.Description("Unique entity identifier")),
			mcp.WithString("kind", mcp.Required(), mcp.Description("Entity kind (e.g. server, vm, rack)")),
			mcp.WithString("name", mcp.Required(), mcp.Description("Human-readable name")),
			mcp.WithString("owner", mcp.Description("Parent entity ID for ownership")),
			mcp.WithString("description", mcp.Description("Description")),
			mcp.WithString("status", mcp.Description("Lifecycle status (planned, active, maintenance, deprecated, offline)")),
			mcp.WithArray("tags", mcp.Description("Tags"), mcp.WithStringItems()),
			mcp.WithString("labels_json", mcp.Description("Labels as JSON string, e.g. {\"region\":\"us-east-1\"}")),
			mcp.WithString("properties_json", mcp.Description("Spec properties as JSON string")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			sd := getOrCreateSession(ctx, sm)

			id, err := req.RequireString("id")
			if err != nil {
				return toolError(err.Error()), nil
			}
			kind, err := req.RequireString("kind")
			if err != nil {
				return toolError(err.Error()), nil
			}
			name, err := req.RequireString("name")
			if err != nil {
				return toolError(err.Error()), nil
			}

			e := core.NewEntity(id, core.EntityKind(kind), name)

			if owner := req.GetString("owner", ""); owner != "" {
				e.SetOwner(owner)
			}
			if desc := req.GetString("description", ""); desc != "" {
				e.Description = desc
			}
			if status := req.GetString("status", ""); status != "" {
				e.SetStatus(core.Status(status))
			}
			if tags := req.GetStringSlice("tags", nil); len(tags) > 0 {
				for _, t := range tags {
					e.AddTag(t)
				}
			}
			if labelsJSON := req.GetString("labels_json", ""); labelsJSON != "" {
				var labels map[string]string
				if err := json.Unmarshal([]byte(labelsJSON), &labels); err != nil {
					return toolError(fmt.Sprintf("invalid labels_json: %v", err)), nil
				}
				e.Labels = labels
			}
			if propsJSON := req.GetString("properties_json", ""); propsJSON != "" {
				var props map[string]interface{}
				if err := json.Unmarshal([]byte(propsJSON), &props); err != nil {
					return toolError(fmt.Sprintf("invalid properties_json: %v", err)), nil
				}
				e.Properties = props
			}

			if err := sd.Graph.AddEntity(e); err != nil {
				return toolError(fmt.Sprintf("failed to add entity: %v", err)), nil
			}
			return toolResult(fmt.Sprintf("Added entity %s (kind=%s)", id, kind)), nil
		},
	)

	s.AddTool(
		mcp.NewTool("get_entity",
			mcp.WithDescription("Get an entity by ID."),
			mcp.WithString("id", mcp.Required(), mcp.Description("Entity ID")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			sd := getOrCreateSession(ctx, sm)
			id, err := req.RequireString("id")
			if err != nil {
				return toolError(err.Error()), nil
			}

			e, ok := sd.Graph.GetEntity(id)
			if !ok {
				return toolError(fmt.Sprintf("entity not found: %s", id)), nil
			}
			data, err := json.MarshalIndent(query.EntityJSONMap(e), "", "  ")
			if err != nil {
				return toolError(fmt.Sprintf("failed to marshal response: %v", err)), nil
			}
			return toolResult(string(data)), nil
		},
	)

	s.AddTool(
		mcp.NewTool("update_entity",
			mcp.WithDescription("Update an existing entity's properties."),
			mcp.WithString("id", mcp.Required(), mcp.Description("Entity ID to update")),
			mcp.WithString("name", mcp.Description("New name")),
			mcp.WithString("owner", mcp.Description("New owner entity ID")),
			mcp.WithString("description", mcp.Description("New description")),
			mcp.WithString("status", mcp.Description("New status")),
			mcp.WithArray("tags", mcp.Description("New tags"), mcp.WithStringItems()),
			mcp.WithString("labels_json", mcp.Description("New labels as JSON string")),
			mcp.WithString("properties_json", mcp.Description("New spec properties as JSON string")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			sd := getOrCreateSession(ctx, sm)
			id, err := req.RequireString("id")
			if err != nil {
				return toolError(err.Error()), nil
			}

			e, ok := sd.Graph.GetEntity(id)
			if !ok {
				return toolError(fmt.Sprintf("entity not found: %s", id)), nil
			}

			if v := req.GetString("name", ""); v != "" {
				e.Name = v
			}
			if v := req.GetString("owner", ""); v != "" {
				e.SetOwner(v)
			}
			if v := req.GetString("description", ""); v != "" {
				e.Description = v
			}
			if v := req.GetString("status", ""); v != "" {
				e.SetStatus(core.Status(v))
			}
			if tags := req.GetStringSlice("tags", nil); len(tags) > 0 {
				e.Tags = tags
			}
			if labelsJSON := req.GetString("labels_json", ""); labelsJSON != "" {
				var labels map[string]string
				if err := json.Unmarshal([]byte(labelsJSON), &labels); err != nil {
					return toolError(fmt.Sprintf("invalid labels_json: %v", err)), nil
				}
				e.Labels = labels
			}
			if propsJSON := req.GetString("properties_json", ""); propsJSON != "" {
				var newProps map[string]interface{}
				if err := json.Unmarshal([]byte(propsJSON), &newProps); err != nil {
					return toolError(fmt.Sprintf("invalid properties_json: %v", err)), nil
				}
				if e.Properties == nil {
					e.Properties = make(map[string]interface{})
				}
				for k, v := range newProps {
					e.Properties[k] = v
				}
			}

			if err := sd.Graph.UpdateEntity(e); err != nil {
				return toolError(fmt.Sprintf("failed to update entity: %v", err)), nil
			}
			return toolResult(fmt.Sprintf("Updated entity %s", id)), nil
		},
	)

	s.AddTool(
		mcp.NewTool("remove_entity",
			mcp.WithDescription("Remove an entity from the graph by ID."),
			mcp.WithString("id", mcp.Required(), mcp.Description("Entity ID to remove")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			sd := getOrCreateSession(ctx, sm)
			id, err := req.RequireString("id")
			if err != nil {
				return toolError(err.Error()), nil
			}

			if !sd.Graph.RemoveEntity(id) {
				return toolError(fmt.Sprintf("entity not found: %s", id)), nil
			}
			return toolResult(fmt.Sprintf("Removed entity %s", id)), nil
		},
	)

	s.AddTool(
		mcp.NewTool("list_entities",
			mcp.WithDescription("List all entities, optionally filtered by kind."),
			mcp.WithString("kind", mcp.Description("Filter by entity kind")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			sd := getOrCreateSession(ctx, sm)
			kind := req.GetString("kind", "")

			var entities []*core.Entity
			if kind != "" {
				entities = sd.Graph.EntitiesByKind(core.EntityKind(kind))
			} else {
				entities = sd.Graph.Entities()
			}

			type entitySummary struct {
				ID    string          `json:"id"`
				Kind  core.EntityKind `json:"kind"`
				Name  string          `json:"name"`
				Owner string          `json:"owner,omitempty"`
			}

			summaries := make([]entitySummary, len(entities))
			for i, e := range entities {
				summaries[i] = entitySummary{
					ID: e.ID, Kind: e.Kind, Name: e.Name, Owner: e.Owner,
				}
			}
			data, err := json.MarshalIndent(summaries, "", "  ")
			if err != nil {
				return toolError(fmt.Sprintf("failed to marshal response: %v", err)), nil
			}
			return toolResult(string(data)), nil
		},
	)
}
