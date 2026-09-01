package schema

import (
	"testing"

	"github.com/bababa/Niigata_Real_IaC/src/core"
	"github.com/bababa/Niigata_Real_IaC/src/core/kinds"
	"github.com/bababa/Niigata_Real_IaC/src/core/types"
)

func TestNewSchema(t *testing.T) {
	s := NewSchema("1.0.0", "1.0")
	if s == nil {
		t.Fatal("expected non-nil schema")
	}
	if s.Version.SchemaVersion != "1.0.0" {
		t.Errorf("expected schema version 1.0.0, got %s", s.Version.SchemaVersion)
	}
	if s.Version.SpecVersion != "1.0" {
		t.Errorf("expected spec version 1.0, got %s", s.Version.SpecVersion)
	}
	if len(s.EntityKinds) != 0 {
		t.Errorf("expected empty entity kinds, got %d", len(s.EntityKinds))
	}
	if len(s.RelationTypes) != 0 {
		t.Errorf("expected empty relation types, got %d", len(s.RelationTypes))
	}
}

func TestCoreSchemaEntityKinds(t *testing.T) {
	s := CoreSchema()

	expectedKinds := []core.EntityKind{
		kinds.Area, kinds.Ground, kinds.Terrain, kinds.WaterBody,
		kinds.Forest, kinds.Species, kinds.Population, kinds.TourismSpot,
		kinds.HotSpring, kinds.CulturalAsset, kinds.Event,
	}

	if len(s.EntityKinds) != len(expectedKinds) {
		t.Fatalf("expected %d entity kinds, got %d", len(expectedKinds), len(s.EntityKinds))
	}

	for _, kind := range expectedKinds {
		if !s.HasEntityKind(kind) {
			t.Errorf("expected entity kind %q to be defined", kind)
		}
	}
}

func TestCoreSchemaRelationTypes(t *testing.T) {
	s := CoreSchema()

	expectedTypes := []core.RelationType{
		types.LocatedIn, types.Inhabits, types.Near,
		types.DependsOn, types.BelongsTo, types.FlowsInto,
	}

	if len(s.RelationTypes) != len(expectedTypes) {
		t.Fatalf("expected %d relation types, got %d", len(expectedTypes), len(s.RelationTypes))
	}

	for _, relType := range expectedTypes {
		if !s.HasRelationType(relType) {
			t.Errorf("expected relation type %q to be defined", relType)
		}
	}
}

func TestCoreSchemaRelationDirections(t *testing.T) {
	s := CoreSchema()

	tests := []struct {
		relType  core.RelationType
		expected DirectionType
	}{
		{types.LocatedIn, DirectionDirected},
		{types.Inhabits, DirectionDirected},
		{types.Near, DirectionSymmetric},
		{types.DependsOn, DirectionDirected},
		{types.BelongsTo, DirectionDirected},
		{types.FlowsInto, DirectionDirected},
	}

	for _, tt := range tests {
		def, ok := s.GetRelationTypeDef(tt.relType)
		if !ok {
			t.Fatalf("relation type %q not found", tt.relType)
		}
		if def.Direction != tt.expected {
			t.Errorf("relation type %q: expected direction %q, got %q", tt.relType, tt.expected, def.Direction)
		}
	}
}

