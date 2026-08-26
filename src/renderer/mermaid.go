package renderer

import (
	"fmt"
	"strings"

	"IACForge/src/core"
	"IACForge/src/view"
)

// MermaidRenderer renders views as Mermaid diagrams.
type MermaidRenderer struct {
	id     string
	name   string
	format string
}

// NewMermaidRenderer creates a new Mermaid renderer.
func NewMermaidRenderer() *MermaidRenderer {
	return &MermaidRenderer{
		id:     "mermaid",
		name:   "Mermaid Renderer",
		format: "mmd",
	}
}

// ID returns the renderer identifier.
func (r *MermaidRenderer) ID() string {
	return r.id
}

// Name returns the renderer name.
func (r *MermaidRenderer) Name() string {
	return r.name
}

// Format returns the output format.
func (r *MermaidRenderer) Format() string {
	return r.format
}

// Render renders a view as Mermaid diagram.
func (r *MermaidRenderer) Render(v *view.ViewResult, opts *RenderOptions) (*Artifact, error) {
	if opts == nil {
		opts = NewRenderOptions()
	}

	direction := "TB"
	if opts.Options != nil {
		if d, ok := opts.Options["direction"].(string); ok {
			direction = d
		}
	}

	var mermaid strings.Builder

	mermaid.WriteString("graph ")
	mermaid.WriteString(direction)
	mermaid.WriteString("\n")

	r.writeOwnershipTree(&mermaid, v.VisibleEntities)

	for _, group := range v.Groups {
		fmt.Fprintf(&mermaid, "    subgraph %s[\"%s\"]\n", sanitizeID(group.ID), escapeMermaid(group.Name))
		for _, memberID := range group.Members {
			fmt.Fprintf(&mermaid, "        %s\n", sanitizeID(memberID))
		}
		mermaid.WriteString("    end\n")
	}

	// Lifted groups are structural containers (e.g. hidden clusters hosting
	// visible applications) introduced by the relation lift step. They must be
	// declared before edges reference them.
	for _, group := range v.LiftedGroups {
		fmt.Fprintf(&mermaid, "    subgraph %s[\"%s\"]\n", sanitizeID(group.ID), escapeMermaid(group.Name))
		for _, memberID := range group.Members {
			fmt.Fprintf(&mermaid, "        %s\n", sanitizeID(memberID))
		}
		mermaid.WriteString("    end\n")
	}

	nodeSet := make(map[string]struct{}, len(v.VisibleEntities))
	for _, entity := range v.VisibleEntities {
		nodeSet[entity.ID] = struct{}{}
	}

	for _, rel := range v.VisibleRelations {
		source, target := rel.Source(), rel.Target()
		if source == "" || target == "" {
			// Symmetric relations (e.g. connects) carry a participant list
			// rather than source/target. Draw the edge between the first two.
			// Relations with more than two participants are only partially
			// drawn; Mermaid has no native hyperedge support.
			ids := rel.ParticipantIDs()
			if len(ids) >= 2 {
				source, target = ids[0], ids[1]
			}
		}
		if source == "" || target == "" {
			// Skip malformed relations without two endpoints.
			continue
		}
		if _, ok := nodeSet[source]; !ok {
			continue
		}
		if _, ok := nodeSet[target]; !ok {
			continue
		}
		source = sanitizeID(source)
		target = sanitizeID(target)
		edgeLabel := string(rel.Type)
		if rel.Description != "" {
			edgeLabel = rel.Description
		}
		fmt.Fprintf(&mermaid, "    %s -->|%s| %s\n", source, escapeMermaid(edgeLabel), target)
	}

	r.writeLiftedRelations(&mermaid, v)

	artifact := NewArtifact(
		fmt.Sprintf("artifact-%s-%s", r.id, v.ViewID),
		r.id,
		v.ViewID,
		r.format,
		mermaid.String(),
	)
	artifact.Metadata["title"] = v.Title
	artifact.Metadata["description"] = v.Description

	return artifact, nil
}

// writeOwnershipTree declares visible entities as Mermaid nodes nested in
// subgraphs according to the ownership hierarchy. Parents are declared
// before their children so containment renders correctly.
func (r *MermaidRenderer) writeOwnershipTree(mermaid *strings.Builder, entities []*core.Entity) {
	var write func(nodes []*OwnershipNode, depth int)
	write = func(nodes []*OwnershipNode, depth int) {
		indent := strings.Repeat("    ", depth+1)
		for _, node := range nodes {
			id := sanitizeID(node.Entity.ID)
			name := escapeMermaid(node.Entity.Name)
			if len(node.Children) > 0 {
				fmt.Fprintf(mermaid, "%ssubgraph %s[\"%s\"]\n", indent, id, name)
				write(node.Children, depth+1)
				fmt.Fprintf(mermaid, "%send\n", indent)
			} else {
				fmt.Fprintf(mermaid, "%s%s[\"%s\"]\n", indent, id, name)
			}
		}
	}

	write(buildOwnershipTree(entities), 0)
}

// sanitizeID replaces special characters for Mermaid IDs.
func sanitizeID(id string) string {
	result := strings.ReplaceAll(id, "-", "_")
	result = strings.ReplaceAll(result, ".", "_")
	result = strings.ReplaceAll(result, "/", "_")
	return result
}

// writeLiftedRelations emits the derived edges produced by the relation lift
// step. Lifted edges are drawn dashed to distinguish them from explicitly
// modeled relations. Endpoints may reference visible entities or lifted
// structural groups; references that resolve to nothing are skipped.
func (r *MermaidRenderer) writeLiftedRelations(mermaid *strings.Builder, v *view.ViewResult) {
	if len(v.LiftedRelations) == 0 {
		return
	}

	refs := make(map[string]struct{}, len(v.VisibleEntities)+len(v.Groups)+len(v.LiftedGroups))
	for _, entity := range v.VisibleEntities {
		refs[entity.ID] = struct{}{}
	}
	for _, group := range v.Groups {
		refs[group.ID] = struct{}{}
	}
	for _, group := range v.LiftedGroups {
		refs[group.ID] = struct{}{}
	}

	for _, lr := range v.LiftedRelations {
		if _, ok := refs[lr.SourceRef]; !ok {
			continue
		}
		if _, ok := refs[lr.TargetRef]; !ok {
			continue
		}
		source := sanitizeID(lr.SourceRef)
		target := sanitizeID(lr.TargetRef)
		label := string(lr.Type)
		if lr.AggregatedCount > 1 {
			label = fmt.Sprintf("%s ×%d", label, lr.AggregatedCount)
		}
		arrow := "-.->"
		if lr.Direction == core.DirectionSymmetric {
			arrow = "-.-"
		}
		fmt.Fprintf(mermaid, "    %s %s|%s| %s\n", source, arrow, escapeMermaid(label), target)
	}
}

// escapeMermaid escapes special Mermaid characters.
func escapeMermaid(s string) string {
	s = strings.ReplaceAll(s, "\"", "'")
	s = strings.ReplaceAll(s, "\n", " ")
	return s
}
