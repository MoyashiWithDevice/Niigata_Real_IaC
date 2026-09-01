package extension

import (
	"fmt"

	"github.com/bababa/Niigata_Real_IaC/src/core"
	"github.com/bababa/Niigata_Real_IaC/src/core/types"
	"github.com/bababa/Niigata_Real_IaC/src/schema"
)

// EntityKindsExtensionPoint manages entity kind extensions.
type EntityKindsExtensionPoint struct {
	schema *schema.Schema
	kinds  map[core.EntityKind]string // kind -> extension ID that registered it
}

// NewEntityKindsExtensionPoint creates a new entity kinds extension point.
func NewEntityKindsExtensionPoint(s *schema.Schema) *EntityKindsExtensionPoint {
	return &EntityKindsExtensionPoint{
		schema: s,
		kinds:  make(map[core.EntityKind]string),
	}
}

// Type returns the extension point type.
func (ep *EntityKindsExtensionPoint) Type() ExtensionPointType {
	return ExtensionPointEntityKinds
}

// Register registers all entity kind contributions from the given extension.
func (ep *EntityKindsExtensionPoint) Register(ext *Extension) error {
	for _, contrib := range ext.EntityKinds {
		if ep.schema.HasEntityKind(contrib.Kind) {
			return fmt.Errorf("%w: entity kind %q already defined in schema", ErrCoreConflict, contrib.Kind)
		}
		ep.schema.AddEntityKind(contrib.Kind, contrib.Definition)
		ep.kinds[contrib.Kind] = ext.Manifest.ID

		if err := ep.registerNesting(ext, contrib); err != nil {
			return err
		}
	}
	return nil
}

// registerNesting registers global nesting definitions that make the contributed
// kind nestable under its declared parent kinds. Parent kinds must exist in the
// schema; the nest key defaults to the kind name pluralized with a trailing "s".
func (ep *EntityKindsExtensionPoint) registerNesting(ext *Extension, contrib EntityKindContribution) error {
	if len(contrib.ParentKinds) == 0 {
		return nil
	}

	nestKey := contrib.NestKey
	if nestKey == "" {
		nestKey = string(contrib.Kind) + "s"
	}

	for _, parent := range contrib.ParentKinds {
		if !ep.schema.HasEntityKind(parent) {
			return fmt.Errorf("%w: nesting parent kind %q is not defined in schema", ErrInvalidExtension, parent)
		}
		ep.schema.AddNestingDef(parent, schema.NestingDefinition{
			NestKey:            nestKey,
			ChildKind:          contrib.Kind,
			AutoRelationType:   types.BelongsTo,
			AutoRelationSource: "child",
		})
	}
	return nil
}

// GetEntityKindsByExtension returns all entity kinds registered by a specific extension.
func (ep *EntityKindsExtensionPoint) GetEntityKindsByExtension(extensionID string) []core.EntityKind {
	var result []core.EntityKind
	for kind, extID := range ep.kinds {
		if extID == extensionID {
			result = append(result, kind)
		}
	}
	return result
}

// GetExtensionForKind returns the extension ID that registered the given kind.
func (ep *EntityKindsExtensionPoint) GetExtensionForKind(kind core.EntityKind) (string, bool) {
	extID, ok := ep.kinds[kind]
	return extID, ok
}

// AllExtendedKinds returns all entity kinds added by extensions.
func (ep *EntityKindsExtensionPoint) AllExtendedKinds() map[core.EntityKind]string {
	result := make(map[core.EntityKind]string, len(ep.kinds))
	for k, v := range ep.kinds {
		result[k] = v
	}
	return result
}