func TestCoreSchemaEntityKindProperties(t *testing.T) {
	s := CoreSchema()

	// Population should have count property (required integer, min 0)
	populationDef, ok := s.GetEntityKindDef(kinds.Population)
	if !ok {
		t.Fatal("population kind not found")
	}
	foundCount := false
	for _, p := range populationDef.Properties {
		if p.Name == "count" {
			foundCount = true
			if p.Type != PropertyTypeInteger {
				t.Errorf("expected count type integer, got %s", p.Type)
			}
			if !p.Required {
				t.Error("expected count property to be required")
			}
			if p.Constraints == nil || p.Constraints.Min == nil {
				t.Error("expected count property to have min constraint")
			} else if *p.Constraints.Min != 0 {
				t.Errorf("expected count min 0, got %v", *p.Constraints.Min)
			}
		}
	}
	if !foundCount {
		t.Error("population kind missing count property")
	}

	// Event should have held_month property with range 1-12
	eventDef, ok := s.GetEntityKindDef(kinds.Event)
	if !ok {
		t.Fatal("event kind not found")
	}
	foundMonth := false
	for _, p := range eventDef.Properties {
		if p.Name == "held_month" {
			foundMonth = true
			if p.Required {
				t.Error("expected held_month property to be optional")
			}
			if p.Constraints == nil || p.Constraints.Min == nil || p.Constraints.Max == nil {
				t.Error("expected held_month property to have min and max constraints")
			}
		}
	}
	if !foundMonth {
		t.Error("event kind missing held_month property")
	}

	// Area should have latitude/longitude constraints
	areaDef, ok := s.GetEntityKindDef(kinds.Area)
	if !ok {
		t.Fatal("area kind not found")
	}
	foundLat, foundLon := false, false
	for _, p := range areaDef.Properties {
		switch p.Name {
		case "latitude":
			foundLat = true
			if p.Type != PropertyTypeNumber {
				t.Errorf("expected latitude type number, got %s", p.Type)
			}
			if p.Constraints == nil || p.Constraints.Min == nil || p.Constraints.Max == nil {
				t.Fatal("latitude should have min/max constraints")
			}
			if *p.Constraints.Min != -90 || *p.Constraints.Max != 90 {
				t.Errorf("expected latitude range [-90,90], got [%v,%v]", *p.Constraints.Min, *p.Constraints.Max)
			}
		case "longitude":
			foundLon = true
			if p.Constraints == nil || p.Constraints.Min == nil || p.Constraints.Max == nil {
				t.Fatal("longitude should have min/max constraints")
			}
			if *p.Constraints.Min != -180 || *p.Constraints.Max != 180 {
				t.Errorf("expected longitude range [-180,180], got [%v,%v]", *p.Constraints.Min, *p.Constraints.Max)
			}
		}
	}
	if !foundLat {
		t.Error("area kind missing latitude property")
	}
	if !foundLon {
		t.Error("area kind missing longitude property")
	}
}

// TestCoreSchemaGeoPropertiesOnPlaceKinds verifies that place-bearing entity
// kinds declare optional latitude/longitude properties with WGS84 bounds.
func TestCoreSchemaGeoPropertiesOnPlaceKinds(t *testing.T) {
	s := CoreSchema()

	placeKinds := []core.EntityKind{
		kinds.Ground, kinds.Terrain, kinds.WaterBody, kinds.Forest,
		kinds.TourismSpot, kinds.HotSpring, kinds.CulturalAsset,
	}

	for _, kind := range placeKinds {
		def, ok := s.GetEntityKindDef(kind)
		if !ok {
			t.Fatalf("%s kind not found", kind)
		}
		props := make(map[string]PropertyDefinition, len(def.Properties))
		for _, p := range def.Properties {
			props[p.Name] = p
		}
		for _, name := range []string{"latitude", "longitude"} {
			p, ok := props[name]
			if !ok {
				t.Errorf("%s kind missing %s property", kind, name)
				continue
			}
			if p.Required {
				t.Errorf("%s kind %s should be optional", kind, name)
			}
			if p.Type != PropertyTypeNumber {
				t.Errorf("%s kind %s should be number, got %s", kind, name, p.Type)
			}
			if p.Constraints == nil || p.Constraints.Min == nil || p.Constraints.Max == nil {
				t.Fatalf("%s kind %s should have min/max constraints", kind, name)
			}
		}
		if p, ok := props["latitude"]; ok && (*p.Constraints.Min != -90 || *p.Constraints.Max != 90) {
			t.Errorf("%s kind latitude range = [%v,%v], expected [-90,90]", kind, *p.Constraints.Min, *p.Constraints.Max)
		}
		if p, ok := props["longitude"]; ok && (*p.Constraints.Min != -180 || *p.Constraints.Max != 180) {
			t.Errorf("%s kind longitude range = [%v,%v], expected [-180,180]", kind, *p.Constraints.Min, *p.Constraints.Max)
		}
	}
}

