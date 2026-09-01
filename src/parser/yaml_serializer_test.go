package parser

import (
	"strings"
	"testing"

	"github.com/bababa/Niigata_Real_IaC/src/core"
	"github.com/bababa/Niigata_Real_IaC/src/core/kinds"
	"github.com/bababa/Niigata_Real_IaC/src/core/types"
)

func TestSerializeBasicEntity(t *testing.T) {
	g := core.NewGraph()
	e := core.NewEntity("area-sado-island", kinds.Area, "Sado Island")
	if err := g.AddEntity(e); err != nil {
		t.Fatalf("failed to add entity: %v", err)
	}

	serializer := NewSerializer()
	data, err := serializer.Serialize(g)
	if err != nil {
		t.Fatalf("failed to serialize: %v", err)
	}

	// Parse back and verify
	parser := NewParser()
	g2, err := parser.Parse(data)
	if err != nil {
		t.Fatalf("failed to parse serialized data: %v", err)
	}

	e2, ok := g2.GetEntity("area-sado-island")
	if !ok {
		t.Fatal("entity area-sado-island not found in parsed data")
	}

	if e2.ID != e.ID {
		t.Errorf("expected ID %s, got %s", e.ID, e2.ID)
	}
	if e2.Kind != e.Kind {
		t.Errorf("expected kind %s, got %s", e.Kind, e2.Kind)
	}
	if e2.Name != e.Name {
		t.Errorf("expected name %s, got %s", e.Name, e2.Name)
	}
}

func TestSerializeEntityWithAllProperties(t *testing.T) {
	g := core.NewGraph()
	e := core.NewEntity("sapi-01", kinds.Species, "Toki (Crested Ibis)")
	e.Description = "Reintroduced crested ibis population on Sado Island"
	e.SetStatus(core.StatusActive)
	e.AddTag("endangered")
	e.AddTag("bird")
	e.SetLabel("habitat", "wetland")
	e.SetLabel("monitoring", "annual")
	e.Extensions = map[string]interface{}{"surveyor": "sado-city"}
	e.SetProperty("scientific_name", "Nipponia nippon")
	e.SetProperty("category", "bird")
	e.SetProperty("red_list_status", "endangered")
	e.SetProperty("recent_counts", []interface{}{
		map[string]interface{}{"count": 43, "survey_date": "2023-11-15", "method": "visual_count"},
		map[string]interface{}{"count": 39, "survey_date": "2022-11-15", "method": "visual_count"},
	})

	if err := g.AddEntity(e); err != nil {
		t.Fatalf("failed to add entity: %v", err)
	}

	serializer := NewSerializer()
	data, err := serializer.Serialize(g)
	if err != nil {
		t.Fatalf("failed to serialize: %v", err)
	}

	// Parse back and verify
	parser := NewParser()
	g2, err := parser.Parse(data)
	if err != nil {
		t.Fatalf("failed to parse serialized data: %v", err)
	}

	e2, ok := g2.GetEntity("sapi-01")
	if !ok {
		t.Fatal("entity sapi-01 not found in parsed data")
	}

	if e2.Description != e.Description {
		t.Errorf("expected description %s, got %s", e.Description, e2.Description)
	}
	if e2.Status != e.Status {
		t.Errorf("expected status %s, got %s", e.Status, e2.Status)
	}
	if len(e2.Tags) != 2 {
		t.Errorf("expected 2 tags, got %d", len(e2.Tags))
	}
	if e2.Labels["habitat"] != "wetland" {
		t.Errorf("expected label habitat=wetland, got %s", e2.Labels["habitat"])
	}
	if sciName, ok := e2.GetProperty("scientific_name"); !ok || sciName != "Nipponia nippon" {
		t.Errorf("expected property scientific_name=Nipponia nippon, got %v", sciName)
	}
}

