package query

import (
	"fmt"

	"IACForge/src/condition"
	"IACForge/src/core"
)

// applyProject applies a project clause to results.
func (e *Engine) applyProject(items []*ResultItem, project *ProjectClause) ([]*ResultItem, error) {
	switch project.Type {
	case ProjectionTypeObjects:
		return items, nil

	case ProjectionTypeProperties:
		return e.projectProperties(items, project.Properties)

	case ProjectionTypePaths:
		return e.projectPaths(items)

	case ProjectionTypeIDs:
		return e.projectIDs(items)

	case ProjectionTypeSummary:
		return e.projectSummary(items, project)

	default:
		return nil, fmt.Errorf("unsupported projection type: %s", project.Type)
	}
}

// projectProperties projects specific properties from results.
func (e *Engine) projectProperties(items []*ResultItem, properties []PropertyProjection) ([]*ResultItem, error) {
	var results []*ResultItem

	for _, item := range items {
		projected := &ResultItem{
			ID:   item.ID,
			Type: item.Type,
			Path: item.Path,
		}

		// Create a map with only the requested properties
		propMap := make(map[string]interface{})
		for _, prop := range properties {
			value := e.getItemProperty(item, prop.Name)
			if prop.Transform != "" {
				var err error
				value, err = e.applyTransform(value, prop.Transform)
				if err != nil {
					return nil, err
				}
			}
			propMap[prop.Name] = value
		}

		projected.Object = propMap
		results = append(results, projected)
	}

	return results, nil
}

// projectPaths projects only paths from results.
func (e *Engine) projectPaths(items []*ResultItem) ([]*ResultItem, error) {
	var results []*ResultItem

	for _, item := range items {
		projected := &ResultItem{
			ID:     item.ID,
			Type:   item.Type,
			Path:   item.Path,
			Object: item.Path,
		}
		results = append(results, projected)
	}

	return results, nil
}

// projectIDs projects only IDs from results.
func (e *Engine) projectIDs(items []*ResultItem) ([]*ResultItem, error) {
	var results []*ResultItem

	for _, item := range items {
		projected := &ResultItem{
			ID:     item.ID,
			Type:   item.Type,
			Path:   item.Path,
			Object: item.ID,
		}
		results = append(results, projected)
	}

	return results, nil
}

// projectSummary projects summary information.
func (e *Engine) projectSummary(items []*ResultItem, project *ProjectClause) ([]*ResultItem, error) {
	if project.Aggregation != nil {
		return e.projectAggregation(items, project.Aggregation)
	}

	// Simple summary: group by kind
	summary := make(map[string][]*ResultItem)
	for _, item := range items {
		kind := e.getItemKind(item)
		summary[kind] = append(summary[kind], item)
	}

	var results []*ResultItem
	for kind, group := range summary {
		result := &ResultItem{
			ID:   kind,
			Type: "summary",
			Object: map[string]interface{}{
				"kind":  kind,
				"count": len(group),
				"items": e.extractIDs(group),
			},
		}
		results = append(results, result)
	}

	return results, nil
}

// projectAggregation performs aggregation on results.
func (e *Engine) projectAggregation(items []*ResultItem, agg *Aggregation) ([]*ResultItem, error) {
	// Group by
	groups := make(map[string][]*ResultItem)
	if agg.GroupBy != "" {
		for _, item := range items {
			groupKey := fmt.Sprintf("%v", e.getItemProperty(item, agg.GroupBy))
			groups[groupKey] = append(groups[groupKey], item)
		}
	} else {
		groups["all"] = items
	}

	var results []*ResultItem
	for groupKey, group := range groups {
		summary := map[string]interface{}{
			"group": groupKey,
			"count": len(group),
		}

		if agg.Sum != "" {
			sum, err := e.calculateSum(group, agg.Sum)
			if err != nil {
				return nil, err
			}
			summary["sum_"+agg.Sum] = sum
		}

		if agg.Avg != "" {
			avg, err := e.calculateAvg(group, agg.Avg)
			if err != nil {
				return nil, err
			}
			summary["avg_"+agg.Avg] = avg
		}

		if agg.Min != "" {
			min, err := e.calculateMin(group, agg.Min)
			if err != nil {
				return nil, err
			}
			summary["min_"+agg.Min] = min
		}

		if agg.Max != "" {
			max, err := e.calculateMax(group, agg.Max)
			if err != nil {
				return nil, err
			}
			summary["max_"+agg.Max] = max
		}

		result := &ResultItem{
			ID:     groupKey,
			Type:   "aggregation",
			Object: summary,
		}
		results = append(results, result)
	}

	return results, nil
}

