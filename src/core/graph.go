package core

import (
	"errors"
	"fmt"
	"strings"
)

var (
	ErrEntityNotFound      = errors.New("entity not found")
	ErrRelationNotFound    = errors.New("relation not found")
	ErrDuplicateEntityID   = errors.New("duplicate entity id")
	ErrDuplicateRelationID = errors.New("duplicate relation id")
	ErrOwnerNotFound       = errors.New("owner entity not found")
	ErrOwnershipCycle      = errors.New("ownership cycle detected")
	ErrInvalidReference    = errors.New("invalid reference to non-existent object")
)

type Graph struct {
	entities  map[string]*Entity
	relations map[string]*Relation
}

func NewGraph() *Graph {
	return &Graph{
		entities:  make(map[string]*Entity),
		relations: make(map[string]*Relation),
	}
}

func (g *Graph) AddEntity(e *Entity) error {
	if err := e.Validate(); err != nil {
		return err
	}
	if _, exists := g.entities[e.ID]; exists {
		return fmt.Errorf("%w: %s", ErrDuplicateEntityID, e.ID)
	}
	g.entities[e.ID] = e
	return nil
}

// ForceAddEntity adds an entity to the graph without validation.
// Useful for loading partially-valid graphs for validation engine testing.
func (g *Graph) ForceAddEntity(e *Entity) {
	g.entities[e.ID] = e
}

// ForceAddRelation adds a relation to the graph without reference validation.
// Useful for loading partially-valid graphs for validation engine testing.
func (g *Graph) ForceAddRelation(r *Relation) {
	g.relations[r.ID] = r
}

func (g *Graph) GetEntity(id string) (*Entity, bool) {
	e, ok := g.entities[id]
	return e, ok
}

func (g *Graph) RemoveEntity(id string) bool {
	_, exists := g.entities[id]
	if exists {
		delete(g.entities, id)
	}
	return exists
}

func (g *Graph) UpdateEntity(e *Entity) error {
	if err := e.Validate(); err != nil {
		return err
	}
	g.entities[e.ID] = e
	return nil
}

func (g *Graph) Entities() []*Entity {
	result := make([]*Entity, 0, len(g.entities))
	for _, e := range g.entities {
		if !e.IsInternal() {
			result = append(result, e)
		}
	}
	return result
}

func (g *Graph) EntitiesByKind(kind EntityKind) []*Entity {
	var result []*Entity
	for _, e := range g.entities {
		if e.Kind == kind && !e.IsInternal() {
			result = append(result, e)
		}
	}
	return result
}

func (g *Graph) EntityCount() int {
	return len(g.entities)
}

func (g *Graph) AddRelation(r *Relation) error {
	if err := r.Validate(); err != nil {
		return err
	}
	if _, exists := g.relations[r.ID]; exists {
		return fmt.Errorf("%w: %s", ErrDuplicateRelationID, r.ID)
	}
	for _, pid := range r.ParticipantIDs() {
		if !g.isValidReference(pid) {
			return fmt.Errorf("%w: relation %s references entity %s", ErrInvalidReference, r.ID, pid)
		}
	}
	g.relations[r.ID] = r
	return nil
}

// isValidReference checks if a reference string points to a valid entity.
// Supports both simple entity IDs and interface references (entity/interface format).
func (g *Graph) isValidReference(ref string) bool {
	// Direct entity reference
	if _, ok := g.entities[ref]; ok {
		return true
	}
	// Interface reference (entity/interface format)
	if idx := strings.Index(ref, "/"); idx != -1 {
		entityID := ref[:idx]
		if _, ok := g.entities[entityID]; ok {
			return true
		}
	}
	return false
}

func (g *Graph) GetRelation(id string) (*Relation, bool) {
	r, ok := g.relations[id]
	return r, ok
}

func (g *Graph) RemoveRelation(id string) bool {
	_, exists := g.relations[id]
	if exists {
		delete(g.relations, id)
	}
	return exists
}

func (g *Graph) UpdateRelation(r *Relation) error {
	if err := r.Validate(); err != nil {
		return err
	}
	for _, pid := range r.ParticipantIDs() {
		if _, ok := g.entities[pid]; !ok {
			return fmt.Errorf("%w: relation %s references entity %s", ErrInvalidReference, r.ID, pid)
		}
	}
	g.relations[r.ID] = r
	return nil
}

func (g *Graph) Relations() []*Relation {
	result := make([]*Relation, 0, len(g.relations))
	for _, r := range g.relations {
		result = append(result, r)
	}
	return result
}