func TestSerializeDirectedRelation(t *testing.T) {
	g := core.NewGraph()

	species := core.NewEntity("species-toki", kinds.Species, "Toki (Crested Ibis)")
	pond := core.NewEntity("wb-koike", kinds.WaterBody, "Koike Pond")

	if err := g.AddEntity(species); err != nil {
		t.Fatalf("failed to add entity: %v", err)
	}
	if err := g.AddEntity(pond); err != nil {
		t.Fatalf("failed to add entity: %v", err)
	}

	r := core.NewDirectedRelation("rel-inhabits-toki", types.Inhabits, "species-toki", "wb-koike")
	if err := g.AddRelation(r); err != nil {
		t.Fatalf("failed to add relation: %v", err)
	}

	serializer := NewSerializer()
	data, err := serializer.Serialize(g)
	if err != nil {
		t.Fatalf("failed to serialize: %v", err)
	}

	// Parse back and verify
	parser := NewParser()
	g2, err := parser.Parse(data)
	if err != nil {
		t.Fatalf("failed to parse serialized data: %v", err)
	}

	r2, ok := g2.GetRelation("rel-inhabits-toki")
	if !ok {
		t.Fatal("relation rel-inhabits-toki not found in parsed data")
	}

	if r2.Type != types.Inhabits {
		t.Errorf("expected type inhabits, got %s", r2.Type)
	}
	if r2.Direction != core.DirectionDirected {
		t.Errorf("expected direction directed, got %s", r2.Direction)
	}
	if r2.Source() != "species-toki" {
		t.Errorf("expected source species-toki, got %s", r2.Source())
	}
	if r2.Target() != "wb-koike" {
		t.Errorf("expected target wb-koike, got %s", r2.Target())
	}
}

func TestSerializeSymmetricRelation(t *testing.T) {
	g := core.NewGraph()

	// Create parent entities first
	forest := core.NewEntity("forest-sado-01", kinds.Forest, "Sado Forest 01")
	terrain := core.NewEntity("terrain-koshiba", kinds.Terrain, "Koshiba Coast Terrain")
	if err := g.AddEntity(forest); err != nil {
		t.Fatalf("failed to add entity: %v", err)
	}
	if err := g.AddEntity(terrain); err != nil {
		t.Fatalf("failed to add entity: %v", err)
	}

	gr1 := core.NewEntity("gr-loam-plot", kinds.Ground, "Loam Plot")
	gr1.SetOwner("forest-sado-01")
	gr2 := core.NewEntity("gr-sand-shore", kinds.Ground, "Sand Shore")
	gr2.SetOwner("terrain-koshiba")

	if err := g.AddEntity(gr1); err != nil {
		t.Fatalf("failed to add entity: %v", err)
	}
	if err := g.AddEntity(gr2); err != nil {
		t.Fatalf("failed to add entity: %v", err)
	}

	r := core.NewSymmetricRelation("rel-near-grounds", types.Near, []string{"gr-loam-plot", "gr-sand-shore"})
	if err := g.AddRelation(r); err != nil {
		t.Fatalf("failed to add relation: %v", err)
	}

	serializer := NewSerializer()
	data, err := serializer.Serialize(g)
	if err != nil {
		t.Fatalf("failed to serialize: %v", err)
	}

	// Parse back and verify
	parser := NewParser()
	g2, err := parser.Parse(data)
	if err != nil {
		t.Fatalf("failed to parse serialized data: %v", err)
	}

	r2, ok := g2.GetRelation("rel-near-grounds")
	if !ok {
		t.Fatal("relation rel-near-grounds not found in parsed data")
	}

	if r2.Type != types.Near {
		t.Errorf("expected type near, got %s", r2.Type)
	}
	if r2.Direction != core.DirectionSymmetric {
		t.Errorf("expected direction symmetric, got %s", r2.Direction)
	}
	if len(r2.Participants.List) != 2 {
		t.Errorf("expected 2 participants, got %d", len(r2.Participants.List))
	}
}

