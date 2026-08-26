package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"

	"github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"
	"gopkg.in/yaml.v3"

	"IACForge/src/core"
	"IACForge/src/schema"
)

func registerSchemaMCPTools(s *mcpserver.MCPServer, sm *SessionManager) {
	s.AddTool(
		mcp.NewTool("list_entity_kinds",
			mcp.WithDescription("List all entity kinds defined in the schema with their descriptions and nesting keys."),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			sd := getOrCreateSession(ctx, sm)

			type kindSummary struct {
				Kind        string   `json:"kind"`
				Description string   `json:"description,omitempty"`
				NestKeys    []string `json:"nest_keys,omitempty"`
			}

			kinds := make([]kindSummary, 0, len(sd.Schema.EntityKinds))
			names := sortedEntityKinds(sd.Schema)
			for _, kind := range names {
				def := sd.Schema.EntityKinds[kind]
				nestKeys := make([]string, 0, len(def.NestingDefs))
				for _, nd := range sd.Schema.GetNestingDefs(kind) {
					nestKeys = append(nestKeys, nd.NestKey)
				}
				sort.Strings(nestKeys)
				kinds = append(kinds, kindSummary{
					Kind:        string(kind),
					Description: def.Description,
					NestKeys:    nestKeys,
				})
			}

			data, err := json.MarshalIndent(kinds, "", "  ")
			if err != nil {
				return toolError(fmt.Sprintf("failed to marshal response: %v", err)), nil
			}
			return toolResult(string(data)), nil
		},
	)

	s.AddTool(
		mcp.NewTool("get_entity_kind",
			mcp.WithDescription("Get the full schema definition for an entity kind, including properties and nesting definitions."),
			mcp.WithString("kind", mcp.Required(), mcp.Description("Entity kind (e.g. server, vm, switch)")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			sd := getOrCreateSession(ctx, sm)
			kind, err := req.RequireString("kind")
			if err != nil {
				return toolError(err.Error()), nil
			}

			def, ok := sd.Schema.GetEntityKindDef(core.EntityKind(kind))
			if !ok {
				return toolError(fmt.Sprintf("entity kind %q not found in schema", kind)), nil
			}

			type propertySummary struct {
				Name        string      `json:"name"`
				Type        string      `json:"type"`
				Required    bool        `json:"required"`
				Default     interface{} `json:"default,omitempty"`
				Description string      `json:"description,omitempty"`
				Enum        []string    `json:"enum,omitempty"`
			}

			properties := make([]propertySummary, 0, len(def.Properties))
			for _, p := range def.Properties {
				var enum []string
				if p.Constraints != nil {
					enum = p.Constraints.Enum
				}
				properties = append(properties, propertySummary{
					Name:        p.Name,
					Type:        string(p.Type),
					Required:    p.Required,
					Default:     p.Default,
					Description: p.Description,
					Enum:        enum,
				})
			}

			type nestingSummary struct {
				NestKey            string `json:"nest_key"`
				ChildKind          string `json:"child_kind"`
				AutoRelationType   string `json:"auto_relation_type,omitempty"`
				AutoRelationSource string `json:"auto_relation_source,omitempty"`
			}

			nestings := make([]nestingSummary, 0, len(def.NestingDefs))
			for _, nd := range sd.Schema.GetNestingDefs(core.EntityKind(kind)) {
				nestings = append(nestings, nestingSummary{
					NestKey:            nd.NestKey,
					ChildKind:          string(nd.ChildKind),
					AutoRelationType:   string(nd.AutoRelationType),
					AutoRelationSource: nd.AutoRelationSource,
				})
			}

			result := map[string]interface{}{
				"kind":         kind,
				"description":  def.Description,
				"properties":   properties,
				"nesting_defs": nestings,
			}

			data, err := json.MarshalIndent(result, "", "  ")
			if err != nil {
				return toolError(fmt.Sprintf("failed to marshal response: %v", err)), nil
			}
			return toolResult(string(data)), nil
		},
	)

	s.AddTool(
		mcp.NewTool("list_relation_types",
			mcp.WithDescription("List all relation types defined in the schema with their direction and description."),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			sd := getOrCreateSession(ctx, sm)

			type relTypeSummary struct {
				Type        string `json:"type"`
				Direction   string `json:"direction"`
				Description string `json:"description,omitempty"`
			}

			types := make([]relTypeSummary, 0, len(sd.Schema.RelationTypes))
			names := sortedRelationTypes(sd.Schema)
			for _, relType := range names {
				def := sd.Schema.RelationTypes[relType]
				types = append(types, relTypeSummary{
					Type:        string(relType),
					Direction:   string(def.Direction),
					Description: def.Description,
				})
			}

			data, err := json.MarshalIndent(types, "", "  ")
			if err != nil {
				return toolError(fmt.Sprintf("failed to marshal response: %v", err)), nil
			}
			return toolResult(string(data)), nil
		},
	)

	s.AddTool(
		mcp.NewTool("get_schema",
			mcp.WithDescription("Get the full schema (entity kinds, relation types, nesting definitions) as YAML."),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			sd := getOrCreateSession(ctx, sm)

			data, err := yaml.Marshal(sd.Schema)
			if err != nil {
				return toolError(fmt.Sprintf("failed to marshal schema: %v", err)), nil
			}
			return toolResult(string(data)), nil
		},
	)
}

func sortedEntityKinds(s *schema.Schema) []core.EntityKind {
	names := make([]core.EntityKind, 0, len(s.EntityKinds))
	for k := range s.EntityKinds {
		names = append(names, k)
	}
	sort.Slice(names, func(i, j int) bool { return names[i] < names[j] })
	return names
}

func sortedRelationTypes(s *schema.Schema) []core.RelationType {
	names := make([]core.RelationType, 0, len(s.RelationTypes))
	for t := range s.RelationTypes {
		names = append(names, t)
	}
	sort.Slice(names, func(i, j int) bool { return names[i] < names[j] })
	return names
}
