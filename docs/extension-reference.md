# Extension Development Reference

This guide explains how to create plugins (extensions) for IACForge.

## Overview

IACForge supports runtime plugin loading via Go's `plugin` package.
Plugins are `.so` files built with `-buildmode=plugin`. They are loaded from
the directory given by the `IACFORGE_EXTENSIONS` environment variable (and by
the `--extensions` flag of the `validate` command). There is no automatic
startup scan of a fixed directory.

## Quick Start

### 1. Create a plugin project

```bash
mkdir my-plugin
cd my-plugin
go mod init my-plugin
```

### 2. Create `main.go`

```go
//go:build plugin

package main

import (
	"IACForge/src/core"
	"IACForge/src/extension"
	"IACForge/src/schema"
)

func Extension() *extension.Extension {
	return &extension.Extension{
		Manifest: &extension.Manifest{
			ID:              "my-organization.my-plugin",
			Name:            "My Custom Plugin",
			Version:         "1.0.0",
			Description:     "Adds custom entity kinds for our infrastructure",
			Namespace:       "myorg",
			ExtensionPoints: []string{
				string(extension.ExtensionPointEntityKinds),
			},
		},
		EntityKinds: []extension.EntityKindContribution{
			{
				Kind: core.EntityKind("custom_database"),
				Definition: &schema.EntityKindDefinition{
					Description: "Custom database server",
					Properties: []schema.PropertyDefinition{
						{
							Name:     "engine",
							Type:     schema.PropertyTypeString,
							Required: true,
							Description: "Database engine (e.g., postgres, mysql)",
							Constraints: &schema.Constraint{
								Enum: []string{"postgres", "mysql", "mariadb", "sqlite"},
							},
						},
						{
							Name:     "version",
							Type:     schema.PropertyTypeString,
							Required: false,
							Description: "Database version",
						},
					},
				},
			},
		},
	}
}

func main() {}
```

### 3. Build the plugin

```bash
go build -buildmode=plugin -o my-plugin.so .
```

### 4. Deploy

Point `IACFORGE_EXTENSIONS` at a directory containing the built `.so` files.
The directory is scanned for `.so` files on startup; subdirectories and other
files are ignored.

```bash
mkdir -p /path/to/my-extensions
cp my-plugin.so /path/to/my-extensions/
export IACFORGE_EXTENSIONS=/path/to/my-extensions
iacforge validate infra.yaml
```

## Manifest

The `Manifest` struct defines plugin metadata:

```go
type Manifest struct {
	ID              string   `yaml:"id"`
	Name            string   `yaml:"name"`
	Version         string   `yaml:"version"`
	Description     string   `yaml:"description,omitempty"`
	Namespace       string   `yaml:"namespace"`
	Dependencies    []string `yaml:"dependencies,omitempty"`
	ExtensionPoints []string `yaml:"extension_points"`
}
```

| Field | Required | Description |
|-------|----------|-------------|
| `ID` | Yes | Unique identifier (use namespace prefix) |
| `Name` | Yes | Human-readable name |
| `Version` | Yes | Semantic version (e.g., `1.0.0`) |
| `Namespace` | Yes | Prevents naming conflicts |
| `Description` | No | What the plugin does |
| `Dependencies` | No | List of required plugin IDs |
| `ExtensionPoints` | No | Types of extensions provided |

## Extension Points

IACForge supports 4 extension point types:

### entity_kinds

Add custom entity types to the schema.

```go
EntityKinds: []extension.EntityKindContribution{
	{
		Kind: core.EntityKind("my_custom_vm"),
		Definition: &schema.EntityKindDefinition{
			Description: "Description of this entity kind",
			Properties: []schema.PropertyDefinition{
				{
					Name:     "cpu_count",
					Type:     schema.PropertyTypeInteger,
					Required: true,
					Description: "Number of CPU cores",
					Constraints: &schema.Constraint{
						Min: f64Ptr(1),
						Max: f64Ptr(128),
					},
				},
			},
		},
	},
},
```

`f64Ptr` is a local helper you define in your plugin; the schema package does
not export pointer helpers:

```go
func f64Ptr(v float64) *float64 { return &v }
```

**Core entity kinds that CANNOT be redefined:**
`region`, `rack`, `server`, `interface`, `cable`, `power_distribution`, `network`, `vlan`, `switch`, `router`, `firewall`, `acl`, `acl_rule`, `vm`, `container`, `application`, `open_port`, `storage`, `volume`, `cluster`, `availability_zone`

### relation_types

Add custom relation types.

