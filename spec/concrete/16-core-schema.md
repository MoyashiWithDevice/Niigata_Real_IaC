# Core Schema

## Overview

The Core Schema defines the structure of a valid infrastructure model.

The Schema is independent of any serialization format.

The Schema defines what may exist and how it is constrained.

Every compliant implementation MUST support the Core Schema.

---

## Schema Structure

A Schema consists of:

- Version information
- Property type definitions
- Entity kind definitions
- Relation type definitions
- Global nesting definitions
- User-defined validation profiles (extensions)

---

## Global Nesting Definitions

The Schema MAY define nesting definitions that apply to all entity kinds.

These are called Global Nesting Definitions.

Global nesting definitions are merged with per-kind nesting definitions.

When a per-kind definition exists for the same nest key, the per-kind definition takes precedence.

### Global Nesting Definition

| Nest Key | Child Kind | Auto-Relation |
|----------|------------|---------------|
| grounds | ground | belongs_to (child) |
| terrains | terrain | belongs_to (child) |
| water_bodies | water_body | belongs_to (child) |
| forests | forest | belongs_to (child) |
| tourism_spots | tourism_spot | belongs_to (child) |
| species | species | belongs_to (child) |
| populations | population | belongs_to (child) |
| hot_springs | hot_spring | belongs_to (child) |
| events | event | belongs_to (child) |
| cultural_assets | cultural_asset | belongs_to (child) |

コア Schema では、これらのネストは個々の Entity Kind 定義（per-kind）として登録されている。

例えば `populations` は `species` 配下でのみネスト可能である。

---

## Schema Version

Every Schema MUST declare a version.

| Property | Type | Required | Description |
|----------|------|----------|-------------|
| schema_version | string | yes | Semantic version (e.g., "1.0.0") |
| spec_version | string | yes | Specification version (e.g., "1.0") |
| description | string | no | Human-readable description |

---

## Property Types

The Core Schema defines the following property types.

### Primitive Types

| Type | Description | Examples |
|------|-------------|----------|
| string | Text value | "hello", "Nipponia nippon" |
| integer | Whole number | 42, 1000, -1 |
| number | Floating point | 3.14, 1024.5 |
| boolean | True/false | true, false |

### Complex Types

| Type | Description | Examples |
|------|-------------|----------|
| list[Type] | Ordered collection | ["web", "db"], [1, 2, 3] |
| map[Type] | Key-value pairs | {env: "prod", tier: "1"} |
| reference | Reference to another Object | "@area-sado-city", "/area-niigata/forest-myoko/species-toki" |

### Enumerated Types

| Type | Description | Examples |
|------|-------------|----------|
| enum | Predefined values | "active", "planned", "deprecated" |

---

## Common Properties

Every Object (Entity or Relation) shares the following common properties.

### Required Properties

| Property | Type | Description |
|----------|------|-------------|
| id | string | Unique identifier within scope |
| kind | string | Entity kind (for Entities) |
| type | string | Relation type (for Relations) |
| name | string | Human-readable name |

### Optional Properties

| Property | Type | Default | Description |
|----------|------|---------|-------------|
| description | string | - | Human-readable documentation |
| status | enum | - | Lifecycle state |
| tags | list[string] | - | Unordered labels |
| labels | map[string] | - | Key-value metadata |
| extensions | map[string] | - | Implementation-specific data |

---

## Status Enum

The core specification defines the following status values.

| Value | Description |
|-------|-------------|
| planned | Not yet deployed |
| active | Operational |
| maintenance | Under maintenance |
| deprecated | Scheduled for removal |
| offline | Not operational |
| standby | Standby (e.g. redundant member) |

Implementations MAY extend with additional values.

---

## Entity Kind Definitions

Every Entity kind is defined by a Schema.

A definition includes:

- Required properties
- Optional properties
- Property types
- Default values
- Constraints

### Core Entity Kinds

| Kind | Description |
|------|-------------|
| area | Geographic area (city, town, district, island) |
| ground | Ground or soil condition at a site |
| terrain | Landform (mountain, plain, coast, valley) |
| water_body | River, lake, sea area, pond, or marsh |
| forest | Forest or wooded area |
| species | Animal or plant species |
| population | Population record of a species from a survey |
| tourism_spot | Tourist attraction |
| hot_spring | Hot spring source or bath facility |
| cultural_asset | Historic site or cultural property |
| event | Festival or recurring local event |

### Property Type Definitions

Each property is defined with:

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| name | string | yes | Property name |
| type | string | yes | Property type |
| required | boolean | yes | Whether property is required |
| default | any | no | Default value |
| description | string | no | Human-readable description |
| constraints | map | no | Validation constraints |
| properties | list[object] | no | Sub-properties for structured lists (list[object]) |

### Structured Lists

When a property has type `list[object]`, the `properties` field defines the schema for each list element. Each element must be a map conforming to the defined sub-properties.

#### Example

```yaml
properties:
  - name: cpu
    type: list
    description: "CPU configurations"
    properties:
      - name: cores
        type: integer
        description: "Number of CPU cores"
      - name: architecture
        type: string
        description: "CPU architecture"
```