func TestRoundTrip(t *testing.T) {
	yaml := `
objects:
  - id: area-sado-island
    kind: area
    name: Sado Island
    attributes:
      status: active
      labels:
        region: chubu

  - id: forest-sado-beech
    kind: forest
    name: Sado Beech Forest
    attributes:
      owner: area-sado-island
      status: active
    spec:
      forest_type: beech
      area_ha: 120

  - id: species-toki
    kind: species
    name: Toki (Crested Ibis)
    attributes:
      owner: area-sado-island
      status: active
    spec:
      scientific_name: Nipponia nippon
      category: bird
      red_list_status: endangered

  - id: pop-toki-2024
    kind: population
    name: Toki Survey 2024
    attributes:
      owner: species-toki
      status: active
    spec:
      count: 43
      survey_date: "2024-11-15"
      survey_method: visual_count

  - id: rel-inhabits-species-area
    type: inhabits
    participants:
      source: species-toki
      target: area-sado-island
`

	// Parse original
	parser1 := NewParser()
	g1, err := parser1.Parse([]byte(yaml))
	if err != nil {
		t.Fatalf("failed to parse original: %v", err)
	}

	// Serialize
	serializer := NewSerializer()
	data, err := serializer.Serialize(g1)
	if err != nil {
		t.Fatalf("failed to serialize: %v", err)
	}

	// Parse serialized data
	parser2 := NewParser()
	g2, err := parser2.Parse(data)
	if err != nil {
		t.Fatalf("failed to parse serialized data: %v", err)
	}

	// Verify entities
	if g1.EntityCount() != g2.EntityCount() {
		t.Errorf("entity count mismatch: %d vs %d", g1.EntityCount(), g2.EntityCount())
	}

	for _, e1 := range g1.Entities() {
		e2, ok := g2.GetEntity(e1.ID)
		if !ok {
			t.Errorf("entity %s not found in round-trip", e1.ID)
			continue
		}
		if e1.ID != e2.ID {
			t.Errorf("entity ID mismatch: %s vs %s", e1.ID, e2.ID)
		}
		if e1.Kind != e2.Kind {
			t.Errorf("entity kind mismatch for %s: %s vs %s", e1.ID, e1.Kind, e2.Kind)
		}
		if e1.Name != e2.Name {
			t.Errorf("entity name mismatch for %s: %s vs %s", e1.ID, e1.Name, e2.Name)
		}
		if e1.Owner != e2.Owner {
			t.Errorf("entity owner mismatch for %s: %s vs %s", e1.ID, e1.Owner, e2.Owner)
		}
	}

	// Verify relations
	// Auto-generated relations from nesting may differ in count between g1 and g2
	// because the serializer nests entities based on ownership.
	// Verify that all explicit (non-auto) relations from g1 exist in g2.
	for _, r1 := range g1.Relations() {
		if val, ok := r1.GetLabel("auto_generated"); ok && val == "true" {
			continue
		}
		r2, ok := g2.GetRelation(r1.ID)
		if !ok {
			t.Errorf("explicit relation %s not found in round-trip", r1.ID)
			continue
		}
		if r1.Type != r2.Type {
			t.Errorf("relation type mismatch for %s: %s vs %s", r1.ID, r1.Type, r2.Type)
		}
		if r1.Direction != r2.Direction {
			t.Errorf("relation direction mismatch for %s: %s vs %s", r1.ID, r1.Direction, r2.Direction)
		}
		if r1.Source() != r2.Source() {
			t.Errorf("relation source mismatch for %s: %s vs %s", r1.ID, r1.Source(), r2.Source())
		}
		if r1.Target() != r2.Target() {
			t.Errorf("relation target mismatch for %s: %s vs %s", r1.ID, r1.Target(), r2.Target())
		}
	}

	// Verify all non-auto relations in g2 exist in g1 (or are auto-generated)
	for _, r2 := range g2.Relations() {
		if val, ok := r2.GetLabel("auto_generated"); ok && val == "true" {
			continue
		}
		if _, ok := g1.GetRelation(r2.ID); !ok {
			t.Errorf("relation %s in g2 not found in g1", r2.ID)
		}
	}
}

func TestRoundTripSpotNestedHotSpringsAndEvents(t *testing.T) {
	yaml := `
objects:
  - id: area-sado-island
    kind: area
    name: Sado Island
  - id: spot-shukunegi
    kind: tourism_spot
    name: Shukunegi Historic Village
    attributes:
      owner: area-sado-island
    spec:
      spot_type: historic
      annual_visitors: 150000
      events:
        - id: ev-lantern-festival
          name: Shukunegi Lantern Festival
          spec:
            season: autumn
            held_month: 10
            visitor_count: 8000
        - id: ev-morning-market
          name: Morning Market
      hot_springs:
        - id: onsen-shukunegi-01
          name: Shukunegi Onsen 01
          spec:
            spring_quality: chloride
            temperature_c: 72
`
	// Parse original
	parser1 := NewParser()
	g1, err := parser1.Parse([]byte(yaml))
	if err != nil {
		t.Fatalf("failed to parse original: %v", err)
	}

	// Serialize
	serializer := NewSerializer()
	data, err := serializer.Serialize(g1)
	if err != nil {
		t.Fatalf("failed to serialize: %v", err)
	}

	// Parse serialized data
	parser2 := NewParser()
	g2, err := parser2.Parse(data)
	if err != nil {
		t.Fatalf("failed to parse serialized data: %v", err)
	}

	if g1.EntityCount() != g2.EntityCount() {
		t.Errorf("entity count mismatch: %d vs %d", g1.EntityCount(), g2.EntityCount())
	}

	// Verify all entities survive round-trip with same owner
	for _, e1 := range g1.Entities() {
		e2, ok := g2.GetEntity(e1.ID)
		if !ok {
			t.Errorf("entity %s not found in round-trip", e1.ID)
			continue
		}
		if e1.Kind != e2.Kind {
			t.Errorf("kind mismatch for %s: %s vs %s", e1.ID, e1.Kind, e2.Kind)
		}
		if e1.Owner != e2.Owner {
			t.Errorf("owner mismatch for %s: %s vs %s", e1.ID, e1.Owner, e2.Owner)
		}
	}

	// Auto-generated belongs_to relations must be preserved
	for _, relID := range []string{
		"rel-auto-belongs_to-ev-lantern-festival-spot-shukunegi",
		"rel-auto-belongs_to-ev-morning-market-spot-shukunegi",
		"rel-auto-belongs_to-onsen-shukunegi-01-spot-shukunegi",
	} {
		r2, ok := g2.GetRelation(relID)
		if !ok {
			t.Errorf("auto relation %s not found after round-trip", relID)
			continue
		}
		if r2.Type != types.BelongsTo {
			t.Errorf("relation %s: expected belongs_to, got %s", relID, r2.Type)
		}
		if r2.Source() == "" || r2.Target() != "spot-shukunegi" {
			t.Errorf("relation %s: expected source member -> target spot-shukunegi, got %s -> %s",
				relID, r2.Source(), r2.Target())
		}
	}
}

