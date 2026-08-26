package extension

import (
	"testing"

	"IACForge/src/schema"
)

func TestNewSetup(t *testing.T) {
	setup, err := NewSetup("")
	if err != nil {
		t.Fatalf("NewSetup failed: %v", err)
	}
	if setup.Schema == nil {
		t.Fatal("expected non-nil schema")
	}
	if setup.Validation == nil {
		t.Fatal("expected non-nil validation engine")
	}
	if setup.Manager == nil {
		t.Fatal("expected non-nil extension manager")
	}
	if !setup.Schema.HasEntityKind("server") {
		t.Error("core entity kind 'server' missing from setup schema")
	}
	if !setup.Manager.IsLoaded() {
		t.Error("expected extension manager to be loaded")
	}
}

func TestNewSetupEmptyExtDir(t *testing.T) {
	dir := t.TempDir()
	setup, err := NewSetup(dir)
	if err != nil {
		t.Fatalf("NewSetup with empty dir failed: %v", err)
	}
	if !setup.Manager.IsLoaded() {
		t.Error("expected extension manager to be loaded")
	}
}

func TestNewSetupNonexistentExtDir(t *testing.T) {
	if _, err := NewSetup("/nonexistent/path"); err == nil {
		t.Fatal("expected error for nonexistent extension directory")
	}
}

func TestRegisterBuiltin(t *testing.T) {
	ext := &Extension{
		Manifest: &Manifest{
			ID:              "test-builtin-setup",
			Name:            "Test Builtin",
			Version:         "1.0.0",
			Namespace:       "testns",
			ExtensionPoints: []string{string(ExtensionPointEntityKinds)},
		},
		EntityKinds: []EntityKindContribution{
			{
				Kind: "test_builtin_setup_kind",
				Definition: &schema.EntityKindDefinition{
					Description: "Test builtin kind",
				},
			},
		},
	}
	RegisterBuiltin(ext)
	RegisterBuiltin(ext) // registering twice must be a no-op

	setup, err := NewSetup("")
	if err != nil {
		t.Fatalf("NewSetup failed: %v", err)
	}
	if !setup.Schema.HasEntityKind("test_builtin_setup_kind") {
		t.Error("registered builtin kind not in setup schema")
	}
	if got := len(BuiltinExtensions()); got != 1 {
		t.Errorf("expected exactly 1 builtin extension registered, got %d", got)
	}
}
