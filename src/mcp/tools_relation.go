package mcp

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"

	"IACForge/src/core"
)

func registerRelationMCPTools(s *mcpserver.MCPServer, sm *SessionManager) {
	s.AddTool(
		mcp.NewTool("add_relation",
			mcp.WithDescription("Add a relation to the graph."),
			mcp.WithString("id", mcp.Required(), mcp.Description("Unique relation identifier")),
			mcp.WithString("type", mcp.Required(), mcp.Description("Relation type (connects, hosts, depends_on, belongs_to, etc.)")),
			mcp.WithArray("participants", mcp.Required(), mcp.Description("List of participant entity IDs (for symmetric relations like connects)")),
			mcp.WithString("source", mcp.Description("Source participant entity ID (for directed relations)")),
			mcp.WithString("target", mcp.Description("Target participant entity ID (for directed relations)")),
			mcp.WithString("description", mcp.Description("Description")),
			mcp.WithString("status", mcp.Description("Lifecycle status")),
			mcp.WithArray("tags", mcp.Description("Tags"), mcp.WithStringItems()),
			mcp.WithString("labels_json", mcp.Description("Labels as JSON string")),
			mcp.WithString("properties_json", mcp.Description("Spec properties as JSON string")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			sd := getOrCreateSession(ctx, sm)

			id, err := req.RequireString("id")
			if err != nil {
				return toolError(err.Error()), nil
			}
			relType, err := req.RequireString("type")
			if err != nil {
				return toolError(err.Error()), nil
			}

			var r *core.Relation

			source := req.GetString("source", "")
			target := req.GetString("target", "")
			participants := req.GetStringSlice("participants", nil)

			if source != "" && target != "" {
				r = core.NewDirectedRelation(id, core.RelationType(relType), source, target)
			} else if len(participants) >= 2 {
				r = core.NewSymmetricRelation(id, core.RelationType(relType), participants)
			} else {
				return toolError("provide either source+target for directed relations or participants list for symmetric relations"), nil
			}

			if desc := req.GetString("description", ""); desc != "" {
				r.Description = desc
			}
			if status := req.GetString("status", ""); status != "" {
				r.SetStatus(core.Status(status))
			}
			if tags := req.GetStringSlice("tags", nil); len(tags) > 0 {
				for _, t := range tags {
					r.AddTag(t)
				}
			}
			if labelsJSON := req.GetString("labels_json", ""); labelsJSON != "" {
				var labels map[string]string
				if err := json.Unmarshal([]byte(labelsJSON), &labels); err != nil {
					return toolError(fmt.Sprintf("invalid labels_json: %v", err)), nil
				}
				r.Labels = labels
			}
			if propsJSON := req.GetString("properties_json", ""); propsJSON != "" {
				var props map[string]interface{}
				if err := json.Unmarshal([]byte(propsJSON), &props); err != nil {
					return toolError(fmt.Sprintf("invalid properties_json: %v", err)), nil
				}
				r.Properties = props
			}

			if err := sd.Graph.AddRelation(r); err != nil {
				return toolError(fmt.Sprintf("failed to add relation: %v", err)), nil
			}
			return toolResult(fmt.Sprintf("Added relation %s (type=%s)", id, relType)), nil
		},
	)

	s.AddTool(
		mcp.NewTool("get_relation",
			mcp.WithDescription("Get a relation by ID."),
			mcp.WithString("id", mcp.Required(), mcp.Description("Relation ID")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			sd := getOrCreateSession(ctx, sm)
			id, err := req.RequireString("id")
			if err != nil {
				return toolError(err.Error()), nil
			}

			r, ok := sd.Graph.GetRelation(id)
			if !ok {
				return toolError(fmt.Sprintf("relation not found: %s", id)), nil
			}
			data, err := json.MarshalIndent(r, "", "  ")
			if err != nil {
				return toolError(fmt.Sprintf("failed to marshal response: %v", err)), nil
			}
			return toolResult(string(data)), nil
		},
	)

	s.AddTool(
		mcp.NewTool("update_relation",
			mcp.WithDescription("Update an existing relation's properties."),
			mcp.WithString("id", mcp.Required(), mcp.Description("Relation ID to update")),
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

			r, ok := sd.Graph.GetRelation(id)
			if !ok {
				return toolError(fmt.Sprintf("relation not found: %s", id)), nil
			}

			if v := req.GetString("description", ""); v != "" {
				r.Description = v
			}
			if v := req.GetString("status", ""); v != "" {
				r.SetStatus(core.Status(v))
			}
			if tags := req.GetStringSlice("tags", nil); len(tags) > 0 {
				r.Tags = tags
			}
			if labelsJSON := req.GetString("labels_json", ""); labelsJSON != "" {
				var labels map[string]string
				if err := json.Unmarshal([]byte(labelsJSON), &labels); err != nil {
					return toolError(fmt.Sprintf("invalid labels_json: %v", err)), nil
				}
				r.Labels = labels
			}
			if propsJSON := req.GetString("properties_json", ""); propsJSON != "" {
				var props map[string]interface{}
				if err := json.Unmarshal([]byte(propsJSON), &props); err != nil {
					return toolError(fmt.Sprintf("invalid properties_json: %v", err)), nil
				}
				r.Properties = props
			}

			if err := sd.Graph.UpdateRelation(r); err != nil {
				return toolError(fmt.Sprintf("failed to update relation: %v", err)), nil
			}
			return toolResult(fmt.Sprintf("Updated relation %s", id)), nil
		},
	)

	s.AddTool(
		mcp.NewTool("remove_relation",
			mcp.WithDescription("Remove a relation from the graph by ID."),
			mcp.WithString("id", mcp.Required(), mcp.Description("Relation ID to remove")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			sd := getOrCreateSession(ctx, sm)
			id, err := req.RequireString("id")
			if err != nil {
				return toolError(err.Error()), nil
			}

			if !sd.Graph.RemoveRelation(id) {
				return toolError(fmt.Sprintf("relation not found: %s", id)), nil
			}
			return toolResult(fmt.Sprintf("Removed relation %s", id)), nil
		},
	)

	s.AddTool(
		mcp.NewTool("list_relations",
			mcp.WithDescription("List all relations, optionally filtered by type."),
			mcp.WithString("type", mcp.Description("Filter by relation type")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			sd := getOrCreateSession(ctx, sm)
			relType := req.GetString("type", "")

			var relations []*core.Relation
			if relType != "" {
				relations = sd.Graph.RelationsByType(core.RelationType(relType))
			} else {
				relations = sd.Graph.Relations()
			}

			type relationSummary struct {
				ID        string            `json:"id"`
				Type      core.RelationType `json:"type"`
				Source    string            `json:"source,omitempty"`
				Target    string            `json:"target,omitempty"`
				Direction core.Direction    `json:"direction"`
			}

			summaries := make([]relationSummary, len(relations))
			for i, r := range relations {
				summaries[i] = relationSummary{
					ID:        r.ID,
					Type:      r.Type,
					Source:    r.Source(),
					Target:    r.Target(),
					Direction: r.Direction,
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