func TestCoreSchemaRelationParticipantConstraints(t *testing.T) {
	s := CoreSchema()

	// FlowsInto should be restricted to water_body -> water_body
	flowsDef, ok := s.GetRelationTypeDef(types.FlowsInto)
	if !ok {
		t.Fatal("flows_into relation type not found")
	}
	if flowsDef.Participants == nil {
		t.Fatal("flows_into missing participant constraints")
	}
	if len(flowsDef.Participants.SourceKinds) != 1 || flowsDef.Participants.SourceKinds[0] != kinds.WaterBody {
		t.Errorf("flows_into source kinds should be [water_body], got %v", flowsDef.Participants.SourceKinds)
	}
	if len(flowsDef.Participants.TargetKinds) != 1 || flowsDef.Participants.TargetKinds[0] != kinds.WaterBody {
		t.Errorf("flows_into target kinds should be [water_body], got %v", flowsDef.Participants.TargetKinds)
	}

	// Inhabits should have species and population as source
	inhabitsDef, ok := s.GetRelationTypeDef(types.Inhabits)
	if !ok {
		t.Fatal("inhabits relation type not found")
	}
	if inhabitsDef.Participants == nil {
		t.Fatal("inhabits missing participant constraints")
	}
	if len(inhabitsDef.Participants.SourceKinds) != 2 {
		t.Errorf("inhabits should have 2 source kinds, got %d", len(inhabitsDef.Participants.SourceKinds))
	}

	// LocatedIn targets should be area or terrain
	locDef, ok := s.GetRelationTypeDef(types.LocatedIn)
	if !ok {
		t.Fatal("located_in relation type not found")
	}
	if locDef.Participants == nil {
		t.Fatal("located_in missing participant constraints")
	}
	validTargets := map[core.EntityKind]bool{kinds.Area: true, kinds.Terrain: true}
	if len(locDef.Participants.TargetKinds) != len(validTargets) {
		t.Fatalf("located_in should have %d target kinds, got %d", len(validTargets), len(locDef.Participants.TargetKinds))
	}
	for _, k := range locDef.Participants.TargetKinds {
		if !validTargets[k] {
			t.Errorf("located_in unexpected target kind %q", k)
		}
	}
}

func TestAddEntityKind(t *testing.T) {
	s := NewSchema("1.0.0", "1.0")
	s.AddEntityKind("custom_kind", &EntityKindDefinition{
		Description: "Custom kind",
	})

	if !s.HasEntityKind("custom_kind") {
		t.Error("expected custom_kind to exist")
	}

	def, ok := s.GetEntityKindDef("custom_kind")
	if !ok {
		t.Fatal("custom_kind not found")
	}
	if def.Description != "Custom kind" {
		t.Errorf("expected description 'Custom kind', got %q", def.Description)
	}
}

func TestAddRelationType(t *testing.T) {
	s := NewSchema("1.0.0", "1.0")
	s.AddRelationType("custom_type", &RelationTypeDefinition{
		Direction: DirectionDirected,
	})

	if !s.HasRelationType("custom_type") {
		t.Error("expected custom_type to exist")
	}

	def, ok := s.GetRelationTypeDef("custom_type")
	if !ok {
		t.Fatal("custom_type not found")
	}
	if def.Direction != DirectionDirected {
		t.Errorf("expected direction directed, got %s", def.Direction)
	}
}

func TestCoreSchemaNestingDefinitions(t *testing.T) {
	s := CoreSchema()

	tests := []struct {
		parentKind  core.EntityKind
		nestKey     string
		childKind   core.EntityKind
		autoRelType core.RelationType
	}{
		// Area can nest grounds, terrains, water bodies, forests, tourism spots, and species
		{kinds.Area, "grounds", kinds.Ground, types.BelongsTo},
		{kinds.Area, "terrains", kinds.Terrain, types.BelongsTo},
		{kinds.Area, "water_bodies", kinds.WaterBody, types.BelongsTo},
		{kinds.Area, "forests", kinds.Forest, types.BelongsTo},
		{kinds.Area, "tourism_spots", kinds.TourismSpot, types.BelongsTo},
		{kinds.Area, "species", kinds.Species, types.BelongsTo},
		// Terrain can nest grounds
		{kinds.Terrain, "grounds", kinds.Ground, types.BelongsTo},
		// Forest can nest grounds
		{kinds.Forest, "grounds", kinds.Ground, types.BelongsTo},
		// Species can nest populations
		{kinds.Species, "populations", kinds.Population, types.BelongsTo},
		// Tourism spot can nest hot springs and events
		{kinds.TourismSpot, "hot_springs", kinds.HotSpring, types.BelongsTo},
		{kinds.TourismSpot, "events", kinds.Event, types.BelongsTo},
		// Hot spring can nest events
		{kinds.HotSpring, "events", kinds.Event, types.BelongsTo},
	}

	for _, tt := range tests {
		defs := s.GetNestingDefs(tt.parentKind)
		found := false
		for _, d := range defs {
			if d.NestKey == tt.nestKey && d.ChildKind == tt.childKind {
				if d.AutoRelationType != tt.autoRelType {
					t.Errorf("%s/%s: expected auto relation %s, got %s",
						tt.parentKind, tt.nestKey, tt.autoRelType, d.AutoRelationType)
				}
				found = true
				break
			}
		}
		if !found {
			t.Errorf("%s/%s -> %s nesting definition not found", tt.parentKind, tt.nestKey, tt.childKind)
		}
	}
}

