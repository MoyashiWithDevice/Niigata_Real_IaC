package renderer

import (
	"fmt"
	"sort"
	"strings"

	"github.com/bababa/Niigata_Real_IaC/src/core"
	"github.com/bababa/Niigata_Real_IaC/src/view"
)

// MarkdownRenderer renders views as Markdown documents.
type MarkdownRenderer struct {
	id     string
	name   string
	format string
}

// NewMarkdownRenderer creates a new Markdown renderer.
func NewMarkdownRenderer() *MarkdownRenderer {
	return &MarkdownRenderer{
		id:     "markdown",
		name:   "Markdown Renderer",
		format: "md",
	}
}

// ID returns the renderer identifier.
func (r *MarkdownRenderer) ID() string {
	return r.id
}

// Name returns the renderer name.
func (r *MarkdownRenderer) Name() string {
	return r.name
}

// Format returns the output format.
func (r *MarkdownRenderer) Format() string {
	return r.format
}

// Render renders a view as Markdown.
func (r *MarkdownRenderer) Render(v *view.ViewResult, opts *RenderOptions) (*Artifact, error) {
	if opts == nil {
		opts = NewRenderOptions()
	}
	theme := ResolveTheme(opts)

	var md strings.Builder

	includeTOC := true
	if opts.Options != nil {
		if toc, ok := opts.Options["toc"].(bool); ok {
			includeTOC = toc
		}
	}

	md.WriteString("# ")
	md.WriteString(v.Title)
	md.WriteString("\n\n")

	if v.Description != "" {
		md.WriteString(v.Description)
		md.WriteString("\n\n")
	}

	r.writeSummary(&md, v)

	kindGroups := groupEntitiesByKind(v.VisibleEntities)

	if includeTOC && len(v.VisibleEntities) > 0 {
		md.WriteString("## Table of Contents\n\n")
		keys := sortedKeys(kindGroups)
		for _, kind := range keys {
			ids := kindGroups[kind]
			fmt.Fprintf(&md, "- %s**%s** (`%s`, %d)\n", kindIconPrefix(theme, kind), kind, kind, len(ids))
		}
		md.WriteString("\n")
	}

	for _, group := range v.Groups {
		fmt.Fprintf(&md, "## %s\n\n", group.Name)
		fmt.Fprintf(&md, "**Kind:** %s\n\n", group.Kind)
		if count, ok := group.Properties["count"]; ok {
			fmt.Fprintf(&md, "**Members:** %v\n\n", count)
		}
	}

	if len(v.VisibleEntities) > 0 {
		md.WriteString("## Entities\n\n")
		r.writeEntityHierarchy(&md, v.VisibleEntities, theme)
	}

	if len(v.VisibleRelations) > 0 {
		md.WriteString("## Relations\n\n")
		md.WriteString("| ID | Type | Source | Target |\n")
		md.WriteString("|----|------|--------|--------|\n")
		for _, rel := range v.VisibleRelations {
			source, target := relationEndpoints(rel)
			fmt.Fprintf(&md, "| %s | %s | %s | %s |\n",
				escapePipe(rel.ID),
				relTypeLabel(rel),
				escapePipe(source),
				escapePipe(target))
		}
		md.WriteString("\n")
	}

	r.writeEntityDetails(&md, v.VisibleEntities, theme)

	if len(v.Annotations) > 0 {
		md.WriteString("## Annotations\n\n")
		md.WriteString("| Entity | Property | Value |\n")
		md.WriteString("|--------|----------|-------|\n")
		for entityID, annotations := range v.Annotations {
			for prop, value := range annotations {
				fmt.Fprintf(&md, "| %s | %s | %v |\n", entityID, prop, value)
			}
		}
		md.WriteString("\n")
	}

	artifact := NewArtifact(
		fmt.Sprintf("artifact-%s-%s", r.id, v.ViewID),
		r.id,
		v.ViewID,
		r.format,
		md.String(),
	)
	artifact.Metadata["title"] = v.Title
	artifact.Metadata["description"] = v.Description

	return artifact, nil
}

// writeSummary prints a compact overview line with entity/relation totals.
func (r *MarkdownRenderer) writeSummary(m *strings.Builder, v *view.ViewResult) {
	fmt.Fprintf(m, "**%d entities** · **%d relations**\n\n", len(v.VisibleEntities), len(v.VisibleRelations))
}

