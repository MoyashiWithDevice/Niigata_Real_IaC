package parser

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bababa/Niigata_Real_IaC/src/core"
	"github.com/bababa/Niigata_Real_IaC/src/core/kinds"
	"github.com/bababa/Niigata_Real_IaC/src/core/types"
	"github.com/bababa/Niigata_Real_IaC/src/schema"
	"github.com/bababa/Niigata_Real_IaC/src/validation"
)

func TestParseBasicEntity(t *testing.T) {
	yaml := `
objects:
  - id: area-niigata-prefecture
    kind: area
    name: Niigata Prefecture
`

	parser := NewParser()
	g, err := parser.Parse([]byte(yaml))
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	if g.EntityCount() != 1 {
		t.Fatalf("expected 1 entity, got %d", g.EntityCount())
	}

	e, ok := g.GetEntity("area-niigata-prefecture")
	if !ok {
		t.Fatal("entity area-niigata-prefecture not found")
	}

	if e.ID != "area-niigata-prefecture" {
		t.Errorf("expected ID area-niigata-prefecture, got %s", e.ID)
	}
	if e.Kind != kinds.Area {
		t.Errorf("expected kind area, got %s", e.Kind)
	}
	if e.Name != "Niigata Prefecture" {
		t.Errorf("expected name 'Niigata Prefecture', got %s", e.Name)
	}
}

func TestParseEntityWithAllProperties(t *testing.T) {
	yaml := `
objects:
  - id: spot-yahiko-shrine
    kind: tourism_spot
    name: Yahiko Shrine
    attributes:
      description: "Historic shrine at the foot of Mount Yahiko"
      status: active
      tags:
        - culture
        - landmark
      labels:
        area: chuetsu
        category: heritage
      extensions:
        operator: yahiko-jinja
        founded_year: 753
    spec:
      spot_type: shrine_temple
      annual_visitors: 1200000
`

	parser := NewParser()
	g, err := parser.Parse([]byte(yaml))
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	e, ok := g.GetEntity("spot-yahiko-shrine")
	if !ok {
		t.Fatal("entity spot-yahiko-shrine not found")
	}

	if e.Description != "Historic shrine at the foot of Mount Yahiko" {
		t.Errorf("expected description 'Historic shrine at the foot of Mount Yahiko', got %s", e.Description)
	}
	if e.Status != core.StatusActive {
		t.Errorf("expected status active, got %s", e.Status)
	}
	if len(e.Tags) != 2 {
		t.Errorf("expected 2 tags, got %d", len(e.Tags))
	}
	if e.Labels["area"] != "chuetsu" {
		t.Errorf("expected label area=chuetsu, got %s", e.Labels["area"])
	}
	if e.Extensions["operator"] != "yahiko-jinja" {
		t.Errorf("expected extensions operator=yahiko-jinja, got %v", e.Extensions["operator"])
	}

	// Check kind-specific properties
	if spotType, ok := e.GetProperty("spot_type"); !ok || spotType != "shrine_temple" {
		t.Errorf("expected property spot_type=shrine_temple, got %v", spotType)
	}
	if visitors, ok := e.GetProperty("annual_visitors"); !ok || visitors != 1200000 {
		t.Errorf("expected property annual_visitors=1200000, got %v", visitors)
	}
}

func TestParseEntityWithOwnership(t *testing.T) {
	yaml := `
objects:
  - id: area-sado-island
    kind: area
    name: Sado Island

  - id: fst-beech-hill
    kind: forest
    name: Beech Hill Forest
    attributes:
      owner: area-sado-island

  - id: grnd-valley-floor
    kind: ground
    name: Valley Floor Ground
    attributes:
      owner: fst-beech-hill
`

	parser := NewParser()
	g, err := parser.Parse([]byte(yaml))
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	if g.EntityCount() != 3 {
		t.Fatalf("expected 3 entities, got %d", g.EntityCount())
	}

	// Check ownership hierarchy
	area, _ := g.GetEntity("area-sado-island")
	if !area.IsRoot() {
		t.Error("area should be root")
	}

	forest, _ := g.GetEntity("fst-beech-hill")
	if forest.Owner != "area-sado-island" {
		t.Errorf("expected forest owner area-sado-island, got %s", forest.Owner)
	}

	ground, _ := g.GetEntity("grnd-valley-floor")
	if ground.Owner != "fst-beech-hill" {
		t.Errorf("expected ground owner fst-beech-hill, got %s", ground.Owner)
	}

	// Check paths were built
	if area.Path() != "/area-sado-island" {
		t.Errorf("expected area path /area-sado-island, got %s", area.Path())
	}
	if forest.Path() != "/area-sado-island/fst-beech-hill" {
		t.Errorf("expected forest path /area-sado-island/fst-beech-hill, got %s", forest.Path())
	}
	if ground.Path() != "/area-sado-island/fst-beech-hill/grnd-valley-floor" {
		t.Errorf("expected ground path /area-sado-island/fst-beech-hill/grnd-valley-floor, got %s", ground.Path())
	}
}

func TestParseDirectedRelation(t *testing.T) {
	yaml := `
objects:
  - id: spx-ayu
    kind: species
    name: Ayu Sweetfish

  - id: wb-shinano-river
    kind: water_body
    name: Shinano River

  - id: rel-inhabits-ayu-river
    type: inhabits
    participants:
      source: spx-ayu
      target: wb-shinano-river
`

	parser := NewParser()
	g, err := parser.Parse([]byte(yaml))
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	if g.RelationCount() != 1 {
		t.Fatalf("expected 1 relation, got %d", g.RelationCount())
	}

	r, ok := g.GetRelation("rel-inhabits-ayu-river")
	if !ok {
		t.Fatal("relation rel-inhabits-ayu-river not found")
	}

	if r.Type != types.Inhabits {
		t.Errorf("expected type inhabits, got %s", r.Type)
	}
	if r.Direction != core.DirectionDirected {
		t.Errorf("expected direction directed, got %s", r.Direction)
	}
	if r.Source() != "spx-ayu" {
		t.Errorf("expected source spx-ayu, got %s", r.Source())
	}
	if r.Target() != "wb-shinano-river" {
		t.Errorf("expected target wb-shinano-river, got %s", r.Target())
	}
}

func TestParseSymmetricRelation(t *testing.T) {
	yaml := `
objects:
  - id: spot-tsukioka
    kind: tourism_spot
    name: Tsukioka Onsen

  - id: spot-hyoso
    kind: tourism_spot
    name: Hyoso Hot Spring

  - id: rel-near-spots
    type: near
    participants:
      - spot-tsukioka
      - spot-hyoso
`

	parser := NewParser()
	g, err := parser.Parse([]byte(yaml))
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	r, ok := g.GetRelation("rel-near-spots")
	if !ok {
		t.Fatal("relation rel-near-spots not found")
	}

	if r.Type != types.Near {
		t.Errorf("expected type near, got %s", r.Type)
	}
	if r.Direction != core.DirectionSymmetric {
		t.Errorf("expected direction symmetric, got %s", r.Direction)
	}
	if len(r.Participants.List) != 2 {
		t.Errorf("expected 2 participants, got %d", len(r.Participants.List))
	}
}

func TestParseRelationWithAllProperties(t *testing.T) {
	yaml := `
objects:
  - id: spring-tsukioka
    kind: hot_spring
    name: Tsukioka Spring

  - id: wb-sabiura-river
    kind: water_body
    name: Sabiura River

  - id: rel-depends-source
    type: depends_on
    participants:
      source: spring-tsukioka
      target: wb-sabiura-river
    attributes:
      description: "Hot spring draws from the river aquifer"
      status: active
      tags:
        - water_source
      labels:
        source_type: hot_spring
        target_type: water_body
`

	parser := NewParser()
	g, err := parser.Parse([]byte(yaml))
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	r, ok := g.GetRelation("rel-depends-source")
	if !ok {
		t.Fatal("relation rel-depends-source not found")
	}

	if r.Description != "Hot spring draws from the river aquifer" {
		t.Errorf("expected description 'Hot spring draws from the river aquifer', got %s", r.Description)
	}
	if r.Status != core.StatusActive {
		t.Errorf("expected status active, got %s", r.Status)
	}
	if len(r.Tags) != 1 || r.Tags[0] != "water_source" {
		t.Errorf("expected tags [water_source], got %v", r.Tags)
	}
	if r.Labels["source_type"] != "hot_spring" {
		t.Errorf("expected label source_type=hot_spring, got %s", r.Labels["source_type"])
	}
}

func TestParseCompleteExample(t *testing.T) {
	yaml := `
objects:
  # Areas
  - id: area-sado
    kind: area
    name: Sado Island
    attributes:
      status: active
      labels:
        area_type: island

  # Tourism Spots
  - id: spot-tsukioka
    kind: tourism_spot
    name: Tsukioka Onsen
    attributes:
      owner: area-sado
      status: active
      labels:
        zone: onsen
    spec:
      spot_type: scenic
      annual_visitors: 450000

  # Hot Springs
  - id: spring-tsukioka-main
    kind: hot_spring
    name: Tsukioka Main Spring
    attributes:
      owner: spot-tsukioka
      status: active
    spec:
      spring_quality: sulfur
      temperature_c: 96.0
      source_count: 3

  # Events
  - id: evt-yukake
    kind: event
    name: Yukake Festival
    attributes:
      owner: spring-tsukioka-main
      status: active
    spec:
      season: summer
      held_month: 8
      visitor_count: 12000

  # Water Bodies
  - id: wb-sabiura
    kind: water_body
    name: Sabiura River
    attributes:
      status: active
    spec:
      water_type: river
      length_km: 52.4

  # Dependency Relations
  - id: rel-depends-spring-water
    type: depends_on
    participants:
      source: spring-tsukioka-main
      target: wb-sabiura

  - id: rel-depends-event-spring
    type: depends_on
    participants:
      source: evt-yukake
      target: spring-tsukioka-main
`

	parser := NewParser()
	g, err := parser.Parse([]byte(yaml))
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	if g.EntityCount() != 5 {
		t.Fatalf("expected 5 entities, got %d", g.EntityCount())
	}
	if g.RelationCount() != 2 {
		t.Fatalf("expected 2 relations, got %d", g.RelationCount())
	}

	// Verify ownership hierarchy
	spring, _ := g.GetEntity("spring-tsukioka-main")
	if spring.Owner != "spot-tsukioka" {
		t.Errorf("expected spring owner spot-tsukioka, got %s", spring.Owner)
	}

	event, _ := g.GetEntity("evt-yukake")
	if event.Owner != "spring-tsukioka-main" {
		t.Errorf("expected event owner spring-tsukioka-main, got %s", event.Owner)
	}
}

