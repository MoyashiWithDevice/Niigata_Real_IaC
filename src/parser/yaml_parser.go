package parser

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"

	"IACForge/src/core"
	"IACForge/src/schema"
)

// ParseError represents a parsing error with location information.
type ParseError struct {
	Line    int
	Column  int
	Message string
}

func (e *ParseError) Error() string {
	return fmt.Sprintf("parse error at line %d, column %d: %s", e.Line, e.Column, e.Message)
}

// Parser parses YAML syntax into a Graph.
type Parser struct {
	graph         *core.Graph
	schema        *schema.Schema
	autoRelConfig *AutoRelationConfig
}

// AutoRelationConfig holds configuration for auto-relation generation.
type AutoRelationConfig struct {
	// Overrides allows customizing auto-relation mapping.
	// Key format: "parentKind.nestKey"
	Overrides map[string]AutoRelationMapping
	// Disabled disables auto-relation generation entirely.
	Disabled bool
}

// AutoRelationMapping defines a custom auto-relation mapping.
type AutoRelationMapping struct {
	RelationType core.RelationType
	Source       string // "parent" or "child"
}

// NewParser creates a new YAML parser with the core schema.
func NewParser() *Parser {
	return &Parser{
		graph:  core.NewGraph(),
		schema: schema.CoreSchema(),
	}
}

// NewParserWithSchema creates a new YAML parser with a custom schema.
func NewParserWithSchema(s *schema.Schema) *Parser {
	return &Parser{
		graph:  core.NewGraph(),
		schema: s,
	}
}

// NewParserWithAutoRelationConfig creates a new YAML parser with custom auto-relation config.
func NewParserWithAutoRelationConfig(s *schema.Schema, config *AutoRelationConfig) *Parser {
	return &Parser{
		graph:         core.NewGraph(),
		schema:        s,
		autoRelConfig: config,
	}
}

// ParseFile reads a YAML file and parses it into a Graph.
func (p *Parser) ParseFile(path string) (*core.Graph, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}
	return p.Parse(data)
}

// Parse parses YAML bytes into a Graph.
func (p *Parser) Parse(data []byte) (*core.Graph, error) {
	p.graph = core.NewGraph()

	var doc struct {
		Objects []map[string]interface{} `yaml:"objects"`
	}

	if err := yaml.Unmarshal(data, &doc); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	if err := p.parseObjects(doc.Objects); err != nil {
		return nil, err
	}

	if err := p.graph.BuildOwnershipPaths(); err != nil {
		return nil, fmt.Errorf("failed to build ownership paths: %w", err)
	}

	return p.graph, nil
}

// ParseDir reads all YAML files from a directory recursively and merges them into a single Graph.
// File names are sorted to ensure deterministic loading order.
func (p *Parser) ParseDir(dir string) (*core.Graph, error) {
	p.graph = core.NewGraph()

	var allObjects []map[string]interface{}

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		ext := strings.ToLower(filepath.Ext(path))
		if ext != ".yaml" && ext != ".yml" {
			return nil
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read %s: %w", path, err)
		}

		var doc struct {
			Objects []map[string]interface{} `yaml:"objects"`
		}
		if err := yaml.Unmarshal(data, &doc); err != nil {
			return fmt.Errorf("failed to parse %s: %w", path, err)
		}

		allObjects = append(allObjects, doc.Objects...)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to walk directory %s: %w", dir, err)
	}

	if err := p.parseObjects(allObjects); err != nil {
		return nil, err
	}

	if err := p.graph.BuildOwnershipPaths(); err != nil {
		return nil, fmt.Errorf("failed to build ownership paths: %w", err)
	}

	return p.graph, nil
}

// Load is a convenience function that loads a path (file or directory) into a Graph.
func (p *Parser) Load(path string) (*core.Graph, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("failed to access %s: %w", path, err)
	}
	if info.IsDir() {
		return p.ParseDir(path)
	}
	return p.ParseFile(path)
}

