package query

import (
	"encoding/json"
	"fmt"
	"sort"

	"IACForge/src/core"
	"IACForge/src/view"
)

// ToViewResult converts a query Result into a ViewResult that existing
// renderers can consume.
//
// Nodes are the entity result items (deduplicated by ID) plus the participant
// entities of relation result items, resolved from the graph.
//
// When includeRelations is true, every graph relation whose participants are
// all present as nodes is included, producing the induced subgraph over the
// matched entities. Relations anchored on objects outside the node set are
// lifted (see view.LiftRelations): their endpoints are mapped up to visible
// ancestors or down to visible hosted entities, and the derived edges are
// attached to the result as LiftedRelations/LiftedGroups so diagrams such as
// application-only views still express connectivity defined on hidden nodes
// or clusters.
//
// Participant references (e.g. interface references like "srv-01/eth0") are
// resolved to entity IDs on the returned relations so that edge endpoints
// always match rendered node IDs. The source Graph is never modified.
//
// Output is deterministic: nodes and relations are ordered by ID.
func (e *Engine) ToViewResult(result *Result, includeRelations bool) *view.ViewResult {
	entities := make(map[string]*core.Entity)
	relations := make(map[string]*core.Relation)

	// Collect entity items and resolve relation participant entities.
	for _, item := range result.Results {
		switch obj := item.Object.(type) {
		case *core.Entity:
			entities[obj.ID] = obj
		case *core.Relation:
			relations[obj.ID] = e.resolvedRelation(obj)
			for _, pid := range obj.ParticipantIDs() {
				if entity, ok := e.resolveParticipant(pid); ok {
					entities[entity.ID] = entity
				}
			}
		}
	}

	// Include graph relations among visible entities (induced subgraph).
	induced := make([]*core.Relation, 0, len(relations))
	if includeRelations {
		for _, rel := range e.graph.Relations() {
			if _, present := relations[rel.ID]; present {
				continue
			}
			if e.allParticipantsVisible(rel, entities) {
				resolved := e.resolvedRelation(rel)
				relations[rel.ID] = resolved
				induced = append(induced, resolved)
			}
		}
	}

	entityIDs := sortedKeysEntities(entities)
	relationIDs := sortedKeysRelations(relations)

	visibleEntities := make([]*core.Entity, 0, len(entityIDs))
	for _, id := range entityIDs {
		visibleEntities = append(visibleEntities, entities[id])
	}

	visibleRelations := make([]*core.Relation, 0, len(relationIDs))
	for _, id := range relationIDs {
		visibleRelations = append(visibleRelations, relations[id])
	}

	viewID := result.QueryID
	if viewID == "" {
		viewID = "query-result"
	}

	vr := &view.ViewResult{
		ViewID:           viewID,
		Title:            fmt.Sprintf("Query %s", viewID),
		Description:      "Query result graph",
		VisibleEntities:  visibleEntities,
		VisibleRelations: visibleRelations,
		Annotations:      make(map[string]map[string]interface{}),
	}

	if includeRelations {
		vr.LiftedGroups, vr.LiftedRelations = view.LiftRelations(
			e.graph, visibleEntities, induced, e.graph.Relations())
	}

	return vr
}

// resolveParticipant resolves a relation participant reference to an entity.
// Supports direct entity IDs and interface references (entity/interface).
func (e *Engine) resolveParticipant(ref string) (*core.Entity, bool) {
	if entity, ok := e.graph.GetEntity(ref); ok {
		return entity, true
	}
	if entity, ok := e.graph.ResolvePathEntity(ref); ok {
		return entity, true
	}
	return nil, false
}

// resolveRef returns the entity ID for a participant reference, falling back
// to the raw reference when it cannot be resolved.
func (e *Engine) resolveRef(ref string) string {
	if entity, ok := e.resolveParticipant(ref); ok {
		return entity.ID
	}
	return ref
}

