package query

import (
	"fmt"
	"strings"

	"IACForge/src/condition"
	"IACForge/src/core"
)

// Result represents the result of a query execution.
type Result struct {
	QueryID   string                 `yaml:"query_id"`
	Results   []*ResultItem          `yaml:"results"`
	Count     int                    `yaml:"count"`
	Truncated bool                   `yaml:"truncated"`
	Metadata  map[string]interface{} `yaml:"metadata,omitempty"`
}

// ResultItem represents a single result item.
type ResultItem struct {
	ID     string      `yaml:"id"`
	Type   string      `yaml:"type"`
	Path   string      `yaml:"path"`
	Object interface{} `yaml:"object"`
}

// Engine executes queries against a Graph.
type Engine struct {
	graph *core.Graph
}

// NewEngine creates a new query engine.
func NewEngine(graph *core.Graph) *Engine {
	return &Engine{graph: graph}
}

// Execute executes a query and returns the result.
func (e *Engine) Execute(q *Query) (*Result, error) {
	if q.Select == nil {
		return nil, fmt.Errorf("query must have a select clause")
	}

	var results []*ResultItem

	// Execute select
	if q.Select.Entities != nil || q.Select.Relations != nil {
		selected, err := e.executeSelect(q.Select)
		if err != nil {
			return nil, fmt.Errorf("select execution failed: %w", err)
		}
		results = append(results, selected...)
	}

	// Apply where filter
	if q.Where != nil {
		filtered, err := e.applyWhere(results, q.Where)
		if err != nil {
			return nil, fmt.Errorf("where filter failed: %w", err)
		}
		results = filtered
	}

	// Apply traverse
	if q.Traverse != nil {
		traversed, err := e.applyTraverse(results, q.Traverse)
		if err != nil {
			return nil, fmt.Errorf("traverse failed: %w", err)
		}
		results = traversed
	}

	// Apply project
	if q.Project != nil {
		projected, err := e.applyProject(results, q.Project)
		if err != nil {
			return nil, fmt.Errorf("project failed: %w", err)
		}
		results = projected
	}

	// Apply limit and offset
	totalCount := len(results)
	if q.Offset > 0 && q.Offset < len(results) {
		results = results[q.Offset:]
	} else if q.Offset >= len(results) {
		results = []*ResultItem{}
	}

	truncated := false
	if q.Limit > 0 && q.Limit < len(results) {
		results = results[:q.Limit]
		truncated = true
	}

	return &Result{
		QueryID:   q.ID,
		Results:   results,
		Count:     totalCount,
		Truncated: truncated,
	}, nil
}

// executeSelect executes the select clause.
func (e *Engine) executeSelect(sel *SelectClause) ([]*ResultItem, error) {
	var results []*ResultItem

	// Select entities
	if sel.Entities != nil {
		for _, entitySel := range sel.Entities {
			items, err := e.selectEntities(entitySel)
			if err != nil {
				return nil, err
			}
			results = append(results, items...)
		}
	}

	// Select relations
	if sel.Relations != nil {
		for _, relSel := range sel.Relations {
			items, err := e.selectRelations(relSel)
			if err != nil {
				return nil, err
			}
			results = append(results, items...)
		}
	}

	return results, nil
}

// selectEntities selects entities matching the selection criteria.
func (e *Engine) selectEntities(sel *EntitySelection) ([]*ResultItem, error) {
	var results []*ResultItem

	// Get all entities of the specified kind
	entities := e.graph.EntitiesByKind(sel.Kind)

	// Apply selection filter if present
	for _, entity := range entities {
		if sel.Where != nil {
			matches, err := e.evaluateWhereOnObject(entity, sel.Where)
			if err != nil {
				return nil, err
			}
			if !matches {
				continue
			}
		}

		results = append(results, &ResultItem{
			ID:     entity.ID,
			Type:   "entity",
			Path:   entity.Path(),
			Object: entity,
		})
	}

	return results, nil
}

// selectRelations selects relations matching the selection criteria.
func (e *Engine) selectRelations(sel *RelationSelection) ([]*ResultItem, error) {
	var results []*ResultItem

	// Get all relations of the specified type
	relations := e.graph.RelationsByType(sel.Type)

	// Apply selection filter if present
	for _, rel := range relations {
		if sel.Where != nil {
			matches, err := e.evaluateWhereOnRelation(rel, sel.Where)
			if err != nil {
				return nil, err
			}
			if !matches {
				continue
			}
		}

		results = append(results, &ResultItem{
			ID:     rel.ID,
			Type:   "relation",
			Path:   "",
			Object: rel,
		})
	}

	return results, nil
}

// applyWhere applies a where clause to filter results.
func (e *Engine) applyWhere(items []*ResultItem, where *WhereClause) ([]*ResultItem, error) {
	var filtered []*ResultItem

	for _, item := range items {
		matches, err := e.evaluateWhereOnItem(item, where)
		if err != nil {
			return nil, err
		}
		if matches {
			filtered = append(filtered, item)
		}
	}

	return filtered, nil
}

// evaluateWhereOnItem evaluates a where clause on a result item.
func (e *Engine) evaluateWhereOnItem(item *ResultItem, where *WhereClause) (bool, error) {
	switch obj := item.Object.(type) {
	case *core.Entity:
		return e.evaluateWhereOnObject(obj, where)
	case *core.Relation:
		return e.evaluateWhereOnRelation(obj, where)
	default:
		return false, fmt.Errorf("unsupported object type: %T", item.Object)
	}
}

