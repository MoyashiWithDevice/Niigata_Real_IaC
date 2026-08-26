package extension

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"IACForge/src/core"
	"IACForge/src/renderer"
	"IACForge/src/schema"
	"IACForge/src/validation"
	"IACForge/src/view"
)

func newTestExtension(id, namespace string, deps []string) *Extension {
	return &Extension{
		Manifest: &Manifest{
			ID:              id,
			Name:            id + " Name",
			Version:         "1.0.0",
			Namespace:       namespace,
			Dependencies:    deps,
			ExtensionPoints: []string{string(ExtensionPointEntityKinds)},
		},
		EntityKinds: []EntityKindContribution{
			{
				Kind: core.EntityKind("ext_" + id),
				Definition: &schema.EntityKindDefinition{
					Description: "Test entity kind from " + id,
				},
			},
		},
	}
}

func newTestRelationExtension(id, namespace string) *Extension {
	return &Extension{
		Manifest: &Manifest{
			ID:              id,
			Name:            id + " Name",
			Version:         "1.0.0",
			Namespace:       namespace,
			ExtensionPoints: []string{string(ExtensionPointRelationTypes)},
		},
		RelationTypes: []RelationTypeContribution{
			{
				Type: core.RelationType("ext_" + id),
				Definition: &schema.RelationTypeDefinition{
					Direction:   schema.DirectionDirected,
					Description: "Test relation type from " + id,
				},
			},
		},
	}
}

// --- Manager Tests ---

func TestNewManager(t *testing.T) {
	m := NewManager()
	if m == nil {
		t.Fatal("NewManager returned nil")
	}
	if m.IsLoaded() {
		t.Error("new manager should not be loaded")
	}
	if len(m.Extensions()) != 0 {
		t.Error("new manager should have no extensions")
	}
}

func TestRegisterExtension(t *testing.T) {
	m := NewManager()
	ext := newTestExtension("test-ext", "test", nil)

	if err := m.Register(ext); err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	if len(m.Extensions()) != 1 {
		t.Errorf("expected 1 extension, got %d", len(m.Extensions()))
	}

	got, ok := m.GetExtension("test-ext")
	if !ok {
		t.Fatal("extension not found after registration")
	}
	if got.Manifest.ID != "test-ext" {
		t.Errorf("expected extension ID 'test-ext', got %q", got.Manifest.ID)
	}
}

func TestRegisterNilExtension(t *testing.T) {
	m := NewManager()
	if err := m.Register(nil); err == nil {
		t.Fatal("expected error when registering nil extension")
	}
}

func TestRegisterNilManifest(t *testing.T) {
	m := NewManager()
	if err := m.Register(&Extension{}); err == nil {
		t.Fatal("expected error when registering extension with nil manifest")
	}
}

func TestRegisterEmptyID(t *testing.T) {
	m := NewManager()
	ext := &Extension{Manifest: &Manifest{Name: "test"}}
	if err := m.Register(ext); err == nil {
		t.Fatal("expected error when registering extension with empty ID")
	}
}

func TestRegisterDuplicate(t *testing.T) {
	m := NewManager()
	ext1 := newTestExtension("test-ext", "test", nil)
	ext2 := newTestExtension("test-ext", "test2", nil)

	if err := m.Register(ext1); err != nil {
		t.Fatalf("first Register failed: %v", err)
	}

	if err := m.Register(ext2); err == nil {
		t.Fatal("expected error when registering duplicate extension")
	}
}

func TestRegisterCoreEntityKindConflict(t *testing.T) {
	m := NewManager()
	ext := &Extension{
		Manifest: &Manifest{
			ID:        "conflict-ext",
			Name:      "Conflict Extension",
			Version:   "1.0.0",
			Namespace: "conflict",
		},
		EntityKinds: []EntityKindContribution{
			{
				Kind: core.EntityKind("server"), // core kind
				Definition: &schema.EntityKindDefinition{
					Description: "Conflicts with core",
				},
			},
		},
	}

	if err := m.Register(ext); err == nil {
		t.Fatal("expected error when extension redefines core entity kind")
	}
}

func TestRegisterCoreRelationTypeConflict(t *testing.T) {
	m := NewManager()
	ext := &Extension{
		Manifest: &Manifest{
			ID:        "conflict-ext",
			Name:      "Conflict Extension",
			Version:   "1.0.0",
			Namespace: "conflict",
		},
		RelationTypes: []RelationTypeContribution{
			{
				Type: core.RelationType("connects"), // core type
				Definition: &schema.RelationTypeDefinition{
					Direction: schema.DirectionDirected,
				},
			},
		},
	}

	if err := m.Register(ext); err == nil {
		t.Fatal("expected error when extension redefines core relation type")
	}
}

func TestRegisterNamespaceConflict(t *testing.T) {
	m := NewManager()
	ext1 := newTestExtension("ext1", "shared-ns", nil)
	ext2 := newTestExtension("ext2", "shared-ns", nil)

	if err := m.Register(ext1); err != nil {
		t.Fatalf("first Register failed: %v", err)
	}

	if err := m.Register(ext2); err == nil {
		t.Fatal("expected error when two extensions share the same namespace")
	}
}

