package validation

import (
	"strings"
	"testing"

	"github.com/bababa/Niigata_Real_IaC/src/core"
	"github.com/bababa/Niigata_Real_IaC/src/core/kinds"
	"github.com/bababa/Niigata_Real_IaC/src/core/types"
	"github.com/bababa/Niigata_Real_IaC/src/schema"
)

// newTestGraph builds a typical valid Niigata graph:
//
//	area sado
//	├── forest f-01
//	│   └── ground g-01
//	├── water_body wb-01
//	├── species toki
//	│   └── population pop-01
//	└── tourism_spot ts-01
//	    └── hot_spring onsen-01
//	        └── event ev-01
//
// with relations located_in, inhabits, near, depends_on.
func newTestGraph() *core.Graph {
	g := core.NewGraph()

	sado := core.NewEntity("sado", kinds.Area, "Sado Island")
	sado.SetProperty("area_type", "island")
	sado.SetProperty("population", 55000)
	sado.SetProperty("latitude", 38.04)
	sado.SetProperty("longitude", 138.28)
	_ = g.AddEntity(sado)

	forest := core.NewEntity("f-01", kinds.Forest, "Beech Forest")
	forest.SetOwner("sado")
	forest.SetProperty("forest_type", "beech")
	forest.SetProperty("area_ha", 120)
	_ = g.AddEntity(forest)

	ground := core.NewEntity("g-01", kinds.Ground, "Forest Soil")
	ground.SetOwner("f-01")
	ground.SetProperty("soil_type", "volcanic_ash")
	ground.SetProperty("elevation_m", 250)
	ground.SetProperty("slope_deg", 15)
	ground.SetProperty("stability", "stable")
	_ = g.AddEntity(ground)

	waterBody := core.NewEntity("wb-01", kinds.WaterBody, "Ryotsu Bay")
	waterBody.SetOwner("sado")
	waterBody.SetProperty("water_type", "sea")
	_ = g.AddEntity(waterBody)

	toki := core.NewEntity("toki", kinds.Species, "Crested Ibis")
	toki.SetOwner("sado")
	toki.SetProperty("scientific_name", "Nipponia nippon")
	toki.SetProperty("category", "bird")
	toki.SetProperty("red_list_status", "endangered")
	_ = g.AddEntity(toki)

	population := core.NewEntity("pop-01", kinds.Population, "Toki Survey 2026")
	population.SetOwner("toki")
	population.SetProperty("count", 500)
	population.SetProperty("survey_date", "2026-06-01")
	population.SetProperty("survey_method", "visual_count")
	_ = g.AddEntity(population)

	spot := core.NewEntity("ts-01", kinds.TourismSpot, "Senkakuwan Bay")
	spot.SetOwner("sado")
	spot.SetProperty("spot_type", "scenic")
	spot.SetProperty("annual_visitors", 1200000)
	_ = g.AddEntity(spot)

	onsen := core.NewEntity("onsen-01", kinds.HotSpring, "Ono Onsen")
	onsen.SetOwner("ts-01")
	onsen.SetProperty("spring_quality", "sulfur")
	onsen.SetProperty("temperature_c", 78.5)
	onsen.SetProperty("source_count", 3)
	_ = g.AddEntity(onsen)

	event := core.NewEntity("ev-01", kinds.Event, "Winter Onsen Festival")
	event.SetOwner("onsen-01")
	event.SetProperty("season", "winter")
	event.SetProperty("held_month", 3)
	event.SetProperty("visitor_count", 80000)
	_ = g.AddEntity(event)

	locForest := core.NewDirectedRelation("loc-f01", types.LocatedIn, "f-01", "sado")
	locForest.SetProperty("distance_km", 12.5)
	_ = g.AddRelation(locForest)

	locWater := core.NewDirectedRelation("loc-wb01", types.LocatedIn, "wb-01", "sado")
	locWater.SetProperty("distance_km", 2.0)
	_ = g.AddRelation(locWater)

	inhabits := core.NewDirectedRelation("inh-toki", types.Inhabits, "toki", "f-01")
	_ = g.AddRelation(inhabits)

	nearRel := core.NewSymmetricRelation("near-ts-wb", types.Near, []string{"ts-01", "wb-01"})
	nearRel.SetProperty("walking_minutes", 20)
	_ = g.AddRelation(nearRel)

	dependsOn := core.NewDirectedRelation("dep-onsen-wb", types.DependsOn, "onsen-01", "wb-01")
	dependsOn.SetProperty("dependency_type", "source")
	dependsOn.SetProperty("critical", true)
	_ = g.AddRelation(dependsOn)

	return g
}

func newTestEngine() *Engine {
	s := schema.CoreSchema()
	e := NewEngine(s)
	RegisterCoreRules(e)
	return e
}

func hasFindingByRule(result *Result, ruleID string) bool {
	for _, f := range result.Findings {
		if f.RuleID == ruleID {
			return true
		}
	}
	return false
}

func hasFindingFor(result *Result, ruleID, objectID string) bool {
	for _, f := range result.Findings {
		if f.RuleID == ruleID && f.ObjectID == objectID {
			return true
		}
	}
	return false
}

func findingsByRule(result *Result, ruleID string) []Finding {
	var found []Finding
	for _, f := range result.Findings {
		if f.RuleID == ruleID {
			found = append(found, f)
		}
	}
	return found
}

func TestEngineCoreRulesRegistered(t *testing.T) {
	e := newTestEngine()
	expectedRules := []string{
		"unique-id", "valid-reference", "valid-owner", "single-owner",
		"valid-property",
		"required-kind", "required-name", "valid-kind", "valid-status",
		"no-slash-in-id", "valid-nesting-parent",
		"required-type", "required-participants", "valid-type", "valid-direction",
		"valid-cardinality", "valid-participant-kind",
		"ownership-tree", "no-ownership-cycle", "root-entity",
		"dangling-reference", "invalid-path",
		"population-requires-species", "positive-count", "valid-niigata-coordinates",
	}

	for _, ruleID := range expectedRules {
		if _, ok := e.ruleDefs[ruleID]; !ok {
			t.Errorf("expected rule %q to be registered", ruleID)
		}
	}

	removedRules := []string{
		"valid-port-range", "valid-acl-rule-parent",
		"valid-ip-format", "ip-requires-network", "network-reference-kind",
		"ip-in-cidr", "network-cidr-required", "gateway-in-cidr",
		"ip-unique-in-network",
	}
	for _, ruleID := range removedRules {
		if _, ok := e.ruleDefs[ruleID]; ok {
			t.Errorf("expected rule %q to be removed from the engine", ruleID)
		}
	}
}

func TestValidateValidGraph(t *testing.T) {
	e := newTestEngine()
	graph := newTestGraph()

	result := e.Validate(graph, nil)
	if !result.Passed {
		t.Errorf("expected validation to pass, but found errors:")
		for _, f := range result.Findings {
			if f.Severity == SeverityError {
				t.Errorf("  %s: %s", f.RuleID, f.Message)
			}
		}
	}
	if result.Summary.Warnings != 0 {
		t.Errorf("expected zero warnings for compliant graph, got %d:", result.Summary.Warnings)
		for _, f := range result.Findings {
			if f.Severity == SeverityWarning {
				t.Errorf("  %s: %s", f.RuleID, f.Message)
			}
		}
	}
}

func TestValidateDuplicateEntityID(t *testing.T) {
	e := newTestEngine()
	graph := core.NewGraph()

	area := core.NewEntity("dup-01", kinds.Area, "Area 1")
	if err := graph.AddEntity(area); err != nil {
		t.Fatalf("failed to addentity: %v", err)
	}

	result := e.Validate(graph, nil)
	for _, f := range result.Findings {
		if f.RuleID == "unique-id" && f.Severity == SeverityError {
			t.Errorf("unexpected unique-id error: %s", f.Message)
		}
	}
}

func TestValidateDuplicateRelationID(t *testing.T) {
	e := newTestEngine()
	graph := core.NewGraph()

	area := core.NewEntity("sado", kinds.Area, "Sado Island")
	if err := graph.AddEntity(area); err != nil {
		t.Fatalf("failed to addentity: %v", err)
	}
	spot := core.NewEntity("ts-01", kinds.TourismSpot, "Spot 1")
	spot.SetOwner("sado")
	if err := graph.AddEntity(spot); err != nil {
		t.Fatalf("failed to addentity: %v", err)
	}

	r1 := core.NewDirectedRelation("loc-01", types.LocatedIn, "ts-01", "sado")
	if err := graph.AddRelation(r1); err != nil {
		t.Fatalf("failed to addrelation: %v", err)
	}

	result := e.Validate(graph, nil)
	for _, f := range result.Findings {
		if f.RuleID == "unique-id" && f.ObjectType == ObjectTypeRelation {
			t.Errorf("unexpected unique-id error: %s", f.Message)
		}
	}
}

