package extension

import (
	"fmt"

	"IACForge/src/core"
	"IACForge/src/schema"
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