func TestLoadAllWithDependencies(t *testing.T) {
	m := NewManager()

	extA := newTestExtension("ext-a", "ns-a", nil)
	extB := newTestExtension("ext-b", "ns-b", []string{"ext-a"})
	extC := newTestExtension("ext-c", "ns-c", []string{"ext-a", "ext-b"})

	if err := m.Register(extA); err != nil {
		t.Fatalf("Register ext-a failed: %v", err)
	}
	if err := m.Register(extB); err != nil {
		t.Fatalf("Register ext-b failed: %v", err)
	}
	if err := m.Register(extC); err != nil {
		t.Fatalf("Register ext-c failed: %v", err)
	}

	if err := m.LoadAll(); err != nil {
		t.Fatalf("LoadAll failed: %v", err)
	}

	if !m.IsLoaded() {
		t.Error("manager should be loaded after LoadAll")
	}

	order := m.LoadOrder()
	if len(order) != 3 {
		t.Fatalf("expected 3 extensions in load order, got %d", len(order))
	}

	// ext-a must come before ext-b and ext-c
	idxA := indexOf(order, "ext-a")
	idxB := indexOf(order, "ext-b")
	idxC := indexOf(order, "ext-c")

	if idxA >= idxB {
		t.Error("ext-a should come before ext-b in load order")
	}
	if idxA >= idxC {
		t.Error("ext-a should come before ext-c in load order")
	}
	if idxB >= idxC {
		t.Error("ext-b should come before ext-c in load order")
	}
}

func TestLoadAllMissingDependency(t *testing.T) {
	m := NewManager()

	ext := newTestExtension("ext-a", "ns-a", []string{"ext-missing"})
	if err := m.Register(ext); err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	if err := m.LoadAll(); err == nil {
		t.Fatal("expected error when dependency is missing")
	}
}

func TestLoadAllCircularDependency(t *testing.T) {
	m := NewManager()

	extA := newTestExtension("ext-a", "ns-a", []string{"ext-b"})
	extB := newTestExtension("ext-b", "ns-b", []string{"ext-a"})

	if err := m.Register(extA); err != nil {
		t.Fatalf("Register ext-a failed: %v", err)
	}
	if err := m.Register(extB); err != nil {
		t.Fatalf("Register ext-b failed: %v", err)
	}

	if err := m.LoadAll(); err == nil {
		t.Fatal("expected error for circular dependency")
	}
}

func TestLoadConvenienceMethod(t *testing.T) {
	m := NewManager()
	ext := newTestExtension("test-ext", "test", nil)

	if err := m.Load(ext); err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if !m.IsLoaded() {
		t.Error("manager should be loaded after Load")
	}
}

func TestNamespaces(t *testing.T) {
	m := NewManager()
	ext1 := newTestExtension("ext1", "alpha", nil)
	ext2 := newTestExtension("ext2", "beta", nil)
	ext3 := newTestExtension("ext3", "alpha", nil) // same namespace as ext1

	if err := m.Register(ext1); err != nil {
		t.Fatalf("Register ext1 failed: %v", err)
	}
	if err := m.Register(ext2); err != nil {
		t.Fatalf("Register ext2 failed: %v", err)
	}
	if err := m.Register(ext3); err == nil {
		t.Fatal("expected namespace conflict error for ext3")
	}

	ns := m.Namespaces()
	if len(ns) != 2 {
		t.Errorf("expected 2 namespaces, got %d", len(ns))
	}
	if ns[0] != "alpha" || ns[1] != "beta" {
		t.Errorf("unexpected namespaces: %v", ns)
	}
}

func TestExtensionPointRegistration(t *testing.T) {
	m := NewManager()
	s := schema.NewSchema("1.0", "1.0")
	ep := NewEntityKindsExtensionPoint(s)

	m.RegisterExtensionPoint(ep)

	got, ok := m.GetExtensionPoint(ExtensionPointEntityKinds)
	if !ok {
		t.Fatal("extension point not found")
	}
	if got.Type() != ExtensionPointEntityKinds {
		t.Errorf("expected type %q, got %q", ExtensionPointEntityKinds, got.Type())
	}
}

// --- EntityKindsExtensionPoint Tests ---

func TestEntityKindsExtensionPointRegister(t *testing.T) {
	s := schema.NewSchema("1.0", "1.0")
	ep := NewEntityKindsExtensionPoint(s)

	ext := &Extension{
		Manifest: &Manifest{ID: "test-ext", Namespace: "test"},
		EntityKinds: []EntityKindContribution{
			{
				Kind: "custom_vm",
				Definition: &schema.EntityKindDefinition{
					Description: "Custom VM kind",
				},
			},
		},
	}

	if err := ep.Register(ext); err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	if !s.HasEntityKind("custom_vm") {
		t.Error("entity kind not added to schema")
	}

	kinds := ep.GetEntityKindsByExtension("test-ext")
	if len(kinds) != 1 || kinds[0] != "custom_vm" {
		t.Errorf("unexpected kinds for extension: %v", kinds)
	}

	extID, ok := ep.GetExtensionForKind("custom_vm")
	if !ok || extID != "test-ext" {
		t.Errorf("expected extension ID 'test-ext', got %q", extID)
	}
}