func TestParseInvalidEntity(t *testing.T) {
	yaml := `
objects:
  - id: test-entity
    kind: species
`

	parser := NewParser()
	_, err := parser.Parse([]byte(yaml))
	if err == nil {
		t.Fatal("expected error for missing name field")
	}
}

func TestParseInvalidRelation(t *testing.T) {
	yaml := `
objects:
  - id: spx-01
    kind: species
    name: Species 01

  - id: test-relation
    type: inhabits
`

	parser := NewParser()
	_, err := parser.Parse([]byte(yaml))
	if err == nil {
		t.Fatal("expected error for missing participants")
	}
}

func TestParseInvalidYAML(t *testing.T) {
	yaml := `
objects:
  - id: test
    kind: area
    name: Test
    invalid: [unclosed
`

	parser := NewParser()
	_, err := parser.Parse([]byte(yaml))
	if err == nil {
		t.Fatal("expected error for invalid YAML")
	}
}

func TestParseEmptyDocument(t *testing.T) {
	yaml := `
objects: []
`

	parser := NewParser()
	g, err := parser.Parse([]byte(yaml))
	if err != nil {
		t.Fatalf("failed to parse empty document: %v", err)
	}

	if g.EntityCount() != 0 {
		t.Errorf("expected 0 entities, got %d", g.EntityCount())
	}
	if g.RelationCount() != 0 {
		t.Errorf("expected 0 relations, got %d", g.RelationCount())
	}
}

func TestParseNestedEntities(t *testing.T) {
	yaml := `
objects:
  - id: area-sado
    kind: area
    name: Sado Island
    spec:
      species:
        - id: spx-ibis
          name: Crested Ibis
          spec:
            scientific_name: Nipponia nippon
            red_list_status: endangered
            populations:
              - id: pop-wild
                name: Wild Population
                spec:
                  count: 190
                  survey_method: drone
      tourism_spots:
        - id: spot-shukunegi
          name: Shukunegi Village
          spec:
            spot_type: historic
`

	parser := NewParser()
	g, err := parser.Parse([]byte(yaml))
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	// 1 area + 1 species + 1 population + 1 tourism_spot = 4
	if g.EntityCount() != 4 {
		t.Fatalf("expected 4 entities, got %d", g.EntityCount())
	}

	// Check parent area
	area, ok := g.GetEntity("area-sado")
	if !ok {
		t.Fatal("entity area-sado not found")
	}
	if area.Owner != "" {
		t.Errorf("area should be root, got owner %s", area.Owner)
	}

	// Check nested species
	species, ok := g.GetEntity("spx-ibis")
	if !ok {
		t.Fatal("entity spx-ibis not found")
	}
	if species.Owner != "area-sado" {
		t.Errorf("expected species owner area-sado, got %s", species.Owner)
	}
	if species.Kind != kinds.Species {
		t.Errorf("expected kind species, got %s", species.Kind)
	}

	// Check nested population - owned by species, not area
	pop, ok := g.GetEntity("pop-wild")
	if !ok {
		t.Fatal("entity pop-wild not found")
	}
	if pop.Owner != "spx-ibis" {
		t.Errorf("expected population owner spx-ibis, got %s", pop.Owner)
	}
	if pop.Kind != kinds.Population {
		t.Errorf("expected kind population, got %s", pop.Kind)
	}

	// Check nested tourism spot
	spot, ok := g.GetEntity("spot-shukunegi")
	if !ok {
		t.Fatal("entity spot-shukunegi not found")
	}
	if spot.Owner != "area-sado" {
		t.Errorf("expected tourism spot owner area-sado, got %s", spot.Owner)
	}

	// Check population properties
	count, ok := pop.GetProperty("count")
	if !ok || count != 190 {
		t.Errorf("expected count 190, got %v", count)
	}
}

func TestParseNestedChildrenExplicitIDs(t *testing.T) {
	yaml := `
objects:
  - id: fst-cedar-stand
    kind: forest
    name: Cedar Stand
    spec:
      grounds:
        - id: grnd-upper
          name: Upper Terrace
          spec:
            soil_type: volcanic_ash
            slope_deg: 15
        - id: grnd-lower
          name: Lower Terrace
          spec:
            soil_type: loam
            slope_deg: 5
`

	parser := NewParser()
	g, err := parser.Parse([]byte(yaml))
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	// 1 forest + 2 grounds = 3
	if g.EntityCount() != 3 {
		t.Fatalf("expected 3 entities, got %d", g.EntityCount())
	}

	// Check explicit IDs
	g1, ok := g.GetEntity("grnd-upper")
	if !ok {
		t.Fatal("entity grnd-upper not found")
	}
	if g1.Owner != "fst-cedar-stand" {
		t.Errorf("expected owner fst-cedar-stand, got %s", g1.Owner)
	}
	if g1.Name != "Upper Terrace" {
		t.Errorf("expected name Upper Terrace, got %s", g1.Name)
	}

	g2, ok := g.GetEntity("grnd-lower")
	if !ok {
		t.Fatal("entity grnd-lower not found")
	}
	if g2.Owner != "fst-cedar-stand" {
		t.Errorf("expected owner fst-cedar-stand, got %s", g2.Owner)
	}
}

func TestParseNestedPopulationKindInference(t *testing.T) {
	yaml := `
objects:
  - id: spx-japanese-cedar
    kind: species
    name: Japanese Cedar
    spec:
      category: plant
      survey_policy: annual
      populations:
        - id: pop-2024
          name: 2024 Survey
          spec:
            count: 4200
            survey_date: "2024-10-01"
        - id: pop-2023
          name: 2023 Survey
          spec:
            count: 4100
            survey_date: "2023-10-01"
`

	parser := NewParser()
	g, err := parser.Parse([]byte(yaml))
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	// 1 species + 2 populations = 3
	if g.EntityCount() != 3 {
		t.Fatalf("expected 3 entities, got %d", g.EntityCount())
	}

	// Check first population
	p1, ok := g.GetEntity("pop-2024")
	if !ok {
		t.Fatal("entity pop-2024 not found")
	}
	if p1.Owner != "spx-japanese-cedar" {
		t.Errorf("expected owner spx-japanese-cedar, got %s", p1.Owner)
	}
	if p1.Kind != kinds.Population {
		t.Errorf("expected kind population, got %s", p1.Kind)
	}
}

func TestParseFlatAndNestedMixed(t *testing.T) {
	yaml := `
objects:
  # Flat definition
  - id: fst-cedar
    kind: forest
    name: Cedar Forest
    attributes:
      owner: area-myoko

  - id: area-myoko
    kind: area
    name: Myoko Area

  # Nested definition
  - id: spot-myoko
    kind: tourism_spot
    name: Myoko Ski Resort
    attributes:
      owner: area-myoko
    spec:
      spot_type: ski_resort
      events:
        - id: evt-ski-opening
          name: Ski Season Opening
`

	parser := NewParser()
	g, err := parser.Parse([]byte(yaml))
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	// area + forest + spot + event = 4
	if g.EntityCount() != 4 {
		t.Fatalf("expected 4 entities, got %d", g.EntityCount())
	}

	// Check flat ownership
	forest, ok := g.GetEntity("fst-cedar")
	if !ok {
		t.Fatal("forest not found")
	}
	if forest.Owner != "area-myoko" {
		t.Errorf("expected forest owner area-myoko, got %s", forest.Owner)
	}

	// Check nested ownership
	event, ok := g.GetEntity("evt-ski-opening")
	if !ok {
		t.Fatal("event not found")
	}
	if event.Owner != "spot-myoko" {
		t.Errorf("expected event owner spot-myoko, got %s", event.Owner)
	}
}

func TestParseNestedSpeciesWithAndWithoutID(t *testing.T) {
	yaml := `
objects:
  - id: area-sado
    kind: area
    name: Sado Island
    spec:
      species:
        - id: spx-ibis
          name: Crested Ibis
          spec:
            scientific_name: Nipponia nippon
            populations:
              - id: pop-wild
                name: Wild Population
                spec:
                  count: 190
        - name: Sika Deer
          spec:
            category: mammal
            populations:
              - name: 2025 Survey
                spec:
                  count: 3200
`

	parser := NewParser()
	g, err := parser.Parse([]byte(yaml))
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	// 1 area + 2 species + 2 populations = 5
	if g.EntityCount() != 5 {
		t.Fatalf("expected 5 entities, got %d", g.EntityCount())
	}

	// Both species should be owned by the area
	for _, id := range []string{"spx-ibis", "_area-sado-species"} {
		e, ok := g.GetEntity(id)
		if !ok {
			t.Fatalf("entity %s not found", id)
		}
		if e.Owner != "area-sado" {
			t.Errorf("entity %s expected owner area-sado, got %s", id, e.Owner)
		}
	}

	// pop-wild should be owned by spx-ibis
	popWild, ok := g.GetEntity("pop-wild")
	if !ok {
		t.Fatal("entity pop-wild not found")
	}
	if popWild.Owner != "spx-ibis" {
		t.Errorf("expected pop-wild owner spx-ibis, got %s", popWild.Owner)
	}

	// The auto-ID'd population should be owned by the auto-ID'd species
	autoPop, ok := g.GetEntity("_area-sado-species-population")
	if !ok {
		t.Fatal("entity _area-sado-species-population not found")
	}
	if autoPop.Owner != "_area-sado-species" {
		t.Errorf("expected owner _area-sado-species, got %s", autoPop.Owner)
	}
}