func TestRoundTripForestTerrainGrounds(t *testing.T) {
	yaml := `
objects:
  - id: forest-sado-01
    kind: forest
    name: Sado National Forest 01
    spec:
      forest_type: mixed
      area_ha: 24
      grounds:
        - id: gr-loam
          name: Loam Plot
          spec:
            soil_type: loam
            slope_deg: 10
            stability: stable
        - id: gr-clay
          name: Clay Plot
          spec:
            soil_type: clay
            slope_deg: 5
            stability: watch

  - id: terrain-koshiba
    kind: terrain
    name: Koshiba Coast Terrain
    spec:
      terrain_type: coast
      grounds:
        - id: gr-0/0
          name: Shore Section 0/0
          spec:
            soil_type: sand
`
	// Parse original
	parser1 := NewParser()
	g1, err := parser1.Parse([]byte(yaml))
	if err != nil {
		t.Fatalf("failed to parse original: %v", err)
	}

	// Serialize
	serializer := NewSerializer()
	data, err := serializer.Serialize(g1)
	if err != nil {
		t.Fatalf("failed to serialize: %v", err)
	}

	// The serialized output must use the `grounds` nest key for forest/terrain,
	// not any other nest key from unrelated parent kinds.
	if !strings.Contains(string(data), "grounds:") {
		t.Errorf("serialized output missing 'grounds:' nest key:\n%s", data)
	}
	if strings.Contains(string(data), "water_bodies:") {
		t.Errorf("serialized output should not use 'water_bodies:' nest key:\n%s", data)
	}

	// Parse serialized data
	parser2 := NewParser()
	g2, err := parser2.Parse(data)
	if err != nil {
		t.Fatalf("failed to parse serialized data: %v", err)
	}

	if g1.EntityCount() != g2.EntityCount() {
		t.Errorf("entity count mismatch: %d vs %d", g1.EntityCount(), g2.EntityCount())
	}

	// Verify all entities survive round-trip with same owner and kind
	for _, e1 := range g1.Entities() {
		e2, ok := g2.GetEntity(e1.ID)
		if !ok {
			t.Errorf("entity %s not found in round-trip", e1.ID)
			continue
		}
		if e1.Kind != e2.Kind {
			t.Errorf("kind mismatch for %s: %s vs %s", e1.ID, e1.Kind, e2.Kind)
		}
		if e1.Owner != e2.Owner {
			t.Errorf("owner mismatch for %s: %s vs %s", e1.ID, e1.Owner, e2.Owner)
		}
	}

	// area_ha property must survive round-trip
	forest2, ok := g2.GetEntity("forest-sado-01")
	if !ok {
		t.Fatal("entity forest-sado-01 not found after round-trip")
	}
	areaHa, ok := forest2.GetProperty("area_ha")
	if !ok || areaHa != 24 {
		t.Errorf("expected area_ha 24 after round-trip, got %v", areaHa)
	}

	// Auto-generated belongs_to relations must be preserved
	for _, relID := range []string{
		"rel-auto-belongs_to-gr-loam-forest-sado-01",
		"rel-auto-belongs_to-gr-clay-forest-sado-01",
		"rel-auto-belongs_to-gr-0/0-terrain-koshiba",
	} {
		r2, ok := g2.GetRelation(relID)
		if !ok {
			t.Errorf("auto relation %s not found after round-trip", relID)
			continue
		}
		if r2.Type != types.BelongsTo {
			t.Errorf("relation %s: expected belongs_to, got %s", relID, r2.Type)
		}
	}
}

