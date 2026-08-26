package extension

import (
	"IACForge/src/core"
	"IACForge/src/validation"
)

// RootKindsExtensionPoint manages root-kind contributions from extensions.
// Root kinds grant root authority to specific entity kinds, relaxing the
// exactly-one-root ownership rules for kinds such as aws.organization.
type RootKindsExtensionPoint struct {
	engine *validation.Engine
	kinds  map[core.EntityKind]string // kind -> extension ID that granted it
}

// NewRootKindsExtensionPoint creates a new root kinds extension point.
func NewRootKindsExtensionPoint(engine *validation.Engine) *RootKindsExtensionPoint {
	return &RootKindsExtensionPoint{
		engine: engine,
		kinds:  make(map[core.EntityKind]string),
	}
}

// Type returns the extension point type.
func (ep *RootKindsExtensionPoint) Type() ExtensionPointType {
	return ExtensionPointRootKinds
}

// Register grants root authority to every root kind declared by the extension.
func (ep *RootKindsExtensionPoint) Register(ext *Extension) error {
	for _, kind := range ext.RootKinds {
		ep.engine.AddAllowedRootKind(kind)
		ep.kinds[kind] = ext.Manifest.ID
	}
	return nil
}

// GetRootKindsByExtension returns all root kinds granted by a specific extension.
func (ep *RootKindsExtensionPoint) GetRootKindsByExtension(extensionID string) []core.EntityKind {
	var result []core.EntityKind
	for kind, extID := range ep.kinds {
		if extID == extensionID {
			result = append(result, kind)
		}
	}
	return result
}

// GetExtensionForRootKind returns the extension ID that granted the given kind.
func (ep *RootKindsExtensionPoint) GetExtensionForRootKind(kind core.EntityKind) (string, bool) {
	extID, ok := ep.kinds[kind]
	return extID, ok
}