func TestFindNestingByChildKind(t *testing.T) {
	s := CoreSchema()

	// area nests grounds
	def, ok := s.FindNestingByChildKind(kinds.Area, kinds.Ground)
	if !ok {
		t.Fatal("expected nesting definition for area -> ground")
	}
	if def.NestKey != "grounds" {
		t.Errorf("expected nest key 'grounds', got %q", def.NestKey)
	}
	if def.ChildKind != kinds.Ground {
		t.Errorf("expected child kind ground, got %s", def.ChildKind)
	}
	if def.AutoRelationType != types.BelongsTo {
		t.Errorf("expected auto relation belongs_to, got %s", def.AutoRelationType)
	}

	// species nests populations
	def, ok = s.FindNestingByChildKind(kinds.Species, kinds.Population)
	if !ok {
		t.Fatal("expected nesting definition for species -> population")
	}
	if def.NestKey != "populations" {
		t.Errorf("expected nest key 'populations', got %q", def.NestKey)
	}

	// terrain does not nest populations
	if _, ok := s.FindNestingByChildKind(kinds.Terrain, kinds.Population); ok {
		t.Error("expected no nesting definition for terrain -> population")
	}

	// unknown parent kind yields no nesting
	if _, ok := s.FindNestingByChildKind("unknown_kind", kinds.Ground); ok {
		t.Error("expected no nesting definition for unknown parent kind")
	}
}

func TestCoreSchemaFindNestingByNestKey(t *testing.T) {
	s := CoreSchema()

	def, ok := s.FindNestingByNestKey(kinds.TourismSpot, "hot_springs")
	if !ok {
		t.Fatal("expected nesting definition for tourism_spot/hot_springs")
	}
	if def.ChildKind != kinds.HotSpring {
		t.Errorf("expected child kind hot_spring, got %s", def.ChildKind)
	}

	if _, ok := s.FindNestingByNestKey(kinds.TourismSpot, "nonexistent"); ok {
		t.Error("expected no nesting definition for nonexistent nest key")
	}
}

func TestValidateProperty(t *testing.T) {
	s := CoreSchema()

	populationDef, _ := s.GetEntityKindDef(kinds.Population)

	// Find count property (required integer, min 0)
	var countProp *PropertyDefinition
	for i := range populationDef.Properties {
		if populationDef.Properties[i].Name == "count" {
			countProp = &populationDef.Properties[i]
			break
		}
	}
	if countProp == nil {
		t.Fatal("count property not found")
	}

	// Required property with nil value should fail
	if err := s.ValidateProperty(countProp, nil); err == nil {
		t.Error("expected error for nil value on required count property")
	}

	// Valid integer value
	if err := s.ValidateProperty(countProp, 42); err != nil {
		t.Errorf("expected no error for valid integer value, got %v", err)
	}

	// Value below minimum
	if err := s.ValidateProperty(countProp, -1); err == nil {
		t.Error("expected error for negative count value")
	}

	// Non-numeric value (should fail for integer type)
	if err := s.ValidateProperty(countProp, "many"); err == nil {
		t.Error("expected error for string value on integer property")
	}

	// Find survey_method property (optional string enum)
	var methodProp *PropertyDefinition
	for i := range populationDef.Properties {
		if populationDef.Properties[i].Name == "survey_method" {
			methodProp = &populationDef.Properties[i]
			break
		}
	}
	if methodProp == nil {
		t.Fatal("survey_method property not found")
	}

	// Optional property with nil value should pass
	if err := s.ValidateProperty(methodProp, nil); err != nil {
		t.Errorf("expected no error for nil value on optional property, got %v", err)
	}

	// Valid enum value
	if err := s.ValidateProperty(methodProp, "transect"); err != nil {
		t.Errorf("expected no error for valid enum value, got %v", err)
	}

	// Invalid enum value
	if err := s.ValidateProperty(methodProp, "guess"); err == nil {
		t.Error("expected error for invalid enum value")
	}
}

