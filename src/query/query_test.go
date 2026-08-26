package query

import (
	"testing"

	"IACForge/src/core"
)

// createTestGraph creates a test graph with entities and relations.
func createTestGraph() *core.Graph {
	g := core.NewGraph()

	// Create entities
	region := core.NewEntity("region-ap-northeast-1", "region", "Tokyo Region")
	region.SetStatus(core.StatusActive)
	region.SetLabel("region", "ap-northeast-1")
	_ = g.AddEntity(region)

	rack := core.NewEntity("rack-a01", "rack", "Rack A01")
	rack.SetOwner("region-ap-northeast-1")
	rack.SetStatus(core.StatusActive)
	_ = g.AddEntity(rack)

	server1 := core.NewEntity("srv-proxmox-01", "server", "Proxmox Node 01")
	server1.SetOwner("rack-a01")
	server1.SetStatus(core.StatusActive)
	server1.SetProperty("cpu_cores", 16)
	server1.SetProperty("memory", []interface{}{
		map[string]interface{}{"size_gb": 32, "speed": 3200, "type": "ddr4"},
		map[string]interface{}{"size_gb": 32, "speed": 3200, "type": "ddr4"},
	})
	server1.AddTag("production")
	_ = g.AddEntity(server1)

	server2 := core.NewEntity("srv-proxmox-02", "server", "Proxmox Node 02")
	server2.SetOwner("rack-a01")
	server2.SetStatus(core.StatusActive)
	server2.SetProperty("cpu_cores", 8)
	server2.SetProperty("memory", []interface{}{
		map[string]interface{}{"size_gb": 16, "speed": 3200, "type": "ddr4"},
		map[string]interface{}{"size_gb": 16, "speed": 3200, "type": "ddr4"},
	})
	server2.AddTag("development")
	_ = g.AddEntity(server2)

	vm1 := core.NewEntity("vm-web-01", "vm", "Web Server 01")
	vm1.SetOwner("srv-proxmox-01")
	vm1.SetStatus(core.StatusActive)
	vm1.SetProperty("cpu_cores", 4)
	vm1.SetProperty("memory", []interface{}{
		map[string]interface{}{"size_gb": 8, "speed": 3200, "type": "ddr4"},
	})
	vm1.AddTag("production")
	_ = g.AddEntity(vm1)

	vm2 := core.NewEntity("vm-api-01", "vm", "API Server 01")
	vm2.SetOwner("srv-proxmox-01")
	vm2.SetStatus(core.StatusActive)
	vm2.SetProperty("cpu_cores", 2)
	vm2.SetProperty("memory", []interface{}{
		map[string]interface{}{"size_gb": 4, "speed": 3200, "type": "ddr4"},
	})
	vm2.AddTag("production")
	_ = g.AddEntity(vm2)

	vm3 := core.NewEntity("vm-dev-01", "vm", "Dev Server 01")
	vm3.SetOwner("srv-proxmox-02")
	vm3.SetStatus(core.StatusMaintenance)
	vm3.SetProperty("cpu_cores", 2)
	vm3.SetProperty("memory", []interface{}{
		map[string]interface{}{"size_gb": 4, "speed": 3200, "type": "ddr4"},
	})
	vm3.AddTag("development")
	_ = g.AddEntity(vm3)

	app1 := core.NewEntity("app-web", "application", "Web Application")
	app1.SetOwner("vm-web-01")
	app1.SetStatus(core.StatusActive)
	_ = g.AddEntity(app1)

	app2 := core.NewEntity("app-api", "application", "API Application")
	app2.SetOwner("vm-api-01")
	app2.SetStatus(core.StatusActive)
	_ = g.AddEntity(app2)

	// Build ownership paths
	_ = g.BuildOwnershipPaths()

	// Create relations
	rel1 := core.NewDirectedRelation("rel-hosts-1", "hosts", "srv-proxmox-01", "vm-web-01")
	_ = g.AddRelation(rel1)

	rel2 := core.NewDirectedRelation("rel-hosts-2", "hosts", "srv-proxmox-01", "vm-api-01")
	_ = g.AddRelation(rel2)

	rel3 := core.NewDirectedRelation("rel-hosts-3", "hosts", "srv-proxmox-02", "vm-dev-01")
	_ = g.AddRelation(rel3)

	rel4 := core.NewDirectedRelation("rel-depends-1", "depends_on", "app-web", "app-api")
	_ = g.AddRelation(rel4)

	rel5 := core.NewDirectedRelation("rel-connects-1", "connects", "srv-proxmox-01", "srv-proxmox-02")
	rel5.SetProperty("connection_type", "network")
	_ = g.AddRelation(rel5)

	return g
}