func TestValidateNoSlashInID(t *testing.T) {
	e := newTestEngine()
	graph := core.NewGraph()

	bad := core.NewEntity("sado/north", kinds.Area, "North Sado")
	graph.ForceAddEntity(bad)

	result := e.Validate(graph, nil)
	if !hasFindingByRule(result, "no-slash-in-id") {
		t.Error("expected no-slash-in-id error for ID containing slash")
	}
}

func TestValidateMissingKind(t *testing.T) {
	e := newTestEngine()
	graph := core.NewGraph()

	area := core.NewEntity("sado", "", "Sado Island")
	graph.ForceAddEntity(area)

	result := e.Validate(graph, nil)
	found := false
	for _, f := range result.Findings {
		if f.RuleID == "required-kind" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected required-kind error")
	}
}

func TestValidateMissingName(t *testing.T) {
	e := newTestEngine()
	graph := core.NewGraph()

	area := core.NewEntity("sado", kinds.Area, "")
	graph.ForceAddEntity(area)

	result := e.Validate(graph, nil)
	found := false
	for _, f := range result.Findings {
		if f.RuleID == "required-name" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected required-name error")
	}
}

func TestValidateInvalidKind(t *testing.T) {
	e := newTestEngine()
	graph := core.NewGraph()

	area := core.NewEntity("sado", "nonexistent_kind", "Sado Island")
	graph.ForceAddEntity(area)

	result := e.Validate(graph, nil)
	found := false
	for _, f := range result.Findings {
		if f.RuleID == "valid-kind" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected valid-kind error")
	}
}

