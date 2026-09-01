package extension

import (
	"strings"
	"testing"

	"github.com/bababa/Niigata_Real_IaC/src/core"
	"github.com/bababa/Niigata_Real_IaC/src/core/kinds"
	"github.com/bababa/Niigata_Real_IaC/src/parser"
	"github.com/bababa/Niigata_Real_IaC/src/schema"
	"github.com/bababa/Niigata_Real_IaC/src/validation"
)

func TestAgriWildlifeExtensionManifest(t *testing.T) {
	ext := AgriWildlifeExtension()
	if ext.Manifest == nil {
		t.Fatal("expected manifest")
	}
	if ext.Manifest.ID != "niigata.agri-wildlife" {
		t.Errorf("unexpected ID: %s", ext.Manifest.ID)
	}
	if ext.Manifest.Namespace != "niigata.agri" {
		t.Errorf("unexpected namespace: %s", ext.Manifest.Namespace)
	}
	if len(ext.EntityKinds) != 2 {
		t.Errorf("expected 2 entity kinds, got %d", len(ext.EntityKinds))
	}
	if len(ext.ValidationRules) != 1 {
		t.Errorf("expected 1 validation rule, got %d", len(ext.ValidationRules))
	}
	if len(ext.RelationTypes) != 1 || !ext.RelationTypes[0].Augment {
		t.Errorf("expected 1 augmenting relation type contribution, got %d", len(ext.RelationTypes))
	}
	if len(ext.RootKinds) != 1 || ext.RootKinds[0] != kinds.Area {
		t.Errorf("expected area root kind grant, got %v", ext.RootKinds)
	}
}

// TestAgriWildlifeIsBuiltin ensures the extension is available through the
// standard NewSetup path without any external plugin directory.
func TestAgriWildlifeIsBuiltin(t *testing.T) {
	setup, err := NewSetup("")
	if err != nil {
		t.Fatalf("NewSetup failed: %v", err)
	}
	if !setup.Schema.HasEntityKind(KindWildlifeIncident) {
		t.Error("wildlife_incident kind missing from setup schema")
	}
	if !setup.Schema.HasEntityKind(KindCropHarvest) {
		t.Error("crop_harvest kind missing from setup schema")
	}
	if _, ok := setup.Manager.GetExtension("niigata.agri-wildlife"); !ok {
		t.Error("niigata.agri-wildlife extension not registered in manager")
	}
}

// TestAgriWildlifeAugmentsBelongsTo ensures the contributed record kinds can
// participate in the core belongs_to auto-relations.
func TestAgriWildlifeAugmentsBelongsTo(t *testing.T) {
	setup, err := NewSetup("")
	if err != nil {
		t.Fatalf("NewSetup failed: %v", err)
	}
	def, ok := setup.Schema.GetRelationTypeDef("belongs_to")
	if !ok {
		t.Fatal("belongs_to type missing")
	}
	for _, kind := range []core.EntityKind{KindWildlifeIncident, KindCropHarvest} {
		found := false
		for _, k := range def.Participants.SourceKinds {
			if k == kind {
				found = true
			}
		}
		if !found {
			t.Errorf("belongs_to source kinds missing %s", kind)
		}
	}
	// Core semantics must be untouched.
	found := false
	for _, k := range def.Participants.SourceKinds {
		if k == kinds.Population {
			found = true
		}
	}
	if !found {
		t.Error("belongs_to augmentation dropped core source kinds")
	}
}

// TestAgriWildlifeGrantsAreaRootAuthority ensures multiple area roots validate.
func TestAgriWildlifeGrantsAreaRootAuthority(t *testing.T) {
	setup, err := NewSetup("")
	if err != nil {
		t.Fatalf("NewSetup failed: %v", err)
	}
	if !setup.Validation.IsAllowedRootKind(kinds.Area) {
		t.Error("area kind has no root authority")
	}
}

func TestAgriWildlifeNestingUnderArea(t *testing.T) {
	setup, err := NewSetup("")
	if err != nil {
		t.Fatalf("NewSetup failed: %v", err)
	}

	for _, child := range []core.EntityKind{KindWildlifeIncident, KindCropHarvest} {
		nd, ok := setup.Schema.FindNestingByChildKind(kinds.Area, child)
		if !ok {
			t.Errorf("no nesting def for area -> %s", child)
			continue
		}
		if nd.AutoRelationType != "belongs_to" {
			t.Errorf("unexpected auto relation type for %s: %s", child, nd.AutoRelationType)
		}
	}

	if nd, ok := setup.Schema.FindNestingByNestKey(kinds.Area, "wildlife_incidents"); !ok || nd.ChildKind != KindWildlifeIncident {
		t.Errorf("nest key wildlife_incidents not resolved for area")
	}
	if nd, ok := setup.Schema.FindNestingByNestKey(kinds.Area, "crop_harvests"); !ok || nd.ChildKind != KindCropHarvest {
		t.Errorf("nest key crop_harvests not resolved for area")
	}

	// The global defs must not leak to unrelated parent kinds.
	if _, ok := setup.Schema.FindNestingByChildKind(kinds.Species, KindWildlifeIncident); ok {
		t.Error("wildlife_incident nesting def leaked to species")
	}
}