// parseObjects processes a list of object definitions.
func (p *Parser) parseObjects(objects []map[string]interface{}) error {
	// First pass: parse all entities (including nested ones)
	entities := make(map[string]*core.Entity)
	var relations []*rawRelation
	nestedIDs := make(map[string]bool) // track entities created via nesting

	for i, obj := range objects {
		if obj == nil {
			continue
		}

		if _, hasKind := obj["kind"]; hasKind {
			entity, nestedEntities, err := p.parseEntity(obj, "")
			if err != nil {
				return fmt.Errorf("object at index %d: %w", i, err)
			}
			entities[entity.ID] = entity
			for _, ne := range nestedEntities {
				entities[ne.ID] = ne
				nestedIDs[ne.ID] = true
			}
		} else if _, hasType := obj["type"]; hasType {
			relations = append(relations, &rawRelation{index: i, data: obj})
		}
	}

	// Add entities to graph
	for _, entity := range entities {
		if err := p.graph.AddEntity(entity); err != nil {
			return fmt.Errorf("failed to add entity %s: %w", entity.ID, err)
		}
	}

	// Generate auto-relations from nesting (before explicit relations)
	autoRelations := p.generateAutoRelations(entities, nestedIDs)

	// Second pass: parse relations (need entities to be added first for reference validation)
	for _, raw := range relations {
		relation, err := p.parseRelation(raw.data, entities)
		if err != nil {
			return fmt.Errorf("object at index %d: %w", raw.index, err)
		}
		if err := p.graph.AddRelation(relation); err != nil {
			return fmt.Errorf("failed to add relation %s: %w", relation.ID, err)
		}
	}

	// Add auto-relations (skip duplicates)
	for _, ar := range autoRelations {
		if hasDuplicateRelation(ar, relations) {
			continue
		}
		if err := p.graph.AddRelation(ar); err != nil {
			return fmt.Errorf("failed to add auto-relation %s: %w", ar.ID, err)
		}
	}

	return nil
}

// rawRelation holds a relation definition before parsing.
type rawRelation struct {
	index int
	data  map[string]interface{}
}

// generateAutoRelations generates relations automatically from nesting relationships.
func (p *Parser) generateAutoRelations(entities map[string]*core.Entity, nestedIDs map[string]bool) []*core.Relation {
	if p.autoRelConfig != nil && p.autoRelConfig.Disabled {
		return nil
	}

	var relations []*core.Relation

	for _, entity := range entities {
		if entity.Owner == "" {
			continue
		}
		// Only generate auto-relations for entities created via nesting
		if !nestedIDs[entity.ID] {
			continue
		}
		parent, ok := entities[entity.Owner]
		if !ok {
			continue
		}

		nd := p.findAutoRelationDef(parent.Kind, entity.Kind)
		if nd == nil {
			continue
		}

		rel := p.buildAutoRelation(nd, parent, entity)
		if rel != nil {
			relations = append(relations, rel)
		}
	}

	return relations
}

// findAutoRelationDef finds the nesting definition with auto-relation for given parent/child kinds.
func (p *Parser) findAutoRelationDef(parentKind, childKind core.EntityKind) *schema.NestingDefinition {
	// Check for override first
	if p.autoRelConfig != nil && p.autoRelConfig.Overrides != nil {
		key := string(parentKind) + ".*"
		if override, ok := p.autoRelConfig.Overrides[key]; ok {
			// Find the nesting definition that matches this child kind
			defs := p.schema.GetNestingDefs(parentKind)
			for i := range defs {
				if defs[i].ChildKind == childKind {
					return &schema.NestingDefinition{
						NestKey:            defs[i].NestKey,
						ChildKind:          childKind,
						AutoRelationType:   override.RelationType,
						AutoRelationSource: override.Source,
					}
				}
			}
		}
	}

	// Use schema defaults
	defs := p.schema.GetNestingDefs(parentKind)
	for i := range defs {
		if defs[i].ChildKind == childKind && defs[i].AutoRelationType != "" {
			return &defs[i]
		}
	}
	return nil
}

