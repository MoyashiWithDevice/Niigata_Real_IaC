package query

import (
	"IACForge/src/core"
)

// Query represents a complete query against a Graph.
type Query struct {
	ID       string          `yaml:"id,omitempty"`
	Select   *SelectClause   `yaml:"select"`
	Where    *WhereClause    `yaml:"where,omitempty"`
	Traverse *TraverseClause `yaml:"traverse,omitempty"`
	Project  *ProjectClause  `yaml:"project,omitempty"`
	Limit    int             `yaml:"limit,omitempty"`
	Offset   int             `yaml:"offset,omitempty"`
}

// SelectClause defines what Objects to select.
type SelectClause struct {
	Entities  []*EntitySelection   `yaml:"entities,omitempty"`
	Relations []*RelationSelection `yaml:"relations,omitempty"`
}

// EntitySelection defines an entity selection with optional filtering.
type EntitySelection struct {
	Kind  core.EntityKind `yaml:"kind"`
	Where *WhereClause    `yaml:"where,omitempty"`
}

// RelationSelection defines a relation selection with optional filtering.
type RelationSelection struct {
	Type  core.RelationType `yaml:"type"`
	Where *WhereClause      `yaml:"where,omitempty"`
}

// WhereClause defines filtering conditions.
type WhereClause struct {
	Conditions []*Condition `yaml:"conditions,omitempty"`
	Logical    *LogicalOp   `yaml:"logical,omitempty"`
}

// Condition represents a single filter condition.
type Condition struct {
	Field    string      `yaml:"field"`
	Operator Operator    `yaml:"operator"`
	Value    interface{} `yaml:"value"`
}

// Operator represents a comparison operator.
type Operator string

const (
	OperatorEq         Operator = "eq"
	OperatorNe         Operator = "ne"
	OperatorIn         Operator = "in"
	OperatorNin        Operator = "nin"
	OperatorGt         Operator = "gt"
	OperatorGe         Operator = "ge"
	OperatorLt         Operator = "lt"
	OperatorLe         Operator = "le"
	OperatorContains   Operator = "contains"
	OperatorStartsWith Operator = "starts_with"
	OperatorEndsWith   Operator = "ends_with"
	OperatorMatches    Operator = "matches"
	OperatorDefined    Operator = "defined"
	OperatorUndefined  Operator = "undefined"
)

// LogicalOp represents a logical operation (and, or, not).
type LogicalOp struct {
	Type  LogicalOpType  `yaml:"type"`
	Rules []*WhereClause `yaml:"rules"`
}

// LogicalOpType represents the type of logical operation.
type LogicalOpType string

const (
	LogicalOpAnd LogicalOpType = "and"
	LogicalOpOr  LogicalOpType = "or"
	LogicalOpNot LogicalOpType = "not"
)

// TraverseClause defines how to navigate the Graph.
type TraverseClause struct {
	From         string            `yaml:"from"`
	Operation    TraverseOp        `yaml:"operation"`
	RelationType core.RelationType `yaml:"relation_type,omitempty"`
	Depth        int               `yaml:"depth,omitempty"`
	MaxDepth     int               `yaml:"max_depth,omitempty"`
}

// TraverseOp represents a traversal operation.
type TraverseOp string

const (
	TraverseOpChildren         TraverseOp = "children"
	TraverseOpParent           TraverseOp = "parent"
	TraverseOpAncestors        TraverseOp = "ancestors"
	TraverseOpDescendants      TraverseOp = "descendants"
	TraverseOpRelated          TraverseOp = "related"
	TraverseOpSources          TraverseOp = "sources"
	TraverseOpTargets          TraverseOp = "targets"
	TraverseOpOutgoing         TraverseOp = "outgoing"
	TraverseOpIncoming         TraverseOp = "incoming"
	TraverseOpReverseOwnership TraverseOp = "reverse_ownership"
)

// ProjectClause defines how results are presented.
type ProjectClause struct {
	Type        ProjectionType       `yaml:"type"`
	Properties  []PropertyProjection `yaml:"properties,omitempty"`
	Aggregation *Aggregation         `yaml:"aggregation,omitempty"`
	GroupBy     string               `yaml:"group_by,omitempty"`
}

// ProjectionType represents the type of projection.
type ProjectionType string