func TestEntityKindsExtensionPointConflict(t *testing.T) {
	s := schema.NewSchema("1.0", "1.0")
	s.AddEntityKind("server", &schema.EntityKindDefinition{Description: "Core server"})
	ep := NewEntityKindsExtensionPoint(s)

	ext := &Extension{
		Manifest: &Manifest{ID: "test-ext", Namespace: "test"},
		EntityKinds: []EntityKindContribution{
			{
				Kind: "server", // conflicts with existing
				Definition: &schema.EntityKindDefinition{
					Description: "Conflicts with server",
				},
			},
		},
	}

	if err := ep.Register(ext); err == nil {
		t.Fatal("expected conflict error when registering duplicate entity kind")
	}
}

func TestEntityKindsAllExtendedKinds(t *testing.T) {
	s := schema.NewSchema("1.0", "1.0")
	ep := NewEntityKindsExtensionPoint(s)

	ext1 := &Extension{
		Manifest: &Manifest{ID: "ext1", Namespace: "ns1"},
		EntityKinds: []EntityKindContribution{
			{Kind: "kind_a", Definition: &schema.EntityKindDefinition{}},
			{Kind: "kind_b", Definition: &schema.EntityKindDefinition{}},
		},
	}
	ext2 := &Extension{
		Manifest: &Manifest{ID: "ext2", Namespace: "ns2"},
		EntityKinds: []EntityKindContribution{
			{Kind: "kind_c", Definition: &schema.EntityKindDefinition{}},
		},
	}

	if err := ep.Register(ext1); err != nil {
		t.Fatalf("failed to register: %v", err)
	}
	if err := ep.Register(ext2); err != nil {
		t.Fatalf("failed to register: %v", err)
	}

	allKinds := ep.AllExtendedKinds()
	if len(allKinds) != 3 {
		t.Errorf("expected 3 extended kinds, got %d", len(allKinds))
	}
}

// --- RelationTypesExtensionPoint Tests ---

func TestRelationTypesExtensionPointRegister(t *testing.T) {
	s := schema.NewSchema("1.0", "1.0")
	ep := NewRelationTypesExtensionPoint(s)

	ext := &Extension{
		Manifest: &Manifest{ID: "test-ext", Namespace: "test"},
		RelationTypes: []RelationTypeContribution{
			{
				Type: "custom_connects",
				Definition: &schema.RelationTypeDefinition{
					Direction: schema.DirectionDirected,
				},
			},
		},
	}

	if err := ep.Register(ext); err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	if !s.HasRelationType("custom_connects") {
		t.Error("relation type not added to schema")
	}

	types := ep.GetRelationTypesByExtension("test-ext")
	if len(types) != 1 || types[0] != "custom_connects" {
		t.Errorf("unexpected types for extension: %v", types)
	}

	extID, ok := ep.GetExtensionForType("custom_connects")
	if !ok || extID != "test-ext" {
		t.Errorf("expected extension ID 'test-ext', got %q", extID)
	}
}

func TestRelationTypesExtensionPointConflict(t *testing.T) {
	s := schema.NewSchema("1.0", "1.0")
	s.AddRelationType("connects", &schema.RelationTypeDefinition{Direction: schema.DirectionSymmetric})
	ep := NewRelationTypesExtensionPoint(s)

	ext := &Extension{
		Manifest: &Manifest{ID: "test-ext", Namespace: "test"},
		RelationTypes: []RelationTypeContribution{
			{
				Type:       "connects",
				Definition: &schema.RelationTypeDefinition{Direction: schema.DirectionDirected},
			},
		},
	}

	if err := ep.Register(ext); err == nil {
		t.Fatal("expected conflict error when registering duplicate relation type")
	}
}