func TestSchemaNestingDefinitions(t *testing.T) {
	s := schema.CoreSchema()

	// Global nesting defs should be empty in the Niigata schema
	if len(s.NestingDefs) != 0 {
		t.Fatalf("expected 0 global nesting defs, got %d", len(s.NestingDefs))
	}

	// Area: 0 global + 7 per-kind (grounds, terrains, water_bodies, forests, tourism_spots, species, cultural_assets) = 7
	areaNesting := s.GetNestingDefs(kinds.Area)
	if len(areaNesting) != 7 {
		t.Fatalf("expected 7 nesting defs for area, got %d", len(areaNesting))
	}

	// Ground: 0 global + 0 per-kind = 0
	groundNesting := s.GetNestingDefs(kinds.Ground)
	if len(groundNesting) != 0 {
		t.Fatalf("expected 0 nesting defs for ground, got %d", len(groundNesting))
	}

	// Species: 0 global + 1 per-kind (populations) = 1
	speciesNesting := s.GetNestingDefs(kinds.Species)
	if len(speciesNesting) != 1 {
		t.Fatalf("expected 1 nesting def for species, got %d", len(speciesNesting))
	}

	// Find tourism_spots nesting under area
	nd, ok := s.FindNestingByNestKey(kinds.Area, "tourism_spots")
	if !ok {
		t.Fatal("tourism_spots nesting not found for area")
	}
	if nd.ChildKind != kinds.TourismSpot {
		t.Errorf("expected child kind tourism_spot, got %s", nd.ChildKind)
	}
}

func TestTourismSpotNestingDefinitions(t *testing.T) {
	s := schema.CoreSchema()

	// TourismSpot: 2 per-kind (hot_springs, events)
	spotNesting := s.GetNestingDefs(kinds.TourismSpot)
	if len(spotNesting) != 2 {
		t.Fatalf("expected 2 nesting defs for tourism_spot, got %d", len(spotNesting))
	}

	nd, ok := s.FindNestingByNestKey(kinds.TourismSpot, "hot_springs")
	if !ok {
		t.Fatal("hot_springs nesting not found for tourism_spot")
	}
	if nd.ChildKind != kinds.HotSpring {
		t.Errorf("expected child kind hot_spring, got %s", nd.ChildKind)
	}
}

func TestParseNestedHotSpringEvents(t *testing.T) {
	yaml := `
objects:
  - id: spot-tsukioka
    kind: tourism_spot
    name: Tsukioka Onsen
    spec:
      hot_springs:
        - id: spring-grand
          kind: hot_spring
          name: Grand Bath House
          attributes:
            status: active
          spec:
            spring_quality: sulfur
            temperature_c: 96.5
            events:
              - id: evt-morning-bath
                kind: event
                name: Morning Bath Session
                attributes:
                  status: active
                spec:
                  season: all
              - id: evt-evening-bath
                kind: event
                name: Evening Bath Session
                attributes:
                  status: standby
                spec:
                  season: winter
`

	parser := NewParser()
	g, err := parser.Parse([]byte(yaml))
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	// 1 tourism spot + 1 hot spring + 2 events = 4
	if g.EntityCount() != 4 {
		t.Fatalf("expected 4 entities, got %d", g.EntityCount())
	}

	// Check parent tourism spot
	spot, ok := g.GetEntity("spot-tsukioka")
	if !ok {
		t.Fatal("entity spot-tsukioka not found")
	}
	if spot.Owner != "" {
		t.Errorf("tourism spot should be root, got owner %s", spot.Owner)
	}

	// Check hot spring
	spring, ok := g.GetEntity("spring-grand")
	if !ok {
		t.Fatal("entity spring-grand not found")
	}
	if spring.Owner != "spot-tsukioka" {
		t.Errorf("expected owner spot-tsukioka, got %s", spring.Owner)
	}
	if spring.Kind != kinds.HotSpring {
		t.Errorf("expected kind hot_spring, got %s", spring.Kind)
	}
	if spring.Status != core.StatusActive {
		t.Errorf("expected status active, got %s", spring.Status)
	}
	if quality, ok := spring.GetProperty("spring_quality"); !ok || quality != "sulfur" {
		t.Errorf("expected spring_quality sulfur, got %v", quality)
	}

	// Check morning event
	morning, ok := g.GetEntity("evt-morning-bath")
	if !ok {
		t.Fatal("entity evt-morning-bath not found")
	}
	if morning.Owner != "spring-grand" {
		t.Errorf("expected owner spring-grand, got %s", morning.Owner)
	}
	if morning.Kind != kinds.Event {
		t.Errorf("expected kind event, got %s", morning.Kind)
	}
	if morning.Status != core.StatusActive {
		t.Errorf("expected status active, got %s", morning.Status)
	}

	// Check evening event
	evening, ok := g.GetEntity("evt-evening-bath")
	if !ok {
		t.Fatal("entity evt-evening-bath not found")
	}
	if evening.Owner != "spring-grand" {
		t.Errorf("expected owner spring-grand, got %s", evening.Owner)
	}
	if evening.Status != core.StatusStandby {
		t.Errorf("expected status standby, got %s", evening.Status)
	}
}

func TestParseNestedTerrainGrounds(t *testing.T) {
	yaml := `
objects:
  - id: area-myoko
    kind: area
    name: Myoko Area
    spec:
      terrains:
        - id: ter-myoko
          kind: terrain
          name: Mount Myoko
          spec:
            terrain_type: mountain
            elevation_m: 2459
            grounds:
              - id: grnd-summit
                kind: ground
                name: Summit Slope
                attributes:
                  status: active
                spec:
                  slope_deg: 35
                  stability: stable
                  soil_type: rock
              - id: grnd-scree
                kind: ground
                name: Scree Field
                attributes:
                  status: standby
                spec:
                  slope_deg: 55
                  stability: watch
                  soil_type: gravel
`

	parser := NewParser()
	g, err := parser.Parse([]byte(yaml))
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	// 1 area + 1 terrain + 2 grounds = 4
	if g.EntityCount() != 4 {
		t.Fatalf("expected 4 entities, got %d", g.EntityCount())
	}

	// Check terrain
	terrain, ok := g.GetEntity("ter-myoko")
	if !ok {
		t.Fatal("entity ter-myoko not found")
	}
	if terrain.Owner != "area-myoko" {
		t.Errorf("expected owner area-myoko, got %s", terrain.Owner)
	}
	if ttype, ok := terrain.GetProperty("terrain_type"); !ok || ttype != "mountain" {
		t.Errorf("expected terrain_type mountain, got %v", ttype)
	}

	// Check summit ground
	summit, ok := g.GetEntity("grnd-summit")
	if !ok {
		t.Fatal("entity grnd-summit not found")
	}
	if summit.Owner != "ter-myoko" {
		t.Errorf("expected owner ter-myoko, got %s", summit.Owner)
	}
}

func TestParseNestedGroundSlopeElevation(t *testing.T) {
	yaml := `
objects:
  - id: area-sado
    kind: area
    name: Sado Island
    spec:
      terrains:
        - id: ter-kosado
          kind: terrain
          name: Kosado Ridge
          spec:
            terrain_type: hill
            elevation_m: 440
            trails:
              - Kosado Traverse
            grounds:
              - id: grnd-north
                kind: ground
                name: North Slope
                spec:
                  slope_deg: 25
                  elevation_m: 180
                  soil_type: loam
              - id: grnd-south
                kind: ground
                name: South Slope
                spec:
                  slope_deg: 40
                  elevation_m: 210
                  soil_type: clay
`

	parser := NewParser()
	g, err := parser.Parse([]byte(yaml))
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	// 1 area + 1 terrain + 2 grounds = 4
	if g.EntityCount() != 4 {
		t.Fatalf("expected 4 entities, got %d", g.EntityCount())
	}

	// Check terrain
	terrain, ok := g.GetEntity("ter-kosado")
	if !ok {
		t.Fatal("entity ter-kosado not found")
	}
	if terrain.Owner != "area-sado" {
		t.Errorf("expected owner area-sado, got %s", terrain.Owner)
	}
	if ttype, ok := terrain.GetProperty("terrain_type"); !ok || ttype != "hill" {
		t.Errorf("expected terrain_type hill, got %v", ttype)
	}
	if trails, ok := terrain.GetProperty("trails"); !ok || len(trails.([]interface{})) != 1 {
		t.Errorf("expected 1 trail, got %v", trails)
	}

	// Check north slope ground
	north, ok := g.GetEntity("grnd-north")
	if !ok {
		t.Fatal("entity grnd-north not found")
	}
	if north.Owner != "ter-kosado" {
		t.Errorf("expected owner ter-kosado, got %s", north.Owner)
	}
	if north.Kind != kinds.Ground {
		t.Errorf("expected kind ground, got %s", north.Kind)
	}
	if slope, ok := north.GetProperty("slope_deg"); !ok || slope != 25 {
		t.Errorf("expected slope_deg 25, got %v", slope)
	}
	if soil, ok := north.GetProperty("soil_type"); !ok || soil != "loam" {
		t.Errorf("expected soil_type loam, got %v", soil)
	}

	// Check south slope ground
	south, ok := g.GetEntity("grnd-south")
	if !ok {
		t.Fatal("entity grnd-south not found")
	}
	if south.Owner != "ter-kosado" {
		t.Errorf("expected owner ter-kosado, got %s", south.Owner)
	}
	if slope, ok := south.GetProperty("slope_deg"); !ok || slope != 40 {
		t.Errorf("expected slope_deg 40, got %v", slope)
	}

	// Auto-generated belongs_to relations for nested entities
	if g.RelationCount() != 3 {
		t.Fatalf("expected 3 auto-generated belongs_to relations, got %d", g.RelationCount())
	}
	found := false
	for _, r := range g.Relations() {
		if r.Type == types.BelongsTo && r.Participants.Source == "grnd-north" && r.Participants.Target == "ter-kosado" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected belongs_to relation from grnd-north to ter-kosado")
	}
}

