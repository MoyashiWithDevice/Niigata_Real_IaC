package view

import (
	"testing"

	"IACForge/src/core"
)

// buildLiftGraph constructs:
//
//	site-main
//	├── cluster-a
//	│   └── srv-a
//	│       └── vm-web
//	│           ├── app-web
//	│           └── app-cache
//	└── cluster-b
//	    └── srv-b
//	        └── vm-db
//	            └── app-db
func buildLiftGraph(t *testing.T) *core.Graph {
	t.Helper()
	g := core.NewGraph()

	add := func(id string, kind core.EntityKind, name, owner string) {
		e := core.NewEntity(id, kind, name)
		if owner != "" {
			e.Owner = owner
		}
		if err := g.AddEntity(e); err != nil {
			t.Fatalf("failed to add entity %s: %v", id, err)
		}
	}

	add("site-main", "site", "Main Site", "")
	add("cluster-a", "cluster", "Cluster A", "site-main")
	add("cluster-b", "cluster", "Cluster B", "site-main")
	add("srv-a", "server", "Server A", "cluster-a")
	add("srv-b", "server", "Server B", "cluster-b")
	add("vm-web", "vm", "VM Web", "srv-a")
	add("vm-db", "vm", "VM DB", "srv-b")
	add("app-web", "application", "App Web", "vm-web")
	add("app-cache", "application", "App Cache", "vm-web")
	add("app-db", "application", "App DB", "vm-db")

	rels := []*core.Relation{
		core.NewDirectedRelation("rel-vm-dep", "depends_on", "vm-web", "vm-db"),
		core.NewSymmetricRelation("rel-cluster-link", "connects", []string{"cluster-a", "cluster-b"}),
		core.NewDirectedRelation("rel-app-direct", "depends_on", "app-web", "app-db"),
	}
	for _, r := range rels {
		if err := g.AddRelation(r); err != nil {
			t.Fatalf("failed to add relation %s: %v", r.ID, err)
		}
	}
	return g
}

func applicationsOnly(g *core.Graph) []*core.Entity {
	return g.EntitiesByKind("application")
}

func findLifted(rels []*LiftedRelation, refType core.RelationType, src, dst string) *LiftedRelation {
	for _, r := range rels {
		if r.Type == refType && r.SourceRef == src && r.TargetRef == dst {
			return r
		}
	}
	return nil
}

func TestLiftVMDependencyToApplications(t *testing.T) {
	g := buildLiftGraph(t)
	visible := applicationsOnly(g)

	groups, lifted := LiftRelations(g, visible, nil, g.Relations())

	// vm-web hosts two applications, so the dependency collapses onto a
	// structural group; vm-db hosts a single application which connects
	// directly without a singleton box.
	edge := findLifted(lifted, "depends_on", "vm-web", "app-db")
	if edge == nil {
		t.Fatalf("expected lifted edge vm-web -> app-db, got %v", lifted)
	}
	if edge.Direction != core.DirectionDirected {
		t.Errorf("expected directed edge, got %q", edge.Direction)
	}
	if len(edge.Via) != 1 || edge.Via[0] != "rel-vm-dep" {
		t.Errorf("expected Via [rel-vm-dep], got %v", edge.Via)
	}

	assertGroup(t, groups, "vm-web", []string{"app-cache", "app-web"})
	for _, grp := range groups {
		if grp.ID == "app-db" || grp.ID == "vm-db" {
			t.Errorf("did not expect singleton group %q", grp.ID)
		}
	}
}

func TestLiftClusterRelationToStructuralGroups(t *testing.T) {
	g := buildLiftGraph(t)
	visible := applicationsOnly(g)

	groups, lifted := LiftRelations(g, visible, nil, g.Relations())

	edge := findLifted(lifted, "connects", "cluster-a", "app-db")
	if edge == nil {
		t.Fatalf("expected lifted edge cluster-a --- app-db, got %v", lifted)
	}
	if edge.Direction != core.DirectionSymmetric {
		t.Errorf("expected symmetric edge, got %q", edge.Direction)
	}
	if edge.AggregatedCount != 0 {
		t.Errorf("structural collapse must not report aggregation, got %d", edge.AggregatedCount)
	}

	assertGroup(t, groups, "cluster-a", []string{"app-cache", "app-web"})
}

func TestLiftDirectRelationSuppressedByExistingEdge(t *testing.T) {
	g := buildLiftGraph(t)
	visible := applicationsOnly(g)

	// The direct app-web -> app-db relation is already part of the visible
	// subgraph; the vm-level dependency lifts onto the same pair and must be
	// dropped as redundant.
	existing := []*core.Relation{
		core.NewDirectedRelation("rel-app-direct", "depends_on", "app-web", "app-db"),
	}

	_, lifted := LiftRelations(g, visible, existing, g.Relations())

	if edge := findLifted(lifted, "depends_on", "app-web", "app-db"); edge != nil {
		t.Fatalf("expected duplicate lifted edge to be suppressed")
	}
}