func TestRelationTypesExtensionPointAugment(t *testing.T) {
	s := schema.NewSchema("1.0", "1.0")
	s.AddRelationType("belongs_to", &schema.RelationTypeDefinition{
		Direction: schema.DirectionDirected,
		Participants: &schema.ParticipantConstraints{
			SourceKinds: []core.EntityKind{"vm", "container"},
			TargetKinds: []core.EntityKind{"region"},
		},
	})
	ep := NewRelationTypesExtensionPoint(s)

	ext := &Extension{
		Manifest: &Manifest{ID: "aws-ext", Namespace: "aws"},
		RelationTypes: []RelationTypeContribution{
			{
				Type:    "belongs_to",
				Augment: true,
				Definition: &schema.RelationTypeDefinition{
					Participants: &schema.ParticipantConstraints{
						SourceKinds: []core.EntityKind{"aws.subnet"},
						TargetKinds: []core.EntityKind{"aws.vpc", "vm"},
					},
				},
			},
		},
	}

	if err := ep.Register(ext); err != nil {
		t.Fatalf("Register with Augment failed: %v", err)
	}

	def, ok := s.GetRelationTypeDef("belongs_to")
	if !ok {
		t.Fatal("belongs_to should still exist in schema")
	}

	// Source kinds must include the existing core kinds plus the augmented kind.
	if !containsKind(def.Participants.SourceKinds, "aws.subnet") {
		t.Errorf("belongs_to source kinds should include aws.subnet, got %v", def.Participants.SourceKinds)
	}
	if !containsKind(def.Participants.SourceKinds, "vm") {
		t.Errorf("belongs_to source kinds should retain existing vm, got %v", def.Participants.SourceKinds)
	}
	// Target kinds must include the augmented kinds without duplicates.
	if !containsKind(def.Participants.TargetKinds, "aws.vpc") {
		t.Errorf("belongs_to target kinds should include aws.vpc, got %v", def.Participants.TargetKinds)
	}
	if countKind(def.Participants.TargetKinds, "vm") != 1 {
		t.Errorf("belongs_to target kinds should contain vm exactly once, got %v", def.Participants.TargetKinds)
	}

	// The augmenting extension should be tracked for the type.
	extID, ok := ep.GetExtensionForType("belongs_to")
	if !ok || extID != "aws-ext" {
		t.Errorf("expected extension 'aws-ext' for belongs_to, got %q", extID)
	}
}

func TestRelationTypesExtensionPointAugmentNilDefinition(t *testing.T) {
	s := schema.NewSchema("1.0", "1.0")
	s.AddRelationType("hosts", &schema.RelationTypeDefinition{Direction: schema.DirectionDirected})
	ep := NewRelationTypesExtensionPoint(s)

	ext := &Extension{
		Manifest: &Manifest{ID: "bad-ext", Namespace: "bad"},
		RelationTypes: []RelationTypeContribution{
			{Type: "hosts", Augment: true, Definition: nil},
		},
	}

	if err := ep.Register(ext); err == nil {
		t.Fatal("expected error when augmenting with nil definition")
	}
}

func TestAugmentCoreTypeViaManager(t *testing.T) {
	s := schema.CoreSchema()
	m := NewManager()

	rtEP := NewRelationTypesExtensionPoint(s)
	m.RegisterExtensionPoint(rtEP)

	ext := &Extension{
		Manifest: &Manifest{
			ID:              "aws-ext",
			Name:            "AWS Extension",
			Version:         "1.0.0",
			Namespace:       "aws",
			ExtensionPoints: []string{string(ExtensionPointRelationTypes)},
		},
		RelationTypes: []RelationTypeContribution{
			{
				Type:    "belongs_to",
				Augment: true,
				Definition: &schema.RelationTypeDefinition{
					Participants: &schema.ParticipantConstraints{
						SourceKinds: []core.EntityKind{"aws.subnet"},
						TargetKinds: []core.EntityKind{"aws.vpc"},
					},
				},
			},
		},
	}

	if err := m.Load(ext); err != nil {
		t.Fatalf("Load of augmenting extension failed: %v", err)
	}

	def, ok := s.GetRelationTypeDef("belongs_to")
	if !ok {
		t.Fatal("belongs_to should exist in core schema")
	}
	if !containsKind(def.Participants.SourceKinds, "aws.subnet") {
		t.Errorf("belongs_to source kinds should include aws.subnet, got %v", def.Participants.SourceKinds)
	}
	if !containsKind(def.Participants.TargetKinds, "aws.vpc") {
		t.Errorf("belongs_to target kinds should include aws.vpc, got %v", def.Participants.TargetKinds)
	}
}

func containsKind(kinds []core.EntityKind, k string) bool {
	for _, kind := range kinds {
		if string(kind) == k {
			return true
		}
	}
	return false
}

func countKind(kinds []core.EntityKind, k string) int {
	count := 0
	for _, kind := range kinds {
		if string(kind) == k {
			count++
		}
	}
	return count
}

func TestRelationTypesExtensionPointAugmentDirectionConflict(t *testing.T) {
	s := schema.NewSchema("1.0", "1.0")
	s.AddRelationType("belongs_to", &schema.RelationTypeDefinition{
		Direction: schema.DirectionDirected,
		Participants: &schema.ParticipantConstraints{
			SourceKinds: []core.EntityKind{"vm"},
			TargetKinds: []core.EntityKind{"region"},
		},
	})
	ep := NewRelationTypesExtensionPoint(s)

	ext := &Extension{
		Manifest: &Manifest{ID: "aws-ext", Namespace: "aws"},
		RelationTypes: []RelationTypeContribution{
			{
				Type:    "belongs_to",
				Augment: true,
				Definition: &schema.RelationTypeDefinition{
					Direction: schema.DirectionSymmetric,
					Participants: &schema.ParticipantConstraints{
						SourceKinds: []core.EntityKind{"aws.subnet"},
					},
				},
			},
		},
	}

	err := ep.Register(ext)
	if err == nil {
		t.Fatal("expected error when augmenting changes direction")
	}
	if !errors.Is(err, ErrInvalidExtension) {
		t.Errorf("expected ErrInvalidExtension, got %v", err)
	}
}