func TestParseSinglePondWaterBody(t *testing.T) {
	yaml := `
objects:
  - id: area-yuzawa
    kind: area
    name: Yuzawa Town
    spec:
      water_bodies:
        - id: wb-kotobiki-pond
          kind: water_body
          name: Kotobiki Pond
          spec:
            water_type: pond
            max_depth_m: 8
            inflows:
              - Kiyotsu River
`

	parser := NewParser()
	g, err := parser.Parse([]byte(yaml))
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	// 1 area + 1 pond = 2
	if g.EntityCount() != 2 {
		t.Fatalf("expected 2 entities, got %d", g.EntityCount())
	}

	pond, ok := g.GetEntity("wb-kotobiki-pond")
	if !ok {
		t.Fatal("entity wb-kotobiki-pond not found")
	}
	if pond.Owner != "area-yuzawa" {
		t.Errorf("expected owner area-yuzawa, got %s", pond.Owner)
	}
	if wtype, ok := pond.GetProperty("water_type"); !ok || wtype != "pond" {
		t.Errorf("expected water_type pond, got %v", wtype)
	}
	if inflows, ok := pond.GetProperty("inflows"); !ok || len(inflows.([]interface{})) != 1 {
		t.Errorf("expected 1 inflow, got %v", inflows)
	}

	// Auto-generated belongs_to relation from pond to area
	if g.RelationCount() != 1 {
		t.Fatalf("expected 1 auto-generated belongs_to relation, got %d", g.RelationCount())
	}
}

func TestRoundTripNestedEntities(t *testing.T) {
	yaml := `
objects:
  - id: area-niigata
    kind: area
    name: Niigata City Area
    spec:
      tourism_spots:
        - id: spot-toki-messe
          name: Toki Messe
          spec:
            spot_type: scenic
            events:
              - id: evt-fireworks-nm
                name: Nishimonai Fireworks
                spec:
                  season: summer
                  held_month: 7
`

	// Parse original
	parser1 := NewParser()
	g1, err := parser1.Parse([]byte(yaml))
	if err != nil {
		t.Fatalf("failed to parse original: %v", err)
	}

	if g1.EntityCount() != 3 {
		t.Fatalf("expected 3 entities, got %d", g1.EntityCount())
	}

	// Serialize
	serializer := NewSerializer()
	data, err := serializer.Serialize(g1)
	if err != nil {
		t.Fatalf("failed to serialize: %v", err)
	}

	// Parse serialized
	parser2 := NewParser()
	g2, err := parser2.Parse(data)
	if err != nil {
		t.Fatalf("failed to parse serialized: %v", err)
	}

	if g2.EntityCount() != g1.EntityCount() {
		t.Errorf("entity count mismatch: %d vs %d", g1.EntityCount(), g2.EntityCount())
	}

	// Verify round-trip integrity
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
}

func TestValidationNoSlashInID(t *testing.T) {
	yaml := `
objects:
  - id: bad/entity
    kind: species
    name: Bad Entity
`

	parser := NewParser()
	g, err := parser.Parse([]byte(yaml))
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	s := schema.CoreSchema()
	engine := validation.NewEngine(s)
	validation.RegisterCoreRules(engine)
	result := engine.Validate(g, nil)

	found := false
	for _, f := range result.Findings {
		if f.RuleID == "no-slash-in-id" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected no-slash-in-id finding for entity with slash in ID")
	}
}

func TestValidationNestingParent(t *testing.T) {
	yaml := `
objects:
  - id: area-01
    kind: area
    name: Area 01

  - id: spx-01
    kind: species
    name: Species 01
    attributes:
      owner: area-01
`

	parser := NewParser()
	g, err := parser.Parse([]byte(yaml))
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	s := schema.CoreSchema()
	engine := validation.NewEngine(s)
	validation.RegisterCoreRules(engine)
	result := engine.Validate(g, nil)

	for _, f := range result.Findings {
		if f.RuleID == "valid-nesting-parent" {
			t.Error("valid-nesting-parent warning should not be emitted for species owned by area (species is a valid child of area)")
		}
	}
}

func TestPathReferenceResolution(t *testing.T) {
	yaml := `
objects:
  - id: area-aga
    kind: area
    name: Aga Town
    spec:
      tourism_spots:
        - id: spot-ohide
          name: Ohide Beach
          spec:
            spot_type: beach

  - id: area-yuzawa
    kind: area
    name: Yuzawa Town
    spec:
      tourism_spots:
        - id: spot-naeba
          name: Naeba Ski Resort
          spec:
            spot_type: ski_resort

  - id: rel-near
    type: near
    participants:
      - area-aga/spot-ohide
      - area-yuzawa/spot-naeba
`

	parser := NewParser()
	g, err := parser.Parse([]byte(yaml))
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	errs := ResolveReferences(g)
	if len(errs) != 0 {
		t.Errorf("expected no reference errors, got %v", errs)
	}
}

func TestParseEntityPropertyReference(t *testing.T) {
	yaml := `
objects:
  - id: wb-lake-kamo
    kind: water_body
    name: Lake Kamo
    spec:
      water_type: lake
      max_depth_m: 7

  - id: spx-crucian-carp
    kind: species
    name: Crucian Carp
    spec:
      category: fish
      spawning_ground: "@wb-lake-kamo"
`

	parser := NewParser()
	g, err := parser.Parse([]byte(yaml))
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	species, ok := g.GetEntity("spx-crucian-carp")
	if !ok {
		t.Fatal("expected entity spx-crucian-carp")
	}

	v, ok := species.GetProperty("spawning_ground")
	if !ok {
		t.Fatal("expected property spawning_ground")
	}
	ref, ok := v.(core.ReferenceValue)
	if !ok {
		t.Fatalf("expected ReferenceValue, got %T", v)
	}
	if ref.RefTargetID() != "wb-lake-kamo" {
		t.Errorf("expected reference target wb-lake-kamo, got %s", ref.RefTargetID())
	}

	// Verify reference resolution
	errs := ResolveReferences(g)
	if len(errs) != 0 {
		t.Errorf("expected no reference errors, got %v", errs)
	}
}

func TestParseEntityPropertyReferenceNotFound(t *testing.T) {
	yaml := `
objects:
  - id: spx-crucian-carp
    kind: species
    name: Crucian Carp
    spec:
      category: fish
      spawning_ground: "@nonexistent"
`

	parser := NewParser()
	g, err := parser.Parse([]byte(yaml))
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	errs := ResolveReferences(g)
	if len(errs) == 0 {
		t.Error("expected reference error for nonexistent entity")
	}
	found := false
	for _, e := range errs {
		if strings.Contains(e.Error(), "nonexistent") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected error about nonexistent reference, got %v", errs)
	}
}

func TestResolveReferencesListProperty(t *testing.T) {
	yaml := `
objects:
  - id: wb-pond
    kind: water_body
    name: Irrigation Pond
    spec:
      water_type: pond

  - id: fst-beech
    kind: forest
    name: Beech Forest

  - id: spx-ayu
    kind: species
    name: Ayu
    spec:
      habitats:
        - "@fst-beech"
        - "@ghost"

  - id: spot-ohide
    kind: tourism_spot
    name: Ohide Beach
    spec:
      nearby_nature:
        - "@wb-pond"
`

	parser := NewParser()
	g, err := parser.Parse([]byte(yaml))
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	errs := ResolveReferences(g)
	if len(errs) == 0 {
		t.Fatal("expected reference error for dangling list element")
	}
	found := false
	for _, e := range errs {
		if strings.Contains(e.Error(), "ghost") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected error about dangling list element @ghost, got %v", errs)
	}

	// spot-ohide's valid list reference must not produce an error mentioning it.
	for _, e := range errs {
		if strings.Contains(e.Error(), "spot-ohide") {
			t.Errorf("spot-ohide should not have a reference error, got: %v", e)
		}
	}
}

func TestParseAreaNestingSpecies(t *testing.T) {
	yaml := `
objects:
  - id: area-sado
    kind: area
    name: Sado Island
    species:
      - id: spx-ibis
        kind: species
        name: Crested Ibis
        spec:
          red_list_status: endangered
      - id: spx-deer
        kind: species
        name: Sika Deer
        spec:
          red_list_status: least_concern
`
	parser := NewParser()
	g, err := parser.Parse([]byte(yaml))
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	// Verify nested species are owned by the area
	for _, id := range []string{"spx-ibis", "spx-deer"} {
		e, ok := g.GetEntity(id)
		if !ok {
			t.Fatalf("entity %s not found", id)
		}
		if e.Kind != kinds.Species {
			t.Errorf("entity %s kind should be species, got %s", id, e.Kind)
		}
		if e.Owner != "area-sado" {
			t.Errorf("entity %s owner should be area-sado, got %s", id, e.Owner)
		}
		if v, ok := e.GetProperty("red_list_status"); !ok || v != "endangered" && v != "least_concern" {
			t.Errorf("entity %s unexpected red_list_status, got %v", id, v)
		}
	}

	// Verify auto-generated belongs_to relations (species -> area)
	for _, spxID := range []string{"spx-ibis", "spx-deer"} {
		relID := "rel-auto-belongs_to-" + spxID + "-area-sado"
		rel, ok := g.GetRelation(relID)
		if !ok {
			t.Errorf("missing auto relation %s", relID)
			continue
		}
		if rel.Type != types.BelongsTo {
			t.Errorf("relation %s: expected belongs_to, got %s", relID, rel.Type)
		}
		if rel.Participants.Source != spxID || rel.Participants.Target != "area-sado" {
			t.Errorf("relation %s: expected %s -> area-sado, got %s -> %s",
				relID, spxID, rel.Participants.Source, rel.Participants.Target)
		}
	}
}