func TestNewEngine(t *testing.T) {
	g := createTestGraph()
	engine := NewEngine(g)
	if engine == nil {
		t.Fatal("NewEngine returned nil")
	}
}

func TestSelectAllEntities(t *testing.T) {
	g := createTestGraph()
	engine := NewEngine(g)

	q := NewQuery()
	q.Select = NewSelectClause()
	q.Select.AddEntity("server")

	result, err := engine.Execute(q)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if result.Count != 2 {
		t.Errorf("Expected 2 results, got %d", result.Count)
	}
}

func TestSelectEntitiesByKind(t *testing.T) {
	g := createTestGraph()
	engine := NewEngine(g)

	q := NewQuery()
	q.Select = NewSelectClause()
	q.Select.AddEntity("vm")

	result, err := engine.Execute(q)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if result.Count != 3 {
		t.Errorf("Expected 3 VMs, got %d", result.Count)
	}
}

func TestSelectRelationsByType(t *testing.T) {
	g := createTestGraph()
	engine := NewEngine(g)

	q := NewQuery()
	q.Select = NewSelectClause()
	q.Select.AddRelation("hosts")

	result, err := engine.Execute(q)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if result.Count != 3 {
		t.Errorf("Expected 3 hosts relations, got %d", result.Count)
	}
}

func TestWhereEquals(t *testing.T) {
	g := createTestGraph()
	engine := NewEngine(g)

	q := NewQuery()
	q.Select = NewSelectClause()
	q.Select.AddEntity("server")
	q.Where = NewWhereClause()
	q.Where.AddCondition("status", OperatorEq, "active")

	result, err := engine.Execute(q)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if result.Count != 2 {
		t.Errorf("Expected 2 active servers, got %d", result.Count)
	}
}

func TestWhereNotEquals(t *testing.T) {
	g := createTestGraph()
	engine := NewEngine(g)

	q := NewQuery()
	q.Select = NewSelectClause()
	q.Select.AddEntity("vm")
	q.Where = NewWhereClause()
	q.Where.AddCondition("status", OperatorNe, "maintenance")

	result, err := engine.Execute(q)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if result.Count != 2 {
		t.Errorf("Expected 2 non-maintenance VMs, got %d", result.Count)
	}
}

func TestWhereIn(t *testing.T) {
	g := createTestGraph()
	engine := NewEngine(g)

	q := NewQuery()
	q.Select = NewSelectClause()
	q.Select.AddEntity("server")
	q.Where = NewWhereClause()
	q.Where.AddCondition("status", OperatorIn, []interface{}{"active", "maintenance"})

	result, err := engine.Execute(q)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if result.Count != 2 {
		t.Errorf("Expected 2 servers, got %d", result.Count)
	}
}

func TestWhereGreaterThan(t *testing.T) {
	g := createTestGraph()
	engine := NewEngine(g)

	q := NewQuery()
	q.Select = NewSelectClause()
	q.Select.AddEntity("server")
	q.Where = NewWhereClause()
	q.Where.AddCondition("cpu_cores", OperatorGt, 8)

	result, err := engine.Execute(q)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if result.Count != 1 {
		t.Errorf("Expected 1 server with >8 cores, got %d", result.Count)
	}
}

func TestWhereContains(t *testing.T) {
	g := createTestGraph()
	engine := NewEngine(g)

	q := NewQuery()
	q.Select = NewSelectClause()
	q.Select.AddEntity("vm")
	q.Where = NewWhereClause()
	q.Where.AddCondition("name", OperatorContains, "Web")

	result, err := engine.Execute(q)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if result.Count != 1 {
		t.Errorf("Expected 1 VM with 'Web' in name, got %d", result.Count)
	}
}

func TestWhereStartsWith(t *testing.T) {
	g := createTestGraph()
	engine := NewEngine(g)

	q := NewQuery()
	q.Select = NewSelectClause()
	q.Select.AddEntity("vm")
	q.Where = NewWhereClause()
	q.Where.AddCondition("id", OperatorStartsWith, "vm-web")

	result, err := engine.Execute(q)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if result.Count != 1 {
		t.Errorf("Expected 1 VM with id starting with 'vm-web', got %d", result.Count)
	}
}