```go
RelationTypes: []extension.RelationTypeContribution{
	{
		Type: core.RelationType("custom_manages"),
		Definition: &schema.RelationTypeDefinition{
			Direction: schema.DirectionDirected,
			Description: "Custom management relation",
			Participants: &schema.ParticipantConstraints{
				SourceKinds: []core.EntityKind{"server", "vm"},
				TargetKinds: []core.EntityKind{"application"},
			},
		},
	},
},
```

**Direction types:**
- `directed` - Has source and target
- `symmetric` - Same meaning in both directions

**Core relation types that CANNOT be redefined:**
`connects`, `hosts`, `depends_on`, `belongs_to`, `replicates_to`, `backs_up`, `monitors`, `managed_by`, `mounted_on`, `applies_to`, `listens_on`

### validation_rules

Add custom validation rules.

```go
ValidationRules: []extension.ValidationRuleContribution{
	{
		Rule: &validation.Rule{
			ID:          "my-custom-rule",
			Name:        "Custom Validation Rule",
			Description: "Ensures custom business logic",
			Severity:    validation.SeverityWarning,
		},
		Fn: func(ctx *validation.Context) []validation.Finding {
			g := ctx.Graph.(*core.Graph)
			var findings []validation.Finding

			for _, e := range g.Entities() {
				if e.Kind == "my_custom_vm" {
					// Your validation logic here
				}
			}

			return findings
		},
	},
},
```

**Severity levels:** `info`, `warning`, `error`

### renderers

Add custom output renderers.

```go
Renderers: []extension.RendererContribution{
	{
		Renderer: &MyCustomRenderer{},
	},
},
```

Your renderer must implement:

```go
type Renderer interface {
	Render(v *view.ViewResult, opts *RenderOptions) (*Artifact, error)
	ID() string
	Name() string
	Format() string
}
```

## Dependencies

Declare dependencies on other plugins:

```go
Manifest: &extension.Manifest{
	ID:           "myorg.dependent-plugin",
	Dependencies: []string{"myorg.base-plugin"},
},
```

- Load order is automatically resolved (topological sort)
- Circular dependencies cause an error
- Missing dependencies cause an error

## Namespace Conventions

Use reverse-domain notation to prevent conflicts:

```
com.github.username.plugin-name
org.company.infrastructure-type
```

## Property Types

| Type | Go Type | Description |
|------|---------|-------------|
| `string` | `string` | Text values |
| `integer` | `int`, `int64` | Whole numbers |
| `number` | `float64` | Decimal numbers |
| `boolean` | `bool` | True/false |
| `list` | `[]interface{}` | Array of values |
| `map` | `map[string]interface{}` | Key-value pairs |
| `reference` | `string` | Reference to another entity |

## Constraints

`Constraint` fields that are pointers (`Min`, `Max`, `MinLength`, `MaxLength`,
`Pattern`, `UniqueItems`) require pointer values. The schema package does not
export pointer helpers; define small helpers in your plugin:

```go
func f64Ptr(v float64) *float64 { return &v }
func intPtr(v int) *int         { return &v }
func strPtr(v string) *string   { return &v }
func boolPtr(v bool) *bool      { return &v }
```

```go
&schema.Constraint{
	Min:         f64Ptr(0),        // Minimum numeric value
	Max:         f64Ptr(100),      // Maximum numeric value
	MinLength:   intPtr(1),        // Minimum string length
	MaxLength:   intPtr(255),      // Maximum string length
	Pattern:     strPtr("^[a-z]"), // Regex pattern
	Enum:        []string{"a", "b"}, // Allowed values
	UniqueItems: boolPtr(true),    // List items must be unique
}
```

## Complete Example

A runnable example plugin lives in `testdata/plugins/testplugin/`.

- `testdata/plugins/testplugin/main.go` — the plugin source.
- `testdata/plugins/run-example.sh` — builds the plugin and loads it from a
  standalone host program.

Go plugins require the host binary and the plugin to be built with identical
Go versions and dependency build IDs, so the runtime load is demonstrated via
the script rather than inside `go test` (see `src/extension/plugin_load_test.go`).

```bash
bash testdata/plugins/run-example.sh
```

## Troubleshooting

### Plugin fails to load

1. Ensure Go versions match between host and plugin
2. Verify the plugin exports `Extension() *extension.Extension`
3. Check that dependencies are built with `-buildmode=plugin`

### Namespace conflicts

Each namespace can only be used by one plugin. Use unique namespaces.

### Core conflicts

You cannot redefine core entity kinds or relation types. Use different names with your namespace prefix.