func TestParseRelationPropertyReference(t *testing.T) {
	yaml := `
objects:
  - id: spring-tsukioka
    kind: hot_spring
    name: Tsukioka Spring

  - id: wb-sabiura
    kind: water_body
    name: Sabiura River

  - id: rel-depends
    type: depends_on
    participants:
      source: spring-tsukioka
      target: wb-sabiura
    spec:
      dependency_type: "@wb-sabiura"
`

	parser := NewParser()
	g, err := parser.Parse([]byte(yaml))
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	rel, ok := g.GetRelation("rel-depends")
	if !ok {
		t.Fatal("expected relation rel-depends")
	}

	v, ok := rel.GetProperty("dependency_type")
	if !ok {
		t.Fatal("expected property dependency_type")
	}
	_, ok = v.(core.ReferenceValue)
	if !ok {
		t.Fatalf("expected ReferenceValue, got %T", v)
	}

	errs := ResolveReferences(g)
	if len(errs) != 0 {
		t.Errorf("expected no reference errors, got %v", errs)
	}
}

func TestParsePropertyPlainTextNotAffected(t *testing.T) {
	yaml := `
objects:
  - id: spot-ohide
    kind: tourism_spot
    name: Ohide Beach
    spec:
      spot_type: beach
      description: "A spot without @ prefix"
`

	parser := NewParser()
	g, err := parser.Parse([]byte(yaml))
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	spot, ok := g.GetEntity("spot-ohide")
	if !ok {
		t.Fatal("expected entity spot-ohide")
	}

	// Plain string should NOT be converted to ReferenceValue
	v, ok := spot.GetProperty("spot_type")
	if !ok {
		t.Fatal("expected property spot_type")
	}
	if _, ok := v.(core.ReferenceValue); ok {
		t.Error("plain string should not be converted to ReferenceValue")
	}
	str, ok := v.(string)
	if !ok || str != "beach" {
		t.Errorf("expected plain string beach, got %v", v)
	}
}

func TestConvertPropertyValueRecursive(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		wantRef  bool
		wantType string
	}{
		{
			name:    "simple @ reference",
			input:   "@wb-lake",
			wantRef: true,
		},
		{
			name:     "plain string",
			input:    "hello",
			wantType: "string",
		},
		{
			name: "list with @ reference",
			input: []interface{}{
				"@wb-lake",
				"plain",
			},
			wantType: "list-of-strings",
		},
		{
			name: "nested map with @ reference",
			input: map[string]interface{}{
				"river": "@wb-lake",
				"name":  "mgmt",
			},
			wantType: "map",
		},
		{
			name: "list of maps with @ reference",
			input: []interface{}{
				map[string]interface{}{
					"river":     "@wb-lake",
					"max_depth": float64(100),
				},
			},
			wantType: "list-of-maps",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := convertPropertyValue(tt.input)

			if tt.wantRef {
				ref, ok := result.(core.ReferenceValue)
				if !ok {
					t.Errorf("expected ReferenceValue, got %T", result)
				}
				if ref.RefTargetID() != "wb-lake" {
					t.Errorf("expected reference target wb-lake, got %s", ref.RefTargetID())
				}
				return
			}

			switch tt.wantType {
			case "list-of-strings":
				list, ok := result.([]interface{})
				if !ok {
					t.Fatalf("expected []interface{}, got %T", result)
				}
				if len(list) > 0 {
					if ref, ok := list[0].(core.ReferenceValue); !ok {
						t.Errorf("expected list[0] to be ReferenceValue, got %T", list[0])
					} else if ref.RefTargetID() != "wb-lake" {
						t.Errorf("expected reference target wb-lake, got %s", ref.RefTargetID())
					}
				}
			case "list-of-maps":
				list, ok := result.([]interface{})
				if !ok {
					t.Fatalf("expected []interface{}, got %T", result)
				}
				if len(list) > 0 {
					m, ok := list[0].(map[string]interface{})
					if !ok {
						t.Fatalf("expected list[0] to be map, got %T", list[0])
					}
					if ref, ok := m["river"].(core.ReferenceValue); !ok {
						t.Errorf("expected river to be ReferenceValue, got %T", m["river"])
					} else if ref.RefTargetID() != "wb-lake" {
						t.Errorf("expected reference target wb-lake, got %s", ref.RefTargetID())
					}
				}
			case "map":
				m, ok := result.(map[string]interface{})
				if !ok {
					t.Fatalf("expected map[string]interface{}, got %T", result)
				}
				if ref, ok := m["river"].(core.ReferenceValue); !ok {
					t.Errorf("expected river to be ReferenceValue, got %T", m["river"])
				} else if ref.RefTargetID() != "wb-lake" {
					t.Errorf("expected reference target wb-lake, got %s", ref.RefTargetID())
				}
			case "string":
				str, ok := result.(string)
				if !ok {
					t.Fatalf("expected string, got %T", result)
				}
				if str != "hello" {
					t.Errorf("expected hello, got %s", str)
				}
			}
		})

		t.Run("area nesting tourism_spot generates belongs_to relation", func(t *testing.T) {
			yaml := `
objects:
  - id: area-01
    kind: area
    name: Area 01
    spec:
      tourism_spots:
        - id: spot-01
          name: Tourism Spot 01
          spec:
            spot_type: park
`
			parser := NewParser()
			g, err := parser.Parse([]byte(yaml))
			if err != nil {
				t.Fatalf("failed to parse: %v", err)
			}

			area, ok := g.GetEntity("area-01")
			if !ok {
				t.Fatal("area not found")
			}
			if area.Owner != "" {
				t.Errorf("expected area to be root, got owner %s", area.Owner)
			}

			spot, ok := g.GetEntity("spot-01")
			if !ok {
				t.Fatal("tourism spot not found")
			}
			if spot.Owner != "area-01" {
				t.Errorf("expected tourism spot owner to be area-01, got %s", spot.Owner)
			}

			rel, ok := g.GetRelation("rel-auto-belongs_to-spot-01-area-01")
			if !ok || rel.Type != types.BelongsTo {
				t.Error("missing belongs_to relation from tourism spot to area")
			}
		})

		t.Run("tourism_spot nesting hot_spring generates belongs_to relation", func(t *testing.T) {
			yaml := `
objects:
  - id: area-01
    kind: area
    name: Area 01
    spec:
      tourism_spots:
        - id: spot-01
          name: Tourism Spot 01
          spec:
            hot_springs:
              - id: spring-01
                name: Hot Spring 01
`
			parser := NewParser()
			g, err := parser.Parse([]byte(yaml))
			if err != nil {
				t.Fatalf("failed to parse: %v", err)
			}

			spring, ok := g.GetEntity("spring-01")
			if !ok {
				t.Fatal("hot spring not found")
			}
			if spring.Owner != "spot-01" {
				t.Errorf("expected hot spring owner to be spot-01, got %s", spring.Owner)
			}

			rel, ok := g.GetRelation("rel-auto-belongs_to-spring-01-spot-01")
			if !ok || rel.Type != types.BelongsTo {
				t.Error("missing belongs_to relation from hot spring to tourism spot")
			}
		})

		t.Run("hot_spring nesting event generates belongs_to relation", func(t *testing.T) {
			yaml := `
objects:
  - id: spring-01
    kind: hot_spring
    name: Hot Spring 01
    spec:
      spring_quality: sulfur
      events:
        - id: evt-01
          name: Event 01
          spec:
            season: summer
`
			parser := NewParser()
			g, err := parser.Parse([]byte(yaml))
			if err != nil {
				t.Fatalf("failed to parse: %v", err)
			}

			evt, ok := g.GetEntity("evt-01")
			if !ok {
				t.Fatal("event not found")
			}
			if evt.Owner != "spring-01" {
				t.Errorf("expected event owner to be spring-01, got %s", evt.Owner)
			}

			rel, ok := g.GetRelation("rel-auto-belongs_to-evt-01-spring-01")
			if !ok || rel.Type != types.BelongsTo {
				t.Error("missing belongs_to relation from event to hot spring")
			}
		})

		t.Run("species nesting population generates belongs_to relation", func(t *testing.T) {
			yaml := `
objects:
  - id: spx-01
    kind: species
    name: Species 01
    spec:
      category: bird
      populations:
        - id: pop-01
          name: Population 01
          spec:
            count: 150
`
			parser := NewParser()
			g, err := parser.Parse([]byte(yaml))
			if err != nil {
				t.Fatalf("failed to parse: %v", err)
			}

			pop, ok := g.GetEntity("pop-01")
			if !ok {
				t.Fatal("population not found")
			}
			if pop.Owner != "spx-01" {
				t.Errorf("expected population owner to be spx-01, got %s", pop.Owner)
			}

			rel, ok := g.GetRelation("rel-auto-belongs_to-pop-01-spx-01")
			if !ok || rel.Type != types.BelongsTo {
				t.Error("missing belongs_to relation from population to species")
			}
		})

		t.Run("multi-ground area scenario", func(t *testing.T) {
			yaml := `
objects:
  - id: area-sado
    kind: area
    name: Sado Island
  - id: fst-north
    kind: forest
    name: North Forest
    attributes:
      owner: area-sado
  - id: fst-south
    kind: forest
    name: South Forest
    attributes:
      owner: area-sado
  - id: spot-tsukioka
    kind: tourism_spot
    name: Tsukioka Onsen
    attributes:
      owner: area-sado
    spec:
      hot_springs:
        - id: spring-01
          name: Spring 01
          spec:
            spring_quality: chloride
  - id: rel-located-01
    type: located_in
    participants:
      source: fst-north
      target: area-sado
  - id: rel-located-02
    type: located_in
    participants:
      source: fst-south
      target: area-sado
`
			parser := NewParser()
			g, err := parser.Parse([]byte(yaml))
			if err != nil {
				t.Fatalf("failed to parse: %v", err)
			}

			// Verify entities
			fst1, ok := g.GetEntity("fst-north")
			if !ok || fst1.Owner != "area-sado" {
				t.Error("forest 01 owner mismatch")
			}
			fst2, ok := g.GetEntity("fst-south")
			if !ok || fst2.Owner != "area-sado" {
				t.Error("forest 02 owner mismatch")
			}
			spot, ok := g.GetEntity("spot-tsukioka")
			if !ok || spot.Owner != "area-sado" {
				t.Error("tourism spot owner mismatch")
			}
			spring, ok := g.GetEntity("spring-01")
			if !ok || spring.Owner != "spot-tsukioka" {
				t.Error("hot spring owner mismatch")
			}

			// Verify explicit located_in relations (multi-child)
			rel1, ok := g.GetRelation("rel-located-01")
			if !ok || rel1.Type != types.LocatedIn {
				t.Error("missing explicit located_in relation 01")
			}
			rel2, ok := g.GetRelation("rel-located-02")
			if !ok || rel2.Type != types.LocatedIn {
				t.Error("missing explicit located_in relation 02")
			}

			// Verify auto-relation from nesting
			autoRel, ok := g.GetRelation("rel-auto-belongs_to-spring-01-spot-tsukioka")
			if !ok || autoRel.Type != types.BelongsTo {
				t.Error("missing auto belongs_to relation from hot spring to tourism spot")
			}
		})

		t.Run("area nesting forests and species generates belongs_to relations", func(t *testing.T) {
			yaml := `
objects:
  - id: area-sado
    kind: area
    name: Sado Island
    spec:
      species:
        - id: spx-deer-01
          name: Sika Deer Herd 01
          spec:
            category: mammal
        - id: spx-deer-02
          name: Sika Deer Herd 02
          spec:
            category: mammal
      forests:
        - id: fst-cedar-01
          name: Cedar Stand 01
`
			parser := NewParser()
			g, err := parser.Parse([]byte(yaml))
			if err != nil {
				t.Fatalf("failed to parse: %v", err)
			}

			// Verify nested entities are owned by the area
			for _, id := range []string{"spx-deer-01", "spx-deer-02", "fst-cedar-01"} {
				e, ok := g.GetEntity(id)
				if !ok {
					t.Fatalf("entity %s not found", id)
				}
				if e.Owner != "area-sado" {
					t.Errorf("entity %s owner should be area-sado, got %s", id, e.Owner)
				}
			}

			// Verify auto-generated belongs_to relations (member -> area)
			rels := map[string]string{
				"rel-auto-belongs_to-spx-deer-01-area-sado": "spx-deer-01",
				"rel-auto-belongs_to-spx-deer-02-area-sado": "spx-deer-02",
				"rel-auto-belongs_to-fst-cedar-01-area-sado": "fst-cedar-01",
			}
			for relID, memberID := range rels {
				rel, ok := g.GetRelation(relID)
				if !ok {
					t.Errorf("missing auto relation %s", relID)
					continue
				}
				if rel.Type != types.BelongsTo {
					t.Errorf("relation %s: expected belongs_to, got %s", relID, rel.Type)
				}
				if rel.Participants.Source != memberID || rel.Participants.Target != "area-sado" {
					t.Errorf("relation %s: expected source %s -> target area-sado, got %s -> %s",
						relID, memberID, rel.Participants.Source, rel.Participants.Target)
				}
			}
		})

		t.Run("event depending on multiple hot springs via depends_on relations", func(t *testing.T) {
			yaml := `
objects:
  - id: spring-01
    kind: hot_spring
    name: Hot Spring 01
  - id: spring-02
    kind: hot_spring
    name: Hot Spring 02
  - id: spot-onsen-town
    kind: tourism_spot
    name: Onsen Town
    spec:
      events:
        - id: evt-lantern
          name: Lantern Festival
          spec:
            season: autumn
  - id: rel-dep-01
    type: depends_on
    participants:
      source: spring-01
      target: evt-lantern
  - id: rel-dep-02
    type: depends_on
    participants:
      source: spring-02
      target: evt-lantern
`
			parser := NewParser()
			g, err := parser.Parse([]byte(yaml))
			if err != nil {
				t.Fatalf("failed to parse: %v", err)
			}

			evt, ok := g.GetEntity("evt-lantern")
			if !ok {
				t.Fatal("event not found")
			}
			if evt.Owner != "spot-onsen-town" {
				t.Errorf("expected event owner to be spot-onsen-town, got %s", evt.Owner)
			}

			rel1, ok := g.GetRelation("rel-dep-01")
			if !ok || rel1.Type != types.DependsOn {
				t.Error("missing depends_on relation from spring-01 to event")
			}
			if rel1.Participants.Source != "spring-01" || rel1.Participants.Target != "evt-lantern" {
				t.Errorf("wrong participants: %s -> %s", rel1.Participants.Source, rel1.Participants.Target)
			}

			rel2, ok := g.GetRelation("rel-dep-02")
			if !ok || rel2.Type != types.DependsOn {
				t.Error("missing depends_on relation from spring-02 to event")
			}
			if rel2.Participants.Source != "spring-02" || rel2.Participants.Target != "evt-lantern" {
				t.Errorf("wrong participants: %s -> %s", rel2.Participants.Source, rel2.Participants.Target)
			}

			autoRel, ok := g.GetRelation("rel-auto-belongs_to-evt-lantern-spot-onsen-town")
			if !ok || autoRel.Type != types.BelongsTo {
				t.Error("missing auto belongs_to relation from event to tourism spot")
			}
		})
	}
}

