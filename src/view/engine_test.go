package view

import (
	"testing"

	"IACForge/src/core"
)

func TestNewView(t *testing.T) {
	v := NewView("test-view", "Test View")
	if v.ID != "test-view" {
		t.Errorf("expected ID 'test-view', got '%s'", v.ID)
	}
	if v.Name != "Test View" {
		t.Errorf("expected Name 'Test View', got '%s'", v.Name)
	}
}

func TestViewAddVisibility(t *testing.T) {
	v := NewView("test-view", "Test View")
	rule := NewVisibilityRule(VisibilityTargetEntities, VisibilityActionHide)
	v.AddVisibility(rule)
	if len(v.Visibility) != 1 {
		t.Errorf("expected 1 visibility rule, got %d", len(v.Visibility))
	}
}

func TestViewAddGrouping(t *testing.T) {
	v := NewView("test-view", "Test View")
	rule := &GroupingRule{
		TargetKind: "server",
		GroupKind:  "rack",
		GroupBy:    []string{},
	}
	v.AddGrouping(rule)
	if len(v.Grouping) != 1 {
		t.Errorf("expected 1 grouping rule, got %d", len(v.Grouping))
	}
}

func TestEngineApplyNoRules(t *testing.T) {
	g := core.NewGraph()
	e1 := core.NewEntity("srv-1", "server", "Server 1")
	if err := g.AddEntity(e1); err != nil {
		t.Fatalf("failed to add entity: %v", err)
	}
	e2 := core.NewEntity("srv-2", "server", "Server 2")
	if err := g.AddEntity(e2); err != nil {
		t.Fatalf("failed to add entity: %v", err)
	}

	engine := NewEngine(g)
	v := NewView("test-view", "Test View")

	result, err := engine.Apply(v)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.VisibleEntities) != 2 {
		t.Errorf("expected 2 visible entities, got %d", len(result.VisibleEntities))
	}
}

func TestEngineApplyVisibilityHide(t *testing.T) {
	g := core.NewGraph()
	e1 := core.NewEntity("srv-1", "server", "Server 1")
	if err := g.AddEntity(e1); err != nil {
		t.Fatalf("failed to add entity: %v", err)
	}
	e2 := core.NewEntity("vm-1", "vm", "VM 1")
	if err := g.AddEntity(e2); err != nil {
		t.Fatalf("failed to add entity: %v", err)
	}

	engine := NewEngine(g)
	v := NewView("test-view", "Test View")
	rule := NewVisibilityRule(VisibilityTargetEntities, VisibilityActionHide)
	rule.Kind = "vm"
	v.AddVisibility(rule)

	result, err := engine.Apply(v)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.VisibleEntities) != 1 {
		t.Errorf("expected 1 visible entity, got %d", len(result.VisibleEntities))
	}
	if result.VisibleEntities[0].ID != "srv-1" {
		t.Errorf("expected entity 'srv-1', got '%s'", result.VisibleEntities[0].ID)
	}
}

func TestEngineApplyVisibilityShow(t *testing.T) {
	g := core.NewGraph()
	e1 := core.NewEntity("srv-1", "server", "Server 1")
	if err := g.AddEntity(e1); err != nil {
		t.Fatalf("failed to add entity: %v", err)
	}
	e2 := core.NewEntity("vm-1", "vm", "VM 1")
	if err := g.AddEntity(e2); err != nil {
		t.Fatalf("failed to add entity: %v", err)
	}

	engine := NewEngine(g)
	v := NewView("test-view", "Test View")
	rule := NewVisibilityRule(VisibilityTargetEntities, VisibilityActionShow)
	rule.Kind = "server"
	v.AddVisibility(rule)

	result, err := engine.Apply(v)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.VisibleEntities) != 1 {
		t.Errorf("expected 1 visible entity, got %d", len(result.VisibleEntities))
	}
	if result.VisibleEntities[0].ID != "srv-1" {
		t.Errorf("expected entity 'srv-1', got '%s'", result.VisibleEntities[0].ID)
	}
}

