package aws

import (
	"IACForge/src/core"
	"IACForge/src/extension"
)

func init() {
	extension.RegisterBuiltin(Extension())
}

// Extension returns the AWS extension. The factory is exported so a Go plugin
// (.so) can re-export it from a package main wrapper; this package itself is
// registered in-process as a built-in via RegisterBuiltin.
func Extension() *extension.Extension {
	return &extension.Extension{
		Manifest: &extension.Manifest{
			ID:              "iacforge.aws",
			Name:            "AWS",
			Version:         "1.0.0",
			Namespace:       "aws",
			ExtensionPoints: []string{string(extension.ExtensionPointEntityKinds), string(extension.ExtensionPointRelationTypes), string(extension.ExtensionPointRootKinds)},
		},
		EntityKinds:   KindDefinitions(),
		RelationTypes: append(RelationTypeDefinitions(), AugmentDefinitions()...),
		RootKinds:     []core.EntityKind{Organization},
	}
}