func TestRelationTypesExtensionPointAugmentPropertiesRejected(t *testing.T) {
	s := schema.NewSchema("1.0", "1.0")
	s.AddRelationType("belongs_to", &schema.RelationTypeDefinition{
		Direction: schema.DirectionDirected,
		Participants: &schema.ParticipantConstraints{
			SourceKinds: []core.EntityKind{"vm"},
			TargetKinds: []core.EntityKind{"region"},
		},
	})
	ep := NewRelationTypesExtensionPoint(s)

	ext := &Extension{
		Manifest: &Manifest{ID: "aws-ext", Namespace: "aws"},
		RelationTypes: []RelationTypeContribution{
			{
				Type:    "belongs_to",
				Augment: true,
				Definition: &schema.RelationTypeDefinition{
					Participants: &schema.ParticipantConstraints{
						SourceKinds: []core.EntityKind{"aws.subnet"},
					},
					Properties: []schema.PropertyDefinition{
						{Name: "scope", Type: schema.PropertyTypeString},
					},
				},
			},
		},
	}

	if err := ep.Register(ext); err == nil {
		t.Fatal("expected error when augmenting adds properties")
	}
}

func TestRelationTypesExtensionPointAugmentMinMaxConflict(t *testing.T) {
	s := schema.NewSchema("1.0", "1.0")
	s.AddRelationType("belongs_to", &schema.RelationTypeDefinition{
		Direction: schema.DirectionDirected,
		Participants: &schema.ParticipantConstraints{
			SourceKinds:     []core.EntityKind{"vm"},
			TargetKinds:     []core.EntityKind{"region"},
			MinParticipants: 2,
			MaxParticipants: 2,
		},
	})
	ep := NewRelationTypesExtensionPoint(s)

	ext := &Extension{
		Manifest: &Manifest{ID: "aws-ext", Namespace: "aws"},
		RelationTypes: []RelationTypeContribution{
			{
				Type:    "belongs_to",
				Augment: true,
				Definition: &schema.RelationTypeDefinition{
					Participants: &schema.ParticipantConstraints{
						SourceKinds:     []core.EntityKind{"aws.subnet"},
						MaxParticipants: 4,
					},
				},
			},
		},
	}

	if err := ep.Register(ext); err == nil {
		t.Fatal("expected error when augmenting changes max participants")
	}
}

func TestRelationTypesExtensionPointAugmentDoesNotMutateOriginal(t *testing.T) {
	s := schema.NewSchema("1.0", "1.0")
	original := &schema.RelationTypeDefinition{
		Direction: schema.DirectionDirected,
		Participants: &schema.ParticipantConstraints{
			SourceKinds: []core.EntityKind{"vm", "container"},
			TargetKinds: []core.EntityKind{"region"},
		},
	}
	s.AddRelationType("belongs_to", original)
	ep := NewRelationTypesExtensionPoint(s)

	ext := &Extension{
		Manifest: &Manifest{ID: "aws-ext", Namespace: "aws"},
		RelationTypes: []RelationTypeContribution{
			{
				Type:    "belongs_to",
				Augment: true,
				Definition: &schema.RelationTypeDefinition{
					Participants: &schema.ParticipantConstraints{
						SourceKinds: []core.EntityKind{"aws.subnet"},
						TargetKinds: []core.EntityKind{"aws.vpc"},
					},
				},
			},
		},
	}

	if err := ep.Register(ext); err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	// The original definition must be untouched.
	if len(original.Participants.SourceKinds) != 2 {
		t.Errorf("original source kinds should be unchanged, got %v", original.Participants.SourceKinds)
	}
	if containsKind(original.Participants.SourceKinds, "aws.subnet") {
		t.Error("original source kinds should not contain augmented kind")
	}
	if containsKind(original.Participants.TargetKinds, "aws.vpc") {
		t.Error("original target kinds should not contain augmented kind")
	}

	// The stored definition must reflect the merged result.
	def, ok := s.GetRelationTypeDef("belongs_to")
	if !ok {
		t.Fatal("belongs_to should exist in schema")
	}
	if def == original {
		t.Error("stored definition should be a copy, not the original pointer")
	}
	if !containsKind(def.Participants.SourceKinds, "aws.subnet") {
		t.Errorf("stored source kinds should include aws.subnet, got %v", def.Participants.SourceKinds)
	}
}

// --- ValidationRulesExtensionPoint Tests ---

func TestValidationRulesExtensionPointRegister(t *testing.T) {
	s := schema.NewSchema("1.0", "1.0")
	vEngine := validation.NewEngine(s)
	ep := NewValidationRulesExtensionPoint(vEngine)

	ext := &Extension{
		Manifest: &Manifest{ID: "test-ext", Namespace: "test"},
		ValidationRules: []ValidationRuleContribution{
			{
				Rule: &validation.Rule{
					ID:       "custom-rule-1",
					Name:     "Custom Rule 1",
					Severity: validation.SeverityWarning,
				},
				Fn: func(ctx *validation.Context) []validation.Finding {
					return nil
				},
			},
		},
	}

	if err := ep.Register(ext); err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	rules := ep.GetRulesByExtension("test-ext")
	if len(rules) != 1 || rules[0] != "custom-rule-1" {
		t.Errorf("unexpected rules for extension: %v", rules)
	}

	extID, ok := ep.GetExtensionForRule("custom-rule-1")
	if !ok || extID != "test-ext" {
		t.Errorf("expected extension ID 'test-ext', got %q", extID)
	}
}