// resolvedRelation returns a copy of rel with participant references resolved
// to entity IDs. The original relation is left untouched.
func (e *Engine) resolvedRelation(rel *core.Relation) *core.Relation {
	resolved := *rel
	switch {
	case len(rel.Participants.List) > 0:
		ids := make([]string, 0, len(rel.Participants.List))
		for _, pid := range rel.Participants.List {
			ids = append(ids, e.resolveRef(pid))
		}
		resolved.Participants = core.Participants{List: ids}
	case rel.Participants.Source != "" || rel.Participants.Target != "":
		resolved.Participants = core.Participants{
			Source: e.resolveRef(rel.Participants.Source),
			Target: e.resolveRef(rel.Participants.Target),
		}
	}
	return &resolved
}

// allParticipantsVisible reports whether every participant of rel resolves to
// a visible entity.
func (e *Engine) allParticipantsVisible(rel *core.Relation, entities map[string]*core.Entity) bool {
	for _, pid := range rel.ParticipantIDs() {
		if _, ok := entities[e.resolveRef(pid)]; !ok {
			return false
		}
	}
	return true
}

func sortedKeysEntities(m map[string]*core.Entity) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func sortedKeysRelations(m map[string]*core.Relation) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// JSON serializes the Result as indented JSON with snake_case keys,
// consistent with the structure returned by the MCP query tools:
//
//	{query_id?, count, truncated, results: [{id, type, path, object}]}
func (r *Result) JSON() []byte {
	type resultItem struct {
		ID     string      `json:"id"`
		Type   string      `json:"type"`
		Path   string      `json:"path,omitempty"`
		Object interface{} `json:"object"`
	}

	items := make([]resultItem, 0, len(r.Results))
	for _, item := range r.Results {
		var obj interface{}
		switch o := item.Object.(type) {
		case *core.Entity:
			obj = EntityJSONMap(o)
		case *core.Relation:
			obj = resultRelationJSONMap(o)
		default:
			obj = o
		}
		items = append(items, resultItem{
			ID:     item.ID,
			Type:   item.Type,
			Path:   item.Path,
			Object: obj,
		})
	}

	resp := map[string]interface{}{
		"count":     r.Count,
		"truncated": r.Truncated,
		"results":   items,
	}
	if r.QueryID != "" {
		resp["query_id"] = r.QueryID
	}
	if len(r.Metadata) > 0 {
		resp["metadata"] = r.Metadata
	}

	data, _ := json.MarshalIndent(resp, "", "  ")
	return data
}

// EntityJSONMap converts an Entity to a map with snake_case keys.
func EntityJSONMap(e *core.Entity) map[string]interface{} {
	m := map[string]interface{}{
		"id":   e.ID,
		"kind": string(e.Kind),
		"name": e.Name,
	}
	if e.Owner != "" {
		m["owner"] = e.Owner
	}
	if e.Description != "" {
		m["description"] = e.Description
	}
	if e.Status != "" {
		m["status"] = string(e.Status)
	}
	if len(e.Tags) > 0 {
		m["tags"] = e.Tags
	}
	if len(e.Labels) > 0 {
		m["labels"] = e.Labels
	}
	if len(e.Extensions) > 0 {
		m["extensions"] = e.Extensions
	}
	if len(e.Properties) > 0 {
		m["spec"] = e.Properties
	}
	return m
}

// resultRelationJSONMap converts a Relation to a map with snake_case keys.
func resultRelationJSONMap(r *core.Relation) map[string]interface{} {
	m := map[string]interface{}{
		"id":        r.ID,
		"type":      string(r.Type),
		"direction": string(r.Direction),
	}
	switch r.Direction {
	case core.DirectionSymmetric:
		m["participants"] = r.Participants.List
	default:
		m["source"] = r.Source()
		m["target"] = r.Target()
	}
	if r.Description != "" {
		m["description"] = r.Description
	}
	if r.Status != "" {
		m["status"] = string(r.Status)
	}
	if len(r.Tags) > 0 {
		m["tags"] = r.Tags
	}
	if len(r.Labels) > 0 {
		m["labels"] = r.Labels
	}
	if len(r.Properties) > 0 {
		m["spec"] = r.Properties
	}
	return m
}
