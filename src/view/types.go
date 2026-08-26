package view

import (
	"IACForge/src/core"
)

// View defines how a projected Graph is presented to a consumer.
type View struct {
	ID          string            `yaml:"id"`
	Name        string            `yaml:"name"`
	Description string            `yaml:"description,omitempty"`
	Visibility  []*VisibilityRule `yaml:"visibility,omitempty"`
	Grouping    []*GroupingRule   `yaml:"grouping,omitempty"`
	Annotations []*AnnotationRule `yaml:"annotations,omitempty"`
}

// VisibilityRule determines which objects are shown or hidden.
type VisibilityRule struct {
	Target   VisibilityTarget `yaml:"target"`
	Kind     string           `yaml:"kind,omitempty"`
	Relation string           `yaml:"relation,omitempty"`
	Action   VisibilityAction `yaml:"action"`
	Where    *WhereClause     `yaml:"where,omitempty"`
}

// VisibilityTarget specifies whether the rule applies to entities or relations.
type VisibilityTarget string

const (
	VisibilityTargetEntities  VisibilityTarget = "entities"
	VisibilityTargetRelations VisibilityTarget = "relations"
)

// VisibilityAction specifies whether to show or hide matching objects.
type VisibilityAction string

const (
	VisibilityActionShow VisibilityAction = "show"
	VisibilityActionHide VisibilityAction = "hide"
)

// GroupingRule organizes objects into logical groups.
type GroupingRule struct {
	TargetKind string       `yaml:"target_kind"`
	GroupKind  string       `yaml:"group_kind"`
	GroupBy    []string     `yaml:"group_by"`
	Where      *WhereClause `yaml:"where,omitempty"`
}

// AnnotationRule attaches computed metadata to objects.
type AnnotationRule struct {
	TargetSelector *EntitySelector `yaml:"target_selector"`
	Annotations    []*Annotation   `yaml:"annotations"`
}

// EntitySelector defines entity selection criteria.
type EntitySelector struct {
	Kind  string       `yaml:"kind,omitempty"`
	Where *WhereClause `yaml:"where,omitempty"`
}

// WhereClause defines filtering conditions.
type WhereClause struct {
	Conditions []*Condition `yaml:"conditions,omitempty"`
}

// Condition represents a single filter condition.
type Condition struct {
	Field    string      `yaml:"field"`
	Operator string      `yaml:"operator"`
	Value    interface{} `yaml:"value"`
}

// Annotation defines an annotation to attach.
type Annotation struct {
	Property       string      `yaml:"property"`
	Value          interface{} `yaml:"value,omitempty"`
	Expression     string      `yaml:"expression,omitempty"`
	SourceProperty string      `yaml:"source_property,omitempty"`
}

// ViewResult represents the result of applying a View to a Graph.
//
// In addition to the explicitly visible subgraph, the result carries lifted
// content produced by LiftRelations: edges derived between visible entities
// from relations anchored on hidden objects (e.g. applications connected via
// relations defined on their host nodes or clusters).
type ViewResult struct {
	ViewID           string
	Title            string
	Description      string
	VisibleEntities  []*core.Entity
	VisibleRelations []*core.Relation
	Groups           []*Group
	LiftedGroups     []*Group
	LiftedRelations  []*LiftedRelation
	Annotations      map[string]map[string]interface{}
}

// Group represents a collection of objects.
type Group struct {
	ID         string
	Kind       string
	Name       string
	Members    []string
	Properties map[string]interface{}
}

// LiftedRelation is a derived edge between visible objects produced by the
// relation lift step. Each endpoint references either a visible entity ID or
// the ID of a LiftedGroups entry (structural group), so diagrams can express
// relations between hidden ancestors (nodes, clusters, sites) without showing
// those ancestors.
type LiftedRelation struct {
	ID        string
	Type      core.RelationType
	Direction core.Direction
	SourceRef string
	TargetRef string
	// AggregatedCount reports how many anchor candidates were collapsed onto
	// a single representative endpoint when no structural group was available.
	// Values <= 1 mean no aggregation happened on either side.
	AggregatedCount int
	// Via lists the IDs of the source relations this edge was derived from.
	Via []string
}

// NewView creates a new View.
func NewView(id, name string) *View {
	return &View{
		ID:   id,
		Name: name,
	}
}

// NewVisibilityRule creates a new VisibilityRule.
func NewVisibilityRule(target VisibilityTarget, action VisibilityAction) *VisibilityRule {
	return &VisibilityRule{
		Target: target,
		Action: action,
	}
}

// AddVisibility adds a visibility rule to the view.
func (v *View) AddVisibility(rule *VisibilityRule) {
	v.Visibility = append(v.Visibility, rule)
}

// AddGrouping adds a grouping rule to the view.
func (v *View) AddGrouping(rule *GroupingRule) {
	v.Grouping = append(v.Grouping, rule)
}