const (
	ProjectionTypeObjects    ProjectionType = "objects"
	ProjectionTypeProperties ProjectionType = "properties"
	ProjectionTypePaths      ProjectionType = "paths"
	ProjectionTypeIDs        ProjectionType = "ids"
	ProjectionTypeSummary    ProjectionType = "summary"
)

// PropertyProjection defines a property to project.
type PropertyProjection struct {
	Name      string `yaml:"name"`
	Transform string `yaml:"transform,omitempty"`
}

// Aggregation defines aggregation operations.
type Aggregation struct {
	Count   bool   `yaml:"count,omitempty"`
	Sum     string `yaml:"sum,omitempty"`
	Avg     string `yaml:"avg,omitempty"`
	Min     string `yaml:"min,omitempty"`
	Max     string `yaml:"max,omitempty"`
	GroupBy string `yaml:"group_by,omitempty"`
}

// NewQuery creates a new empty Query.
func NewQuery() *Query {
	return &Query{}
}

// NewSelectClause creates a new SelectClause.
func NewSelectClause() *SelectClause {
	return &SelectClause{}
}

// NewEntitySelection creates a new EntitySelection.
func NewEntitySelection(kind core.EntityKind) *EntitySelection {
	return &EntitySelection{Kind: kind}
}

// NewRelationSelection creates a new RelationSelection.
func NewRelationSelection(relType core.RelationType) *RelationSelection {
	return &RelationSelection{Type: relType}
}

// NewWhereClause creates a new WhereClause.
func NewWhereClause() *WhereClause {
	return &WhereClause{}
}

// NewCondition creates a new Condition.
func NewCondition(field string, op Operator, value interface{}) *Condition {
	return &Condition{
		Field:    field,
		Operator: op,
		Value:    value,
	}
}

// NewLogicalOp creates a new LogicalOp.
func NewLogicalOp(opType LogicalOpType, rules ...*WhereClause) *LogicalOp {
	return &LogicalOp{
		Type:  opType,
		Rules: rules,
	}
}

// NewTraverseClause creates a new TraverseClause.
func NewTraverseClause(from string, op TraverseOp) *TraverseClause {
	return &TraverseClause{
		From:      from,
		Operation: op,
	}
}

// NewProjectClause creates a new ProjectClause.
func NewProjectClause(projType ProjectionType) *ProjectClause {
	return &ProjectClause{Type: projType}
}

// AddEntity adds an entity selection to the select clause.
func (s *SelectClause) AddEntity(kind core.EntityKind) *EntitySelection {
	sel := NewEntitySelection(kind)
	s.Entities = append(s.Entities, sel)
	return sel
}

// AddRelation adds a relation selection to the select clause.
func (s *SelectClause) AddRelation(relType core.RelationType) *RelationSelection {
	sel := NewRelationSelection(relType)
	s.Relations = append(s.Relations, sel)
	return sel
}

// AddCondition adds a condition to the where clause.
func (w *WhereClause) AddCondition(field string, op Operator, value interface{}) *Condition {
	cond := NewCondition(field, op, value)
	w.Conditions = append(w.Conditions, cond)
	return cond
}

// AddConditionWrapper adds a condition to the where clause and returns the clause itself.
func (w *WhereClause) AddConditionWrapper(field string, op Operator, value interface{}) *WhereClause {
	cond := NewCondition(field, op, value)
	w.Conditions = append(w.Conditions, cond)
	return w
}

// SetLogical sets the logical operation for the where clause.
func (w *WhereClause) SetLogical(op *LogicalOp) {
	w.Logical = op
}

// SetDepth sets the traversal depth.
func (t *TraverseClause) SetDepth(depth int) {
	t.Depth = depth
}

// SetMaxDepth sets the maximum traversal depth.
func (t *TraverseClause) SetMaxDepth(maxDepth int) {
	t.MaxDepth = maxDepth
}

// SetProperty adds a property projection.
func (p *ProjectClause) SetProperty(name string) *PropertyProjection {
	prop := &PropertyProjection{Name: name}
	p.Properties = append(p.Properties, *prop)
	return prop
}

// SetAggregation sets the aggregation.
func (p *ProjectClause) SetAggregation(agg *Aggregation) {
	p.Aggregation = agg
}
