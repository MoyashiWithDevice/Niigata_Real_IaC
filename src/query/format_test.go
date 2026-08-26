package query

import (
	"encoding/json"
	"strings"
	"testing"

	"IACForge/src/core"
)

func TestToViewResultEntitiesOnly(t *testing.T) {
	g := createTestGraph()
	eng := NewEngine(g)

	q := NewQuery()
	q.Select = NewSelectClause()
	q.Select.AddEntity("server")
	result, err := eng.Execute(q)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	vr := eng.ToViewResult(result, false)
	if len(vr.VisibleEntities) != 2 {
		t.Fatalf("expected 2 visible entities, got %d", len(vr.VisibleEntities))
	}
	if vr.VisibleEntities[0].ID != "srv-proxmox-01" || vr.VisibleEntities[1].ID != "srv-proxmox-02" {
		t.Errorf("expected sorted entity IDs, got %q, %q", vr.VisibleEntities[0].ID, vr.VisibleEntities[1].ID)
	}
	if len(vr.VisibleRelations) != 0 {
		t.Errorf("expected no relations with includeRelations=false, got %d", len(vr.VisibleRelations))
	}
	if vr.ViewID != "query-result" {
		t.Errorf("expected fallback ViewID \"query-result\", got %q", vr.ViewID)
	}
}

func TestToViewResultInducedSubgraph(t *testing.T) {
	g := createTestGraph()
	eng := NewEngine(g)

	q := NewQuery()
	q.Select = NewSelectClause()
	q.Select.AddEntity("server")
	result, err := eng.Execute(q)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	vr := eng.ToViewResult(result, true)
	if len(vr.VisibleEntities) != 2 {
		t.Fatalf("expected 2 visible entities, got %d", len(vr.VisibleEntities))
	}

	// rel-connects-1 connects both servers and must be included.
	// rel-hosts-* reference VMs outside the node set and must be excluded.
	if len(vr.VisibleRelations) != 1 {
		t.Fatalf("expected 1 visible relation, got %d", len(vr.VisibleRelations))
	}
	if vr.VisibleRelations[0].ID != "rel-connects-1" {
		t.Errorf("expected rel-connects-1, got %s", vr.VisibleRelations[0].ID)
	}
}

func TestToViewResultRelationItems(t *testing.T) {
	g := createTestGraph()
	eng := NewEngine(g)

	q := NewQuery()
	q.Select = NewSelectClause()
	q.Select.AddRelation("hosts")
	result, err := eng.Execute(q)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}
	if result.Count != 3 {
		t.Fatalf("expected 3 relation results, got %d", result.Count)
	}

	vr := eng.ToViewResult(result, true)

	// Participants of the selected relations become nodes.
	expectedEntities := []string{
		"srv-proxmox-01", "srv-proxmox-02", "vm-api-01", "vm-dev-01", "vm-web-01",
	}
	if len(vr.VisibleEntities) != len(expectedEntities) {
		t.Fatalf("expected %d visible entities, got %d", len(expectedEntities), len(vr.VisibleEntities))
	}
	for i, id := range expectedEntities {
		if vr.VisibleEntities[i].ID != id {
			t.Errorf("expected entity %q at index %d, got %q", id, i, vr.VisibleEntities[i].ID)
		}
	}

	// Selected relations plus the induced subgraph relation between the servers.
	expectedRelations := []string{
		"rel-connects-1", "rel-hosts-1", "rel-hosts-2", "rel-hosts-3",
	}
	if len(vr.VisibleRelations) != len(expectedRelations) {
		t.Fatalf("expected %d visible relations, got %d", len(expectedRelations), len(vr.VisibleRelations))
	}
	for i, id := range expectedRelations {
		if vr.VisibleRelations[i].ID != id {
			t.Errorf("expected relation %q at index %d, got %q", id, i, vr.VisibleRelations[i].ID)
		}
	}
}