func TestValidatePropertyRequired(t *testing.T) {
	s := NewSchema("1.0.0", "1.0")
	prop := &PropertyDefinition{
		Name:     "test_prop",
		Type:     PropertyTypeString,
		Required: true,
	}

	// Test nil value on required property
	if err := s.ValidateProperty(prop, nil); err == nil {
		t.Error("expected error for nil value on required property")
	}

	// Test non-nil value on required property
	if err := s.ValidateProperty(prop, "hello"); err != nil {
		t.Errorf("expected no error for valid value, got %v", err)
	}
}

func TestValidatePropertyStringConstraints(t *testing.T) {
	s := NewSchema("1.0.0", "1.0")
	minLen := 3
	maxLen := 10
	prop := &PropertyDefinition{
		Name: "test_prop",
		Type: PropertyTypeString,
		Constraints: &Constraint{
			MinLength: &minLen,
			MaxLength: &maxLen,
		},
	}

	// Too short
	if err := s.ValidateProperty(prop, "ab"); err == nil {
		t.Error("expected error for string too short")
	}

	// Too long
	if err := s.ValidateProperty(prop, "abcdefghijk"); err == nil {
		t.Error("expected error for string too long")
	}

	// Valid
	if err := s.ValidateProperty(prop, "hello"); err != nil {
		t.Errorf("expected no error for valid string, got %v", err)
	}
}

func TestValidatePropertyEnumConstraints(t *testing.T) {
	s := NewSchema("1.0.0", "1.0")
	prop := &PropertyDefinition{
		Name: "soil_type",
		Type: PropertyTypeString,
		Constraints: &Constraint{
			Enum: []string{"loam", "sand"},
		},
	}

	// Valid enum value
	if err := s.ValidateProperty(prop, "loam"); err != nil {
		t.Errorf("expected no error for valid enum value, got %v", err)
	}

	// Invalid enum value
	if err := s.ValidateProperty(prop, "silt"); err == nil {
		t.Error("expected error for invalid enum value")
	}
}

func TestValidatePropertyPatternConstraints(t *testing.T) {
	s := NewSchema("1.0.0", "1.0")
	pattern := `^srv-[a-z0-9-]+$`
	prop := &PropertyDefinition{
		Name: "hostname",
		Type: PropertyTypeString,
		Constraints: &Constraint{
			Pattern: &pattern,
		},
	}

	// Valid match
	if err := s.ValidateProperty(prop, "srv-web-01"); err != nil {
		t.Errorf("expected no error for matching string, got %v", err)
	}

	// Invalid match
	if err := s.ValidateProperty(prop, "web_server!"); err == nil {
		t.Error("expected error for string not matching pattern")
	}

	// Invalid pattern (schema definition error)
	bad := `[`
	badProp := &PropertyDefinition{
		Name: "hostname",
		Type: PropertyTypeString,
		Constraints: &Constraint{
			Pattern: &bad,
		},
	}
	if err := s.ValidateProperty(badProp, "anything"); err == nil {
		t.Error("expected error for invalid pattern")
	}
}

func TestValidatePropertyListUnknownKey(t *testing.T) {
	s := NewSchema("1.0.0", "1.0")
	prop := &PropertyDefinition{
		Name: "observations",
		Type: PropertyTypeList,
		Properties: []PropertyDefinition{
			{Name: "count", Type: PropertyTypeInteger},
			{Name: "method", Type: PropertyTypeString},
		},
	}

	// Valid item
	valid := []interface{}{
		map[string]interface{}{"count": 4, "method": "transect"},
	}
	if err := s.ValidateProperty(prop, valid); err != nil {
		t.Errorf("expected no error for valid list item, got %v", err)
	}

	// Item with unknown key
	invalid := []interface{}{
		map[string]interface{}{"count": 4, "observer": "tanaka"},
	}
	if err := s.ValidateProperty(prop, invalid); err == nil {
		t.Error("expected error for list item with unknown key")
	}
}