func TestWhereEndsWith(t *testing.T) {
	g := createTestGraph()
	engine := NewEngine(g)

	q := NewQuery()
	q.Select = NewSelectClause()
	q.Select.AddEntity("vm")
	q.Where = NewWhereClause()
	q.Where.AddCondition("id", OperatorEndsWith, "01")

	result, err := engine.Execute(q)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if result.Count != 3 {
		t.Errorf("Expected 3 VMs with id ending with '01', got %d", result.Count)
	}
}

func TestWhereMatches(t *testing.T) {
	g := createTestGraph()
	engine := NewEngine(g)

	q := NewQuery()
	q.Select = NewSelectClause()
	q.Select.AddEntity("vm")
	q.Where = NewWhereClause()
	q.Where.AddCondition("id", OperatorMatches, "vm-.*-01")

	result, err := engine.Execute(q)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if result.Count != 3 {
		t.Errorf("Expected 3 VMs matching pattern, got %d", result.Count)
	}
}

func TestWhereDefined(t *testing.T) {
	g := createTestGraph()
	engine := NewEngine(g)

	q := NewQuery()
	q.Select = NewSelectClause()
	q.Select.AddEntity("server")
	q.Where = NewWhereClause()
	q.Where.AddCondition("description", OperatorDefined, true)

	result, err := engine.Execute(q)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if result.Count != 0 {
		t.Errorf("Expected 0 servers with description defined, got %d", result.Count)
	}
}

func TestWhereUndefined(t *testing.T) {
	g := createTestGraph()
	engine := NewEngine(g)

	q := NewQuery()
	q.Select = NewSelectClause()
	q.Select.AddEntity("server")
	q.Where = NewWhereClause()
	q.Where.AddCondition("description", OperatorUndefined, true)

	result, err := engine.Execute(q)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if result.Count != 2 {
		t.Errorf("Expected 2 servers with description undefined, got %d", result.Count)
	}
}

func TestLogicalAnd(t *testing.T) {
	g := createTestGraph()
	engine := NewEngine(g)

	q := NewQuery()
	q.Select = NewSelectClause()
	q.Select.AddEntity("server")
	q.Where = NewWhereClause()
	q.Where.SetLogical(NewLogicalOp(LogicalOpAnd,
		NewWhereClause().AddConditionWrapper("status", OperatorEq, "active"),
		NewWhereClause().AddConditionWrapper("tags", OperatorContains, "production"),
	))

	result, err := engine.Execute(q)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if result.Count != 1 {
		t.Errorf("Expected 1 active production server, got %d", result.Count)
	}
}

func TestLogicalOr(t *testing.T) {
	g := createTestGraph()
	engine := NewEngine(g)

	q := NewQuery()
	q.Select = NewSelectClause()
	q.Select.AddEntity("server")
	q.Where = NewWhereClause()
	q.Where.SetLogical(NewLogicalOp(LogicalOpOr,
		NewWhereClause().AddConditionWrapper("status", OperatorEq, "active"),
		NewWhereClause().AddConditionWrapper("status", OperatorEq, "maintenance"),
	))

	result, err := engine.Execute(q)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if result.Count != 2 {
		t.Errorf("Expected 2 servers (active or maintenance), got %d", result.Count)
	}
}

func TestLogicalNot(t *testing.T) {
	g := createTestGraph()
	engine := NewEngine(g)

	q := NewQuery()
	q.Select = NewSelectClause()
	q.Select.AddEntity("server")
	q.Where = NewWhereClause()
	q.Where.SetLogical(NewLogicalOp(LogicalOpNot,
		NewWhereClause().AddConditionWrapper("status", OperatorEq, "active"),
	))

	result, err := engine.Execute(q)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if result.Count != 0 {
		t.Errorf("Expected 0 non-active servers, got %d", result.Count)
	}
}

func TestTraverseChildren(t *testing.T) {
	g := createTestGraph()
	engine := NewEngine(g)

	q := NewQuery()
	q.Select = NewSelectClause()
	q.Select.AddEntity("rack")
	q.Traverse = NewTraverseClause("rack-a01", TraverseOpChildren)

	result, err := engine.Execute(q)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if result.Count != 2 {
		t.Errorf("Expected 2 children (servers), got %d", result.Count)
	}
}

