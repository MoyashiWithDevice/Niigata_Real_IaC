package mcp

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"

	"IACForge/src/core"
)

func registerGraphMCPTools(s *mcpserver.MCPServer, sm *SessionManager) {
	s.AddTool(
		mcp.NewTool("validate_graph",
			mcp.WithDescription("Validate the current graph for integrity and consistency."),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			sd := getOrCreateSession(ctx, sm)

			result := sd.Validation.Validate(sd.Graph, nil)

			type finding struct {
				RuleID   string `json:"rule_id"`
				Severity string `json:"severity"`
				Message  string `json:"message"`
				ObjectID string `json:"object_id,omitempty"`
			}

			type validationResult struct {
				Passed  bool `json:"passed"`
				Summary struct {
					TotalRules    int `json:"total_rules"`
					TotalFindings int `json:"total_findings"`
					Errors        int `json:"errors"`
					Warnings      int `json:"warnings"`
					Infos         int `json:"infos"`
				} `json:"summary"`
				Findings []finding `json:"findings,omitempty"`
			}

			resp := validationResult{
				Passed: result.Passed,
			}
			resp.Summary.TotalRules = result.Summary.TotalRules
			resp.Summary.TotalFindings = result.Summary.TotalFindings
			resp.Summary.Errors = result.Summary.Errors
			resp.Summary.Warnings = result.Summary.Warnings
			resp.Summary.Infos = result.Summary.Infos

			for _, f := range result.Findings {
				resp.Findings = append(resp.Findings, finding{
					RuleID:   f.RuleID,
					Severity: string(f.Severity),
					Message:  f.Message,
					ObjectID: f.ObjectID,
				})
			}

			data, err := json.MarshalIndent(resp, "", "  ")
			if err != nil {
				return toolError(fmt.Sprintf("failed to marshal response: %v", err)), nil
			}
			return toolResult(string(data)), nil
		},
	)

	s.AddTool(
		mcp.NewTool("graph_summary",
			mcp.WithDescription("Get a summary of the current graph (entity/relation counts by kind/type)."),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			sd := getOrCreateSession(ctx, sm)

			entities := sd.Graph.Entities()
			relations := sd.Graph.Relations()

			kindCounts := make(map[core.EntityKind]int)
			for _, e := range entities {
				kindCounts[e.Kind]++
			}

			relTypeCounts := make(map[core.RelationType]int)
			for _, r := range relations {
				relTypeCounts[r.Type]++
			}

			type summary struct {
				TotalEntities  int            `json:"total_entities"`
				TotalRelations int            `json:"total_relations"`
				EntityKinds    map[string]int `json:"entity_kinds"`
				RelationTypes  map[string]int `json:"relation_types"`
			}

			ek := make(map[string]int)
			for k, v := range kindCounts {
				ek[string(k)] = v
			}
			rt := make(map[string]int)
			for k, v := range relTypeCounts {
				rt[string(k)] = v
			}

			resp := summary{
				TotalEntities:  len(entities),
				TotalRelations: len(relations),
				EntityKinds:    ek,
				RelationTypes:  rt,
			}

			data, err := json.MarshalIndent(resp, "", "  ")
			if err != nil {
				return toolError(fmt.Sprintf("failed to marshal response: %v", err)), nil
			}
			return toolResult(string(data)), nil
		},
	)

	s.AddTool(
		mcp.NewTool("clear_graph",
			mcp.WithDescription("Clear the entire graph, removing all entities and relations."),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			sd := getOrCreateSession(ctx, sm)
			sd.Graph = core.NewGraph()
			return toolResult("Graph cleared"), nil
		},
	)
}