// getItemProperty gets a property from a result item.
func (e *Engine) getItemProperty(item *ResultItem, field string) interface{} {
	switch obj := item.Object.(type) {
	case *core.Entity:
		return e.getObjectField(obj, field)
	case *core.Relation:
		return e.getRelationField(obj, field)
	case map[string]interface{}:
		return obj[field]
	default:
		return nil
	}
}

// getItemKind gets the kind of a result item.
func (e *Engine) getItemKind(item *ResultItem) string {
	switch obj := item.Object.(type) {
	case *core.Entity:
		return string(obj.Kind)
	case *core.Relation:
		return string(obj.Type)
	default:
		return "unknown"
	}
}

// extractIDs extracts IDs from result items.
func (e *Engine) extractIDs(items []*ResultItem) []string {
	ids := make([]string, len(items))
	for i, item := range items {
		ids[i] = item.ID
	}
	return ids
}

// applyTransform applies a transform to a value.
func (e *Engine) applyTransform(value interface{}, transform string) (interface{}, error) {
	switch transform {
	case "to_json":
		// Simple JSON-like representation
		return fmt.Sprintf("%v", value), nil
	case "to_string":
		return fmt.Sprintf("%v", value), nil
	case "to_upper":
		str, ok := value.(string)
		if !ok {
			return value, nil
		}
		return str, nil
	case "to_lower":
		str, ok := value.(string)
		if !ok {
			return value, nil
		}
		return str, nil
	default:
		return value, nil
	}
}

// calculateSum calculates the sum of a numeric property.
func (e *Engine) calculateSum(items []*ResultItem, field string) (float64, error) {
	var sum float64
	for _, item := range items {
		value := e.getItemProperty(item, field)
		if value == nil {
			continue
		}
		f, ok := condition.ToFloat64(value)
		if !ok {
			return 0, fmt.Errorf("cannot sum non-numeric value: %v", value)
		}
		sum += f
	}
	return sum, nil
}

// calculateAvg calculates the average of a numeric property.
func (e *Engine) calculateAvg(items []*ResultItem, field string) (float64, error) {
	var sum float64
	var count int
	for _, item := range items {
		value := e.getItemProperty(item, field)
		if value == nil {
			continue
		}
		f, ok := condition.ToFloat64(value)
		if !ok {
			return 0, fmt.Errorf("cannot average non-numeric value: %v", value)
		}
		sum += f
		count++
	}
	if count == 0 {
		return 0, nil
	}
	return sum / float64(count), nil
}

// calculateMin calculates the minimum of a numeric property.
func (e *Engine) calculateMin(items []*ResultItem, field string) (float64, error) {
	var min float64
	first := true
	for _, item := range items {
		value := e.getItemProperty(item, field)
		if value == nil {
			continue
		}
		f, ok := condition.ToFloat64(value)
		if !ok {
			return 0, fmt.Errorf("cannot find min of non-numeric value: %v", value)
		}
		if first || f < min {
			min = f
			first = false
		}
	}
	return min, nil
}

// calculateMax calculates the maximum of a numeric property.
func (e *Engine) calculateMax(items []*ResultItem, field string) (float64, error) {
	var max float64
	first := true
	for _, item := range items {
		value := e.getItemProperty(item, field)
		if value == nil {
			continue
		}
		f, ok := condition.ToFloat64(value)
		if !ok {
			return 0, fmt.Errorf("cannot find max of non-numeric value: %v", value)
		}
		if first || f > max {
			max = f
			first = false
		}
	}
	return max, nil
}