func TestAutoRelationGeneration(t *testing.T) {
	t.Run("area nesting tourism_spot generates belongs_to relation", func(t *testing.T) {
		yaml := `
objects:
  - id: area-01
    kind: area
    name: Area 01
    spec:
      tourism_spots:
        - id: spot-01
          name: Tourism Spot 01
`
		parser := NewParser()
		g, err := parser.Parse([]byte(yaml))
		if err != nil {
			t.Fatalf("failed to parse: %v", err)
		}

		rel, ok := g.GetRelation("rel-auto-belongs_to-spot-01-area-01")
		if !ok {
			t.Fatal("auto-relation not generated")
		}
		if rel.Type != types.BelongsTo {
			t.Errorf("expected belongs_to, got %s", rel.Type)
		}
		if rel.Participants.Source != "spot-01" {
			t.Errorf("expected source spot-01, got %s", rel.Participants.Source)
		}
		if rel.Participants.Target != "area-01" {
			t.Errorf("expected target area-01, got %s", rel.Participants.Target)
		}
		if val, ok := rel.GetLabel("auto_generated"); !ok || val != "true" {
			t.Error("expected auto_generated label")
		}
	})

	t.Run("area nesting forest generates belongs_to relation", func(t *testing.T) {
		yaml := `
objects:
  - id: area-01
    kind: area
    name: Area 01
    spec:
      forests:
        - id: fst-01
          name: Forest 01
`
		parser := NewParser()
		g, err := parser.Parse([]byte(yaml))
		if err != nil {
			t.Fatalf("failed to parse: %v", err)
		}

		rel, ok := g.GetRelation("rel-auto-belongs_to-fst-01-area-01")
		if !ok {
			t.Fatal("auto-relation not generated")
		}
		if rel.Type != types.BelongsTo {
			t.Errorf("expected belongs_to, got %s", rel.Type)
		}
		if rel.Participants.Source != "fst-01" {
			t.Errorf("expected source fst-01, got %s", rel.Participants.Source)
		}
		if rel.Participants.Target != "area-01" {
			t.Errorf("expected target area-01, got %s", rel.Participants.Target)
		}
	})

	t.Run("species nesting population generates belongs_to relation", func(t *testing.T) {
		yaml := `
objects:
  - id: spx-01
    kind: species
    name: Species 01
    spec:
      populations:
        - id: pop-01
          name: Population 01
`
		parser := NewParser()
		g, err := parser.Parse([]byte(yaml))
		if err != nil {
			t.Fatalf("failed to parse: %v", err)
		}

		rel, ok := g.GetRelation("rel-auto-belongs_to-pop-01-spx-01")
		if !ok {
			t.Fatal("auto-relation not generated")
		}
		if rel.Type != types.BelongsTo {
			t.Errorf("expected belongs_to, got %s", rel.Type)
		}
	})

	t.Run("explicit relation skips auto-relation", func(t *testing.T) {
		yaml := `
objects:
  - id: area-01
    kind: area
    name: Area 01
    spec:
      tourism_spots:
        - id: spot-01
          name: Tourism Spot 01
  - id: rel-member-1
    type: belongs_to
    participants:
      source: spot-01
      target: area-01
`
		parser := NewParser()
		g, err := parser.Parse([]byte(yaml))
		if err != nil {
			t.Fatalf("failed to parse: %v", err)
		}

		// Explicit relation should exist
		_, ok := g.GetRelation("rel-member-1")
		if !ok {
			t.Fatal("explicit relation not found")
		}

		// Auto-relation should not be generated (duplicate skipped)
		autoRel, ok := g.GetRelation("rel-auto-belongs_to-spot-01-area-01")
		if ok {
			t.Errorf("auto-relation should have been skipped, but found: %v", autoRel)
		}
	})

	t.Run("multi-level nesting", func(t *testing.T) {
		yaml := `
objects:
  - id: area-01
    kind: area
    name: Area 01
    spec:
      tourism_spots:
        - id: spot-01
          name: Tourism Spot 01
          spec:
            hot_springs:
              - id: spring-01
                name: Hot Spring 01
`
		parser := NewParser()
		g, err := parser.Parse([]byte(yaml))
		if err != nil {
			t.Fatalf("failed to parse: %v", err)
		}

		// tourism_spot -> area (belongs_to)
		rel1, ok := g.GetRelation("rel-auto-belongs_to-spot-01-area-01")
		if !ok {
			t.Fatal("auto-relation spot->area not generated")
		}
		if rel1.Type != types.BelongsTo {
			t.Errorf("expected belongs_to, got %s", rel1.Type)
		}

		// hot_spring -> tourism_spot (belongs_to)
		rel2, ok := g.GetRelation("rel-auto-belongs_to-spring-01-spot-01")
		if !ok {
			t.Fatal("auto-relation spring->spot not generated")
		}
		if rel2.Type != types.BelongsTo {
			t.Errorf("expected belongs_to, got %s", rel2.Type)
		}
	})
}