// buildAutoRelation creates a Relation from a nesting definition.
func (p *Parser) buildAutoRelation(nd *schema.NestingDefinition, parent, child *core.Entity) *core.Relation {
	relType := nd.AutoRelationType

	// Check for per-type override
	if p.autoRelConfig != nil && p.autoRelConfig.Overrides != nil {
		key := string(parent.Kind) + "." + nd.NestKey
		if override, ok := p.autoRelConfig.Overrides[key]; ok {
			relType = override.RelationType
		}
	}

	relTypeDef, ok := p.schema.GetRelationTypeDef(relType)
	if !ok {
		return nil
	}

	var source, target string
	if nd.AutoRelationSource == "parent" {
		source = parent.ID
		target = child.ID
	} else {
		source = child.ID
		target = parent.ID
	}

	relID := generateAutoRelationID(relType, source, target)

	rel := &core.Relation{
		ID:        relID,
		Type:      relType,
		Direction: core.Direction(relTypeDef.Direction),
		Participants: core.Participants{
			Source: source,
			Target: target,
		},
		Properties: make(map[string]interface{}),
	}
	rel.SetLabel("auto_generated", "true")
	return rel
}

// hasDuplicateRelation checks if a relation with the same type and participants already exists.
func hasDuplicateRelation(newRel *core.Relation, existing []*rawRelation) bool {
	for _, raw := range existing {
		typeStr, _ := raw.data["type"].(string)
		if typeStr != string(newRel.Type) {
			continue
		}
		participantsRaw, ok := raw.data["participants"]
		if !ok {
			continue
		}
		switch p := participantsRaw.(type) {
		case map[string]interface{}:
			src, _ := p["source"].(string)
			tgt, _ := p["target"].(string)
			if src == newRel.Participants.Source && tgt == newRel.Participants.Target {
				return true
			}
		case []interface{}:
			if len(p) == len(newRel.Participants.List) {
				ids := make([]string, len(p))
				for i, v := range p {
					ids[i], _ = v.(string)
				}
				match := true
				for i, id := range ids {
					if id != newRel.Participants.List[i] {
						match = false
						break
					}
				}
				if match {
					return true
				}
			}
		}
	}
	return false
}

// generateAutoRelationID generates a deterministic ID for an auto-generated relation.
func generateAutoRelationID(relType core.RelationType, source, target string) string {
	return fmt.Sprintf("rel-auto-%s-%s-%s", relType, source, target)
}

// parseEntity parses an entity from its YAML representation.
// Returns the entity, any nested child entities, and an error.
func (p *Parser) parseEntity(obj map[string]interface{}, parentID string) (*core.Entity, []*core.Entity, error) {
	// Required fields at top level
	id, err := getString(obj, "id")
	if err != nil {
		if parentID == "" {
			return nil, nil, fmt.Errorf("entity missing required field 'id': %w", err)
		}
		// Nested entity without id: will be auto-generated
		id = ""
	}

	kindStr, err := getString(obj, "kind")
	if err != nil {
		return nil, nil, fmt.Errorf("entity missing required field 'kind': %w", err)
	}

	name, nameProvided := getStringOptional(obj, "name")
	kind := core.EntityKind(kindStr)

	// For nested entities: if id is empty, generate a scoped ID
	if id == "" && parentID != "" {
		id = p.generateScopedID(parentID, kind)
	}

	if !nameProvided {
		if parentID == "" {
			return nil, nil, fmt.Errorf("entity missing required field 'name'")
		}
		// Nested entities default name to id
		name = id
	}

	entity := core.NewEntity(id, kind, name)

	// Mark as internal if using scoped ID (starts with _)
	if len(id) > 0 && id[0] == '_' {
		entity.SetInternal(true)
	}

	// Set owner from parent
	if parentID != "" {
		entity.SetOwner(parentID)
	}

	// Parse attributes sub-key
	if attrs, ok := getMapOptional(obj, "attributes"); ok {
		if owner, ok := getStringOptional(attrs, "owner"); ok {
			entity.SetOwner(owner)
		}
		ca := parseCommonAttributes(attrs)
		if ca.description != "" {
			entity.Description = ca.description
		}
		if ca.status != "" {
			entity.SetStatus(ca.status)
		}
		for _, tag := range ca.tags {
			entity.AddTag(tag)
		}
		for k, v := range ca.labels {
			entity.SetLabel(k, v)
		}
		if ca.extensions != nil {
			entity.Extensions = ca.extensions
		}
	}

	// Parse spec sub-key for kind-specific properties and nested entities
	var nestedEntities []*core.Entity
	nestingDefs := p.schema.GetNestingDefs(kind)
	nestKeySet := make(map[string]bool)
	for _, nd := range nestingDefs {
		nestKeySet[nd.NestKey] = true
	}

	// First, check for nesting keys at the entity definition level (outside spec)
	for k, v := range obj {
		if nestKeySet[k] && k != "kind" && k != "id" && k != "name" && k != "attributes" && k != "spec" {
			children, err := p.parseNestedEntities(v, id, k, kind)
			if err != nil {
				return nil, nil, fmt.Errorf("nested key %q: %w", k, err)
			}
			nestedEntities = append(nestedEntities, children...)
		}
	}

	// Then, parse spec sub-key for kind-specific properties and nested entities
	if spec, ok := getMapOptional(obj, "spec"); ok {
		for k, v := range spec {
			if nestKeySet[k] {
				children, err := p.parseNestedEntities(v, id, k, kind)
				if err != nil {
					return nil, nil, fmt.Errorf("nested key %q in spec: %w", k, err)
				}
				nestedEntities = append(nestedEntities, children...)
			} else {
				entity.SetProperty(k, convertPropertyValue(v))
			}
		}
	}

	return entity, nestedEntities, nil
}