func TestTraverseDescendants(t *testing.T) {
	g := createTestGraph()
	engine := NewEngine(g)

	q := NewQuery()
	q.Select = NewSelectClause()
	q.Select.AddEntity("server")
	q.Traverse = NewTraverseClause("srv-proxmox-01", TraverseOpDescendants)

	result, err := engine.Execute(q)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	// Should find vm-web-01, vm-api-01, app-web, app-api
	if result.Count != 4 {
		t.Errorf("Expected 4 descendants, got %d", result.Count)
	}
}

func TestTraverseParent(t *testing.T) {
	g := createTestGraph()
	engine := NewEngine(g)

	q := NewQuery()
	q.Select = NewSelectClause()
	q.Select.AddEntity("vm")
	q.Traverse = NewTraverseClause("vm-web-01", TraverseOpParent)

	result, err := engine.Execute(q)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if result.Count != 1 {
		t.Errorf("Expected 1 parent (server), got %d", result.Count)
	}
}

func TestTraverseAncestors(t *testing.T) {
	g := createTestGraph()
	engine := NewEngine(g)

	q := NewQuery()
	q.Select = NewSelectClause()
	q.Select.AddEntity("application")
	q.Traverse = NewTraverseClause("app-web", TraverseOpAncestors)

	result, err := engine.Execute(q)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	// Should find vm-web-01, srv-proxmox-01, rack-a01, region-ap-northeast-1
	if result.Count != 4 {
		t.Errorf("Expected 4 ancestors, got %d", result.Count)
	}
}

func TestTraverseOutgoing(t *testing.T) {
	g := createTestGraph()
	engine := NewEngine(g)

	q := NewQuery()
	q.Select = NewSelectClause()
	q.Select.AddEntity("server")
	q.Traverse = NewTraverseClause("srv-proxmox-01", TraverseOpOutgoing)
	q.Traverse.RelationType = "hosts"

	result, err := engine.Execute(q)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	// srv-proxmox-01 hosts vm-web-01 and vm-api-01
	if result.Count != 2 {
		t.Errorf("Expected 2 outgoing targets, got %d", result.Count)
	}
}

func TestTraverseIncoming(t *testing.T) {
	g := createTestGraph()
	engine := NewEngine(g)

	q := NewQuery()
	q.Select = NewSelectClause()
	q.Select.AddEntity("vm")
	q.Traverse = NewTraverseClause("vm-web-01", TraverseOpIncoming)
	q.Traverse.RelationType = "hosts"

	result, err := engine.Execute(q)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	// vm-web-01 is hosted by srv-proxmox-01
	if result.Count != 1 {
		t.Errorf("Expected 1 incoming source, got %d", result.Count)
	}
}

func TestTraverseRelated(t *testing.T) {
	g := createTestGraph()
	engine := NewEngine(g)

	q := NewQuery()
	q.Select = NewSelectClause()
	q.Select.AddEntity("server")
	q.Traverse = NewTraverseClause("srv-proxmox-01", TraverseOpRelated)

	result, err := engine.Execute(q)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	// srv-proxmox-01 is related to vm-web-01, vm-api-01 (via hosts), and srv-proxmox-02 (via connects)
	if result.Count != 3 {
		t.Errorf("Expected 3 related entities, got %d", result.Count)
	}
}

func TestProjectIDs(t *testing.T) {
	g := createTestGraph()
	engine := NewEngine(g)

	q := NewQuery()
	q.Select = NewSelectClause()
	q.Select.AddEntity("server")
	q.Project = NewProjectClause(ProjectionTypeIDs)

	result, err := engine.Execute(q)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if result.Count != 2 {
		t.Errorf("Expected 2 results, got %d", result.Count)
	}

	// Check that result objects are IDs
	for _, item := range result.Results {
		id, ok := item.Object.(string)
		if !ok {
			t.Errorf("Expected object to be string ID, got %T", item.Object)
		}
		if id == "" {
			t.Error("Expected non-empty ID")
		}
	}
}

func TestProjectPaths(t *testing.T) {
	g := createTestGraph()
	engine := NewEngine(g)

	q := NewQuery()
	q.Select = NewSelectClause()
	q.Select.AddEntity("vm")
	q.Project = NewProjectClause(ProjectionTypePaths)

	result, err := engine.Execute(q)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if result.Count != 3 {
		t.Errorf("Expected 3 results, got %d", result.Count)
	}

	// Check that result objects are paths
	for _, item := range result.Results {
		path, ok := item.Object.(string)
		if !ok {
			t.Errorf("Expected object to be string path, got %T", item.Object)
		}
		if path == "" {
			t.Error("Expected non-empty path")
		}
	}
}

