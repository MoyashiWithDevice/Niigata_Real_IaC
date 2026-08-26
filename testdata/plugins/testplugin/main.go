package main

import (
	"IACForge/src/core"
	"IACForge/src/extension"
	"IACForge/src/schema"
)

func Extension() *extension.Extension {
	return &extension.Extension{
		Manifest: &extension.Manifest{
			ID:        "testplugin.sample",
			Name:      "Test Plugin",
			Version:   "1.0.0",
			Namespace: "testplugin",
		},
		EntityKinds: []extension.EntityKindContribution{
			{
				Kind: core.EntityKind("testplugin.widget"),
				Definition: &schema.EntityKindDefinition{
					Description: "A widget contributed by the test plugin",
				},
			},
		},
	}
}

func main() {}