// generateScopedID generates an internal ID for a nested entity that doesn't specify one.
// This ID is for internal processing only and should not be used directly by users.
// Users should reference nested entities using path notation (e.g., parent/child).
func (p *Parser) generateScopedID(parentID string, childKind core.EntityKind) string {
	// Strip leading underscores from parentID to avoid doubling
	cleanParentID := parentID
	for len(cleanParentID) > 0 && cleanParentID[0] == '_' {
		cleanParentID = cleanParentID[1:]
	}
	// Use _ prefix to indicate this is an internal scoped ID
	return "_" + cleanParentID + "-" + string(childKind)
}

// parseNestedEntities parses a list of nested child entities under a parent.
func (p *Parser) parseNestedEntities(value interface{}, parentID string, nestKey string, parentKind core.EntityKind) ([]*core.Entity, error) {
	children, ok := value.([]interface{})
	if !ok {
		return nil, fmt.Errorf("expected list for nest key %q, got %T", nestKey, value)
	}

	nestingDef, ok := p.schema.FindNestingByNestKey(parentKind, nestKey)
	if !ok {
		return nil, fmt.Errorf("no nesting definition for parent kind %q and nest key %q", parentKind, nestKey)
	}

	var result []*core.Entity

	for i, child := range children {
		childObj, ok := child.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("child at index %d in %q: expected map, got %T", i, nestKey, child)
		}

		// Determine child kind: use explicit kind or inferred from nesting
		childKind := nestingDef.ChildKind
		if explicitKind, ok := getStringOptional(childObj, "kind"); ok {
			childKind = core.EntityKind(explicitKind)
		}

		// Build child object with kind injected
		childObjWithKind := make(map[string]interface{})
		for k, v := range childObj {
			childObjWithKind[k] = v
		}
		childObjWithKind["kind"] = string(childKind)

		// If no ID provided, generate scoped ID
		if _, hasID := childObj["id"]; !hasID {
			childID := p.generateScopedID(parentID, childKind)
			childObjWithKind["id"] = childID
		}

		entity, nestedEntities, err := p.parseEntity(childObjWithKind, parentID)
		if err != nil {
			return nil, fmt.Errorf("child at index %d in %q: %w", i, nestKey, err)
		}

		result = append(result, entity)
		result = append(result, nestedEntities...)
	}

	return result, nil
}

