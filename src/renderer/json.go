package renderer

import (
	"encoding/json"
	"fmt"
	"strings"

	"IACForge/src/view"
)

// JSONRenderer renders views as JSON documents.
type JSONRenderer struct {
	id     string
	name   string
	format string
}

// NewJSONRenderer creates a new JSON renderer.
func NewJSONRenderer() *JSONRenderer {
	return &JSONRenderer{
		id:     "json",
		name:   "JSON Renderer",
		format: "json",
	}
}

// ID returns the renderer identifier.
func (r *JSONRenderer) ID() string {
	return r.id
}

// Name returns the renderer name.
func (r *JSONRenderer) Name() string {
	return r.name
}

// Format returns the output format.
func (r *JSONRenderer) Format() string {
	return r.format
}

// Render renders a view as JSON.
func (r *JSONRenderer) Render(v *view.ViewResult, opts *RenderOptions) (*Artifact, error) {
	if opts == nil {
		opts = NewRenderOptions()
	}

	indent := 2
	if opts.Options != nil {
		if i, ok := opts.Options["indent"].(int); ok {
			indent = i
		}
	}

	includeMetadata := true
	if opts.Options != nil {
		if m, ok := opts.Options["include_metadata"].(bool); ok {
			includeMetadata = m
		}
	}

	data := make(map[string]interface{})
	data["view_id"] = v.ViewID
	data["title"] = v.Title
	data["description"] = v.Description

	if includeMetadata {
		data["entity_count"] = len(v.VisibleEntities)
		data["relation_count"] = len(v.VisibleRelations)
		data["group_count"] = len(v.Groups)
		data["annotation_count"] = len(v.Annotations)
	}

	entities := make([]map[string]interface{}, 0, len(v.VisibleEntities))
	for _, entity := range v.VisibleEntities {
		e := map[string]interface{}{
			"id":   entity.ID,
			"kind": entity.Kind,
			"name": entity.Name,
		}
		if entity.Owner != "" {
			e["owner"] = entity.Owner
		}
		if entity.Status != "" {
			e["status"] = entity.Status
		}
		if len(entity.Tags) > 0 {
			e["tags"] = entity.Tags
		}
		if len(entity.Properties) > 0 {
			e["properties"] = entity.Properties
		}
		entities = append(entities, e)
	}
	data["entities"] = entities

	hierarchy := make([]map[string]interface{}, 0, len(v.VisibleEntities))
	for _, root := range buildOwnershipTree(v.VisibleEntities) {
		hierarchy = append(hierarchy, entityHierarchyNode(root))
	}
	data["hierarchy"] = hierarchy

	relations := make([]map[string]interface{}, 0, len(v.VisibleRelations))
	for _, rel := range v.VisibleRelations {
		r := map[string]interface{}{
			"id":   rel.ID,
			"type": rel.Type,
			"participants": map[string]interface{}{
				"source": rel.Source(),
				"target": rel.Target(),
			},
		}
		if rel.Direction != "" {
			r["direction"] = rel.Direction
		}
		if rel.Status != "" {
			r["status"] = rel.Status
		}
		relations = append(relations, r)
	}
	data["relations"] = relations

	groups := make([]map[string]interface{}, 0, len(v.Groups))
	for _, group := range v.Groups {
		g := map[string]interface{}{
			"id":      group.ID,
			"kind":    group.Kind,
			"name":    group.Name,
			"members": group.Members,
		}
		if len(group.Properties) > 0 {
			g["properties"] = group.Properties
		}
		groups = append(groups, g)
	}
	data["groups"] = groups

	if includeMetadata {
		data["lifted_relation_count"] = len(v.LiftedRelations)
		data["lifted_group_count"] = len(v.LiftedGroups)
	}

	liftedRelations := make([]map[string]interface{}, 0, len(v.LiftedRelations))
	for _, rel := range v.LiftedRelations {
		lr := map[string]interface{}{
			"id":               rel.ID,
			"type":             rel.Type,
			"source_ref":       rel.SourceRef,
			"target_ref":       rel.TargetRef,
			"via":              rel.Via,
			"aggregated_count": rel.AggregatedCount,
		}
		if rel.Direction != "" {
			lr["direction"] = rel.Direction
		}
		liftedRelations = append(liftedRelations, lr)
	}
	if len(liftedRelations) > 0 {
		data["lifted_relations"] = liftedRelations
	}

	if len(v.LiftedGroups) > 0 {
		liftedGroups := make([]map[string]interface{}, 0, len(v.LiftedGroups))
		for _, group := range v.LiftedGroups {
			liftedGroups = append(liftedGroups, map[string]interface{}{
				"id":      group.ID,
				"kind":    group.Kind,
				"name":    group.Name,
				"members": group.Members,
			})
		}
		data["lifted_groups"] = liftedGroups
	}

	data["annotations"] = v.Annotations

	jsonBytes, err := json.MarshalIndent(data, "", strings.Repeat(" ", indent))
	if err != nil {
		return nil, fmt.Errorf("failed to marshal JSON: %w", err)
	}

	artifact := NewArtifact(
		fmt.Sprintf("artifact-%s-%s", r.id, v.ViewID),
		r.id,
		v.ViewID,
		r.format,
		string(jsonBytes),
	)
	artifact.Metadata["title"] = v.Title
	artifact.Metadata["description"] = v.Description

	return artifact, nil
}

// entityHierarchyNode converts an OwnershipNode into a nested JSON structure.
func entityHierarchyNode(node *OwnershipNode) map[string]interface{} {
	out := map[string]interface{}{
		"id":   node.Entity.ID,
		"kind": node.Entity.Kind,
		"name": node.Entity.Name,
	}
	if len(node.Children) > 0 {
		children := make([]map[string]interface{}, 0, len(node.Children))
		for _, child := range node.Children {
			children = append(children, entityHierarchyNode(child))
		}
		out["children"] = children
	}
	return out
}