func TestValidatePropertyMapValueProperty(t *testing.T) {
	s := NewSchema("1.0.0", "1.0")
	valueProp := &PropertyDefinition{Name: "entry", Type: PropertyTypeInteger}
	prop := &PropertyDefinition{
		Name:          "annual_visitors_by_year",
		Type:          PropertyTypeMap,
		ValueProperty: valueProp,
	}

	// Valid map values
	valid := map[string]interface{}{"2024": 8080, "2025": 5432}
	if err := s.ValidateProperty(prop, valid); err != nil {
		t.Errorf("expected no error for valid map values, got %v", err)
	}

	// map[string]string values must fail against integer value schema
	invalid := map[string]string{"2024": "8080"}
	if err := s.ValidateProperty(prop, invalid); err == nil {
		t.Error("expected error for map value not matching value property type")
	}

	// Non-numeric value in a numeric map
	badVal := map[string]interface{}{"2024": "8080"}
	if err := s.ValidateProperty(prop, badVal); err == nil {
		t.Error("expected error for non-integer map value")
	}

	// Map without value property still validates as a map
	plain := &PropertyDefinition{Name: "tags", Type: PropertyTypeMap}
	if err := s.ValidateProperty(plain, map[string]interface{}{"a": 1, "b": "two"}); err != nil {
		t.Errorf("expected no error for map without value property, got %v", err)
	}
}

func TestValidatePropertyNumericConstraints(t *testing.T) {
	s := NewSchema("1.0.0", "1.0")
	min := float64(1)
	max := float64(12)
	prop := &PropertyDefinition{
		Name: "held_month",
		Type: PropertyTypeInteger,
		Constraints: &Constraint{
			Min: &min,
			Max: &max,
		},
	}

	// Below min
	if err := s.ValidateProperty(prop, 0); err == nil {
		t.Error("expected error for value below min")
	}

	// Above max
	if err := s.ValidateProperty(prop, 13); err == nil {
		t.Error("expected error for value above max")
	}

	// Valid
	if err := s.ValidateProperty(prop, 8); err != nil {
		t.Errorf("expected no error for valid value, got %v", err)
	}
}

func TestProfile(t *testing.T) {
	p := NewProfile("test-profile")
	if p.Name != "test-profile" {
		t.Errorf("expected name 'test-profile', got %q", p.Name)
	}

	p.Rules = append(p.Rules, "unique-id", "valid-reference")

	if !p.HasRule("unique-id") {
		t.Error("expected profile to have rule 'unique-id'")
	}
	if !p.HasRule("valid-reference") {
		t.Error("expected profile to have rule 'valid-reference'")
	}
	if p.HasRule("nonexistent") {
		t.Error("expected profile not to have rule 'nonexistent'")
	}

	p.AddRequiredKind("area")
	if len(p.RequiredKinds) != 1 || p.RequiredKinds[0] != "area" {
		t.Error("expected profile to require kind 'area'")
	}

	p.AddRequiredRelation("located_in")
	if len(p.RequiredRelations) != 1 || p.RequiredRelations[0] != "located_in" {
		t.Error("expected profile to require relation 'located_in'")
	}
}

func TestValidatePropertyReferenceType(t *testing.T) {
	s := NewSchema("1.0.0", "1.0")
	propDef := &PropertyDefinition{
		Name:     "related_spot",
		Type:     PropertyTypeReference,
		Required: false,
	}

	// Valid ReferenceValue
	ref := core.NewReferenceValue("@spot-gappo")
	if err := s.ValidateProperty(propDef, ref); err != nil {
		t.Errorf("expected no error for valid ReferenceValue, got %v", err)
	}

	// Valid @ prefix string (not yet converted)
	if err := s.ValidateProperty(propDef, "@spot-gappo"); err != nil {
		t.Errorf("expected no error for @ prefix string, got %v", err)
	}

	// Invalid: plain string without @ prefix
	if err := s.ValidateProperty(propDef, "spot-gappo"); err == nil {
		t.Error("expected error for plain string without @ prefix")
	}

	// Invalid: non-string value
	if err := s.ValidateProperty(propDef, 42); err == nil {
		t.Error("expected error for non-string value")
	}

	// nil is OK (not required)
	if err := s.ValidateProperty(propDef, nil); err != nil {
		t.Errorf("expected no error for nil value, got %v", err)
	}
}