// writeEntityHierarchy writes visible entities as a nested list reflecting the
// ownership hierarchy. Children are indented inside their parents.
func (r *MarkdownRenderer) writeEntityHierarchy(m *strings.Builder, entities []*core.Entity, theme *Theme) {
	roots := buildOwnershipTree(entities)

	var write func(nodes []*OwnershipNode, depth int)
	write = func(nodes []*OwnershipNode, depth int) {
		indent := strings.Repeat("  ", depth)
		for _, node := range nodes {
			icon := kindIconPrefix(theme, string(node.Entity.Kind))
			fmt.Fprintf(m, "%s- %s**%s** (`%s`, %s)\n",
				indent, icon, node.Entity.Name, node.Entity.ID, node.Entity.Kind)
			if len(node.Children) > 0 {
				write(node.Children, depth+1)
			}
		}
	}

	write(roots, 0)
	m.WriteString("\n")
}

// writeEntityDetails renders a per-entity card (description, status, tags and
// notable properties) grouped by kind. It supplements the tree so readers get
// value without opening the YAML.
func (r *MarkdownRenderer) writeEntityDetails(m *strings.Builder, entities []*core.Entity, theme *Theme) {
	if len(entities) == 0 {
		return
	}
	kindGroups := groupEntitiesByKind(entities)
	keys := sortedKeys(kindGroups)

	m.WriteString("## Entity Details\n\n")
	for _, kind := range keys {
		fmt.Fprintf(m, "### %s%s\n\n", kindIconPrefix(theme, kind), kind)
		for _, e := range kindGroups[kind] {
			fmt.Fprintf(m, "**%s** (`%s`)\n\n", e.Name, e.ID)
			if e.Description != "" {
				fmt.Fprintf(m, "%s\n\n", e.Description)
			}
			var badges []string
			if e.Status != "" {
				badges = append(badges, fmt.Sprintf("status: `%s`", e.Status))
			}
			for _, tag := range e.Tags {
				badges = append(badges, fmt.Sprintf("`%s`", tag))
			}
			if len(badges) > 0 {
				fmt.Fprintf(m, "%s\n\n", strings.Join(badges, " · "))
			}
			if len(e.Properties) > 0 {
				fmt.Fprintf(m, "```\n%s\n```\n\n", formatProperties(e.Properties))
			}
		}
	}
}

// formatProperties renders entity spec properties as compact key=value lines.
func formatProperties(props map[string]interface{}) string {
	keys := make([]string, 0, len(props))
	for k := range props {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var b strings.Builder
	for i, k := range keys {
		if i > 0 {
			b.WriteString("\n")
		}
		fmt.Fprintf(&b, "%s: %v", k, props[k])
	}
	return b.String()
}

// groupEntitiesByKind buckets entities by their kind, preserving input order.
func groupEntitiesByKind(entities []*core.Entity) map[string][]*core.Entity {
	groups := make(map[string][]*core.Entity)
	for _, e := range entities {
		groups[string(e.Kind)] = append(groups[string(e.Kind)], e)
	}
	return groups
}

// relationEndpoints renders the source/target cells of a relation. Symmetric
// relations carry a participant list instead of a source/target pair, so those
// are joined with a bidirectional arrow.
func relationEndpoints(rel *core.Relation) (string, string) {
	if rel.IsSymmetric() {
		ids := rel.ParticipantIDs()
		switch len(ids) {
		case 0:
			return "", ""
		case 1:
			return ids[0], ids[0]
		default:
			return ids[0], ids[1]
		}
	}
	return rel.Source(), rel.Target()
}

// kindIconPrefix returns the kind glyph followed by a single space, or "" when
// the kind has no icon. Callers interpolate it directly before bold text.
func kindIconPrefix(theme *Theme, kind string) string {
	if icon := theme.KindIcon(kind); icon != "" {
		return icon + " "
	}
	return ""
}

// relTypeLabel renders a relation type with a small glyph for quick scanning.
func relTypeLabel(rel *core.Relation) string {
	glyph, ok := relGlyphs[string(rel.Type)]
	if !ok {
		glyph = "🔗"
	}
	return fmt.Sprintf("%s `%s`", glyph, rel.Type)
}

// relGlyphs maps common relation types to a short glyph.
var relGlyphs = map[string]string{
	"belongs_to": "🏠",
	"located_in": "📍",
	"inhabits":   "🐾",
	"flows_into": "➡️",
	"near":       "↔️",
	"depends_on": "🔗",
	"connects":   "🔗",
}

// escaped constants
func escapePipe(s string) string {
	return strings.ReplaceAll(s, "|", "\\|")
}

// sortedKeys returns the keys of a string-indexed map in sorted order.
func sortedKeys(m map[string][]*core.Entity) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