func TestEngineApplyGrouping(t *testing.T) {
	g := core.NewGraph()
	e1 := core.NewEntity("srv-1", "server", "Server 1")
	if err := g.AddEntity(e1); err != nil {
		t.Fatalf("failed to add entity: %v", err)
	}
	e2 := core.NewEntity("srv-2", "server", "Server 2")
	if err := g.AddEntity(e2); err != nil {
		t.Fatalf("failed to add entity: %v", err)
	}

	engine := NewEngine(g)
	v := NewView("test-view", "Test View")
	rule := &GroupingRule{
		TargetKind: "server",
		GroupKind:  "server_group",
		GroupBy:    []string{},
	}
	v.AddGrouping(rule)

	result, err := engine.Apply(v)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Groups) != 1 {
		t.Errorf("expected 1 group, got %d", len(result.Groups))
	}
	if len(result.Groups[0].Members) != 2 {
		t.Errorf("expected 2 members, got %d", len(result.Groups[0].Members))
	}
}

func TestEngineApplyAnnotations(t *testing.T) {
	g := core.NewGraph()
	e1 := core.NewEntity("srv-1", "server", "Server 1")
	e1.SetProperty("cpu_cores", 8)
	if err := g.AddEntity(e1); err != nil {
		t.Fatalf("failed to add entity: %v", err)
	}

	engine := NewEngine(g)
	v := NewView("test-view", "Test View")

	rule := &AnnotationRule{
		TargetSelector: &EntitySelector{Kind: "server"},
		Annotations: []*Annotation{
			{Property: "cpu_count", SourceProperty: "cpu_cores"},
		},
	}
	v.Annotations = append(v.Annotations, rule)

	result, err := engine.Apply(v)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, ok := result.Annotations["srv-1"]; !ok {
		t.Error("expected annotation for entity 'srv-1'")
	}
	if v, ok := result.Annotations["srv-1"]["cpu_count"]; !ok || v != 8 {
		t.Errorf("expected cpu_count=8, got %v", v)
	}
}

func TestEngineApplyGroupingByProperty(t *testing.T) {
	g := core.NewGraph()
	e1 := core.NewEntity("srv-1", "server", "Server 1")
	e1.SetLabel("rack", "rack-A")
	if err := g.AddEntity(e1); err != nil {
		t.Fatalf("failed to add entity: %v", err)
	}
	e2 := core.NewEntity("srv-2", "server", "Server 2")
	e2.SetLabel("rack", "rack-B")
	if err := g.AddEntity(e2); err != nil {
		t.Fatalf("failed to add entity: %v", err)
	}
	e3 := core.NewEntity("srv-3", "server", "Server 3")
	e3.SetLabel("rack", "rack-A")
	if err := g.AddEntity(e3); err != nil {
		t.Fatalf("failed to add entity: %v", err)
	}

	engine := NewEngine(g)
	v := NewView("test-view", "Test View")
	rule := &GroupingRule{
		TargetKind: "server",
		GroupKind:  "rack_group",
		GroupBy:    []string{"rack"},
	}
	v.AddGrouping(rule)

	result, err := engine.Apply(v)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Groups) != 2 {
		t.Errorf("expected 2 groups, got %d", len(result.Groups))
	}
}

func TestWhereClause(t *testing.T) {
	w := &WhereClause{Conditions: []*Condition{{Field: "status", Operator: "eq", Value: "active"}}}
	if len(w.Conditions) != 1 {
		t.Errorf("expected 1 condition, got %d", len(w.Conditions))
	}
	if w.Conditions[0].Field != "status" {
		t.Errorf("expected field 'status', got '%s'", w.Conditions[0].Field)
	}
}

func TestEvaluateCondition(t *testing.T) {
	tests := []struct {
		value    interface{}
		cond     *Condition
		expected bool
	}{
		{"active", &Condition{Field: "status", Operator: "eq", Value: "active"}, true},
		{"active", &Condition{Field: "status", Operator: "ne", Value: "active"}, false},
		{"hello world", &Condition{Field: "name", Operator: "contains", Value: "world"}, true},
		{8, &Condition{Field: "cpu", Operator: "gt", Value: 4}, true},
		{4, &Condition{Field: "cpu", Operator: "lt", Value: 8}, true},
		{4, &Condition{Field: "cpu", Operator: "ge", Value: 4}, true},
		{4, &Condition{Field: "cpu", Operator: "le", Value: 4}, true},
	}

	for _, test := range tests {
		result := evaluateCondition(test.value, test.cond)
		if result != test.expected {
			t.Errorf("expected %v for %v %s %v, got %v", test.expected, test.value, test.cond.Operator, test.cond.Value, result)
		}
	}
}
