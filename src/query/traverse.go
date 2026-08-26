package query

import (
	"fmt"

	"IACForge/src/core"
)

// applyTraverse applies a traverse clause to results.
func (e *Engine) applyTraverse(items []*ResultItem, traverse *TraverseClause) ([]*ResultItem, error) {
	if traverse.From == "" {
		return nil, fmt.Errorf("traverse clause must have a 'from' field")
	}

	// Resolve the starting point
	startItems := e.resolveTraverseFrom(traverse.From, items)
	if len(startItems) == 0 {
		return []*ResultItem{}, nil
	}

	var results []*ResultItem
	visited := make(map[string]bool)

	for _, startItem := range startItems {
		traversed, err := e.traverseFromItem(startItem, traverse, visited)
		if err != nil {
			return nil, err
		}
		results = append(results, traversed...)
	}

	return results, nil
}

// resolveTraverseFrom resolves the 'from' field of a traverse clause.
func (e *Engine) resolveTraverseFrom(from string, currentItems []*ResultItem) []*ResultItem {
	// Check if it's a query reference
	if len(from) > 6 && from[:6] == "query." {
		queryID := from[6:]
		// For now, return current items if query reference matches
		// In a full implementation, this would reference named queries
		_ = queryID
		return currentItems
	}

	// Try to find the entity directly
	if entity, ok := e.graph.GetEntity(from); ok {
		return []*ResultItem{
			{
				ID:     entity.ID,
				Type:   "entity",
				Path:   entity.Path(),
				Object: entity,
			},
		}
	}

	// Try to find in current items
	for _, item := range currentItems {
		if item.ID == from {
			return []*ResultItem{item}
		}
	}

	return nil
}

// traverseFromItem performs traversal starting from a result item.
func (e *Engine) traverseFromItem(item *ResultItem, traverse *TraverseClause, visited map[string]bool) ([]*ResultItem, error) {
	entityID := item.ID
	if item.Type != "entity" {
		return nil, fmt.Errorf("traversal can only start from entities, got %s", item.Type)
	}

	maxDepth := traverse.MaxDepth
	if maxDepth == 0 && traverse.Depth > 0 {
		maxDepth = traverse.Depth
	}
	if maxDepth == 0 {
		maxDepth = 100 // Default max depth
	}

	depth := traverse.Depth
	if depth == 0 {
		depth = 1
	}

	return e.traverseEntity(entityID, traverse, visited, 0, depth, maxDepth)
}

// traverseEntity performs the actual traversal from an entity.
func (e *Engine) traverseEntity(entityID string, traverse *TraverseClause, visited map[string]bool, currentDepth, requestedDepth, maxDepth int) ([]*ResultItem, error) {
	if visited[entityID] {
		return nil, nil
	}
	visited[entityID] = true

	if currentDepth > maxDepth {
		return nil, nil
	}

	var results []*ResultItem

	switch traverse.Operation {
	case TraverseOpChildren:
		if currentDepth < requestedDepth {
			children := e.graph.Children(entityID)
			for _, child := range children {
				results = append(results, &ResultItem{
					ID:     child.ID,
					Type:   "entity",
					Path:   child.Path(),
					Object: child,
				})
				// Recursively get descendants if depth allows
				if currentDepth+1 < requestedDepth {
					childResults, err := e.traverseEntity(child.ID, traverse, visited, currentDepth+1, requestedDepth, maxDepth)
					if err != nil {
						return nil, err
					}
					results = append(results, childResults...)
				}
			}
		}

	case TraverseOpParent:
		if parent, ok := e.graph.Parent(entityID); ok {
			results = append(results, &ResultItem{
				ID:     parent.ID,
				Type:   "entity",
				Path:   parent.Path(),
				Object: parent,
			})
		}

	case TraverseOpAncestors:
		ancestors := e.graph.Ancestors(entityID)
		for _, ancestor := range ancestors {
			results = append(results, &ResultItem{
				ID:     ancestor.ID,
				Type:   "entity",
				Path:   ancestor.Path(),
				Object: ancestor,
			})
		}

	case TraverseOpDescendants:
		descendants := e.graph.Descendants(entityID)
		for _, desc := range descendants {
			results = append(results, &ResultItem{
				ID:     desc.ID,
				Type:   "entity",
				Path:   desc.Path(),
				Object: desc,
			})
		}

	case TraverseOpRelated:
		relations := e.graph.RelationsForEntity(entityID)
		for _, rel := range relations {
			// Get the other participant
			otherIDs := e.getOtherParticipants(rel, entityID)
			for _, otherID := range otherIDs {
				if entity, ok := e.graph.GetEntity(otherID); ok {
					results = append(results, &ResultItem{
						ID:     entity.ID,
						Type:   "entity",
						Path:   entity.Path(),
						Object: entity,
					})
				}
			}
		}

	case TraverseOpSources:
		relations := e.graph.RelationsForEntity(entityID)
		for _, rel := range relations {
			if rel.IsDirected() && rel.Target() == entityID {
				if source, ok := e.graph.GetEntity(rel.Source()); ok {
					results = append(results, &ResultItem{
						ID:     source.ID,
						Type:   "entity",
						Path:   source.Path(),
						Object: source,
					})
				}
			}
		}

	case TraverseOpTargets:
		relations := e.graph.RelationsForEntity(entityID)
		for _, rel := range relations {
			if rel.IsDirected() && rel.Source() == entityID {
				if target, ok := e.graph.GetEntity(rel.Target()); ok {
					results = append(results, &ResultItem{
						ID:     target.ID,
						Type:   "entity",
						Path:   target.Path(),
						Object: target,
					})
				}
			}
		}

	case TraverseOpOutgoing:
		relations := e.graph.RelationsForEntity(entityID)
		for _, rel := range relations {
			if traverse.RelationType != "" && rel.Type != traverse.RelationType {
				continue
			}
			if rel.IsDirected() && rel.Source() == entityID {
				if target, ok := e.graph.GetEntity(rel.Target()); ok {
					results = append(results, &ResultItem{
						ID:     target.ID,
						Type:   "entity",
						Path:   target.Path(),
						Object: target,
					})
				}
			}
		}

	case TraverseOpIncoming:
		relations := e.graph.RelationsForEntity(entityID)
		for _, rel := range relations {
			if traverse.RelationType != "" && rel.Type != traverse.RelationType {
				continue
			}
			if rel.IsDirected() && rel.Target() == entityID {
				if source, ok := e.graph.GetEntity(rel.Source()); ok {
					results = append(results, &ResultItem{
						ID:     source.ID,
						Type:   "entity",
						Path:   source.Path(),
						Object: source,
					})
				}
			}
		}

	case TraverseOpReverseOwnership:
		// Get parent
		if parent, ok := e.graph.Parent(entityID); ok {
			results = append(results, &ResultItem{
				ID:     parent.ID,
				Type:   "entity",
				Path:   parent.Path(),
				Object: parent,
			})
		}

	default:
		return nil, fmt.Errorf("unsupported traverse operation: %s", traverse.Operation)
	}

	return results, nil
}

// getOtherParticipants returns the participant IDs other than the given entity.
func (e *Engine) getOtherParticipants(rel *core.Relation, entityID string) []string {
	var others []string
	for _, pid := range rel.ParticipantIDs() {
		if pid != entityID {
			others = append(others, pid)
		}
	}
	return others
}