func TestValidationRulesExtensionPointNilRule(t *testing.T) {
	s := schema.NewSchema("1.0", "1.0")
	vEngine := validation.NewEngine(s)
	ep := NewValidationRulesExtensionPoint(vEngine)

	ext := &Extension{
		Manifest: &Manifest{ID: "test-ext", Namespace: "test"},
		ValidationRules: []ValidationRuleContribution{
			{Rule: nil, Fn: nil},
		},
	}

	if err := ep.Register(ext); err == nil {
		t.Fatal("expected error when registering nil validation rule")
	}
}

func TestValidationRulesAllExtendedRules(t *testing.T) {
	s := schema.NewSchema("1.0", "1.0")
	vEngine := validation.NewEngine(s)
	ep := NewValidationRulesExtensionPoint(vEngine)

	ext1 := &Extension{
		Manifest: &Manifest{ID: "ext1", Namespace: "ns1"},
		ValidationRules: []ValidationRuleContribution{
			{
				Rule: &validation.Rule{ID: "rule-a", Name: "Rule A", Severity: validation.SeverityWarning},
				Fn:   func(ctx *validation.Context) []validation.Finding { return nil },
			},
			{
				Rule: &validation.Rule{ID: "rule-b", Name: "Rule B", Severity: validation.SeverityError},
				Fn:   func(ctx *validation.Context) []validation.Finding { return nil },
			},
		},
	}

	if err := ep.Register(ext1); err != nil {
		t.Fatalf("failed to register: %v", err)
	}

	if extID, ok := ep.GetExtensionForRule("rule-a"); !ok || extID != "ext1" {
		t.Errorf("expected ext1 for rule-a, got %q (%v)", extID, ok)
	}
	if extID, ok := ep.GetExtensionForRule("rule-b"); !ok || extID != "ext1" {
		t.Errorf("expected ext1 for rule-b, got %q (%v)", extID, ok)
	}
	if _, ok := ep.GetExtensionForRule("unknown"); ok {
		t.Error("expected no extension for unknown rule")
	}
}

// --- RootKindsExtensionPoint Tests ---

func TestRootKindsExtensionPointRegister(t *testing.T) {
	s := schema.NewSchema("1.0", "1.0")
	vEngine := validation.NewEngine(s)
	ep := NewRootKindsExtensionPoint(vEngine)

	ext := &Extension{
		Manifest: &Manifest{ID: "test-ext", Namespace: "test"},
		RootKinds: []core.EntityKind{
			"aws.organization",
			"aws.account",
		},
	}

	if err := ep.Register(ext); err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	if !vEngine.IsAllowedRootKind("aws.organization") {
		t.Error("aws.organization was not granted root authority")
	}
	if !vEngine.IsAllowedRootKind("aws.account") {
		t.Error("aws.account was not granted root authority")
	}
	if vEngine.IsAllowedRootKind("aws.vpc") {
		t.Error("aws.vpc must not be granted root authority")
	}

	kinds := ep.GetRootKindsByExtension("test-ext")
	if len(kinds) != 2 {
		t.Errorf("expected 2 root kinds for test-ext, got %v", kinds)
	}

	extID, ok := ep.GetExtensionForRootKind("aws.organization")
	if !ok || extID != "test-ext" {
		t.Errorf("expected test-ext for aws.organization, got %q (%v)", extID, ok)
	}
}

func TestRootKindsExtensionPointDuplicate(t *testing.T) {
	s := schema.NewSchema("1.0", "1.0")
	vEngine := validation.NewEngine(s)
	ep := NewRootKindsExtensionPoint(vEngine)

	ext := &Extension{
		Manifest: &Manifest{ID: "test-ext", Namespace: "test"},
		RootKinds: []core.EntityKind{
			"aws.organization",
			"aws.organization",
		},
	}

	if err := ep.Register(ext); err != nil {
		t.Fatalf("Register failed: %v", err)
	}
	if got := len(ep.GetRootKindsByExtension("test-ext")); got != 1 {
		t.Errorf("expected 1 unique root kind after duplicate registration, got %d", got)
	}
}

func TestRootKindsExtensionPointNoRootKinds(t *testing.T) {
	s := schema.NewSchema("1.0", "1.0")
	vEngine := validation.NewEngine(s)
	ep := NewRootKindsExtensionPoint(vEngine)

	ext := &Extension{
		Manifest: &Manifest{ID: "test-ext", Namespace: "test"},
	}

	if err := ep.Register(ext); err != nil {
		t.Fatalf("Register failed: %v", err)
	}
	if got := len(ep.GetRootKindsByExtension("test-ext")); got != 0 {
		t.Errorf("expected no root kinds, got %v", got)
	}
}

// --- RendererExtensionPoint Tests ---