func TestSerializeFile(t *testing.T) {
	g := core.NewGraph()
	e := core.NewEntity("test-entity", kinds.Area, "Test Area")
	if err := g.AddEntity(e); err != nil {
		t.Fatalf("failed to add entity: %v", err)
	}

	serializer := NewSerializer()
	err := serializer.SerializeFile(g, "/tmp/test_serialize.yaml")
	if err != nil {
		t.Fatalf("failed to serialize to file: %v", err)
	}

	// Parse back and verify
	parser := NewParser()
	g2, err := parser.ParseFile("/tmp/test_serialize.yaml")
	if err != nil {
		t.Fatalf("failed to parse file: %v", err)
	}

	e2, ok := g2.GetEntity("test-entity")
	if !ok {
		t.Fatal("entity test-entity not found in file")
	}

	if e2.ID != e.ID {
		t.Errorf("expected ID %s, got %s", e.ID, e2.ID)
	}
}

func TestRoundTripPropertyReference(t *testing.T) {
	input := `
objects:
  - id: wb-shinano
    kind: water_body
    name: Shinano River
    spec:
      water_type: river
      length_km: 367

  - id: pond-100
    kind: water_body
    name: Irrigation Pond 100
    spec:
      water_type: pond
      feeder_channel: "@wb-shinano"
`
	parser := NewParser()
	g, err := parser.Parse([]byte(input))
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	serializer := NewSerializer()
	data, err := serializer.Serialize(g)
	if err != nil {
		t.Fatalf("failed to serialize: %v", err)
	}

	// Verify @ prefix is preserved in output (YAML may use single or double quotes)
	output := string(data)
	if !strings.Contains(output, "@wb-shinano") {
		t.Errorf("serialized output should contain @wb-shinano reference, got:\n%s", output)
	}

	// Parse back and verify round-trip
	g2, err := parser.Parse(data)
	if err != nil {
		t.Fatalf("failed to re-parse: %v", err)
	}

	pond, ok := g2.GetEntity("pond-100")
	if !ok {
		t.Fatal("expected entity pond-100")
	}

	v, ok := pond.GetProperty("feeder_channel")
	if !ok {
		t.Fatal("expected property feeder_channel")
	}
	ref, ok := v.(core.ReferenceValue)
	if !ok {
		t.Fatalf("expected ReferenceValue after round-trip, got %T", v)
	}
	if ref.RefTargetID() != "wb-shinano" {
		t.Errorf("expected reference target wb-shinano, got %s", ref.RefTargetID())
	}
}

func TestRoundTripHotSpringWaterBodyReference(t *testing.T) {
	input := `
objects:
  - id: wb-kamo
    kind: water_body
    name: Kamo River
    spec:
      water_type: river

  - id: onsen-01
    kind: hot_spring
    name: Onsen 01
    spec:
      water_source: "@wb-kamo"
      source_temperatures:
        - 78.5
`
	parser := NewParser()
	g, err := parser.Parse([]byte(input))
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	serializer := NewSerializer()
	data, err := serializer.Serialize(g)
	if err != nil {
		t.Fatalf("failed to serialize: %v", err)
	}

	output := string(data)
	if !strings.Contains(output, "@wb-kamo") {
		t.Errorf("serialized output should contain @wb-kamo reference, got:\n%s", output)
	}

	g2, err := parser.Parse(data)
	if err != nil {
		t.Fatalf("failed to re-parse: %v", err)
	}

	onsen, ok := g2.GetEntity("onsen-01")
	if !ok {
		t.Fatal("expected entity onsen-01")
	}

	v, ok := onsen.GetProperty("water_source")
	if !ok {
		t.Fatal("expected property water_source")
	}
	ref, ok := v.(core.ReferenceValue)
	if !ok {
		t.Fatalf("expected ReferenceValue after round-trip, got %T", v)
	}
	if ref.RefTargetID() != "wb-kamo" {
		t.Errorf("expected reference target wb-kamo, got %s", ref.RefTargetID())
	}
}