// evaluateWhereOnObject evaluates a where clause on an entity.
func (e *Engine) evaluateWhereOnObject(entity *core.Entity, where *WhereClause) (bool, error) {
	if where.Logical != nil {
		return e.evaluateLogicalOpOnObject(entity, where.Logical)
	}

	if len(where.Conditions) == 0 {
		return true, nil
	}

	for _, cond := range where.Conditions {
		matches, err := e.evaluateConditionOnObject(entity, cond)
		if err != nil {
			return false, err
		}
		if !matches {
			return false, nil
		}
	}

	return true, nil
}

// evaluateWhereOnRelation evaluates a where clause on a relation.
func (e *Engine) evaluateWhereOnRelation(rel *core.Relation, where *WhereClause) (bool, error) {
	if where.Logical != nil {
		return e.evaluateLogicalOpOnRelation(rel, where.Logical)
	}

	if len(where.Conditions) == 0 {
		return true, nil
	}

	for _, cond := range where.Conditions {
		matches, err := e.evaluateConditionOnRelation(rel, cond)
		if err != nil {
			return false, err
		}
		if !matches {
			return false, nil
		}
	}

	return true, nil
}

// evaluateConditionOnObject evaluates a condition on an entity.
func (e *Engine) evaluateConditionOnObject(entity *core.Entity, cond *Condition) (bool, error) {
	value := e.getObjectField(entity, cond.Field)
	return e.evaluateCondition(value, cond)
}

// evaluateConditionOnRelation evaluates a condition on a relation.
func (e *Engine) evaluateConditionOnRelation(rel *core.Relation, cond *Condition) (bool, error) {
	value := e.getRelationField(rel, cond.Field)
	return e.evaluateCondition(value, cond)
}

// getObjectField gets a field value from an entity.
func (e *Engine) getObjectField(entity *core.Entity, field string) interface{} {
	switch field {
	case "id":
		return entity.ID
	case "kind":
		return entity.Kind
	case "name":
		return entity.Name
	case "owner":
		return entity.Owner
	case "description":
		return entity.Description
	case "status":
		return entity.Status
	case "tags":
		return entity.Tags
	case "labels":
		return entity.Labels
	case "extensions":
		return entity.Extensions
	default:
		// Check properties with dot-notation support
		if strings.Contains(field, ".") {
			if val := entity.ResolvePropertyPath(field); val != nil {
				return val
			}
		}
		// Check properties
		if val, ok := entity.GetProperty(field); ok {
			return val
		}
		// Check labels
		if val, ok := entity.GetLabel(field); ok {
			return val
		}
		return nil
	}
}

// getRelationField gets a field value from a relation.
func (e *Engine) getRelationField(rel *core.Relation, field string) interface{} {
	switch field {
	case "id":
		return rel.ID
	case "type":
		return rel.Type
	case "direction":
		return rel.Direction
	case "source":
		return rel.Source()
	case "target":
		return rel.Target()
	case "description":
		return rel.Description
	case "status":
		return rel.Status
	case "tags":
		return rel.Tags
	case "labels":
		return rel.Labels
	case "extensions":
		return rel.Extensions
	default:
		// Check properties
		if val, ok := rel.GetProperty(field); ok {
			return val
		}
		// Check labels
		if val, ok := rel.GetLabel(field); ok {
			return val
		}
		return nil
	}
}

// evaluateCondition evaluates a condition against a value.
func (e *Engine) evaluateCondition(value interface{}, cond *Condition) (bool, error) {
	return condition.Evaluate(value, cond.Value, string(cond.Operator))
}

// evaluateLogicalOpOnObject evaluates a logical operation on an entity.
func (e *Engine) evaluateLogicalOpOnObject(entity *core.Entity, op *LogicalOp) (bool, error) {
	switch op.Type {
	case LogicalOpAnd:
		for _, rule := range op.Rules {
			matches, err := e.evaluateWhereOnObject(entity, rule)
			if err != nil {
				return false, err
			}
			if !matches {
				return false, nil
			}
		}
		return true, nil
	case LogicalOpOr:
		for _, rule := range op.Rules {
			matches, err := e.evaluateWhereOnObject(entity, rule)
			if err != nil {
				return false, err
			}
			if matches {
				return true, nil
			}
		}
		return false, nil
	case LogicalOpNot:
		if len(op.Rules) == 0 {
			return true, nil
		}
		matches, err := e.evaluateWhereOnObject(entity, op.Rules[0])
		if err != nil {
			return false, err
		}
		return !matches, nil
	default:
		return false, fmt.Errorf("unsupported logical operator: %s", op.Type)
	}
}

// evaluateLogicalOpOnRelation evaluates a logical operation on a relation.
func (e *Engine) evaluateLogicalOpOnRelation(rel *core.Relation, op *LogicalOp) (bool, error) {
	switch op.Type {
	case LogicalOpAnd:
		for _, rule := range op.Rules {
			matches, err := e.evaluateWhereOnRelation(rel, rule)
			if err != nil {
				return false, err
			}
			if !matches {
				return false, nil
			}
		}
		return true, nil
	case LogicalOpOr:
		for _, rule := range op.Rules {
			matches, err := e.evaluateWhereOnRelation(rel, rule)
			if err != nil {
				return false, err
			}
			if matches {
				return true, nil
			}
		}
		return false, nil
	case LogicalOpNot:
		if len(op.Rules) == 0 {
			return true, nil
		}
		matches, err := e.evaluateWhereOnRelation(rel, op.Rules[0])
		if err != nil {
			return false, err
		}
		return !matches, nil
	default:
		return false, fmt.Errorf("unsupported logical operator: %s", op.Type)
	}
}
