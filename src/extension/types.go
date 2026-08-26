package extension

import (
	"IACForge/src/core"
	"IACForge/src/renderer"
	"IACForge/src/schema"
	"IACForge/src/validation"
)

// Manifest holds machine-readable metadata for an extension.
type Manifest struct {
	ID              string   `yaml:"id"`
	Name            string   `yaml:"name"`
	Version         string   `yaml:"version"`
	Description     string   `yaml:"description,omitempty"`
	Namespace       string   `yaml:"namespace"`
	Dependencies    []string `yaml:"dependencies,omitempty"`
	ExtensionPoints []string `yaml:"extension_points"`
}

// ExtensionPointType identifies the type of an extension point.
type ExtensionPointType string

const (
	ExtensionPointEntityKinds     ExtensionPointType = "entity_kinds"
	ExtensionPointRelationTypes   ExtensionPointType = "relation_types"
	ExtensionPointValidationRules ExtensionPointType = "validation_rules"
	ExtensionPointRenderers       ExtensionPointType = "renderers"
	ExtensionPointRootKinds       ExtensionPointType = "root_kinds"
)

// ExtensionPoint defines the interface that all extension points must implement.
type ExtensionPoint interface {
	Type() ExtensionPointType
	Register(ext *Extension) error
}

// EntityKindContribution represents a single entity kind contributed by an extension.
type EntityKindContribution struct {
	Kind       core.EntityKind
	Definition *schema.EntityKindDefinition
}

// RelationTypeContribution represents a single relation type contributed by an extension.
// When Augment is true, the contribution adds participant kinds to an existing relation
// type definition (e.g. a core type like belongs_to) instead of defining a new type.
// Only Participant.SourceKinds/TargetKinds are merged; attempting to change other
// definition fields (Direction, Properties, min/max participants) is rejected.
type RelationTypeContribution struct {
	Type       core.RelationType
	Definition *schema.RelationTypeDefinition
	Augment    bool
}

// ValidationRuleContribution represents a single validation rule contributed by an extension.
type ValidationRuleContribution struct {
	Rule *validation.Rule
	Fn   validation.RuleFunc
}

// RendererContribution represents a single renderer contributed by an extension.
type RendererContribution struct {
	Renderer renderer.Renderer
}

// RootKinds grants root authority to the listed entity kinds. A graph may have
// multiple root entities only when every root's kind has been granted root
// authority. Root kinds are applied to the validation engine via the root_kinds
// extension point.
type Extension struct {
	Manifest        *Manifest
	EntityKinds     []EntityKindContribution
	RelationTypes   []RelationTypeContribution
	ValidationRules []ValidationRuleContribution
	Renderers       []RendererContribution
	RootKinds       []core.EntityKind
}