func TestToViewResultInterfaceReferences(t *testing.T) {
	g := core.NewGraph()
	srv := core.NewEntity("srv-01", "server", "Server 01")
	eno1 := core.NewEntity("eno1", "interface", "eno1")
	sw := core.NewEntity("sw-01", "switch", "Switch 01")
	port1 := core.NewEntity("port1", "interface", "port1")
	for _, e := range []*core.Entity{srv, eno1, sw, port1} {
		if err := g.AddEntity(e); err != nil {
			t.Fatalf("failed to add entity: %v", err)
		}
	}
	rel := core.NewSymmetricRelation("rel-conn", "connects", []string{"srv-01/eno1", "sw-01/port1"})
	if err := g.AddRelation(rel); err != nil {
		t.Fatalf("failed to add relation: %v", err)
	}

	eng := NewEngine(g)

	// Only the two hosts are selected: the interface-level edge is outside the
	// induced subgraph and must be excluded.
	hostOnly := &Result{
		QueryID: "q-hosts",
		Results: []*ResultItem{
			{ID: "srv-01", Type: "entity", Object: srv},
			{ID: "sw-01", Type: "entity", Object: sw},
		},
	}
	vrHost := eng.ToViewResult(hostOnly, true)
	if len(vrHost.VisibleRelations) != 0 {
		t.Errorf("expected no relations for host-only result, got %d", len(vrHost.VisibleRelations))
	}

	// When the interface nodes are part of the result, the edge appears with
	// participant references resolved to entity IDs.
	ifaceResult := &Result{
		QueryID: "q-interfaces",
		Results: []*ResultItem{
			{ID: "eno1", Type: "entity", Object: eno1},
			{ID: "port1", Type: "entity", Object: port1},
		},
	}
	vrIface := eng.ToViewResult(ifaceResult, true)
	if len(vrIface.VisibleEntities) != 2 {
		t.Fatalf("expected 2 visible entities, got %d", len(vrIface.VisibleEntities))
	}
	if len(vrIface.VisibleRelations) != 1 {
		t.Fatalf("expected 1 visible relation, got %d", len(vrIface.VisibleRelations))
	}
	edge := vrIface.VisibleRelations[0]
	if edge.ID != "rel-conn" {
		t.Errorf("expected relation rel-conn, got %s", edge.ID)
	}
	ids := edge.ParticipantIDs()
	if len(ids) != 2 || ids[0] != "eno1" || ids[1] != "port1" {
		t.Errorf("expected resolved participants [eno1 port1], got %v", ids)
	}

	// The source graph relation must be left unmodified.
	orig, _ := g.GetRelation("rel-conn")
	if orig.ParticipantIDs()[0] != "srv-01/eno1" {
		t.Errorf("source graph relation was modified: %v", orig.ParticipantIDs())
	}
}

func TestToViewResultDeterministic(t *testing.T) {
	g := createTestGraph()
	eng := NewEngine(g)

	q := NewQuery()
	q.Select = NewSelectClause()
	q.Select.AddEntity("vm")
	result, err := eng.Execute(q)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	first := eng.ToViewResult(result, true)
	second := eng.ToViewResult(result, true)

	if len(first.VisibleEntities) != len(second.VisibleEntities) {
		t.Fatalf("entity counts differ: %d vs %d", len(first.VisibleEntities), len(second.VisibleEntities))
	}
	for i := range first.VisibleEntities {
		if first.VisibleEntities[i].ID != second.VisibleEntities[i].ID {
			t.Errorf("entity order differs at %d: %q vs %q", i, first.VisibleEntities[i].ID, second.VisibleEntities[i].ID)
		}
	}
	if len(first.VisibleRelations) != len(second.VisibleRelations) {
		t.Fatalf("relation counts differ: %d vs %d", len(first.VisibleRelations), len(second.VisibleRelations))
	}
	for i := range first.VisibleRelations {
		if first.VisibleRelations[i].ID != second.VisibleRelations[i].ID {
			t.Errorf("relation order differs at %d: %q vs %q", i, first.VisibleRelations[i].ID, second.VisibleRelations[i].ID)
		}
	}
}

func TestResultJSON(t *testing.T) {
	g := createTestGraph()
	eng := NewEngine(g)

	q := NewQuery()
	q.Select = NewSelectClause()
	q.Select.AddEntity("server")
	result, err := eng.Execute(q)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(result.JSON(), &parsed); err != nil {
		t.Fatalf("failed to unmarshal JSON: %v", err)
	}

	if parsed["count"].(float64) != 2 {
		t.Errorf("expected count 2, got %v", parsed["count"])
	}
	if parsed["truncated"] != false {
		t.Errorf("expected truncated false, got %v", parsed["truncated"])
	}
	results, ok := parsed["results"].([]interface{})
	if !ok || len(results) != 2 {
		t.Fatalf("expected 2 results, got %v", results)
	}

	ids := make(map[string]bool)
	for _, raw := range results {
		item, ok := raw.(map[string]interface{})
		if !ok {
			t.Fatalf("expected result object, got %T", raw)
		}
		id, _ := item["id"].(string)
		ids[id] = true
		obj, ok := item["object"].(map[string]interface{})
		if !ok {
			t.Fatalf("expected object map, got %T", item["object"])
		}
		if obj["kind"] != "server" {
			t.Errorf("expected kind server, got %v", obj["kind"])
		}
		if _, hasName := obj["name"]; !hasName {
			t.Errorf("expected name field in object:\n%v", obj)
		}
	}
	for _, want := range []string{"srv-proxmox-01", "srv-proxmox-02"} {
		if !ids[want] {
			t.Errorf("expected id %q in results", want)
		}
	}
	if !strings.Contains(string(result.JSON()), `"spec"`) {
		t.Errorf("expected spec property in JSON:\n%s", result.JSON())
	}
}

func TestResultJSONEmptyQueryID(t *testing.T) {
	r := &Result{
		Results: []*ResultItem{},
	}
	data := r.JSON()
	if !strings.Contains(string(data), `"count": 0`) {
		t.Errorf("expected count 0 in JSON:\n%s", data)
	}
	if strings.Contains(string(data), "query_id") {
		t.Errorf("did not expect query_id for empty QueryID:\n%s", data)
	}
}