func (g *Graph) RelationsByType(relType RelationType) []*Relation {
	var result []*Relation
	for _, r := range g.relations {
		if r.Type == relType {
			result = append(result, r)
		}
	}
	return result
}

func (g *Graph) RelationsForEntity(entityID string) []*Relation {
	var result []*Relation
	for _, r := range g.relations {
		for _, pid := range r.ParticipantIDs() {
			if pid == entityID {
				result = append(result, r)
				break
			}
		}
	}
	return result
}

func (g *Graph) RelationCount() int {
	return len(g.relations)
}

func (g *Graph) ResolveReference(id string) (interface{}, bool) {
	if e, ok := g.ResolvePathEntity(id); ok {
		return e, true
	}
	if r, ok := g.relations[id]; ok {
		return r, true
	}
	return nil, false
}

// ResolvePathEntity resolves a reference string to an entity.
// Resolution strategy:
//  1. Direct ID match
//  2. Path notation: last segment is the entity ID, parent segments are verified
//  3. Legacy interface notation: entityID/interfaceName (still supported for backward compat)
func (g *Graph) ResolvePathEntity(ref string) (*Entity, bool) {
	// 1. Direct ID match
	if e, ok := g.entities[ref]; ok {
		return e, true
	}

	// 2. Path notation (contains "/")
	if idx := strings.Index(ref, "/"); idx != -1 {
		segments := strings.Split(ref, "/")
		lastSegment := segments[len(segments)-1]
		entityID := lastSegment

		// Try the last segment as a direct entity ID
		if e, ok := g.entities[entityID]; ok {
			// Verify parent relationship if there are more than 2 segments
			if len(segments) > 2 {
				if err := g.verifyPathOwnership(segments); err != nil {
					return nil, false
				}
			}
			return e, true
		}

		// Legacy interface reference: first segment is entity, rest is interface name
		entityID = segments[0]
		if e, ok := g.entities[entityID]; ok {
			return e, true
		}
	}

	return nil, false
}

// verifyPathOwnership verifies that the ownership chain in the path is valid.
func (g *Graph) verifyPathOwnership(segments []string) error {
	for i := 1; i < len(segments); i++ {
		childID := segments[i]
		parentID := segments[i-1]

		child, ok := g.entities[childID]
		if !ok {
			return fmt.Errorf("entity %q not found in path", childID)
		}
		if child.Owner != parentID {
			return fmt.Errorf("entity %q is not owned by %q", childID, parentID)
		}
	}
	return nil
}

func (g *Graph) BuildOwnershipPaths() error {
	for _, e := range g.entities {
		path, err := g.computePath(e)
		if err != nil {
			return err
		}
		e.SetPath(path)
	}
	return nil
}

func (g *Graph) computePath(e *Entity) (string, error) {
	if e.IsRoot() {
		return "/" + e.ID, nil
	}

	visited := make(map[string]bool)
	var pathParts []string
	current := e

	for current != nil {
		if visited[current.ID] {
			return "", fmt.Errorf("%w: %s", ErrOwnershipCycle, current.ID)
		}
		visited[current.ID] = true
		pathParts = append([]string{current.ID}, pathParts...)

		if current.IsRoot() {
			break
		}

		owner, ok := g.entities[current.Owner]
		if !ok {
			return "", fmt.Errorf("%w: entity %s references owner %s", ErrOwnerNotFound, current.ID, current.Owner)
		}
		current = owner
	}

	return "/" + strings.Join(pathParts, "/"), nil
}

func (g *Graph) Children(entityID string) []*Entity {
	var result []*Entity
	for _, e := range g.entities {
		if e.Owner == entityID {
			result = append(result, e)
		}
	}
	return result
}

func (g *Graph) Parent(entityID string) (*Entity, bool) {
	e, ok := g.entities[entityID]
	if !ok || e.IsRoot() {
		return nil, false
	}
	parent, ok := g.entities[e.Owner]
	return parent, ok
}

func (g *Graph) Ancestors(entityID string) []*Entity {
	var result []*Entity
	current := entityID
	visited := make(map[string]bool)

	for {
		visited[current] = true

		e, ok := g.entities[current]
		if !ok || e.IsRoot() {
			break
		}
		parent, ok := g.entities[e.Owner]
		if !ok {
			break
		}
		result = append(result, parent)
		current = parent.ID
	}
	return result
}

func (g *Graph) Descendants(entityID string) []*Entity {
	var result []*Entity
	queue := []string{entityID}
	visited := make(map[string]bool)
	visited[entityID] = true

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		for _, child := range g.Children(current) {
			if !visited[child.ID] {
				visited[child.ID] = true
				result = append(result, child)
				queue = append(queue, child.ID)
			}
		}
	}
	return result
}