type mockRenderer struct {
	id     string
	name   string
	format string
}

func (m *mockRenderer) Render(v *view.ViewResult, opts *renderer.RenderOptions) (*renderer.Artifact, error) {
	return nil, nil
}
func (m *mockRenderer) ID() string     { return m.id }
func (m *mockRenderer) Name() string   { return m.name }
func (m *mockRenderer) Format() string { return m.format }

func TestRendererExtensionPointRegister(t *testing.T) {
	registry := NewRendererRegistry()
	ep := NewRendererExtensionPoint(registry)

	r := &mockRenderer{id: "svg-ext", name: "SVG Extension Renderer", format: "svg"}

	ext := &Extension{
		Manifest: &Manifest{ID: "test-ext", Namespace: "test"},
		Renderers: []RendererContribution{
			{Renderer: r},
		},
	}

	if err := ep.Register(ext); err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	got, ok := registry.Get("svg-ext")
	if !ok {
		t.Fatal("renderer not added to registry")
	}
	if got.Name() != "SVG Extension Renderer" {
		t.Errorf("unexpected renderer name: %q", got.Name())
	}

	renderers := ep.GetRenderersByExtension("test-ext")
	if len(renderers) != 1 || renderers[0] != "svg-ext" {
		t.Errorf("unexpected renderers for extension: %v", renderers)
	}

	extID, ok := ep.GetExtensionForRenderer("svg-ext")
	if !ok || extID != "test-ext" {
		t.Errorf("expected extension ID 'test-ext', got %q", extID)
	}
}

func TestRendererExtensionPointNilRenderer(t *testing.T) {
	registry := NewRendererRegistry()
	ep := NewRendererExtensionPoint(registry)

	ext := &Extension{
		Manifest: &Manifest{ID: "test-ext", Namespace: "test"},
		Renderers: []RendererContribution{
			{Renderer: nil},
		},
	}

	if err := ep.Register(ext); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got := len(ep.GetRenderersByExtension("test-ext")); got != 0 {
		t.Errorf("nil renderer should not be registered, got %d", got)
	}
}

func TestRendererRegistryAll(t *testing.T) {
	registry := NewRendererRegistry()

	r1 := &mockRenderer{id: "r1", name: "Renderer 1", format: "svg"}
	r2 := &mockRenderer{id: "r2", name: "Renderer 2", format: "json"}

	registry.Register(r1)
	registry.Register(r2)

	all := registry.All()
	if len(all) != 2 {
		t.Errorf("expected 2 renderers, got %d", len(all))
	}
}

// --- Integration: Manager + Extension Points ---

func TestIntegrationEntityKinds(t *testing.T) {
	s := schema.NewSchema("1.0", "1.0")
	m := NewManager()

	ep := NewEntityKindsExtensionPoint(s)
	m.RegisterExtensionPoint(ep)

	ext := newTestExtension("custom-infra", "custom", nil)
	if err := m.Load(ext); err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if !s.HasEntityKind("ext_custom-infra") {
		t.Error("entity kind not added to schema via integration")
	}
}

func TestIntegrationRelationTypes(t *testing.T) {
	s := schema.NewSchema("1.0", "1.0")
	m := NewManager()

	ep := NewRelationTypesExtensionPoint(s)
	m.RegisterExtensionPoint(ep)

	ext := newTestRelationExtension("custom-relation", "custom")
	if err := m.Load(ext); err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if !s.HasRelationType("ext_custom-relation") {
		t.Error("relation type not added to schema via integration")
	}
}

func TestIntegrationMultipleExtensions(t *testing.T) {
	s := schema.NewSchema("1.0", "1.0")
	m := NewManager()

	ekEP := NewEntityKindsExtensionPoint(s)
	rtEP := NewRelationTypesExtensionPoint(s)
	m.RegisterExtensionPoint(ekEP)
	m.RegisterExtensionPoint(rtEP)

	extA := newTestExtension("ext-a", "ns-a", nil)
	extB := newTestRelationExtension("ext-b", "ns-b")
	extC := newTestExtension("ext-c", "ns-c", []string{"ext-a"})

	if err := m.Register(extA); err != nil {
		t.Fatalf("failed to register: %v", err)
	}
	if err := m.Register(extB); err != nil {
		t.Fatalf("failed to register: %v", err)
	}
	if err := m.Register(extC); err != nil {
		t.Fatalf("failed to register: %v", err)
	}

	if err := m.LoadAll(); err != nil {
		t.Fatalf("LoadAll failed: %v", err)
	}

	if !s.HasEntityKind("ext_ext-a") {
		t.Error("ext-a entity kind not in schema")
	}
	if !s.HasRelationType("ext_ext-b") {
		t.Error("ext-b relation type not in schema")
	}
	if !s.HasEntityKind("ext_ext-c") {
		t.Error("ext-c entity kind not in schema")
	}
}

