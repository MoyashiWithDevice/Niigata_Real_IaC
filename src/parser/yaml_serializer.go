package parser

import (
	"bytes"
	"fmt"
	"os"
	"sort"

	"gopkg.in/yaml.v3"

	"IACForge/src/core"
	"IACForge/src/schema"
)

// Serializer serializes a Graph to YAML syntax.
type Serializer struct {
	indent int
	schema *schema.Schema
}

// NewSerializer creates a new YAML serializer with the core schema.
func NewSerializer() *Serializer {
	return &Serializer{
		indent: 2,
		schema: schema.CoreSchema(),
	}
}

// NewSerializerWithSchema creates a new YAML serializer with a custom schema.
func NewSerializerWithSchema(s *schema.Schema) *Serializer {
	return &Serializer{
		indent: 2,
		schema: s,
	}
}

// SerializeFile writes a Graph to a YAML file.
func (s *Serializer) SerializeFile(g *core.Graph, path string) error {
	data, err := s.Serialize(g)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

// Serialize writes a Graph to YAML bytes.
func (s *Serializer) Serialize(g *core.Graph) ([]byte, error) {
	doc := s.buildDocument(g)

	var buf bytes.Buffer
	encoder := yaml.NewEncoder(&buf)
	encoder.SetIndent(s.indent)

	if err := encoder.Encode(doc); err != nil {
		return nil, fmt.Errorf("failed to encode YAML: %w", err)
	}

	if err := encoder.Close(); err != nil {
		return nil, fmt.Errorf("failed to close YAML encoder: %w", err)
	}

	return buf.Bytes(), nil
}

// buildDocument constructs the YAML document structure.
func (s *Serializer) buildDocument(g *core.Graph) map[string]interface{} {
	objects := make([]interface{}, 0)

	// Build a map of children grouped by parent ID and nest key
	childrenByParent := make(map[string][]childEntry)

	for _, e := range g.Entities() {
		if e.IsRoot() {
			continue
		}
		parent, ok := g.GetEntity(e.Owner)
		if !ok {
			continue
		}
		nd, ok := s.schema.FindNestingByChildKind(parent.Kind, e.Kind)
		if !ok {
			continue
		}
		childrenByParent[e.Owner] = append(childrenByParent[e.Owner], childEntry{
			entity:  e,
			nestKey: nd.NestKey,
		})
	}

	// Sort children by nest key, then by ID for deterministic output
	for parentID := range childrenByParent {
		entries := childrenByParent[parentID]
		sort.Slice(entries, func(i, j int) bool {
			if entries[i].nestKey != entries[j].nestKey {
				return entries[i].nestKey < entries[j].nestKey
			}
			return entries[i].entity.ID < entries[j].entity.ID
		})
		childrenByParent[parentID] = entries
	}

	// Collect set of entity IDs that are nested children (to exclude from top-level)
	nestedIDs := make(map[string]bool)
	for _, entries := range childrenByParent {
		for _, entry := range entries {
			nestedIDs[entry.entity.ID] = true
		}
	}

	// Add entities that are not nested children (roots and non-nestable owned entities)
	entities := g.Entities()
	sort.Slice(entities, func(i, j int) bool {
		return entities[i].ID < entities[j].ID
	})

	for _, e := range entities {
		if nestedIDs[e.ID] {
			continue
		}
		objects = append(objects, s.buildEntityWithChildren(e, childrenByParent, false))
	}

	// Add relations (sorted by ID for deterministic output)
	relations := g.Relations()
	sort.Slice(relations, func(i, j int) bool {
		return relations[i].ID < relations[j].ID
	})

	for _, r := range relations {
		objects = append(objects, s.buildRelation(r))
	}

	return map[string]interface{}{
		"objects": objects,
	}
}

// childEntry holds a nested child entity and its nest key.
type childEntry struct {
	entity  *core.Entity
	nestKey string
}

// buildEntityWithChildren constructs the YAML representation of an entity,
// including nested children in the spec section.
// When isNested is true, the output omits owner and kind (inferred from context).
func (s *Serializer) buildEntityWithChildren(e *core.Entity, childrenByParent map[string][]childEntry, isNested bool) map[string]interface{} {
	obj := make(map[string]interface{})

	// Always include id
	obj["id"] = e.ID

	// Only output kind for top-level entities (nested kind is inferred from nest key)
	if !isNested {
		obj["kind"] = string(e.Kind)
	}

	// Always include name for top-level entities; for nested, only if different from id
	if !isNested || e.Name != e.ID {
		obj["name"] = e.Name
	}

	// Build attributes sub-key (omit owner for nested entities)
	if attrs := buildAttributes(e.Description, e.Status, e.Tags, e.Labels, e.Extensions, !isNested, e.Owner); attrs != nil {
		obj["attributes"] = attrs
	}

	// Build spec sub-key for kind-specific properties
	spec := buildSpec(e.Properties)

	// Group children by nest key and recurse
	directChildren := childrenByParent[e.ID]
	if len(directChildren) > 0 {
		nestKeyGroups := make(map[string][]childEntry)
		for _, c := range directChildren {
			nestKeyGroups[c.nestKey] = append(nestKeyGroups[c.nestKey], c)
		}

		nestKeys := make([]string, 0, len(nestKeyGroups))
		for k := range nestKeyGroups {
			nestKeys = append(nestKeys, k)
		}
		sort.Strings(nestKeys)

		for _, nestKey := range nestKeys {
			group := nestKeyGroups[nestKey]
			nestedList := make([]interface{}, 0, len(group))
			for _, ce := range group {
				nestedList = append(nestedList, s.buildEntityWithChildren(ce.entity, childrenByParent, true))
			}
			spec[nestKey] = nestedList
		}
	}

	if len(spec) > 0 {
		obj["spec"] = spec
	}

	return obj
}

// buildRelation constructs the YAML representation of a relation.
func (s *Serializer) buildRelation(r *core.Relation) map[string]interface{} {
	obj := make(map[string]interface{})

	// Required fields at top level
	obj["id"] = r.ID
	obj["type"] = string(r.Type)

	// Participants at top level
	if r.Direction == core.DirectionSymmetric {
		obj["participants"] = r.Participants.List
	} else {
		participants := make(map[string]interface{})
		if r.Participants.Source != "" {
			participants["source"] = r.Participants.Source
		}
		if r.Participants.Target != "" {
			participants["target"] = r.Participants.Target
		}
		obj["participants"] = participants
	}

	// Build attributes sub-key
	if attrs := buildAttributes(r.Description, r.Status, r.Tags, r.Labels, r.Extensions, false, ""); attrs != nil {
		obj["attributes"] = attrs
	}

	// Build spec sub-key for relation-type-specific properties
	if spec := buildSpec(r.Properties); len(spec) > 0 {
		obj["spec"] = spec
	}

	return obj
}

// buildAttributes constructs the shared attributes map for entities and relations.
// Owner is included only when includeOwner is true and owner is non-empty.
// Returns nil when no attributes are present.
func buildAttributes(description string, status core.Status, tags []string, labels map[string]string, extensions map[string]interface{}, includeOwner bool, owner string) map[string]interface{} {
	attrs := make(map[string]interface{})
	if includeOwner && owner != "" {
		attrs["owner"] = owner
	}
	if description != "" {
		attrs["description"] = description
	}
	if status != "" {
		attrs["status"] = string(status)
	}
	if len(tags) > 0 {
		attrs["tags"] = tags
	}
	if len(labels) > 0 {
		attrs["labels"] = sortMap(labels)
	}
	if len(extensions) > 0 {
		attrs["extensions"] = sortInterfaceMap(extensions)
	}
	if len(attrs) == 0 {
		return nil
	}
	return attrs
}

// buildSpec constructs a spec map from kind-specific properties, sorted for
// deterministic YAML output. Returns an empty map when no properties exist.
func buildSpec(properties map[string]interface{}) map[string]interface{} {
	spec := make(map[string]interface{})
	sortedKeys := make([]string, 0, len(properties))
	for k := range properties {
		sortedKeys = append(sortedKeys, k)
	}
	sort.Strings(sortedKeys)
	for _, k := range sortedKeys {
		spec[k] = serializePropertyValue(properties[k])
	}
	return spec
}

// sortMap returns a sorted copy of a string map for deterministic YAML output.
func sortMap(m map[string]string) map[string]string {
	if m == nil {
		return nil
	}
	result := make(map[string]string, len(m))
	for k, v := range m {
		result[k] = v
	}
	return result
}

// serializePropertyValue converts a property value for YAML output.
// ReferenceValue is serialized with the @ prefix.
func serializePropertyValue(v interface{}) interface{} {
	if ref, ok := v.(core.ReferenceValue); ok {
		return ref.String()
	}
	return v
}

// sortInterfaceMap returns a sorted copy of a map[string]interface{} for deterministic YAML output.
func sortInterfaceMap(m map[string]interface{}) map[string]interface{} {
	if m == nil {
		return nil
	}
	result := make(map[string]interface{}, len(m))
	for k, v := range m {
		result[k] = v
	}
	return result
}