func TestAutoRelationConfig(t *testing.T) {
	t.Run("disabled auto-relation", func(t *testing.T) {
		yaml := `
objects:
  - id: area-01
    kind: area
    name: Area 01
    spec:
      forests:
        - id: fst-01
          name: Forest 01
`
		config := &AutoRelationConfig{
			Disabled: true,
		}
		parser := NewParserWithAutoRelationConfig(schema.CoreSchema(), config)
		g, err := parser.Parse([]byte(yaml))
		if err != nil {
			t.Fatalf("failed to parse: %v", err)
		}

		_, ok := g.GetRelation("rel-auto-belongs_to-fst-01-area-01")
		if ok {
			t.Error("auto-relation should not be generated when disabled")
		}
	})

	t.Run("custom override", func(t *testing.T) {
		yaml := `
objects:
  - id: area-01
    kind: area
    name: Area 01
    spec:
      tourism_spots:
        - id: spot-01
          name: Tourism Spot 01
`
		config := &AutoRelationConfig{
			Overrides: map[string]AutoRelationMapping{
				"area.tourism_spots": {
					RelationType: types.DependsOn,
					Source:       "child",
				},
			},
		}
		parser := NewParserWithAutoRelationConfig(schema.CoreSchema(), config)
		g, err := parser.Parse([]byte(yaml))
		if err != nil {
			t.Fatalf("failed to parse: %v", err)
		}

		rel, ok := g.GetRelation("rel-auto-depends_on-spot-01-area-01")
		if !ok {
			t.Fatal("auto-relation not generated")
		}
		if rel.Type != types.DependsOn {
			t.Errorf("expected depends_on, got %s", rel.Type)
		}
	})
}

func TestAutoRelationIntegration(t *testing.T) {
	t.Run("nesting to flat round-trip preserves auto-relations", func(t *testing.T) {
		nestedYAML := `
objects:
  - id: area-01
    kind: area
    name: Area 01
    spec:
      tourism_spots:
        - id: spot-01
          name: Tourism Spot 01
          spec:
            hot_springs:
              - id: spring-01
                name: Hot Spring 01
                spec:
                  events:
                    - id: evt-01
                      name: Event 01
`
		parser := NewParser()
		g1, err := parser.Parse([]byte(nestedYAML))
		if err != nil {
			t.Fatalf("failed to parse: %v", err)
		}

		if g1.EntityCount() != 4 {
			t.Errorf("expected 4 entities, got %d", g1.EntityCount())
		}
		if g1.RelationCount() != 3 {
			t.Errorf("expected 3 auto-relations, got %d", g1.RelationCount())
		}

		// Verify specific relations
		rel1, ok := g1.GetRelation("rel-auto-belongs_to-spot-01-area-01")
		if !ok || rel1.Type != types.BelongsTo {
			t.Error("missing belongs_to spot->area")
		}

		rel2, ok := g1.GetRelation("rel-auto-belongs_to-spring-01-spot-01")
		if !ok || rel2.Type != types.BelongsTo {
			t.Error("missing belongs_to spring->spot")
		}

		rel3, ok := g1.GetRelation("rel-auto-belongs_to-evt-01-spring-01")
		if !ok || rel3.Type != types.BelongsTo {
			t.Error("missing belongs_to event->spring")
		}

		// Serialize to flat YAML
		serializer := NewSerializer()
		data, err := serializer.Serialize(g1)
		if err != nil {
			t.Fatalf("failed to serialize: %v", err)
		}

		// Re-parse should produce equivalent graph
		parser2 := NewParser()
		g2, err := parser2.Parse(data)
		if err != nil {
			t.Fatalf("failed to re-parse: %v", err)
		}

		// All 4 entities should exist
		if g2.EntityCount() != 4 {
			t.Errorf("expected 4 entities after round-trip, got %d", g2.EntityCount())
		}

		// Verify ownership preserved
		spot, _ := g2.GetEntity("spot-01")
		if spot.Owner != "area-01" {
			t.Errorf("tourism spot owner should be area-01, got %s", spot.Owner)
		}
		spring, _ := g2.GetEntity("spring-01")
		if spring.Owner != "spot-01" {
			t.Errorf("hot spring owner should be spot-01, got %s", spring.Owner)
		}
		evt, _ := g2.GetEntity("evt-01")
		if evt.Owner != "spring-01" {
			t.Errorf("event owner should be spring-01, got %s", evt.Owner)
		}
	})

	t.Run("all nesting types generate correct relations", func(t *testing.T) {
		yaml := `
objects:
  - id: area-01
    kind: area
    name: Area 01
    spec:
      tourism_spots:
        - id: spot-01
          name: Tourism Spot 01
          spec:
            hot_springs:
              - id: spring-01
                name: Hot Spring 01
                spec:
                  events:
                    - id: evt-01
                      name: Event 01
                      spec:
                        season: spring
      forests:
        - id: fst-01
          name: Forest 01
          spec:
            forest_type: beech
`
		parser := NewParser()
		g, err := parser.Parse([]byte(yaml))
		if err != nil {
			t.Fatalf("failed to parse: %v", err)
		}

		// area -> tourism_spot (belongs_to, child source)
		r1, ok := g.GetRelation("rel-auto-belongs_to-spot-01-area-01")
		if !ok || r1.Type != types.BelongsTo {
			t.Error("missing belongs_to spot->area")
		}
		if r1.Participants.Source != "spot-01" || r1.Participants.Target != "area-01" {
			t.Errorf("wrong participants: %s -> %s", r1.Participants.Source, r1.Participants.Target)
		}

		// area -> forest (belongs_to, child source)
		r2, ok := g.GetRelation("rel-auto-belongs_to-fst-01-area-01")
		if !ok || r2.Type != types.BelongsTo {
			t.Error("missing belongs_to forest->area")
		}

		// tourism_spot -> hot_spring (belongs_to, child source)
		r3, ok := g.GetRelation("rel-auto-belongs_to-spring-01-spot-01")
		if !ok || r3.Type != types.BelongsTo {
			t.Error("missing belongs_to spring->spot")
		}

		// hot_spring -> event (belongs_to, child source)
		r4, ok := g.GetRelation("rel-auto-belongs_to-evt-01-spring-01")
		if !ok || r4.Type != types.BelongsTo {
			t.Error("missing belongs_to event->spring")
		}
	})

	t.Run("explicit relation prevents auto-generation", func(t *testing.T) {
		yaml := `
objects:
  - id: fst-01
    kind: forest
    name: Forest 01
    spec:
      grounds:
        - id: grnd-01
          name: Ground 01
  - id: rel-custom
    type: belongs_to
    participants:
      source: grnd-01
      target: fst-01
`
		parser := NewParser()
		g, err := parser.Parse([]byte(yaml))
		if err != nil {
			t.Fatalf("failed to parse: %v", err)
		}

		// Only the explicit relation should exist
		if g.RelationCount() != 1 {
			t.Errorf("expected 1 relation, got %d", g.RelationCount())
		}
		_, ok := g.GetRelation("rel-custom")
		if !ok {
			t.Error("explicit relation not found")
		}
		_, ok = g.GetRelation("rel-auto-belongs_to-grnd-01-fst-01")
		if ok {
			t.Error("auto-relation should not exist when explicit one is present")
		}
	})
}

