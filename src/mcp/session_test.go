package mcp

import (
	"testing"

	"github.com/bababa/Niigata_Real_IaC/src/extension"
)

func TestSessionExtensionManager(t *testing.T) {
	sm := NewSessionManager()
	sd := sm.GetOrCreate("ext-session")

	if sd.Extensions == nil {
		t.Fatal("expected non-nil extension manager in session data")
	}
	if !sd.Extensions.IsLoaded() {
		t.Error("expected extension manager to be loaded")
	}
	if sd.Schema == nil {
		t.Fatal("expected non-nil schema in session data")
	}
	if !sd.Schema.HasEntityKind("area") {
		t.Error("expected core kind 'area' in session schema")
	}
	if sd.Validation == nil {
		t.Fatal("expected non-nil validation engine in session data")
	}
	if _, ok := sd.Extensions.GetExtensionPoint(extension.ExtensionPointEntityKinds); !ok {
		t.Error("expected entity kinds extension point to be registered")
	}
	if _, ok := sd.Extensions.GetExtensionPoint(extension.ExtensionPointRelationTypes); !ok {
		t.Error("expected relation types extension point to be registered")
	}
	if _, ok := sd.Extensions.GetExtensionPoint(extension.ExtensionPointValidationRules); !ok {
		t.Error("expected validation rules extension point to be registered")
	}
	if _, ok := sd.Extensions.GetExtensionPoint(extension.ExtensionPointRenderers); !ok {
		t.Error("expected renderers extension point to be registered")
	}
}

func TestSessionGetOrCreateFallsBackOnInvalidExtensionDir(t *testing.T) {
	t.Setenv("NIIGATA_EXTENSIONS", t.TempDir()+"/missing-plugins")

	sm := NewSessionManager()
	sd := sm.GetOrCreate("fallback-session")

	if sd == nil {
		t.Fatal("expected non-nil session data when extension directory is invalid")
	}
	if !sd.Schema.HasEntityKind("species") {
		t.Error("expected core kind 'species' in fallback session schema")
	}
	if sd.Schema.HasEntityKind("aws.ec2") {
		t.Error("expected aws.ec2 kind to be absent from fallback session schema")
	}
	if !sd.Extensions.IsLoaded() {
		t.Error("expected fallback extension manager to be loaded")
	}
}