**注意:** コア Schema（Niigata Nature and Tourism Resource Schema）は構造化リストを使用しない。すべての kind 固有プロパティはプリミティブ型である。上記は `list[object]` 型の一般的な例としてのみ示す。

### Constraint Types

| Constraint | Applicable Types | Description |
|------------|------------------|-------------|
| min | integer, number | Minimum value |
| max | integer, number | Maximum value |
| minLength | string | Minimum string length |
| maxLength | string | Maximum string length |
| pattern | string | Regex pattern |
| enum | any | Allowed values |
| uniqueItems | list | List items must be unique |

---

## Relation Type Definitions

Every Relation type is defined by a Schema.

A definition includes:

- Directionality
- Participant constraints
- Cardinality
- Properties

### Core Relation Types

| Type | Direction | Description |
|------|-----------|-------------|
| located_in | directed | Entity is located within an area or terrain |
| inhabits | directed | Species lives in a habitat |
| near | symmetric | Two resources are geographically close |
| depends_on | directed | Directional dependency (e.g. hot spring on its source) |
| belongs_to | directed | Logical membership or association |
| flows_into | directed | River or water flow destination |

### Direction Enum

| Value | Description |
|-------|-------------|
| directed | Has source and target |
| symmetric | All participants equal |

### Participant Constraints

| Field | Type | Description |
|-------|------|-------------|
| source_kinds | list[string] | Allowed source Entity Kinds |
| target_kinds | list[string] | Allowed target Entity Kinds |
| min_participants | integer | Minimum participant count |
| max_participants | integer | Maximum participant count |

---

## Validation Profiles

A Profile defines a subset of the Core Schema for specific use cases.

Profiles are user-defined extensions.

The Core Schema does not define specific profiles.

### Profile Structure

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| name | string | yes | Profile name |
| description | string | no | Profile description |
| required_kinds | list[string] | no | Entity Kinds that must exist |
| required_relations | list[string] | no | Relation Types that must exist |
| required_properties | map | no | Properties that must be specified per Kind |
| optional_overrides | map | no | Override default values |

### Profile Example

```yaml
- name: my-profile
  description: Custom validation profile
  required_kinds:
    - area
    - species
  required_properties:
    area: ["name"]
    species: ["name"]
    population: ["name", "count"]
```

---

## Graph Constraints

The Core Schema defines the following graph-level constraints.

### Ownership Constraints

| Constraint | Description |
|------------|-------------|
| single_owner | Every Entity except root has exactly one owner specified |
| tree_structure | Ownership forms exactly one tree |
| no_cycles | Ownership cannot contain cycles |
| owner_exists | Owner identifier MUST reference an existing Entity |

### Reference Constraints

| Constraint | Description |
|------------|-------------|
| valid_reference | References MUST point to existing Objects |
| unique_id | Object identifiers MUST be unique |
| relation_exists | Relations MUST reference existing Objects |

### Cardinality Constraints

| Constraint | Description |
|------------|-------------|
| min_participants | Relations MUST have minimum required participants |
| max_participants | Relations MUST not exceed maximum participants |

---

## Defaults

Schemas MAY define default values.

Missing properties are interpreted as unknown unless a default exists.

Default values never imply that a property was explicitly specified.

### Default Value Rules

| Rule | Description |
|------|-------------|
| type_match | Default value MUST match property type |
| immutable | Default values MUST NOT change after definition |
| documented | Default values SHOULD be documented |

---

## Schema Evolution

Schema evolution SHOULD preserve compatibility.

### Compatibility Rules

| Change Type | Compatibility |
|-------------|---------------|
| Adding optional property | Compatible |
| Adding optional Entity Kind | Compatible |
| Adding optional Relation Type | Compatible |
| Adding enum value | Compatible |
| Removing property | Breaking |
| Renaming property | Breaking |
| Changing property type | Breaking |
| Changing cardinality | Breaking |

### Versioning

Schemas follow semantic versioning.

- Major version: Breaking changes
- Minor version: New features
- Patch version: Bug fixes

---

## Example Schema Definition

```yaml
schema:
  schema_version: "1.0.0"
  spec_version: "1.0"
  description: "Niigata Nature and Tourism Resource Schema"

entity_kinds:
  species:
    description: "Animal or plant species living in the area"
    properties:
      - name: scientific_name
        type: string
        required: false
        description: "Scientific (Latin) name"
      - name: category
        type: string
        required: false
        constraints:
          enum:
            - mammal
            - bird
            - reptile
            - amphibian
            - fish
            - insect
            - plant
            - other
        description: "Biological category"
      - name: red_list_status
        type: string
        required: false
        constraints:
          enum:
            - extinct
            - extinct_in_wild
            - critically_endangered
            - endangered
            - vulnerable
            - near_threatened
            - least_concern
            - data_deficient
        description: "Red List conservation status"
    nesting:
      populations:
        child_kind: population
        auto_relation_type: belongs_to
        auto_relation_source: child

relation_types:
  inhabits:
    direction: directed
    description: "Species lives in a habitat"
    participants:
      source_kinds:
        - species
        - population
      target_kinds:
        - ground
        - terrain
        - water_body
        - forest
        - area
      min_participants: 2
      max_participants: 2

profiles: []
```