func TestParseResortEventReferences(t *testing.T) {
	yaml := `
objects:
  - id: area-yuzawa
    kind: area
    name: Yuzawa Town
    spec:
      tourism_spots:
        - id: spot-naeba
          kind: tourism_spot
          name: Naeba Ski Resort
          spec:
            spot_type: ski_resort
            events:
              - id: evt-fireworks
                kind: event
                name: Summer Fireworks
                spec:
                  season: summer
                  held_month: 7
                  free_admission: false
                  partner_event: "@evt-snowboard"
              - id: evt-snowboard
                kind: event
                name: Snowboard Contest
                spec:
                  season: winter
                  held_month: 2
                  free_admission: true
                  partner_event: "@evt-fireworks"

  - id: spot-gala
    kind: tourism_spot
    name: Gala Yuzawa
    spec:
      spot_type: ski_resort

  - id: spot-mitsumata
    kind: tourism_spot
    name: Mitsumata Ropeway
    spec:
      spot_type: scenic
`

	parser := NewParser()
	g, err := parser.Parse([]byte(yaml))
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	// 1 area + 1 resort + 2 events + 2 spots = 6
	if g.EntityCount() != 6 {
		t.Fatalf("expected 6 entities, got %d", g.EntityCount())
	}

	// Check resort
	resort, ok := g.GetEntity("spot-naeba")
	if !ok {
		t.Fatal("entity spot-naeba not found")
	}
	if resort.Owner != "area-yuzawa" {
		t.Errorf("expected owner area-yuzawa, got %s", resort.Owner)
	}
	spotType, ok := resort.GetProperty("spot_type")
	if !ok || spotType != "ski_resort" {
		t.Errorf("expected spot_type ski_resort, got %v", spotType)
	}

	// Check events nested under the resort
	fireworks, ok := g.GetEntity("evt-fireworks")
	if !ok {
		t.Fatal("entity evt-fireworks not found")
	}
	if fireworks.Owner != "spot-naeba" {
		t.Errorf("expected owner spot-naeba, got %s", fireworks.Owner)
	}
	freeFireworks, ok := fireworks.GetProperty("free_admission")
	if !ok || freeFireworks != false {
		t.Errorf("expected free_admission false for evt-fireworks, got %v", freeFireworks)
	}

	snowboard, ok := g.GetEntity("evt-snowboard")
	if !ok {
		t.Fatal("entity evt-snowboard not found")
	}
	freeSnowboard, ok := snowboard.GetProperty("free_admission")
	if !ok || freeSnowboard != true {
		t.Errorf("expected free_admission true for evt-snowboard, got %v", freeSnowboard)
	}

	// Check auto-generated belongs_to relations
	if g.RelationCount() < 2 {
		t.Errorf("expected at least 2 auto-generated belongs_to relations, got %d", g.RelationCount())
	}
}

func TestParseAreasNestedChildren(t *testing.T) {
	yaml := `
objects:
  - id: area-niigata-city
    kind: area
    name: Niigata City
    spec:
      population: 810000
      tourism_spots:
        - id: spot-minatopia
          name: Minatopia History Museum
          spec:
            spot_type: museum
            annual_visitors: 300000
        - id: spot-furumachi
          name: Furumachi Historic District
          spec:
            spot_type: historic
            annual_visitors: 150000

  - id: area-aga
    kind: area
    name: Aga Town
    spec:
      forests:
        - id: fst/2025-survey
          name: Aga Cedar Stand
          spec:
            forest_type: cedar_plantation
`

	parser := NewParser()
	g, err := parser.Parse([]byte(yaml))
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	city, ok := g.GetEntity("area-niigata-city")
	if !ok {
		t.Fatal("entity area-niigata-city not found")
	}
	population, ok := city.GetProperty("population")
	if !ok || population != 810000 {
		t.Errorf("expected population 810000, got %v", population)
	}

	// Tourism spots are tourism_spot entities owned by the area
	for _, sid := range []string{"spot-minatopia", "spot-furumachi"} {
		spot, ok := g.GetEntity(sid)
		if !ok {
			t.Fatalf("entity %s not found", sid)
		}
		if spot.Kind != kinds.TourismSpot {
			t.Errorf("spot %s: expected kind tourism_spot, got %s", sid, spot.Kind)
		}
		if spot.Owner != "area-niigata-city" {
			t.Errorf("spot %s: expected owner area-niigata-city, got %s", sid, spot.Owner)
		}
	}

	// The surveyed forest is a forest entity owned by the other area
	fst, ok := g.GetEntity("fst/2025-survey")
	if !ok {
		t.Fatal("entity fst/2025-survey not found")
	}
	if fst.Kind != kinds.Forest {
		t.Errorf("survey forest: expected kind forest, got %s", fst.Kind)
	}
	if fst.Owner != "area-aga" {
		t.Errorf("survey forest: expected owner area-aga, got %s", fst.Owner)
	}

	// Auto-generated belongs_to relations for nested children
	belongsTo := g.RelationsByType(core.RelationType("belongs_to"))
	if len(belongsTo) < 3 {
		t.Errorf("expected at least 3 auto-generated belongs_to relations, got %d", len(belongsTo))
	}
}

func TestParseAreaPopulationPropertyOnly(t *testing.T) {
	yaml := `
objects:
  - id: area-nagaoka
    kind: area
    name: Nagaoka City
    spec:
      population: 275000
`

	parser := NewParser()
	g, err := parser.Parse([]byte(yaml))
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	area, ok := g.GetEntity("area-nagaoka")
	if !ok {
		t.Fatal("entity area-nagaoka not found")
	}
	population, ok := area.GetProperty("population")
	if !ok || population != 275000 {
		t.Errorf("expected population 275000, got %v", population)
	}
}

func TestParseAreaSpotsCountRejected(t *testing.T) {
	// The `tourism_spots` key is a nest key for area; an integer count must no
	// longer be accepted (previously broken `tourism_spots: 12` property).
	yaml := `
objects:
  - id: area-01
    kind: area
    name: Area 01
    spec:
      tourism_spots: 12
`

	parser := NewParser()
	if _, err := parser.Parse([]byte(yaml)); err == nil {
		t.Error("expected error when area declares tourism_spots as an integer count")
	}
}

func TestParseDirCrossFileReferences(t *testing.T) {
	dir := t.TempDir()

	writeTestFile := func(relPath, content string) {
		t.Helper()
		full := filepath.Join(dir, relPath)
		if err := os.MkdirAll(filepath.Dir(full), 0755); err != nil {
			t.Fatalf("failed to create dir: %v", err)
		}
		if err := os.WriteFile(full, []byte(content), 0644); err != nil {
			t.Fatalf("failed to write %s: %v", relPath, err)
		}
	}

	// fileA defines entities that are referenced from the other files.
	writeTestFile("fileA.yaml", `
objects:
  - id: area-sado
    kind: area
    name: Sado Island

  - id: spx-ibis
    kind: species
    name: Crested Ibis
    attributes:
      owner: area-sado
    spec:
      scientific_name: Nipponia nippon
      populations:
        - id: pop-wild
          name: Wild Population
          spec:
            count: 190

  - id: wb-lake-kamo
    kind: water_body
    name: Lake Kamo
    spec:
      water_type: lake
      max_depth_m: 7
`)

	// fileB references entities from fileA (owner, relation participant,
	// path-based participant, and @-prefixed property reference).
	writeTestFile("fileB.yaml", `
objects:
  - id: fst-beech
    kind: forest
    name: Beech Forest
    attributes:
      owner: area-sado
    spec:
      forest_type: beech
      area_ha: 120

  - id: spx-crucian-carp
    kind: species
    name: Crucian Carp
    attributes:
      owner: area-sado
    spec:
      category: fish
      spawning_ground: "@wb-lake-kamo"

  - id: rel-inhabits
    type: inhabits
    participants:
      source: pop-wild
      target: wb-lake-kamo

  - id: rel-near
    type: near
    participants:
      - area-sado/spx-ibis/pop-wild
      - fst-mixed/grnd-01
`)

	// fileC lives in a nested subdirectory to exercise recursive walking.
	writeTestFile("nested/fileC.yaml", `
objects:
  - id: fst-mixed
    kind: forest
    name: Mixed Forest
    attributes:
      owner: area-sado
    spec:
      forest_type: mixed
      grounds:
        - id: grnd-01
          name: Valley Floor
`)

	p := NewParser()
	g, err := p.ParseDir(dir)
	if err != nil {
		t.Fatalf("failed to parse dir: %v", err)
	}

	// All entities from all files are merged into a single graph.
	expectedEntities := []string{"area-sado", "spx-ibis", "pop-wild", "wb-lake-kamo", "fst-beech", "spx-crucian-carp", "fst-mixed", "grnd-01"}
	for _, id := range expectedEntities {
		if _, ok := g.GetEntity(id); !ok {
			t.Errorf("expected merged entity %s, not found", id)
		}
	}

	// Explicit cross-file relations are present (nesting may also generate
	// auto-relations, so verify by ID rather than exact count).
	if _, ok := g.GetRelation("rel-inhabits"); !ok {
		t.Error("expected merged relation rel-inhabits, not found")
	}
	if _, ok := g.GetRelation("rel-near"); !ok {
		t.Error("expected merged relation rel-near, not found")
	}

	// Cross-file ownership resolves.
	forest, ok := g.GetEntity("fst-beech")
	if !ok {
		t.Fatal("expected entity fst-beech")
	}
	if forest.Owner != "area-sado" {
		t.Errorf("expected fst-beech owner area-sado (defined in fileA), got %s", forest.Owner)
	}

	// Cross-file @-prefixed property reference resolves.
	species, ok := g.GetEntity("spx-crucian-carp")
	if !ok {
		t.Fatal("expected entity spx-crucian-carp")
	}
	v, ok := species.GetProperty("spawning_ground")
	if !ok {
		t.Fatal("expected property spawning_ground")
	}
	ref, ok := v.(core.ReferenceValue)
	if !ok {
		t.Fatalf("expected ReferenceValue, got %T", v)
	}
	if ref.RefTargetID() != "wb-lake-kamo" {
		t.Errorf("expected reference target wb-lake-kamo (defined in fileA), got %s", ref.RefTargetID())
	}

	// All references resolve across the merged graph, including path references
	// spanning nested entities defined in different files.
	if errs := ResolveReferences(g); len(errs) != 0 {
		t.Errorf("expected no reference errors, got %v", errs)
	}

	e, ok := g.ResolvePathEntity("area-sado/spx-ibis/pop-wild")
	if !ok || e.ID != "pop-wild" {
		t.Errorf("expected path reference to resolve to pop-wild, got %v (ok=%v)", e, ok)
	}
}