// parseRelation parses a relation from its YAML representation.
func (p *Parser) parseRelation(obj map[string]interface{}, entities map[string]*core.Entity) (*core.Relation, error) {
	// Required fields at top level
	id, err := getString(obj, "id")
	if err != nil {
		return nil, fmt.Errorf("relation missing required field 'id': %w", err)
	}

	typeStr, err := getString(obj, "type")
	if err != nil {
		return nil, fmt.Errorf("relation missing required field 'type': %w", err)
	}

	relType := core.RelationType(typeStr)

	// Parse participants
	participants, direction, err := p.parseParticipants(obj, relType, entities)
	if err != nil {
		return nil, fmt.Errorf("failed to parse participants: %w", err)
	}

	relation := &core.Relation{
		ID:           id,
		Type:         relType,
		Direction:    direction,
		Participants: *participants,
		Properties:   make(map[string]interface{}),
	}

	// Parse attributes sub-key
	if attrs, ok := getMapOptional(obj, "attributes"); ok {
		ca := parseCommonAttributes(attrs)
		if ca.description != "" {
			relation.Description = ca.description
		}
		if ca.status != "" {
			relation.SetStatus(ca.status)
		}
		for _, tag := range ca.tags {
			relation.AddTag(tag)
		}
		for k, v := range ca.labels {
			relation.SetLabel(k, v)
		}
		if ca.extensions != nil {
			relation.Extensions = ca.extensions
		}
	}

	// Parse spec sub-key for relation-type-specific properties
	if spec, ok := getMapOptional(obj, "spec"); ok {
		for k, v := range spec {
			relation.SetProperty(k, convertPropertyValue(v))
		}
	}

	return relation, nil
}

// parseParticipants parses the participants field from a relation.
func (p *Parser) parseParticipants(obj map[string]interface{}, relType core.RelationType, entities map[string]*core.Entity) (*core.Participants, core.Direction, error) {
	participantsRaw, ok := obj["participants"]
	if !ok {
		return nil, "", fmt.Errorf("missing required field 'participants'")
	}

	switch participants := participantsRaw.(type) {
	case []interface{}:
		// List format (symmetric relations)
		ids := make([]string, 0, len(participants))
		for _, p := range participants {
			idStr, ok := p.(string)
			if !ok {
				return nil, "", fmt.Errorf("participant must be a string, got %T", p)
			}
			ids = append(ids, idStr)
		}
		return &core.Participants{List: ids}, core.DirectionSymmetric, nil

	case map[string]interface{}:
		// Map format (directed relations)
		source, _ := participants["source"].(string)
		target, _ := participants["target"].(string)

		if source == "" || target == "" {
			return nil, "", fmt.Errorf("directed relation must have both 'source' and 'target' participants")
		}

		return &core.Participants{
			Source: source,
			Target: target,
		}, core.DirectionDirected, nil

	default:
		return nil, "", fmt.Errorf("invalid participants format: expected list or map, got %T", participantsRaw)
	}
}

// rawCommonAttributes holds the common entity/relation attributes parsed from YAML.
type rawCommonAttributes struct {
	description string
	status      core.Status
	tags        []string
	labels      map[string]string
	extensions  map[string]interface{}
}

// parseCommonAttributes extracts the attributes shared by entities and relations
// from an attributes map (description, status, tags, labels, extensions).
func parseCommonAttributes(attrs map[string]interface{}) rawCommonAttributes {
	var ca rawCommonAttributes
	if description, ok := getStringOptional(attrs, "description"); ok {
		ca.description = description
	}
	if statusStr, ok := getStringOptional(attrs, "status"); ok {
		ca.status = core.Status(statusStr)
	}
	if tags, ok := getSliceOptional(attrs, "tags"); ok {
		for _, tag := range tags {
			if tagStr, ok := tag.(string); ok {
				ca.tags = append(ca.tags, tagStr)
			}
		}
	}
	if labels, ok := getMapOptional(attrs, "labels"); ok {
		ca.labels = make(map[string]string, len(labels))
		for k, v := range labels {
			if vStr, ok := v.(string); ok {
				ca.labels[k] = vStr
			}
		}
	}
	if extensions, ok := getMapOptional(attrs, "extensions"); ok {
		ca.extensions = extensions
	}
	return ca
}