func TestCoreSchemaGroundSoilTypeProperty(t *testing.T) {
	s := CoreSchema()
	def, ok := s.GetEntityKindDef(kinds.Ground)
	if !ok {
		t.Fatal("ground kind not found")
	}

	var soilProp *PropertyDefinition
	for _, p := range def.Properties {
		if p.Name == "soil_type" {
			soilProp = &p
			break
		}
	}
	if soilProp == nil {
		t.Fatal("soil_type property not found on ground")
	}
	if soilProp.Type != PropertyTypeString {
		t.Errorf("expected type string, got %s", soilProp.Type)
	}
	if soilProp.Required {
		t.Error("soil_type should be optional")
	}
	if soilProp.Constraints == nil || len(soilProp.Constraints.Enum) != 7 {
		t.Fatalf("expected 7 enum values, got %v", soilProp.Constraints)
	}

	validSoils := map[string]bool{
		"loam": true, "sand": true, "clay": true, "gravel": true,
		"rock": true, "volcanic_ash": true, "peat": true,
	}
	for _, v := range soilProp.Constraints.Enum {
		if !validSoils[v] {
			t.Errorf("unexpected enum value %q", v)
		}
	}
}

func TestCoreSchemaDependsOnCriticalProperty(t *testing.T) {
	s := CoreSchema()
	def, ok := s.GetRelationTypeDef(types.DependsOn)
	if !ok {
		t.Fatal("depends_on relation type not found")
	}

	var criticalProp *PropertyDefinition
	for _, p := range def.Properties {
		if p.Name == "critical" {
			criticalProp = &p
			break
		}
	}
	if criticalProp == nil {
		t.Fatal("critical property not found on depends_on")
	}
	if criticalProp.Type != PropertyTypeBoolean {
		t.Errorf("expected type boolean, got %s", criticalProp.Type)
	}
	if criticalProp.Default != false {
		t.Errorf("expected default false, got %v", criticalProp.Default)
	}

	// dependency_type should be an enum with 5 values
	var depTypeProp *PropertyDefinition
	for _, p := range def.Properties {
		if p.Name == "dependency_type" {
			depTypeProp = &p
			break
		}
	}
	if depTypeProp == nil {
		t.Fatal("dependency_type property not found on depends_on")
	}
	if depTypeProp.Constraints == nil || len(depTypeProp.Constraints.Enum) != 5 {
		t.Fatalf("expected 5 enum values, got %v", depTypeProp.Constraints)
	}
	validDeps := map[string]bool{
		"source": true, "landscape": true, "access": true,
		"ecosystem": true, "event": true,
	}
	for _, v := range depTypeProp.Constraints.Enum {
		if !validDeps[v] {
			t.Errorf("unexpected enum value %q", v)
		}
	}
}

func TestCoreSchemaSpeciesCategoryEnum(t *testing.T) {
	s := CoreSchema()
	def, ok := s.GetEntityKindDef(kinds.Species)
	if !ok {
		t.Fatal("species kind not found")
	}

	var categoryProp *PropertyDefinition
	for _, p := range def.Properties {
		if p.Name == "category" {
			categoryProp = &p
			break
		}
	}
	if categoryProp == nil {
		t.Fatal("category property not found on species")
	}
	if categoryProp.Constraints == nil || len(categoryProp.Constraints.Enum) != 8 {
		t.Fatalf("expected 8 enum values, got %v", categoryProp.Constraints)
	}

	validCategories := map[string]bool{
		"mammal": true, "bird": true, "reptile": true, "amphibian": true,
		"fish": true, "insect": true, "plant": true, "other": true,
	}
	for _, v := range categoryProp.Constraints.Enum {
		if !validCategories[v] {
			t.Errorf("unexpected enum value %q", v)
		}
	}
}

func TestCoreSchemaEventHeldMonthProperty(t *testing.T) {
	s := CoreSchema()
	def, ok := s.GetEntityKindDef(kinds.Event)
	if !ok {
		t.Fatal("event kind not found")
	}

	var monthProp *PropertyDefinition
	for _, p := range def.Properties {
		if p.Name == "held_month" {
			monthProp = &p
			break
		}
	}
	if monthProp == nil {
		t.Fatal("held_month property not found on event")
	}
	if monthProp.Type != PropertyTypeInteger {
		t.Errorf("expected type integer, got %s", monthProp.Type)
	}
	if monthProp.Required {
		t.Errorf("held_month should be optional, but is required")
	}
	if monthProp.Constraints == nil || monthProp.Constraints.Min == nil || monthProp.Constraints.Max == nil {
		t.Fatal("held_month should have min/max constraints")
	}
	if *monthProp.Constraints.Min != 1 {
		t.Errorf("expected min 1, got %v", *monthProp.Constraints.Min)
	}
	if *monthProp.Constraints.Max != 12 {
		t.Errorf("expected max 12, got %v", *monthProp.Constraints.Max)
	}
}