func TestValidateInvalidStatus(t *testing.T) {
	e := newTestEngine()
	graph := core.NewGraph()

	area := core.NewEntity("sado", kinds.Area, "Sado Island")
	area.SetStatus("invalid_status")
	if err := graph.AddEntity(area); err != nil {
		t.Fatalf("failed to addentity: %v", err)
	}

	result := e.Validate(graph, nil)
	found := false
	for _, f := range result.Findings {
		if f.RuleID == "valid-status" && f.Severity == SeverityWarning {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected valid-status warning")
	}
}

func TestValidateNestingParentValid(t *testing.T) {
	e := newTestEngine()
	graph := newTestGraph()

	result := e.Validate(graph, nil)
	for _, f := range result.Findings {
		if f.RuleID == "valid-nesting-parent" {
			t.Errorf("unexpected valid-nesting-parent warning: %s", f.Message)
		}
	}
}

func TestValidateNestingParentInvalid(t *testing.T) {
	e := newTestEngine()
	graph := core.NewGraph()

	area := core.NewEntity("sado", kinds.Area, "Sado Island")
	if err := graph.AddEntity(area); err != nil {
		t.Fatalf("failed to addentity: %v", err)
	}
	// an area does not nest populations directly; populations belong to species
	orphan := core.NewEntity("pop-01", kinds.Population, "Stray Population")
	orphan.SetOwner("sado")
	orphan.SetProperty("count", 10)
	if err := graph.AddEntity(orphan); err != nil {
		t.Fatalf("failed to addentity: %v", err)
	}

	result := e.Validate(graph, nil)
	found := false
	for _, f := range result.Findings {
		if f.RuleID == "valid-nesting-parent" && f.Severity == SeverityWarning && f.ObjectID == "pop-01" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected valid-nesting-parent warning for population nested under area")
	}
}

func TestValidateMissingRelationType(t *testing.T) {
	e := newTestEngine()
	graph := core.NewGraph()

	area := core.NewEntity("sado", kinds.Area, "Sado Island")
	if err := graph.AddEntity(area); err != nil {
		t.Fatalf("failed to addentity: %v", err)
	}
	spot := core.NewEntity("ts-01", kinds.TourismSpot, "Spot 1")
	spot.SetOwner("sado")
	if err := graph.AddEntity(spot); err != nil {
		t.Fatalf("failed to addentity: %v", err)
	}

	r := core.NewRelation("rel-01", "", core.DirectionDirected)
	r.Participants.Source = "ts-01"
	r.Participants.Target = "sado"
	graph.ForceAddRelation(r)

	result := e.Validate(graph, nil)
	found := false
	for _, f := range result.Findings {
		if f.RuleID == "required-type" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected required-type error")
	}
}

func TestValidateInvalidRelationType(t *testing.T) {
	e := newTestEngine()
	graph := core.NewGraph()

	area := core.NewEntity("sado", kinds.Area, "Sado Island")
	if err := graph.AddEntity(area); err != nil {
		t.Fatalf("failed to addentity: %v", err)
	}
	spot := core.NewEntity("ts-01", kinds.TourismSpot, "Spot 1")
	spot.SetOwner("sado")
	if err := graph.AddEntity(spot); err != nil {
		t.Fatalf("failed to addentity: %v", err)
	}

	r := core.NewDirectedRelation("rel-01", "nonexistent_type", "ts-01", "sado")
	if err := graph.AddRelation(r); err != nil {
		t.Fatalf("failed to addrelation: %v", err)
	}

	result := e.Validate(graph, nil)
	found := false
	for _, f := range result.Findings {
		if f.RuleID == "valid-type" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected valid-type error")
	}
}

func TestValidateDirectedRelationMissingTarget(t *testing.T) {
	e := newTestEngine()
	graph := core.NewGraph()

	area := core.NewEntity("sado", kinds.Area, "Sado Island")
	if err := graph.AddEntity(area); err != nil {
		t.Fatalf("failed to addentity: %v", err)
	}
	forest := core.NewEntity("f-01", kinds.Forest, "Forest 1")
	forest.SetOwner("sado")
	if err := graph.AddEntity(forest); err != nil {
		t.Fatalf("failed to addentity: %v", err)
	}

	r := core.NewRelation("rel-01", types.LocatedIn, core.DirectionDirected)
	r.Participants.Source = "f-01"
	graph.ForceAddRelation(r)

	result := e.Validate(graph, nil)
	found := false
	for _, f := range result.Findings {
		if f.RuleID == "valid-direction" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected valid-direction error for directed relation without target")
	}
}

func TestValidateParticipantKindWarning(t *testing.T) {
	e := newTestEngine()
	graph := core.NewGraph()

	area := core.NewEntity("sado", kinds.Area, "Sado Island")
	if err := graph.AddEntity(area); err != nil {
		t.Fatalf("failed to addentity: %v", err)
	}
	toki := core.NewEntity("toki", kinds.Species, "Crested Ibis")
	toki.SetOwner("sado")
	if err := graph.AddEntity(toki); err != nil {
		t.Fatalf("failed to addentity: %v", err)
	}

	// flows_into should have water_body participants, not species/area
	r := core.NewDirectedRelation("rel-01", types.FlowsInto, "toki", "sado")
	if err := graph.AddRelation(r); err != nil {
		t.Fatalf("failed to addrelation: %v", err)
	}

	result := e.Validate(graph, nil)
	found := false
	for _, f := range result.Findings {
		if f.RuleID == "valid-participant-kind" && f.Severity == SeverityWarning {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected valid-participant-kind warning for species/area in flows_into relation")
	}
}

func TestValidateParticipantKindDirection(t *testing.T) {
	e := newTestEngine()
	graph := core.NewGraph()

	area := core.NewEntity("sado", kinds.Area, "Sado Island")
	if err := graph.AddEntity(area); err != nil {
		t.Fatalf("failed to addentity: %v", err)
	}
	terrain := core.NewEntity("mt-01", kinds.Terrain, "Mount Kimpoku")
	terrain.SetOwner("sado")
	terrain.SetProperty("terrain_type", "mountain")
	if err := graph.AddEntity(terrain); err != nil {
		t.Fatalf("failed to addentity: %v", err)
	}

	// located_in source kinds are {ground,terrain,water_body,forest,tourism,
	// species,population} and target kinds are {area,terrain}, so a source of
	// kind area is only caught when the direction is respected.
	r := core.NewDirectedRelation("rel-rev", types.LocatedIn, "sado", "mt-01")
	if err := graph.AddRelation(r); err != nil {
		t.Fatalf("failed to addrelation: %v", err)
	}

	result := e.Validate(graph, nil)
	found := findingsByRule(result, "valid-participant-kind")
	if len(found) != 1 {
		t.Fatalf("expected exactly 1 valid-participant-kind warning for reversed located_in relation, got %d", len(found))
	}
	if !strings.Contains(found[0].Message, "source") {
		t.Errorf("expected source role in message, got: %s", found[0].Message)
	}
}

func TestValidateParticipantKindDirectionValid(t *testing.T) {
	e := newTestEngine()
	graph := core.NewGraph()

	area := core.NewEntity("sado", kinds.Area, "Sado Island")
	if err := graph.AddEntity(area); err != nil {
		t.Fatalf("failed to addentity: %v", err)
	}
	terrain := core.NewEntity("mt-01", kinds.Terrain, "Mount Kimpoku")
	terrain.SetOwner("sado")
	terrain.SetProperty("terrain_type", "mountain")
	if err := graph.AddEntity(terrain); err != nil {
		t.Fatalf("failed to addentity: %v", err)
	}

	r := core.NewDirectedRelation("rel-ok", types.LocatedIn, "mt-01", "sado")
	if err := graph.AddRelation(r); err != nil {
		t.Fatalf("failed to addrelation: %v", err)
	}

	result := e.Validate(graph, nil)
	for _, f := range result.Findings {
		if f.RuleID == "valid-participant-kind" && f.Severity == SeverityWarning {
			t.Errorf("unexpected valid-participant-kind warning: %s", f.Message)
		}
	}
}

func TestValidateParticipantKindSymmetricFallback(t *testing.T) {
	e := newTestEngine()
	graph := core.NewGraph()

	area := core.NewEntity("sado", kinds.Area, "Sado Island")
	if err := graph.AddEntity(area); err != nil {
		t.Fatalf("failed to addentity: %v", err)
	}
	waterBody := core.NewEntity("wb-01", kinds.WaterBody, "Lake Kamo")
	waterBody.SetOwner("sado")
	waterBody.SetProperty("water_type", "lake")
	if err := graph.AddEntity(waterBody); err != nil {
		t.Fatalf("failed to addentity: %v", err)
	}
	toki := core.NewEntity("toki", kinds.Species, "Crested Ibis")
	toki.SetOwner("sado")
	if err := graph.AddEntity(toki); err != nil {
		t.Fatalf("failed to addentity: %v", err)
	}

	// near allows tourism/nature/area participants; symmetric relations fall
	// back to the union of source and target kinds, so species is flagged.
	bad := core.NewSymmetricRelation("near-bad", types.Near, []string{"toki", "wb-01"})
	if err := graph.AddRelation(bad); err != nil {
		t.Fatalf("failed to addrelation: %v", err)
	}
	good := core.NewSymmetricRelation("near-ok", types.Near, []string{"sado", "wb-01"})
	if err := graph.AddRelation(good); err != nil {
		t.Fatalf("failed to addrelation: %v", err)
	}

	result := e.Validate(graph, nil)
	found := findingsByRule(result, "valid-participant-kind")
	if len(found) != 1 {
		t.Fatalf("expected exactly 1 valid-participant-kind warning, got %d", len(found))
	}
	if !strings.Contains(found[0].Message, "toki") {
		t.Errorf("expected warning for toki participant, got: %s", found[0].Message)
	}
}

func TestValidateCardinalityTooFewParticipants(t *testing.T) {
	e := newTestEngine()
	graph := core.NewGraph()

	area := core.NewEntity("sado", kinds.Area, "Sado Island")
	if err := graph.AddEntity(area); err != nil {
		t.Fatalf("failed to addentity: %v", err)
	}

	r := core.NewRelation("rel-01", types.LocatedIn, core.DirectionDirected)
	r.Participants.Source = "sado"
	graph.ForceAddRelation(r)

	result := e.Validate(graph, nil)
	if !hasFindingByRule(result, "valid-cardinality") {
		t.Error("expected valid-cardinality error for relation with fewer than minimum participants")
	}
}

func TestValidateCardinalityTooManyParticipants(t *testing.T) {
	e := newTestEngine()
	graph := core.NewGraph()

	area := core.NewEntity("sado", kinds.Area, "Sado Island")
	if err := graph.AddEntity(area); err != nil {
		t.Fatalf("failed to addentity: %v", err)
	}
	spot := core.NewEntity("ts-01", kinds.TourismSpot, "Spot 1")
	spot.SetOwner("sado")
	if err := graph.AddEntity(spot); err != nil {
		t.Fatalf("failed to addentity: %v", err)
	}
	waterBody := core.NewEntity("wb-01", kinds.WaterBody, "Water Body 1")
	waterBody.SetOwner("sado")
	if err := graph.AddEntity(waterBody); err != nil {
		t.Fatalf("failed to addentity: %v", err)
	}
	forest := core.NewEntity("f-01", kinds.Forest, "Forest 1")
	forest.SetOwner("sado")
	if err := graph.AddEntity(forest); err != nil {
		t.Fatalf("failed to addentity: %v", err)
	}

	r := core.NewRelation("rel-01", types.Near, core.DirectionSymmetric)
	r.Participants.List = []string{"sado", "ts-01", "wb-01", "f-01"}
	graph.ForceAddRelation(r)

	result := e.Validate(graph, nil)
	if !hasFindingByRule(result, "valid-cardinality") {
		t.Error("expected valid-cardinality error for relation exceeding maximum participants")
	}
}

func TestValidateValidRelationParticipants(t *testing.T) {
	e := newTestEngine()
	graph := newTestGraph()

	result := e.Validate(graph, nil)
	for _, f := range result.Findings {
		if f.RuleID == "required-participants" && f.Severity == SeverityError {
			t.Errorf("unexpected required-participants error: %s", f.Message)
		}
	}
}

func TestValidateOwnershipCycle(t *testing.T) {
	e := newTestEngine()
	graph := core.NewGraph()

	a := core.NewEntity("a", kinds.Forest, "A")
	a.SetOwner("b")
	graph.ForceAddEntity(a)
	b := core.NewEntity("b", kinds.Ground, "B")
	b.SetOwner("a")
	graph.ForceAddEntity(b)

	result := e.Validate(graph, nil)
	found := false
	for _, f := range result.Findings {
		if f.RuleID == "no-ownership-cycle" && f.Severity == SeverityError {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected no-ownership-cycle error")
	}
}

func TestValidateMultipleRoots(t *testing.T) {
	e := newTestEngine()
	graph := core.NewGraph()

	root1 := core.NewEntity("sado", kinds.Area, "Sado Island")
	if err := graph.AddEntity(root1); err != nil {
		t.Fatalf("failed to addentity: %v", err)
	}
	root2 := core.NewEntity("echigo", kinds.Area, "Echigo Plain")
	if err := graph.AddEntity(root2); err != nil {
		t.Fatalf("failed to addentity: %v", err)
	}

	result := e.Validate(graph, nil)
	found := false
	for _, f := range result.Findings {
		if f.RuleID == "root-entity" && f.Severity == SeverityError {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected root-entity error for multiple roots")
	}
}

func TestValidateNoRoot(t *testing.T) {
	e := newTestEngine()
	graph := core.NewGraph()

	a := core.NewEntity("a", kinds.Forest, "A")
	a.SetOwner("b")
	graph.ForceAddEntity(a)
	b := core.NewEntity("b", kinds.Ground, "B")
	b.SetOwner("a")
	graph.ForceAddEntity(b)

	result := e.Validate(graph, nil)
	found := false
	for _, f := range result.Findings {
		if f.RuleID == "root-entity" && f.Severity == SeverityError {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected root-entity error")
	}
}

func TestValidatePathReferenceParticipant(t *testing.T) {
	e := newTestEngine()
	graph := core.NewGraph()

	area := core.NewEntity("sado", kinds.Area, "Sado Island")
	if err := graph.AddEntity(area); err != nil {
		t.Fatalf("failed to addentity: %v", err)
	}
	spot := core.NewEntity("ts-01", kinds.TourismSpot, "Spot 1")
	spot.SetOwner("sado")
	if err := graph.AddEntity(spot); err != nil {
		t.Fatalf("failed to addentity: %v", err)
	}
	event := core.NewEntity("ev-01", kinds.Event, "Event 1")
	event.SetOwner("ts-01")
	if err := graph.AddEntity(event); err != nil {
		t.Fatalf("failed to addentity: %v", err)
	}

	// "ts-01/ev-01" resolves via legacy path notation to the event entity.
	r := core.NewDirectedRelation("rel-01", types.BelongsTo, "ts-01/ev-01", "sado")
	if err := graph.AddRelation(r); err != nil {
		t.Fatalf("failed to addrelation: %v", err)
	}

	result := e.Validate(graph, nil)
	for _, f := range result.Findings {
		if (f.RuleID == "valid-reference" || f.RuleID == "dangling-reference") && f.Severity == SeverityError {
			t.Errorf("unexpected %s error: %s", f.RuleID, f.Message)
		}
	}
}

func TestValidateDanglingReference(t *testing.T) {
	e := newTestEngine()
	graph := core.NewGraph()

	toki := core.NewEntity("toki", kinds.Species, "Crested Ibis")
	if err := graph.AddEntity(toki); err != nil {
		t.Fatalf("failed to addentity: %v", err)
	}

	r := core.NewDirectedRelation("rel-01", types.Inhabits, "toki", "nonexistent")
	graph.ForceAddRelation(r)

	result := e.Validate(graph, nil)
	found := false
	for _, f := range result.Findings {
		if f.RuleID == "dangling-reference" && f.Severity == SeverityError {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected dangling-reference error")
	}
}

func TestValidateSummary(t *testing.T) {
	e := newTestEngine()
	graph := core.NewGraph()

	result := e.Validate(graph, nil)

	if result.Summary.TotalRules == 0 {
		t.Error("expected non-zero total rules")
	}
	if result.Summary.TotalFindings != len(result.Findings) {
		t.Errorf("expected summary total findings %d to match findings count %d",
			result.Summary.TotalFindings, len(result.Findings))
	}
}

func TestValidateWithProfile(t *testing.T) {
	e := newTestEngine()
	graph := newTestGraph()

	profile := schema.NewProfile("minimal")
	profile.Rules = []string{"required-kind", "required-name"}

	result := e.Validate(graph, profile)

	if !result.Passed {
		t.Errorf("expected validation to pass with minimal profile, but found errors:")
		for _, f := range result.Findings {
			if f.Severity == SeverityError {
				t.Errorf("  %s: %s", f.RuleID, f.Message)
			}
		}
	}

	for _, f := range result.Findings {
		if f.RuleID != "required-kind" && f.RuleID != "required-name" {
			t.Errorf("unexpected rule %q found in profile-filtered validation", f.RuleID)
		}
	}
}

func TestValidateProfileRequiredKinds(t *testing.T) {
	e := newTestEngine()
	graph := core.NewGraph()

	area := core.NewEntity("sado", kinds.Area, "Sado Island")
	if err := graph.AddEntity(area); err != nil {
		t.Fatalf("failed to addentity: %v", err)
	}

	profile := schema.NewProfile("with-kinds")
	profile.AddRequiredKind("species")

	result := e.Validate(graph, profile)

	found := false
	for _, f := range result.Findings {
		if f.RuleID == "profile-required-kind" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected profile-required-kind error")
	}
}

func TestValidateProfileRequiredRelations(t *testing.T) {
	e := newTestEngine()
	graph := newTestGraph()

	profile := schema.NewProfile("with-relations")
	profile.AddRequiredRelation("flows_into")

	result := e.Validate(graph, profile)

	found := false
	for _, f := range result.Findings {
		if f.RuleID == "profile-required-relation" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected profile-required-relation error")
	}
}

func TestValidateResultPassed(t *testing.T) {
	e := newTestEngine()
	graph := core.NewGraph()
	if err := graph.AddEntity(core.NewEntity("sado", kinds.Area, "Sado Island")); err != nil {
		t.Fatalf("failed to add entity: %v", err)
	}

	result := e.Validate(graph, nil)

	errorCount := 0
	for _, f := range result.Findings {
		if f.Severity == SeverityError {
			errorCount++
		}
	}

	if errorCount > 0 && result.Passed {
		t.Error("expected Passed=false when there are errors")
	}
	if errorCount == 0 && !result.Passed {
		t.Error("expected Passed=true when there are no errors")
	}
}

func TestDanglingPropertyReference(t *testing.T) {
	e := newTestEngine()
	graph := core.NewGraph()

	area := core.NewEntity("sado", kinds.Area, "Sado Island")
	if err := graph.AddEntity(area); err != nil {
		t.Fatalf("failed to addentity: %v", err)
	}

	event := core.NewEntity("ev-01", kinds.Event, "Drum Festival")
	event.SetOwner("sado")
	event.SetProperty("season", "summer")
	if err := graph.AddEntity(event); err != nil {
		t.Fatalf("failed to addentity: %v", err)
	}

	validAsset := core.NewEntity("ca-100", kinds.CulturalAsset, "Shukunegi Houses")
	validAsset.SetOwner("sado")
	validAsset.SetProperty("designated_level", "national")
	validAsset.SetProperty("designated_date", core.NewReferenceValue("@ev-01"))
	if err := graph.AddEntity(validAsset); err != nil {
		t.Fatalf("failed to addentity: %v", err)
	}

	danglingAsset := core.NewEntity("ca-200", kinds.CulturalAsset, "Old Lighthouse")
	danglingAsset.SetOwner("sado")
	danglingAsset.SetProperty("designated_level", "municipal")
	danglingAsset.SetProperty("designated_date", core.NewReferenceValue("@nonexistent"))
	if err := graph.AddEntity(danglingAsset); err != nil {
		t.Fatalf("failed to addentity: %v", err)
	}

	result := e.Validate(graph, nil)

	foundDangling := false
	for _, f := range result.Findings {
		if f.RuleID == "dangling-reference" && f.ObjectID == "ca-200" {
			foundDangling = true
			break
		}
	}
	if !foundDangling {
		t.Error("expected dangling-reference error for ca-200 property reference")
	}

	// ca-100 should NOT have a dangling reference error
	for _, f := range result.Findings {
		if f.RuleID == "dangling-reference" && f.ObjectID == "ca-100" {
			t.Error("ca-100 should not have dangling-reference error")
		}
	}
}

func TestDanglingReferenceInListProperty(t *testing.T) {
	e := newTestEngine()
	graph := core.NewGraph()

	area := core.NewEntity("sado", kinds.Area, "Sado Island")
	if err := graph.AddEntity(area); err != nil {
		t.Fatalf("failed to addentity: %v", err)
	}

	waterBody := core.NewEntity("wb-01", kinds.WaterBody, "Lake Kamo")
	waterBody.SetOwner("sado")
	waterBody.SetProperty("water_type", "lake")
	if err := graph.AddEntity(waterBody); err != nil {
		t.Fatalf("failed to addentity: %v", err)
	}

	forest := core.NewEntity("f-01", kinds.Forest, "Beech Forest")
	forest.SetOwner("sado")
	forest.SetProperty("forest_type", "beech")
	if err := graph.AddEntity(forest); err != nil {
		t.Fatalf("failed to addentity: %v", err)
	}

	// ts-01 has one valid list reference (f-01) and one dangling (ghost).
	spot := core.NewEntity("ts-01", kinds.TourismSpot, "Scenic Spot")
	spot.SetOwner("sado")
	spot.SetProperty("spot_type", "scenic")
	spot.SetProperty("nearby_resources", []interface{}{
		core.NewReferenceValue("@f-01"),
		core.NewReferenceValue("@ghost"),
	})
	if err := graph.AddEntity(spot); err != nil {
		t.Fatalf("failed to addentity: %v", err)
	}

	// onsen-01 has a nested-map reference that is dangling (missing-source).
	onsen := core.NewEntity("onsen-01", kinds.HotSpring, "Hot Spring")
	onsen.SetOwner("ts-01")
	onsen.SetProperty("spring_quality", "chloride")
	onsen.SetProperty("source_config", map[string]interface{}{
		"feeds": []interface{}{
			core.NewReferenceValue("@wb-01"),
			core.NewReferenceValue("@missing-source"),
		},
	})
	if err := graph.AddEntity(onsen); err != nil {
		t.Fatalf("failed to addentity: %v", err)
	}

	// museum-01 has only valid list references.
	museum := core.NewEntity("museum-01", kinds.TourismSpot, "Museum")
	museum.SetOwner("sado")
	museum.SetProperty("spot_type", "museum")
	museum.SetProperty("nearby_resources", []interface{}{
		core.NewReferenceValue("@wb-01"),
	})
	if err := graph.AddEntity(museum); err != nil {
		t.Fatalf("failed to addentity: %v", err)
	}

	result := e.Validate(graph, nil)

	findings := map[string]bool{}
	for _, f := range result.Findings {
		if f.RuleID == "dangling-reference" {
			findings[f.Message] = true
		}
	}

	if !findings["entity \"ts-01\" property \"nearby_resources\" references non-existent object \"ghost\""] {
		t.Error("expected dangling-reference error for list element @ghost on ts-01")
	}
	if !findings["entity \"onsen-01\" property \"source_config\" references non-existent object \"missing-source\""] {
		t.Error("expected dangling-reference error for nested-map element @missing-source on onsen-01")
	}
	for _, f := range result.Findings {
		if f.RuleID == "dangling-reference" && f.ObjectID == "museum-01" {
			t.Errorf("museum-01 should not have dangling-reference error, got: %v", f.Message)
		}
	}
}

func TestValidPropertyReference(t *testing.T) {
	e := newTestEngine()
	graph := core.NewGraph()

	area := core.NewEntity("sado", kinds.Area, "Sado Island")
	if err := graph.AddEntity(area); err != nil {
		t.Fatalf("failed to addentity: %v", err)
	}

	event := core.NewEntity("ev-01", kinds.Event, "Drum Festival")
	event.SetOwner("sado")
	event.SetProperty("season", "summer")
	if err := graph.AddEntity(event); err != nil {
		t.Fatalf("failed to addentity: %v", err)
	}

	asset := core.NewEntity("ca-100", kinds.CulturalAsset, "Shukunegi Houses")
	asset.SetOwner("sado")
	asset.SetProperty("designated_date", core.NewReferenceValue("@ev-01"))
	if err := graph.AddEntity(asset); err != nil {
		t.Fatalf("failed to addentity: %v", err)
	}

	result := e.Validate(graph, nil)

	for _, f := range result.Findings {
		if f.RuleID == "dangling-reference" && f.ObjectID == "ca-100" {
			t.Error("valid property reference should not cause dangling-reference error")
		}
	}
}

func TestInvalidPathNonExistentEntity(t *testing.T) {
	e := newTestEngine()
	graph := core.NewGraph()

	area := core.NewEntity("sado", kinds.Area, "Sado Island")
	area.SetPath("/sado")
	if err := graph.AddEntity(area); err != nil {
		t.Fatalf("failed to addentity: %v", err)
	}

	forest := core.NewEntity("f-01", kinds.Forest, "Beech Forest")
	forest.SetOwner("sado")
	forest.SetPath("/sado/nonexistent/f-01")
	if err := graph.AddEntity(forest); err != nil {
		t.Fatalf("failed to addentity: %v", err)
	}

	result := e.Validate(graph, nil)
	found := false
	for _, f := range result.Findings {
		if f.RuleID == "invalid-path" && f.Severity == SeverityError {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected invalid-path error for path referencing non-existent entity")
	}
}

func TestInvalidPathWrongOwnership(t *testing.T) {
	e := newTestEngine()
	graph := core.NewGraph()

	area := core.NewEntity("sado", kinds.Area, "Sado Island")
	area.SetPath("/sado")
	if err := graph.AddEntity(area); err != nil {
		t.Fatalf("failed to addentity: %v", err)
	}

	forest := core.NewEntity("f-01", kinds.Forest, "Beech Forest")
	forest.SetOwner("sado")
	forest.SetPath("/sado/f-01")
	if err := graph.AddEntity(forest); err != nil {
		t.Fatalf("failed to addentity: %v", err)
	}

	ground := core.NewEntity("g-01", kinds.Ground, "Forest Soil")
	ground.SetOwner("f-01")
	ground.SetPath("/sado/f-01/g-01")

	// Manually set wrong path (not matching ownership)
	ground.SetPath("/sado/g-01")
	if err := graph.AddEntity(ground); err != nil {
		t.Fatalf("failed to addentity: %v", err)
	}

	result := e.Validate(graph, nil)
	found := false
	for _, f := range result.Findings {
		if f.RuleID == "invalid-path" && f.Severity == SeverityError {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected invalid-path error for path with wrong ownership")
	}
}

func TestValidPath(t *testing.T) {
	e := newTestEngine()
	graph := core.NewGraph()

	area := core.NewEntity("sado", kinds.Area, "Sado Island")
	area.SetPath("/sado")
	if err := graph.AddEntity(area); err != nil {
		t.Fatalf("failed to addentity: %v", err)
	}

	forest := core.NewEntity("f-01", kinds.Forest, "Beech Forest")
	forest.SetOwner("sado")
	forest.SetPath("/sado/f-01")
	if err := graph.AddEntity(forest); err != nil {
		t.Fatalf("failed to addentity: %v", err)
	}

	ground := core.NewEntity("g-01", kinds.Ground, "Forest Soil")
	ground.SetOwner("f-01")
	ground.SetPath("/sado/f-01/g-01")
	if err := graph.AddEntity(ground); err != nil {
		t.Fatalf("failed to addentity: %v", err)
	}

	result := e.Validate(graph, nil)
	for _, f := range result.Findings {
		if f.RuleID == "invalid-path" && f.Severity == SeverityError {
			t.Errorf("unexpected invalid-path error: %s", f.Message)
		}
	}
}

// --- Valid Property Rule ---

func TestValidPropertyCompliant(t *testing.T) {
	e := newTestEngine()
	graph := core.NewGraph()
	area := core.NewEntity("sado", kinds.Area, "Sado Island")
	if err := graph.AddEntity(area); err != nil {
		t.Fatalf("failed to addentity: %v", err)
	}
	area.SetProperty("area_type", "island")
	area.SetProperty("population", 55000)
	area.SetProperty("latitude", 38.04)
	area.SetProperty("longitude", 138.28)
	area.SetProperty("timezone", "Asia/Tokyo")

	result := e.Validate(graph, nil)
	for _, f := range result.Findings {
		if f.RuleID == "valid-property" {
			t.Errorf("unexpected valid-property warning: %s", f.Message)
		}
	}
}

func TestValidPropertyTypeMismatch(t *testing.T) {
	e := newTestEngine()
	graph := core.NewGraph()
	area := core.NewEntity("sado", kinds.Area, "Sado Island")
	if err := graph.AddEntity(area); err != nil {
		t.Fatalf("failed to addentity: %v", err)
	}
	area.SetProperty("area_type", "island")
	area.SetProperty("population", "55000") // defined as integer

	result := e.Validate(graph, nil)
	if !hasFindingFor(result, "valid-property", "sado") {
		t.Error("expected valid-property warning for sado with non-integer population")
	}
}

func TestValidPropertyEnumViolation(t *testing.T) {
	e := newTestEngine()
	graph := core.NewGraph()
	area := core.NewEntity("sado", kinds.Area, "Sado Island")
	if err := graph.AddEntity(area); err != nil {
		t.Fatalf("failed to addentity: %v", err)
	}
	toki := core.NewEntity("toki", kinds.Species, "Crested Ibis")
	toki.SetOwner("sado")
	toki.SetProperty("category", "bird")
	if err := graph.AddEntity(toki); err != nil {
		t.Fatalf("failed to addentity: %v", err)
	}
	goodPop := core.NewEntity("pop-good", kinds.Population, "Survey A")
	goodPop.SetOwner("toki")
	goodPop.SetProperty("count", 100)
	goodPop.SetProperty("survey_method", "drone")
	if err := graph.AddEntity(goodPop); err != nil {
		t.Fatalf("failed to addentity: %v", err)
	}
	badPop := core.NewEntity("pop-bad", kinds.Population, "Survey B")
	badPop.SetOwner("toki")
	badPop.SetProperty("count", 200)
	badPop.SetProperty("survey_method", "satellite") // not in enum
	if err := graph.AddEntity(badPop); err != nil {
		t.Fatalf("failed to addentity: %v", err)
	}

	result := e.Validate(graph, nil)
	if !hasFindingFor(result, "valid-property", "pop-bad") {
		t.Error("expected valid-property warning for pop-bad with invalid survey_method")
	}
	if hasFindingFor(result, "valid-property", "pop-good") {
		t.Error("pop-good with valid survey_method should not have valid-property warning")
	}
}

func TestValidPropertyMissingRequired(t *testing.T) {
	e := newTestEngine()
	graph := core.NewGraph()
	area := core.NewEntity("sado", kinds.Area, "Sado Island")
	if err := graph.AddEntity(area); err != nil {
		t.Fatalf("failed to addentity: %v", err)
	}
	toki := core.NewEntity("toki", kinds.Species, "Crested Ibis")
	toki.SetOwner("sado")
	if err := graph.AddEntity(toki); err != nil {
		t.Fatalf("failed to addentity: %v", err)
	}
	population := core.NewEntity("pop-01", kinds.Population, "Toki Survey")
	population.SetOwner("toki")
	// count is required but not set
	if err := graph.AddEntity(population); err != nil {
		t.Fatalf("failed to addentity: %v", err)
	}

	result := e.Validate(graph, nil)
	if !hasFindingFor(result, "valid-property", "pop-01") {
		t.Error("expected valid-property warning for pop-01 missing required count")
	}
}

func TestValidPropertyUndefinedProperty(t *testing.T) {
	e := newTestEngine()
	graph := core.NewGraph()
	area := core.NewEntity("sado", kinds.Area, "Sado Island")
	if err := graph.AddEntity(area); err != nil {
		t.Fatalf("failed to addentity: %v", err)
	}
	toki := core.NewEntity("toki", kinds.Species, "Crested Ibis")
	toki.SetOwner("sado")
	toki.SetProperty("not_a_defined_property", "value")
	if err := graph.AddEntity(toki); err != nil {
		t.Fatalf("failed to addentity: %v", err)
	}

	result := e.Validate(graph, nil)
	if !hasFindingFor(result, "valid-property", "toki") {
		t.Error("expected valid-property warning for undefined property on toki")
	}
}

func TestValidPropertyRelationProperty(t *testing.T) {
	e := newTestEngine()
	graph := core.NewGraph()
	area := core.NewEntity("sado", kinds.Area, "Sado Island")
	if err := graph.AddEntity(area); err != nil {
		t.Fatalf("failed to addentity: %v", err)
	}
	waterBody := core.NewEntity("wb-01", kinds.WaterBody, "Source River")
	waterBody.SetOwner("sado")
	waterBody.SetProperty("water_type", "river")
	if err := graph.AddEntity(waterBody); err != nil {
		t.Fatalf("failed to addentity: %v", err)
	}
	onsen := core.NewEntity("onsen-01", kinds.HotSpring, "Hot Spring")
	onsen.SetOwner("sado")
	if err := graph.AddEntity(onsen); err != nil {
		t.Fatalf("failed to addentity: %v", err)
	}

	// depends_on defines dependency_type (enum) and critical (boolean).
	bad := core.NewDirectedRelation("rel-bad", types.DependsOn, "onsen-01", "wb-01")
	bad.SetProperty("critical", "not-a-bool")
	if err := graph.AddRelation(bad); err != nil {
		t.Fatalf("failed to addrelation: %v", err)
	}
	badEnum := core.NewDirectedRelation("rel-bad-enum", types.DependsOn, "onsen-01", "wb-01")
	badEnum.SetProperty("dependency_type", "magic") // not in enum
	if err := graph.AddRelation(badEnum); err != nil {
		t.Fatalf("failed to addrelation: %v", err)
	}
	good := core.NewDirectedRelation("rel-good", types.DependsOn, "onsen-01", "wb-01")
	good.SetProperty("dependency_type", "source")
	good.SetProperty("critical", true)
	if err := graph.AddRelation(good); err != nil {
		t.Fatalf("failed to addrelation: %v", err)
	}

	result := e.Validate(graph, nil)
	if !hasFindingFor(result, "valid-property", "rel-bad") {
		t.Error("expected valid-property warning for relation with non-boolean critical")
	}
	if !hasFindingFor(result, "valid-property", "rel-bad-enum") {
		t.Error("expected valid-property warning for relation with invalid dependency_type")
	}
	if hasFindingFor(result, "valid-property", "rel-good") {
		t.Error("relation with valid dependency properties should not have valid-property warning")
	}
}

func TestValidPropertyCoreKindEnumViolation(t *testing.T) {
	e := newTestEngine()
	graph := core.NewGraph()
	area := core.NewEntity("sado", kinds.Area, "Sado Island")
	if err := graph.AddEntity(area); err != nil {
		t.Fatalf("failed to addentity: %v", err)
	}
	bad := core.NewEntity("city-bad", kinds.Area, "Bad City")
	bad.SetOwner("sado")
	bad.SetProperty("area_type", "province") // not in enum
	if err := graph.AddEntity(bad); err != nil {
		t.Fatalf("failed to addentity: %v", err)
	}
	good := core.NewEntity("city-good", kinds.Area, "Good City")
	good.SetOwner("sado")
	good.SetProperty("area_type", "city")
	if err := graph.AddEntity(good); err != nil {
		t.Fatalf("failed to addentity: %v", err)
	}

	result := e.Validate(graph, nil)
	if !hasFindingFor(result, "valid-property", "city-bad") {
		t.Error("expected valid-property warning for city-bad with invalid area_type")
	}
	if hasFindingFor(result, "valid-property", "city-good") {
		t.Error("city-good with valid area_type should not have valid-property warning")
	}
}

func TestValidPropertyIntegerJSONFloat(t *testing.T) {
	e := newTestEngine()
	graph := core.NewGraph()
	area := core.NewEntity("sado", kinds.Area, "Sado Island")
	if err := graph.AddEntity(area); err != nil {
		t.Fatalf("failed to addentity: %v", err)
	}
	toki := core.NewEntity("toki", kinds.Species, "Crested Ibis")
	toki.SetOwner("sado")
	if err := graph.AddEntity(toki); err != nil {
		t.Fatalf("failed to addentity: %v", err)
	}
	population := core.NewEntity("pop-01", kinds.Population, "Toki Survey")
	population.SetOwner("toki")
	population.SetProperty("count", float64(500)) // JSON numbers decode to float64
	if err := graph.AddEntity(population); err != nil {
		t.Fatalf("failed to addentity: %v", err)
	}

	result := e.Validate(graph, nil)
	if hasFindingFor(result, "valid-property", "pop-01") {
		t.Error("integral float64 should not trigger valid-property warning")
	}
}

func TestValidPropertyIntegerNonIntegralFloat(t *testing.T) {
	e := newTestEngine()
	graph := core.NewGraph()
	area := core.NewEntity("sado", kinds.Area, "Sado Island")
	if err := graph.AddEntity(area); err != nil {
		t.Fatalf("failed to addentity: %v", err)
	}
	toki := core.NewEntity("toki", kinds.Species, "Crested Ibis")
	toki.SetOwner("sado")
	if err := graph.AddEntity(toki); err != nil {
		t.Fatalf("failed to addentity: %v", err)
	}
	population := core.NewEntity("pop-01", kinds.Population, "Toki Survey")
	population.SetOwner("toki")
	population.SetProperty("count", 500.5)
	if err := graph.AddEntity(population); err != nil {
		t.Fatalf("failed to addentity: %v", err)
	}

	result := e.Validate(graph, nil)
	if !hasFindingFor(result, "valid-property", "pop-01") {
		t.Error("expected valid-property warning for non-integral float count")
	}
}

func TestValidPropertyStringWithReference(t *testing.T) {
	e := newTestEngine()
	graph := core.NewGraph()
	area := core.NewEntity("sado", kinds.Area, "Sado Island")
	if err := graph.AddEntity(area); err != nil {
		t.Fatalf("failed to addentity: %v", err)
	}
	toki := core.NewEntity("toki", kinds.Species, "Crested Ibis")
	toki.SetOwner("sado")
	// scientific_name is a string property; the parser converts @-prefixed
	// strings to ReferenceValue, which must not trigger a type warning.
	toki.SetProperty("scientific_name", core.NewReferenceValue("@ca-01"))
	if err := graph.AddEntity(toki); err != nil {
		t.Fatalf("failed to addentity: %v", err)
	}

	result := e.Validate(graph, nil)
	if hasFindingFor(result, "valid-property", "toki") {
		t.Error("string property holding a ReferenceValue should not warn")
	}
}

// --- Root-Authorized Kind Hook ---

func TestAllowedRootKindsMethods(t *testing.T) {
	e := NewEngine(schema.CoreSchema())
	if e.IsAllowedRootKind(kinds.Area) {
		t.Error("area should not be an allowed root kind by default")
	}
	e.AddAllowedRootKind("prefecture.root")
	e.AddAllowedRootKind("prefecture.root") // idempotent
	e.AddAllowedRootKind("city.root")

	if !e.IsAllowedRootKind("prefecture.root") {
		t.Error("prefecture.root should be an allowed root kind")
	}
	if e.IsAllowedRootKind(kinds.Area) {
		t.Error("area should not be an allowed root kind by default")
	}
}

func TestMultipleRootsAllAllowedKinds(t *testing.T) {
	e := newTestEngine()
	e.AddAllowedRootKind("prefecture.root")

	graph := core.NewGraph()
	org1 := core.NewEntity("org-1", core.EntityKind("prefecture.root"), "Org 1")
	if err := graph.AddEntity(org1); err != nil {
		t.Fatalf("failed to addentity: %v", err)
	}
	org2 := core.NewEntity("org-2", core.EntityKind("prefecture.root"), "Org 2")
	if err := graph.AddEntity(org2); err != nil {
		t.Fatalf("failed to addentity: %v", err)
	}

	result := e.Validate(graph, nil)
	for _, f := range result.Findings {
		if f.RuleID == "root-entity" || f.RuleID == "single-owner" || f.RuleID == "ownership-tree" {
			t.Errorf("unexpected %s error for all-authorized roots: %s", f.RuleID, f.Message)
		}
	}
}

func TestMultipleRootsMixedKinds(t *testing.T) {
	e := newTestEngine()
	e.AddAllowedRootKind("prefecture.root")

	graph := core.NewGraph()
	org := core.NewEntity("org-1", core.EntityKind("prefecture.root"), "Org 1")
	if err := graph.AddEntity(org); err != nil {
		t.Fatalf("failed to addentity: %v", err)
	}
	area := core.NewEntity("sado", kinds.Area, "Sado Island")
	if err := graph.AddEntity(area); err != nil {
		t.Fatalf("failed to addentity: %v", err)
	}

	result := e.Validate(graph, nil)
	for _, f := range result.Findings {
		if f.RuleID == "root-entity" && f.Severity == SeverityError {
			return // expected
		}
	}
	t.Error("expected root-entity error when a non-authorized root coexists with an authorized root")
}

func TestOwnershipTreeForestWithAuthorizedRoots(t *testing.T) {
	e := newTestEngine()
	e.AddAllowedRootKind("prefecture.root")

	graph := core.NewGraph()
	org1 := core.NewEntity("org-1", core.EntityKind("prefecture.root"), "Org 1")
	if err := graph.AddEntity(org1); err != nil {
		t.Fatalf("failed to addentity: %v", err)
	}
	account1 := core.NewEntity("acct-1", core.EntityKind("city.root"), "Account 1")
	account1.SetOwner("org-1")
	if err := graph.AddEntity(account1); err != nil {
		t.Fatalf("failed to addentity: %v", err)
	}
	org2 := core.NewEntity("org-2", core.EntityKind("prefecture.root"), "Org 2")
	if err := graph.AddEntity(org2); err != nil {
		t.Fatalf("failed to addentity: %v", err)
	}
	account2 := core.NewEntity("acct-2", core.EntityKind("city.root"), "Account 2")
	account2.SetOwner("org-2")
	if err := graph.AddEntity(account2); err != nil {
		t.Fatalf("failed to addentity: %v", err)
	}

	result := e.Validate(graph, nil)
	for _, f := range result.Findings {
		if f.RuleID == "ownership-tree" {
			t.Errorf("unexpected ownership-tree error for connected forest: %s", f.Message)
		}
		if f.RuleID == "single-owner" {
			t.Errorf("unexpected single-owner error for authorized forest: %s", f.Message)
		}
		if f.RuleID == "root-entity" {
			t.Errorf("unexpected root-entity error for authorized forest: %s", f.Message)
		}
	}
}

func TestOwnershipTreeDisconnectedForest(t *testing.T) {
	e := newTestEngine()
	e.AddAllowedRootKind("prefecture.root")

	graph := core.NewGraph()
	org1 := core.NewEntity("org-1", core.EntityKind("prefecture.root"), "Org 1")
	if err := graph.AddEntity(org1); err != nil {
		t.Fatalf("failed to addentity: %v", err)
	}
	// acct-1 has an owner that does not exist (dangling), so it is unreachable
	// from any root and the forest is disconnected.
	orphan := core.NewEntity("orphan-1", core.EntityKind("city.root"), "Orphan")
	orphan.SetOwner("nonexistent")
	graph.ForceAddEntity(orphan)

	result := e.Validate(graph, nil)
	found := false
	for _, f := range result.Findings {
		if f.RuleID == "ownership-tree" && f.Severity == SeverityError {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected ownership-tree error for disconnected forest")
	}
}

// --- Niigata Domain Rules ---

// newPopulationFixture builds a graph with an optional owner for a population
// entity and returns the population entity.
func newPopulationFixture(ownerKind core.EntityKind, ownerID string) (*core.Graph, *core.Entity) {
	g := core.NewGraph()
	area := core.NewEntity("sado", kinds.Area, "Sado Island")
	_ = g.AddEntity(area)

	if ownerKind != "" {
		owner := core.NewEntity(ownerID, ownerKind, "Owner")
		if ownerID != "sado" {
			owner.SetOwner("sado")
		}
		_ = g.AddEntity(owner)
	}

	population := core.NewEntity("pop-01", kinds.Population, "Survey")
	population.SetProperty("count", 100)
	if ownerID != "" {
		population.SetOwner(ownerID)
	}
	_ = g.AddEntity(population)

	return g, population
}

func TestRulePopulationRequiresSpecies(t *testing.T) {
	tests := []struct {
		name        string
		setup       func() *core.Graph
		wantCount   int
		wantObject  string
	}{
		{
			name: "population owned by species is valid",
			setup: func() *core.Graph {
				g, _ := newPopulationFixture(kinds.Species, "toki")
				return g
			},
			wantCount: 0,
		},
		{
			name: "root population without owner is an error",
			setup: func() *core.Graph {
				g, _ := newPopulationFixture("", "")
				return g
			},
			wantCount:  1,
			wantObject: "pop-01",
		},
		{
			name: "population owned by area is an error",
			setup: func() *core.Graph {
				g, _ := newPopulationFixture(kinds.Area, "sado")
				return g
			},
			wantCount:  1,
			wantObject: "pop-01",
		},
		{
			name: "population owned by forest is an error",
			setup: func() *core.Graph {
				g, _ := newPopulationFixture(kinds.Forest, "f-01")
				return g
			},
			wantCount:  1,
			wantObject: "pop-01",
		},
		{
			name: "nonexistent owner is left to valid-owner",
			setup: func() *core.Graph {
				g, _ := newPopulationFixture("", "ghost")
				return g
			},
			wantCount: 0,
		},
		{
			name: "non-population entities are ignored",
			setup: func() *core.Graph {
				g := core.NewGraph()
				area := core.NewEntity("sado", kinds.Area, "Sado Island")
				_ = g.AddEntity(area)
				forest := core.NewEntity("f-01", kinds.Forest, "Forest")
				forest.SetOwner("sado")
				_ = g.AddEntity(forest)
				return g
			},
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := newTestEngine()
			result := e.Validate(tt.setup(), nil)

			findings := findingsByRule(result, "population-requires-species")
			if len(findings) != tt.wantCount {
				t.Fatalf("expected %d population-requires-species findings, got %d: %v",
					tt.wantCount, len(findings), findings)
			}
			for _, f := range findings {
				if f.Severity != SeverityError {
					t.Errorf("expected error severity, got %s: %s", f.Severity, f.Message)
				}
				if tt.wantObject != "" && f.ObjectID != tt.wantObject {
					t.Errorf("expected finding on %q, got %q", tt.wantObject, f.ObjectID)
				}
			}
		})
	}
}

// newAreaWithProps builds a graph whose only entity is an area with the given
// spec properties.
func newAreaWithProps(id string, props map[string]interface{}) *core.Graph {
	g := core.NewGraph()
	area := core.NewEntity(id, kinds.Area, "Area "+id)
	for k, v := range props {
		area.SetProperty(k, v)
	}
	_ = g.AddEntity(area)
	return g
}

func TestRulePositiveCount(t *testing.T) {
	tests := []struct {
		name       string
		count      interface{}
		hasCount   bool
		kind       core.EntityKind
		wantCount  int
		wantSubstr string
	}{
		{name: "zero count is valid", count: 0, hasCount: true, kind: kinds.Population, wantCount: 0},
		{name: "positive count is valid", count: 512, hasCount: true, kind: kinds.Population, wantCount: 0},
		{name: "negative count warns", count: -3, hasCount: true, kind: kinds.Population, wantCount: 1, wantSubstr: "-3"},
		{name: "negative float count warns", count: -2.5, hasCount: true, kind: kinds.Population, wantCount: 1, wantSubstr: "-2.5"},
		{name: "non-numeric count warns", count: "many", hasCount: true, kind: kinds.Population, wantCount: 1, wantSubstr: "non-numeric"},
		{name: "nil count warns as non-numeric", count: nil, hasCount: true, kind: kinds.Population, wantCount: 1, wantSubstr: "non-numeric"},
		{name: "missing count is left to valid-property", hasCount: false, kind: kinds.Population, wantCount: 0},
		{name: "negative value on non-population kind is ignored", count: -10, hasCount: true, kind: kinds.Forest, wantCount: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := core.NewGraph()
			area := core.NewEntity("sado", kinds.Area, "Sado Island")
			_ = g.AddEntity(area)

			holder := area
			if tt.kind != kinds.Area {
				holder = core.NewEntity("holder", tt.kind, "Holder")
				holder.SetOwner("sado")
				_ = g.AddEntity(holder)
			}
			if tt.hasCount {
				holder.SetProperty("count", tt.count)
			}

			e := newTestEngine()
			result := e.Validate(g, nil)

			findings := findingsByRule(result, "positive-count")
			if len(findings) != tt.wantCount {
				t.Fatalf("expected %d positive-count findings, got %d: %v",
					tt.wantCount, len(findings), findings)
			}
			for _, f := range findings {
				if f.Severity != SeverityWarning {
					t.Errorf("expected warning severity, got %s: %s", f.Severity, f.Message)
				}
				if tt.wantSubstr != "" && !strings.Contains(f.Message, tt.wantSubstr) {
					t.Errorf("expected message to contain %q, got: %s", tt.wantSubstr, f.Message)
				}
			}
		})
	}
}

func TestRuleValidNiigataCoordinates(t *testing.T) {
	tests := []struct {
		name       string
		props      map[string]interface{}
		kind       core.EntityKind
		wantCount  int
		wantSubstr string
	}{
		{name: "coordinates inside bounds are valid", kind: kinds.Area, props: map[string]interface{}{"latitude": 37.92, "longitude": 139.04}, wantCount: 0},
		{name: "latitude lower bound is inclusive", kind: kinds.Area, props: map[string]interface{}{"latitude": 36.6}, wantCount: 0},
		{name: "latitude upper bound is inclusive", kind: kinds.Area, props: map[string]interface{}{"latitude": 38.7}, wantCount: 0},
		{name: "longitude lower bound is inclusive", kind: kinds.Area, props: map[string]interface{}{"longitude": 137.9}, wantCount: 0},
		{name: "longitude upper bound is inclusive", kind: kinds.Area, props: map[string]interface{}{"longitude": 139.9}, wantCount: 0},
		{name: "latitude below bounds warns", kind: kinds.Area, props: map[string]interface{}{"latitude": 35.46}, wantCount: 1, wantSubstr: "latitude"},
		{name: "latitude above bounds warns", kind: kinds.Area, props: map[string]interface{}{"latitude": 39.72}, wantCount: 1, wantSubstr: "latitude"},
		{name: "longitude below bounds warns", kind: kinds.Area, props: map[string]interface{}{"longitude": 137.0}, wantCount: 1, wantSubstr: "longitude"},
		{name: "longitude above bounds warns", kind: kinds.Area, props: map[string]interface{}{"longitude": 140.47}, wantCount: 1, wantSubstr: "longitude"},
		{name: "both out of bounds produce two warnings", kind: kinds.Area, props: map[string]interface{}{"latitude": 10.0, "longitude": 10.0}, wantCount: 2},
		{name: "non-numeric latitude warns", kind: kinds.Area, props: map[string]interface{}{"latitude": "north"}, wantCount: 1, wantSubstr: "non-numeric"},
		{name: "non-numeric longitude warns", kind: kinds.Area, props: map[string]interface{}{"longitude": "west"}, wantCount: 1, wantSubstr: "non-numeric"},
		{name: "area without coordinates has no findings", kind: kinds.Area, props: nil, wantCount: 0},
		{name: "out-of-range coordinates on non-area kind are ignored", kind: kinds.TourismSpot, props: map[string]interface{}{"spot_type": "scenic", "latitude": 10.0, "longitude": 10.0}, wantCount: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := core.NewGraph()
			id := "target"
			entity := core.NewEntity(id, tt.kind, "Target")
			for k, v := range tt.props {
				entity.SetProperty(k, v)
			}
			_ = g.AddEntity(entity)

			e := newTestEngine()
			result := e.Validate(g, nil)

			findings := findingsByRule(result, "valid-niigata-coordinates")
			if len(findings) != tt.wantCount {
				t.Fatalf("expected %d valid-niigata-coordinates findings, got %d: %v",
					tt.wantCount, len(findings), findings)
			}
			for _, f := range findings {
				if f.Severity != SeverityWarning {
					t.Errorf("expected warning severity, got %s: %s", f.Severity, f.Message)
				}
				if f.ObjectID != id {
					t.Errorf("expected finding on %q, got %q", id, f.ObjectID)
				}
				if tt.wantSubstr != "" && !strings.Contains(f.Message, tt.wantSubstr) {
					t.Errorf("expected message to contain %q, got: %s", tt.wantSubstr, f.Message)
				}
			}
		})
	}
}

func TestValidPropertyGeoCoordinatesOnPlaceKinds(t *testing.T) {
	tests := []struct {
		name       string
		kind       core.EntityKind
		props      map[string]interface{}
		wantWarn   bool
		wantSubstr string
	}{
		{name: "tourism_spot with valid coordinates", kind: kinds.TourismSpot,
			props: map[string]interface{}{"spot_type": "scenic", "latitude": 37.99, "longitude": 138.32}, wantWarn: false},
		{name: "ground with valid coordinates", kind: kinds.Ground,
			props: map[string]interface{}{"soil_type": "loam", "latitude": 38.2, "longitude": 138.3}, wantWarn: false},
		{name: "water_body with valid coordinates", kind: kinds.WaterBody,
			props: map[string]interface{}{"water_type": "lake", "latitude": 38.05, "longitude": 138.44}, wantWarn: false},
		{name: "cultural_asset with valid coordinates", kind: kinds.CulturalAsset,
			props: map[string]interface{}{"asset_type": "historic_site", "latitude": 38.16, "longitude": 138.28}, wantWarn: false},
		{name: "latitude above global bounds warns", kind: kinds.TourismSpot,
			props: map[string]interface{}{"spot_type": "scenic", "latitude": 91.0}, wantWarn: true, wantSubstr: "latitude"},
		{name: "longitude below global bounds warns", kind: kinds.HotSpring,
			props: map[string]interface{}{"spring_quality": "chloride", "longitude": -181.0}, wantWarn: true, wantSubstr: "longitude"},
		{name: "non-numeric latitude warns", kind: kinds.Terrain,
			props: map[string]interface{}{"terrain_type": "mountain", "latitude": "north"}, wantWarn: true, wantSubstr: "latitude"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := core.NewGraph()
			entity := core.NewEntity("geo-target", tt.kind, "Geo Target")
			for k, v := range tt.props {
				entity.SetProperty(k, v)
			}
			if err := g.AddEntity(entity); err != nil {
				t.Fatalf("failed to add entity: %v", err)
			}

			e := newTestEngine()
			result := e.Validate(g, nil)

			findings := findingsByRule(result, "valid-property")
			if tt.wantWarn && len(findings) == 0 {
				t.Fatalf("expected a valid-property warning, got none")
			}
			if !tt.wantWarn && len(findings) > 0 {
				t.Fatalf("expected no valid-property warnings, got: %v", findings)
			}
			for _, f := range findings {
				if tt.wantSubstr != "" && !strings.Contains(f.Message, tt.wantSubstr) {
					t.Errorf("expected message to contain %q, got: %s", tt.wantSubstr, f.Message)
				}
			}
		})
	}
}