// getString extracts a string value from a map.
func getString(obj map[string]interface{}, key string) (string, error) {
	val, ok := obj[key]
	if !ok {
		return "", fmt.Errorf("field %q not found", key)
	}
	str, ok := val.(string)
	if !ok {
		return "", fmt.Errorf("field %q must be a string, got %T", key, val)
	}
	return str, nil
}

// getStringOptional extracts an optional string value from a map.
func getStringOptional(obj map[string]interface{}, key string) (string, bool) {
	val, ok := obj[key]
	if !ok {
		return "", false
	}
	str, ok := val.(string)
	if !ok {
		return "", false
	}
	return str, true
}

// getSliceOptional extracts an optional slice value from a map.
func getSliceOptional(obj map[string]interface{}, key string) ([]interface{}, bool) {
	val, ok := obj[key]
	if !ok {
		return nil, false
	}
	slice, ok := val.([]interface{})
	if !ok {
		return nil, false
	}
	return slice, true
}

// convertPropertyValue converts a raw YAML value, detecting @ prefix for references.
// If the value is a string starting with "@", it is converted to a core.ReferenceValue.
// For lists and maps, the conversion is applied recursively to all contained values.
func convertPropertyValue(v interface{}) interface{} {
	if str, ok := v.(string); ok && strings.HasPrefix(str, "@") {
		return core.NewReferenceValue(str)
	}
	if list, ok := v.([]interface{}); ok {
		converted := make([]interface{}, len(list))
		for i, item := range list {
			converted[i] = convertPropertyValue(item)
		}
		return converted
	}
	if m, ok := v.(map[string]interface{}); ok {
		converted := make(map[string]interface{}, len(m))
		for k, item := range m {
			converted[k] = convertPropertyValue(item)
		}
		return converted
	}
	return v
}

// getMapOptional extracts an optional map value from a map.
func getMapOptional(obj map[string]interface{}, key string) (map[string]interface{}, bool) {
	val, ok := obj[key]
	if !ok {
		return nil, false
	}
	m, ok := val.(map[string]interface{})
	if !ok {
		return nil, false
	}
	return m, true
}

// ResolveReferences checks that all references in the graph point to existing objects.
// It supports both simple ID references and path-based references.
// It also validates @-prefixed property references.
func ResolveReferences(g *core.Graph) []error {
	var errs []error

	// Check entity owners
	for _, e := range g.Entities() {
		if !e.IsRoot() {
			if _, ok := g.GetEntity(e.Owner); !ok {
				errs = append(errs, fmt.Errorf("entity %s references non-existent owner %s", e.ID, e.Owner))
			}
		}
	}

	// Check relation participants
	for _, r := range g.Relations() {
		for _, pid := range r.ParticipantIDs() {
			if _, err := resolvePathReference(g, pid); err != nil {
				errs = append(errs, fmt.Errorf("relation %s: %w", r.ID, err))
			}
		}
	}

	// Check entity property references (@ prefix)
	for _, e := range g.Entities() {
		for key, value := range e.Properties {
			for _, targetID := range core.ExtractReferenceTargets(value) {
				if _, err := resolvePathReference(g, targetID); err != nil {
					errs = append(errs, fmt.Errorf("entity %s property %q: %w", e.ID, key, err))
				}
			}
		}
	}

	// Check relation property references (@ prefix)
	for _, r := range g.Relations() {
		for key, value := range r.Properties {
			for _, targetID := range core.ExtractReferenceTargets(value) {
				if _, err := resolvePathReference(g, targetID); err != nil {
					errs = append(errs, fmt.Errorf("relation %s property %q: %w", r.ID, key, err))
				}
			}
		}
	}

	return errs
}

// resolvePathReference resolves a reference string to an entity ID.
// Resolution strategy:
// 1. Direct ID match
// 2. Path notation: last segment is the entity ID, parent segments are verified
// 3. Legacy interface notation: entityID/interfaceName (still supported for backward compat)
func resolvePathReference(g *core.Graph, ref string) (*core.Entity, error) {
	if e, ok := g.ResolvePathEntity(ref); ok {
		return e, nil
	}
	return nil, fmt.Errorf("reference %q could not be resolved", ref)
}