func TestIntegrationGraphIntegrity(t *testing.T) {
	s := schema.CoreSchema()
	m := NewManager()

	ekEP := NewEntityKindsExtensionPoint(s)
	m.RegisterExtensionPoint(ekEP)

	ext := &Extension{
		Manifest: &Manifest{ID: "safe-ext", Namespace: "safe"},
		EntityKinds: []EntityKindContribution{
			{
				Kind: "custom_server",
				Definition: &schema.EntityKindDefinition{
					Description: "A custom server kind",
				},
			},
		},
	}

	if err := m.Load(ext); err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	// Core kinds should still be in the schema
	coreKinds := []core.EntityKind{"server", "vm", "network", "firewall"}
	for _, k := range coreKinds {
		if !s.HasEntityKind(k) {
			t.Errorf("core kind %q should still be in schema", k)
		}
	}
}

// --- LoadFromDir Tests ---

func TestLoadAllIdempotent(t *testing.T) {
	s := schema.NewSchema("1.0", "1.0")
	m := NewManager()

	ekEP := NewEntityKindsExtensionPoint(s)
	rtEP := NewRelationTypesExtensionPoint(s)
	m.RegisterExtensionPoint(ekEP)
	m.RegisterExtensionPoint(rtEP)

	extA := newTestExtension("ext-a", "ns-a", nil)
	extB := newTestRelationExtension("ext-b", "ns-b")

	if err := m.Register(extA); err != nil {
		t.Fatalf("failed to register: %v", err)
	}
	if err := m.Register(extB); err != nil {
		t.Fatalf("failed to register: %v", err)
	}

	if err := m.LoadAll(); err != nil {
		t.Fatalf("first LoadAll failed: %v", err)
	}
	if !s.HasEntityKind("ext_ext-a") {
		t.Error("ext-a entity kind not in schema after first LoadAll")
	}

	// Second LoadAll must not re-apply already-loaded extensions.
	if err := m.LoadAll(); err != nil {
		t.Fatalf("second LoadAll failed: %v", err)
	}

	// A newly registered extension should still be applied by a subsequent LoadAll.
	extC := newTestExtension("ext-c", "ns-c", nil)
	if err := m.Register(extC); err != nil {
		t.Fatalf("Register failed: %v", err)
	}
	if err := m.LoadAll(); err != nil {
		t.Fatalf("LoadAll after registering new extension failed: %v", err)
	}
	if !s.HasEntityKind("ext_ext-c") {
		t.Error("ext-c entity kind not in schema after LoadAll on new extension")
	}
}

func TestLoadAllIdempotentWithLoadFromDir(t *testing.T) {
	s := schema.NewSchema("1.0", "1.0")
	m := NewManager()

	ekEP := NewEntityKindsExtensionPoint(s)
	m.RegisterExtensionPoint(ekEP)

	builtin := newTestExtension("builtin-a", "ns-a", nil)
	if err := m.Register(builtin); err != nil {
		t.Fatalf("Register builtin failed: %v", err)
	}

	// LoadFromDir internally calls LoadAll; the already-registered builtin must
	// not be re-applied (which would cause a "already defined" conflict).
	dir := t.TempDir()
	if err := m.LoadFromDir(dir); err != nil {
		t.Fatalf("LoadFromDir failed: %v", err)
	}
	if !s.HasEntityKind("ext_builtin-a") {
		t.Error("builtin entity kind not in schema after LoadFromDir")
	}
}

func TestLoadFromDirEmpty(t *testing.T) {
	dir := t.TempDir()
	m := NewManager()
	if err := m.LoadFromDir(dir); err != nil {
		t.Fatalf("LoadFromDir on empty directory failed: %v", err)
	}
	if len(m.Extensions()) != 0 {
		t.Errorf("expected 0 extensions, got %d", len(m.Extensions()))
	}
}

func TestLoadFromDirNonexistent(t *testing.T) {
	m := NewManager()
	if err := m.LoadFromDir("/nonexistent/path"); err == nil {
		t.Fatal("expected error for nonexistent directory")
	}
}

func TestLoadFromDirSkipsNonPlugins(t *testing.T) {
	dir := t.TempDir()
	// Create a non-.so file
	if err := os.WriteFile(filepath.Join(dir, "readme.txt"), []byte("not a plugin"), 0644); err != nil {
		t.Fatal(err)
	}

	m := NewManager()
	if err := m.LoadFromDir(dir); err != nil {
		t.Fatalf("LoadFromDir failed: %v", err)
	}
	if len(m.Extensions()) != 0 {
		t.Errorf("expected 0 extensions, got %d", len(m.Extensions()))
	}
}

func TestLoadFromDirSkipsSubdirs(t *testing.T) {
	dir := t.TempDir()
	// Create a subdirectory
	if err := os.Mkdir(filepath.Join(dir, "subdir"), 0755); err != nil {
		t.Fatal(err)
	}

	m := NewManager()
	if err := m.LoadFromDir(dir); err != nil {
		t.Fatalf("LoadFromDir failed: %v", err)
	}
	if len(m.Extensions()) != 0 {
		t.Errorf("expected 0 extensions, got %d", len(m.Extensions()))
	}
}

// --- Helper functions ---

func indexOf(slice []string, s string) int {
	for i, v := range slice {
		if v == s {
			return i
		}
	}
	return -1
}