func TestAgriWildlifeNestingParentUnknownKind(t *testing.T) {
	s := schema.NewSchema("1.0.0", "1.0")
	ep := NewEntityKindsExtensionPoint(s)
	ext := &Extension{
		Manifest: &Manifest{ID: "bad-parent", Namespace: "bad"},
		EntityKinds: []EntityKindContribution{
			{
				Kind:        "orphan_kind",
				Definition:  &schema.EntityKindDefinition{},
				ParentKinds: []core.EntityKind{"nonexistent"},
			},
		},
	}
	if err := ep.Register(ext); err == nil {
		t.Fatal("expected error for unknown nesting parent kind")
	} else if !strings.Contains(err.Error(), "not defined in schema") {
		t.Errorf("unexpected error: %v", err)
	}
}

const agriSampleYAML = `
schema_version: "1.0"
objects:
  - id: area-niigata
    kind: area
    name: Niigata Prefecture
    spec:
      area_type: prefecture

  - id: incident-bear-2025
    kind: wildlife_incident
    name: Bear Incidents 2025
    attributes:
      owner: area-niigata
    spec:
      count: 42
      incident_type: sighting
      year: 2025
      location: Myoko foothills

  - id: harvest-rice-2025
    kind: crop_harvest
    name: Rice Harvest 2025
    attributes:
      owner: area-niigata
    spec:
      crop_type: rice
      harvest_t: 438000
      year: 2025
      area_ha: 61200
`

func TestAgriWildlifeValidateSample(t *testing.T) {
	setup, err := NewSetup("")
	if err != nil {
		t.Fatalf("NewSetup failed: %v", err)
	}

	p := parser.NewParserWithSchema(setup.Schema)
	g, err := p.Parse([]byte(agriSampleYAML))
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}

	result := setup.Validation.Validate(g, nil)
	if !result.Passed {
		for _, f := range result.Findings {
			t.Errorf("unexpected finding: %s %s: %s", f.Severity, f.RuleID, f.Message)
		}
	}
}

func TestAgriWildlifeYearRule(t *testing.T) {
	setup, err := NewSetup("")
	if err != nil {
		t.Fatalf("NewSetup failed: %v", err)
	}

	yaml := strings.Replace(agriSampleYAML, "year: 2025\n      location: Myoko foothills", "year: 1800\n      location: Myoko foothills", 1)
	yaml = strings.Replace(yaml, "year: 2025\n      area_ha: 61200", "year: 1800\n      area_ha: 61200", 1)

	p := parser.NewParserWithSchema(setup.Schema)
	g, err := p.Parse([]byte(yaml))
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}

	result := setup.Validation.Validate(g, nil)
	// The rule is a warning, so the graph still passes, but the findings must be present.
	if !result.Passed {
		t.Fatal("expected validation to pass (warnings only)")
	}

	ruleFired := 0
	for _, f := range result.Findings {
		if f.RuleID == "agri-year-plausible" && f.Severity == validation.SeverityWarning {
			ruleFired++
		}
	}
	if ruleFired != 2 {
		t.Errorf("expected 2 agri-year-plausible warnings, got %d", ruleFired)
	}
}

func TestAgriWildlifeRequiredProperties(t *testing.T) {
	setup, err := NewSetup("")
	if err != nil {
		t.Fatalf("NewSetup failed: %v", err)
	}

	yaml := `
schema_version: "1.0"
objects:
  - id: area-niigata
    kind: area
    name: Niigata Prefecture
    spec:
      area_type: prefecture
  - id: harvest-bad
    kind: crop_harvest
    name: Harvest Missing Crop Type
    attributes:
      owner: area-niigata
    spec:
      harvest_t: 100
`
	p := parser.NewParserWithSchema(setup.Schema)
	g, err := p.Parse([]byte(yaml))
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}

	// Missing required properties are reported as warnings; the finding must
	// mention the missing crop_type property.
	result := setup.Validation.Validate(g, nil)
	found := false
	for _, f := range result.Findings {
		if strings.Contains(f.Message, "crop_type") {
			found = true
		}
	}
	if !found {
		t.Error("expected a finding mentioning crop_type")
	}
}
