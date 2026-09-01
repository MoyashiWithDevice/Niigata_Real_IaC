package renderer

import (
	"fmt"
	"strings"

	"github.com/bababa/Niigata_Real_IaC/src/core"
	"github.com/bababa/Niigata_Real_IaC/src/view"
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
	theme := ResolveTheme(opts)

	direction := "TB"
	if opts.Options != nil {
		if d, ok := opts.Options["direction"].(string); ok {
			direction = d
		}
	}

	var mermaid strings.Builder

	r.writeInit(&mermaid, theme)

	mermaid.WriteString("graph ")
	mermaid.WriteString(direction)
	mermaid.WriteString("\n")

	r.writeClassDefs(&mermaid, v, theme)
	r.writeOwnershipTree(&mermaid, v.VisibleEntities, theme)

	for _, group := range v.Groups {
		fmt.Fprintf(&mermaid, "    subgraph %s[\"%s\"]\n", sanitizeID(group.ID), escapeMermaid(group.Name))
		for _, memberID := range group.Members {
			fmt.Fprintf(&mermaid, "        %s:::kind_%s\n", sanitizeID(memberID), kindClassName(group.Kind))
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

	r.writeContainerStyles(&mermaid, v, theme)

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

// writeInit emits a Mermaid init directive that applies the theme colors
// globally (edges, node defaults, background).
func (r *MermaidRenderer) writeInit(m *strings.Builder, theme *Theme) {
	fmt.Fprintf(m, "%%%%{init: {\"theme\": \"base\", \"themeVariables\": {\n")
	fmt.Fprintf(m, "  \"primaryColor\": \"%s\",\n", theme.AccentColor(""))
	fmt.Fprintf(m, "  \"primaryTextColor\": \"%s\",\n", theme.TextColor())
	fmt.Fprintf(m, "  \"primaryBorderColor\": \"%s\",\n", theme.AccentColor(""))
	fmt.Fprintf(m, "  \"lineColor\": \"%s\",\n", theme.LineColor(""))
	fmt.Fprintf(m, "  \"fontFamily\": \"%s\",\n", theme.FontFamily())
	fmt.Fprintf(m, "  \"fontSize\": \"%dpx\",\n", theme.FontSize())
	if b := theme.BackgroundColor(); b != "" {
		fmt.Fprintf(m, "  \"background\": \"%s\",\n", b)
	}
	m.WriteString("  \"clusterBkg\": \"#ffffff\"")
	m.WriteString("\n}}}%%\n")
}

// writeClassDefs emits a class definition per distinct entity kind so nodes are
// colored by their kind. Classes are emitted only for kinds actually present.
func (r *MermaidRenderer) writeClassDefs(m *strings.Builder, v *view.ViewResult, theme *Theme) {
	seen := make(map[string]struct{})
	for _, e := range v.VisibleEntities {
		if _, ok := seen[string(e.Kind)]; ok {
			continue
		}
		seen[string(e.Kind)] = struct{}{}
		color := theme.AccentColor(string(e.Kind))
		fmt.Fprintf(m, "    classDef %s fill:%s,stroke:%s,color:#ffffff\n",
			kindClassName(string(e.Kind)), color, color)
	}
}

// writeOwnershipTree declares visible entities as Mermaid nodes nested in
// subgraphs according to the ownership hierarchy. Parents are declared before
// their children so containment renders correctly.
func (r *MermaidRenderer) writeOwnershipTree(m *strings.Builder, entities []*core.Entity, theme *Theme) {
	var write func(nodes []*OwnershipNode, depth int)
	write = func(nodes []*OwnershipNode, depth int) {
		indent := strings.Repeat("    ", depth+1)
		for _, node := range nodes {
			id := sanitizeID(node.Entity.ID)
			name := escapeMermaid(node.Entity.Name)
			if len(node.Children) > 0 {
				fmt.Fprintf(m, "%ssubgraph %s[\"%s\"]\n", indent, id, name)
				write(node.Children, depth+1)
				fmt.Fprintf(m, "%send\n", indent)
			} else {
				fmt.Fprintf(m, "%s%s[\"%s\"]:::%s\n", indent, id, name, kindClassName(string(node.Entity.Kind)))
			}
		}
	}

	write(buildOwnershipTree(entities), 0)
}

// writeContainerStyles colors the subgraphs that correspond to container
// entities, using the entity kind accent color as the cluster border.
func (r *MermaidRenderer) writeContainerStyles(m *strings.Builder, v *view.ViewResult, theme *Theme) {
	for _, e := range v.VisibleEntities {
		if len(entityChildrenByOwner(v, e.ID)) == 0 {
			continue
		}
		color := theme.AccentColor(string(e.Kind))
		fmt.Fprintf(m, "    style %s fill:%s,stroke:%s\n", sanitizeID(e.ID), theme.ContainerFill(), color)
	}
}

// entityChildrenByOwner returns the visible entities owned by the given ID.
func entityChildrenByOwner(v *view.ViewResult, owner string) []*core.Entity {
	var out []*core.Entity
	for _, e := range v.VisibleEntities {
		if e.Owner == owner {
			out = append(out, e)
		}
	}
	return out
}

// kindClassName returns the Mermaid class name for a kind.
func kindClassName(kind string) string {
	return "kind_" + sanitizeID(kind)
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