func TestLiftSkipsFullyVisibleRelations(t *testing.T) {
	g := buildLiftGraph(t)
	visible := g.Entities()

	_, lifted := LiftRelations(g, visible, nil, g.Relations())

	if len(lifted) != 0 {
		t.Errorf("expected no lifted edges when everything is visible, got %v", lifted)
	}
}

func TestLiftSelfLoopSkippedForSameHostObjects(t *testing.T) {
	g := core.NewGraph()

	add := func(id string, kind core.EntityKind, name, owner string) {
		e := core.NewEntity(id, kind, name)
		if owner != "" {
			e.Owner = owner
		}
		if err := g.AddEntity(e); err != nil {
			t.Fatalf("failed to add entity %s: %v", id, err)
		}
	}
	add("app-web", "application", "App Web", "")
	add("port-80", "open_port", "Port 80", "app-web")
	add("port-443", "open_port", "Port 443", "app-web")

	if err := g.AddRelation(core.NewSymmetricRelation(
		"rel-ports", "connects", []string{"port-80", "port-443"})); err != nil {
		t.Fatalf("failed to add relation: %v", err)
	}

	visible := []*core.Entity{}
	if e, ok := g.GetEntity("app-web"); ok {
		visible = append(visible, e)
	}

	groups, lifted := LiftRelations(g, visible, nil, g.Relations())

	if len(lifted) != 0 {
		t.Errorf("expected self loop to be skipped, got %v", lifted)
	}
	if len(groups) != 0 {
		t.Errorf("expected no groups, got %v", groups)
	}
}

func TestLiftMapsHiddenPortsToOwningApplications(t *testing.T) {
	g := core.NewGraph()

	add := func(id string, kind core.EntityKind, name, owner string) {
		e := core.NewEntity(id, kind, name)
		if owner != "" {
			e.Owner = owner
		}
		if err := g.AddEntity(e); err != nil {
			t.Fatalf("failed to add entity %s: %v", id, err)
		}
	}
	add("app-web", "application", "App Web", "")
	add("app-db", "application", "App DB", "")
	add("port-web", "open_port", "Port 8080", "app-web")
	add("port-db", "open_port", "Port 5432", "app-db")

	if err := g.AddRelation(core.NewSymmetricRelation(
		"rel-port-link", "connects", []string{"port-web", "port-db"})); err != nil {
		t.Fatalf("failed to add relation: %v", err)
	}

	visible := g.EntitiesByKind("application")
	groups, lifted := LiftRelations(g, visible, nil, g.Relations())

	edge := findLifted(lifted, "connects", "app-web", "app-db")
	if edge == nil {
		t.Fatalf("expected lifted edge app-web --- app-db, got %v", lifted)
	}
	if len(groups) != 0 {
		t.Errorf("expected no groups for plain ancestor mapping, got %v", groups)
	}
}

func TestEngineApplyAddsLiftedContent(t *testing.T) {
	g := buildLiftGraph(t)

	v := NewView("apps-only", "Applications")
	rule := NewVisibilityRule(VisibilityTargetEntities, VisibilityActionShow)
	rule.Kind = "application"
	v.AddVisibility(rule)

	result, err := NewEngine(g).Apply(v)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.VisibleEntities) != 3 {
		t.Fatalf("expected 3 visible applications, got %d", len(result.VisibleEntities))
	}
	if len(result.LiftedRelations) == 0 {
		t.Fatal("expected lifted relations to be attached automatically")
	}
	found := false
	for _, lr := range result.LiftedRelations {
		if lr.Type == "depends_on" && lr.SourceRef == "vm-web" && lr.TargetRef == "app-db" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected vm dependency lifted into result, got %v", result.LiftedRelations)
	}
}

func TestEngineApplyNoLiftWhenAllVisible(t *testing.T) {
	g := buildLiftGraph(t)

	v := NewView("all", "Everything")
	v.AddVisibility(NewVisibilityRule(VisibilityTargetEntities, VisibilityActionShow))

	result, err := NewEngine(g).Apply(v)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.LiftedRelations) != 0 || len(result.LiftedGroups) != 0 {
		t.Errorf("expected no lift output when all entities are visible, got relations=%v groups=%v",
			result.LiftedRelations, result.LiftedGroups)
	}
}

func assertGroup(t *testing.T, groups []*Group, id string, wantMembers []string) {
	t.Helper()
	for _, grp := range groups {
		if grp.ID != id {
			continue
		}
		if len(grp.Members) != len(wantMembers) {
			t.Fatalf("group %s: expected members %v, got %v", id, wantMembers, grp.Members)
		}
		for i, m := range wantMembers {
			if grp.Members[i] != m {
				t.Fatalf("group %s: expected members %v, got %v", id, wantMembers, grp.Members)
			}
		}
		if grp.Kind != liftGroupKind {
			t.Errorf("group %s: expected kind %q, got %q", id, liftGroupKind, grp.Kind)
		}
		return
	}
	t.Fatalf("group %s not found in %v", id, groups)
}