func TestProjectProperties(t *testing.T) {
	g := createTestGraph()
	engine := NewEngine(g)

	q := NewQuery()
	q.Select = NewSelectClause()
	q.Select.AddEntity("server")
	q.Project = NewProjectClause(ProjectionTypeProperties)
	q.Project.SetProperty("name")
	q.Project.SetProperty("status")

	result, err := engine.Execute(q)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if result.Count != 2 {
		t.Errorf("Expected 2 results, got %d", result.Count)
	}

	// Check that result objects have only requested properties
	for _, item := range result.Results {
		propMap, ok := item.Object.(map[string]interface{})
		if !ok {
			t.Errorf("Expected object to be map, got %T", item.Object)
			continue
		}
		if _, exists := propMap["name"]; !exists {
			t.Error("Expected 'name' property in projection")
		}
		if _, exists := propMap["status"]; !exists {
			t.Error("Expected 'status' property in projection")
		}
		if _, exists := propMap["id"]; exists {
			t.Error("Unexpected 'id' property in projection")
		}
	}
}

func TestLimit(t *testing.T) {
	g := createTestGraph()
	engine := NewEngine(g)

	q := NewQuery()
	q.Select = NewSelectClause()
	q.Select.AddEntity("vm")
	q.Limit = 2

	result, err := engine.Execute(q)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if result.Count != 3 {
		t.Errorf("Expected total count 3, got %d", result.Count)
	}

	if len(result.Results) != 2 {
		t.Errorf("Expected 2 results (limited), got %d", len(result.Results))
	}

	if !result.Truncated {
		t.Error("Expected truncated to be true")
	}
}

func TestOffset(t *testing.T) {
	g := createTestGraph()
	engine := NewEngine(g)

	q := NewQuery()
	q.Select = NewSelectClause()
	q.Select.AddEntity("vm")
	q.Offset = 1

	result, err := engine.Execute(q)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if result.Count != 3 {
		t.Errorf("Expected total count 3, got %d", result.Count)
	}

	if len(result.Results) != 2 {
		t.Errorf("Expected 2 results (offset), got %d", len(result.Results))
	}
}

func TestLimitAndOffset(t *testing.T) {
	g := createTestGraph()
	engine := NewEngine(g)

	q := NewQuery()
	q.Select = NewSelectClause()
	q.Select.AddEntity("vm")
	q.Limit = 1
	q.Offset = 1

	result, err := engine.Execute(q)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if result.Count != 3 {
		t.Errorf("Expected total count 3, got %d", result.Count)
	}

	if len(result.Results) != 1 {
		t.Errorf("Expected 1 result, got %d", len(result.Results))
	}
}

func TestQueryID(t *testing.T) {
	g := createTestGraph()
	engine := NewEngine(g)

	q := NewQuery()
	q.ID = "test-query"
	q.Select = NewSelectClause()
	q.Select.AddEntity("server")

	result, err := engine.Execute(q)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if result.QueryID != "test-query" {
		t.Errorf("Expected query ID 'test-query', got '%s'", result.QueryID)
	}
}

func TestNoSelectClause(t *testing.T) {
	g := createTestGraph()
	engine := NewEngine(g)

	q := NewQuery()

	_, err := engine.Execute(q)
	if err == nil {
		t.Error("Expected error for missing select clause")
	}
}

func TestSelectWithWhereFilter(t *testing.T) {
	g := createTestGraph()
	engine := NewEngine(g)

	q := NewQuery()
	q.Select = NewSelectClause()
	entitySel := q.Select.AddEntity("vm")
	entitySel.Where = NewWhereClause()
	entitySel.Where.AddCondition("status", OperatorEq, "active")

	result, err := engine.Execute(q)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	// Should only get active VMs
	if result.Count != 2 {
		t.Errorf("Expected 2 active VMs, got %d", result.Count)
	}
}

func TestSelectWithRelationFilter(t *testing.T) {
	g := createTestGraph()
	engine := NewEngine(g)

	q := NewQuery()
	q.Select = NewSelectClause()
	relSel := q.Select.AddRelation("connects")
	relSel.Where = NewWhereClause()
	relSel.Where.AddCondition("connection_type", OperatorEq, "network")

	result, err := engine.Execute(q)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if result.Count != 1 {
		t.Errorf("Expected 1 network connection, got %d", result.Count)
	}
}
