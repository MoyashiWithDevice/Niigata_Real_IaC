package aws

import (
	"path/filepath"
	"testing"

	"IACForge/src/core"
	"IACForge/src/core/types"
	"IACForge/src/parser"
)

func TestExtensionFactory(t *testing.T) {
	ext := Extension()
	if ext == nil {
		t.Fatal("Extension() returned nil")
	}
	if ext.Manifest == nil {
		t.Fatal("Extension() manifest is nil")
	}
	if ext.Manifest.ID != "iacforge.aws" {
		t.Errorf("expected manifest ID iacforge.aws, got %q", ext.Manifest.ID)
	}
	if ext.Manifest.Namespace != "aws" {
		t.Errorf("expected namespace aws, got %q", ext.Manifest.Namespace)
	}
	if got := len(ext.EntityKinds); got != len(AllKinds()) {
		t.Errorf("expected %d entity kinds, got %d", len(AllKinds()), got)
	}
	wantRels := len(RelationTypeDefinitions()) + len(AugmentDefinitions())
	if got := len(ext.RelationTypes); got != wantRels {
		t.Errorf("expected %d relation types, got %d", wantRels, got)
	}
}

func TestExtensionDeclaresRootKinds(t *testing.T) {
	ext := Extension()
	found := false
	for _, k := range ext.RootKinds {
		if k == Organization {
			found = true
		}
	}
	if !found {
		t.Error("aws.organization must be declared as a root kind")
	}
}

func TestExtensionLoadsViaSetup(t *testing.T) {
	s, v, _ := newTestSetup(t)

	if !s.HasEntityKind(EC2) {
		t.Error("aws.ec2 kind missing from setup schema")
	}
	if !s.HasEntityKind(VPC) {
		t.Error("aws.vpc kind missing from setup schema")
	}

	if _, ok := s.GetRelationTypeDef(Subscribes); !ok {
		t.Error("aws.subscribes relation type missing from setup schema")
	}

	// Core relation types must be augmented with AWS participant kinds.
	belongsTo, ok := s.GetRelationTypeDef(types.BelongsTo)
	if !ok {
		t.Fatal("belongs_to missing from setup schema")
	}
	if !containsKind(belongsTo.Participants.SourceKinds, VPC) {
		t.Error("belongs_to was not augmented with aws participant kinds")
	}

	// aws.organization must be granted root authority.
	if !v.IsAllowedRootKind(Organization) {
		t.Error("aws.organization must be a root-authorized kind")
	}
}

func TestSampleModelLoadValidateRoundTrip(t *testing.T) {
	s, v, _ := newTestSetup(t)

	path := filepath.Join("testdata", "aws-example.yaml")
	g, err := parser.NewParserWithSchema(s).Load(path)
	if err != nil {
		t.Fatalf("Load %s failed: %v", path, err)
	}

	result := v.Validate(g, nil)
	if !result.Passed {
		t.Errorf("sample AWS model must validate without errors, got:\n%s", resultSummary(result))
	}

	data, err := parser.NewSerializerWithSchema(s).Serialize(g)
	if err != nil {
		t.Fatalf("Serialize failed: %v", err)
	}

	rg, err := parser.NewParserWithSchema(s).Parse(data)
	if err != nil {
		t.Fatalf("re-parse of serialized model failed: %v\n%s", err, string(data))
	}

	if got, want := len(g.Entities()), len(rg.Entities()); got != want {
		t.Errorf("entity count mismatch after round-trip: got %d, want %d", got, want)
	}

	// The serializer normalizes flat ownership into nested structure, so the
	// re-parsed graph regenerates auto-relations from nesting. The explicit
	// relations must survive, and entity/ownership data must be preserved.
	for _, relID := range []string{"rel-ec2-subnet", "rel-cw-ec2", "rel-sns-sqs"} {
		if _, ok := rg.GetRelation(relID); !ok {
			t.Errorf("relation %q missing after round-trip", relID)
		}
	}

	// Spot check a few entities survive the round-trip.
	for _, id := range []string{"org-01", "vpc-01", "ec2-01", "s3-01"} {
		if _, ok := rg.GetEntity(id); !ok {
			t.Errorf("entity %q missing after round-trip", id)
		}
	}
	if e, ok := rg.GetEntity("ec2-01"); ok && e.Kind != EC2 {
		t.Errorf("ec2-01 kind mismatch after round-trip: got %q", e.Kind)
	}
	if e, ok := rg.GetEntity("subnet-01"); ok && e.Owner != "vpc-01" {
		t.Errorf("subnet-01 owner lost after round-trip: got %q", e.Owner)
	}

	// The round-tripped graph must also validate.
	rres := v.Validate(rg, nil)
	if !rres.Passed {
		t.Errorf("round-tripped graph must validate without errors, got:\n%s", resultSummary(rres))
	}
}

func TestSampleModelRelations(t *testing.T) {
	s, _, _ := newTestSetup(t)

	path := filepath.Join("testdata", "aws-example.yaml")
	g, err := parser.NewParserWithSchema(s).Load(path)
	if err != nil {
		t.Fatalf("Load %s failed: %v", path, err)
	}

	relTypes := make(map[core.RelationType]int)
	for _, r := range g.Relations() {
		relTypes[r.Type]++
	}
	if relTypes[types.BelongsTo] == 0 {
		t.Error("sample model should contain a belongs_to relation")
	}
	if relTypes[types.Monitors] == 0 {
		t.Error("sample model should contain a monitors relation")
	}
	if relTypes[Subscribes] == 0 {
		t.Error("sample model should contain an aws.subscribes relation")
	}
}
