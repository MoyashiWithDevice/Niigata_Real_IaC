package mcp

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"
	"gopkg.in/yaml.v3"

	"IACForge/src/renderer"
	"IACForge/src/view"
)

func registerRenderMCPTools(s *mcpserver.MCPServer, sm *SessionManager) {
	s.AddTool(
		mcp.NewTool("render_graph",
			mcp.WithDescription("Render the current graph to markdown, mermaid, svg, or json. Optionally filter by entity kinds, group by a field, or pass a full view definition in YAML."),
			mcp.WithString("format", mcp.Description("Output format (markdown, mermaid, svg, json). Default: markdown")),
			mcp.WithArray("kinds", mcp.Description("Restrict rendering to these entity kinds"), mcp.WithStringItems()),
			mcp.WithString("group_by", mcp.Description("Field to group entities by (e.g. status)")),
			mcp.WithString("view_yaml", mcp.Description("Full view definition as YAML (id, visibility, grouping, annotations)")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			sd := getOrCreateSession(ctx, sm)

			v, err := buildView(req)
			if err != nil {
				return toolError(fmt.Sprintf("failed to build view: %v", err)), nil
			}

			ve := view.NewEngine(sd.Graph)
			vr, err := ve.Apply(v)
			if err != nil {
				return toolError(fmt.Sprintf("view application failed: %v", err)), nil
			}

			content, err := renderer.RenderFormat(vr, req.GetString("format", "markdown"))
			if err != nil {
				return toolError(fmt.Sprintf("render failed: %v", err)), nil
			}
			return toolResult(content), nil
		},
	)
}

// buildView constructs a View from the render_graph request parameters.
func buildView(req mcp.CallToolRequest) (*view.View, error) {
	if viewYAML := req.GetString("view_yaml", ""); viewYAML != "" {
		var v view.View
		if err := yaml.Unmarshal([]byte(viewYAML), &v); err != nil {
			return nil, fmt.Errorf("invalid view_yaml: %w", err)
		}
		if v.ID == "" {
			return nil, fmt.Errorf("view_yaml must specify an id")
		}
		return &v, nil
	}

	v := view.NewView("mcp-render", "Infrastructure View")

	kinds := req.GetStringSlice("kinds", nil)
	if len(kinds) > 0 {
		// Show only the requested kinds; relations remain visible.
		for _, kind := range kinds {
			rule := view.NewVisibilityRule(view.VisibilityTargetEntities, view.VisibilityActionShow)
			rule.Kind = kind
			v.AddVisibility(rule)
		}
		v.AddVisibility(view.NewVisibilityRule(view.VisibilityTargetRelations, view.VisibilityActionShow))
	} else {
		v.AddVisibility(view.NewVisibilityRule(view.VisibilityTargetEntities, view.VisibilityActionShow))
		v.AddVisibility(view.NewVisibilityRule(view.VisibilityTargetRelations, view.VisibilityActionShow))
	}

	if groupBy := req.GetString("group_by", ""); groupBy != "" {
		gr := &view.GroupingRule{
			TargetKind: "",
			GroupKind:  "group",
			GroupBy:    []string{groupBy},
		}
		v.AddGrouping(gr)
	}

	return v, nil
}