func TestCoreSchemaWaterBodyProperties(t *testing.T) {
	s := CoreSchema()
	def, ok := s.GetEntityKindDef(kinds.WaterBody)
	if !ok {
		t.Fatal("water_body kind not found")
	}

	var waterTypeProp *PropertyDefinition
	for _, p := range def.Properties {
		if p.Name == "water_type" {
			waterTypeProp = &p
			break
		}
	}
	if waterTypeProp == nil {
		t.Fatal("water_type property not found on water_body")
	}
	if waterTypeProp.Constraints == nil || len(waterTypeProp.Constraints.Enum) != 6 {
		t.Fatalf("expected 6 enum values, got %v", waterTypeProp.Constraints)
	}

	validTypes := map[string]bool{
		"river": true, "lake": true, "sea": true,
		"pond": true, "marsh": true, "waterfall": true,
	}
	for _, v := range waterTypeProp.Constraints.Enum {
		if !validTypes[v] {
			t.Errorf("unexpected enum value %q", v)
		}
	}

	// length_km should be numeric with min 0
	var lengthProp *PropertyDefinition
	for _, p := range def.Properties {
		if p.Name == "length_km" {
			lengthProp = &p
			break
		}
	}
	if lengthProp == nil {
		t.Fatal("length_km property not found on water_body")
	}
	if lengthProp.Type != PropertyTypeNumber {
		t.Errorf("expected length_km type number, got %s", lengthProp.Type)
	}
	if lengthProp.Constraints == nil || lengthProp.Constraints.Min == nil || *lengthProp.Constraints.Min != 0 {
		t.Error("length_km should have min constraint of 0")
	}
}

func TestValidatePropertyIntegerFloat(t *testing.T) {
	s := CoreSchema()

	// Integral float64 (as produced by encoding/json) is acceptable for integer.
	integral := &PropertyDefinition{Name: "count", Type: PropertyTypeInteger}
	if err := s.ValidateProperty(integral, float64(42)); err != nil {
		t.Errorf("integer property with integral float64 should be accepted, got %v", err)
	}

	// Non-integral float64 must be rejected.
	if err := s.ValidateProperty(integral, 42.5); err == nil {
		t.Error("integer property with non-integral float64 should be rejected")
	}

	// Plain int remains accepted.
	if err := s.ValidateProperty(integral, 42); err != nil {
		t.Errorf("integer property with int should be accepted, got %v", err)
	}
}

func TestValidatePropertyNumberUnsigned(t *testing.T) {
	s := CoreSchema()
	def := &PropertyDefinition{Name: "elevation_m", Type: PropertyTypeNumber}
	if err := s.ValidateProperty(def, uint(240)); err != nil {
		t.Errorf("number property with uint should be accepted, got %v", err)
	}
	if err := s.ValidateProperty(def, uint64(240)); err != nil {
		t.Errorf("number property with uint64 should be accepted, got %v", err)
	}
}

func TestValidateNumericConstraintsUnsigned(t *testing.T) {
	s := CoreSchema()
	def := &PropertyDefinition{
		Name:        "held_month",
		Type:        PropertyTypeInteger,
		Constraints: &Constraint{Min: intPtr(1), Max: intPtr(12)},
	}
	if err := s.ValidateProperty(def, uint(7)); err != nil {
		t.Errorf("unsigned integer with min/max constraints should pass, got %v", err)
	}
	if err := s.ValidateProperty(def, uint(13)); err == nil {
		t.Error("unsigned integer exceeding max should be rejected")
	}
}

func TestValidatePropertyStringReferenceValue(t *testing.T) {
	minLength := 1
	s := CoreSchema()
	def := &PropertyDefinition{
		Name:        "scientific_name",
		Type:        PropertyTypeString,
		Constraints: &Constraint{MinLength: &minLength},
	}
	if err := s.ValidateProperty(def, core.NewReferenceValue("@species-tancho")); err != nil {
		t.Errorf("string property with ReferenceValue should be accepted, got %v", err)
	}
	if err := s.ValidateProperty(def, "Grus japonensis"); err != nil {
		t.Errorf("string property with plain string should be accepted, got %v", err)
	}
}
