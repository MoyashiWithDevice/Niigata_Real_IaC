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
| interfaces | interface | belongs_to (child) |
| servers | server | belongs_to (child) |
| switches | switch | belongs_to (child) |
| routers | router | belongs_to (child) |
| firewalls | firewall | belongs_to (child) |
| networks | network | belongs_to (child) |

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
| string | Text value | "hello", "10.0.1.10" |
| integer | Whole number | 42, 1000, -1 |
| number | Floating point | 3.14, 1024.5 |
| boolean | True/false | true, false |

### Complex Types

| Type | Description | Examples |
|------|-------------|----------|
| list[Type] | Ordered collection | ["web", "db"], [1, 2, 3] |
| map[Type] | Key-value pairs | {env: "prod", tier: "1"} |
| reference | Reference to another Object | "srv-01", "/region01/rack01/server01" |

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
| region | Physical location |
| rack | Physical rack enclosure |
| server | Physical or virtual compute host |
| interface | Network interface |
| cable | Physical cable |
| power_distribution | Power distribution unit (PDU) |
| network | Logical network |
| vlan | VLAN definition |
| switch | Network switch |
| router | Network router |
| firewall | Network firewall |
| acl | Access Control List |
| acl_rule | Individual ACL rule |
| vm | Virtual machine |
| container | Containerized workload |
| application | Software application |
| open_port | Listening network port |
| storage | Storage system |
| volume | Logical storage volume |
| cluster | Logical compute grouping |
| availability_zone | Logical availability zone |

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
| connects | symmetric | Physical connection |
| hosts | directed | Execution hosting |
| depends_on | directed | Dependency |
| belongs_to | directed | Logical membership |
| replicates_to | directed | Data replication |
| backs_up | directed | Backup relationship |
| monitors | directed | Monitoring |
| managed_by | directed | Management |
| mounted_on | directed | Storage mounting |
| applies_to | directed | ACL application |
| listens_on | directed | Port listening |

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
    - region
    - rack
    - server
  required_properties:
    region: ["name"]
    rack: ["name"]
    server: ["name", "platform"]
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
  description: "Core Infrastructure Schema"

entity_kinds:
  server:
    description: "Physical or virtual compute host"
    properties:
      - name: platform
        type: string
        required: false
        description: "Virtualization platform"
      - name: cpu
        type: list
        required: false
        description: "CPU configurations"
        properties:
          - name: cores
            type: integer
            required: false
            constraints:
              min: 1
              max: 1024
            description: "Number of CPU cores"
          - name: architecture
            type: string
            required: false
            description: "CPU architecture (x86_64, arm64)"
      - name: memory
        type: list
        required: false
        description: "Memory modules"
        properties:
          - name: size_gb
            type: number
            required: true
            description: "Memory module size in GB"
          - name: speed
            type: integer
            required: false
            description: "Memory speed in MHz"
          - name: type
            type: string
            required: false
            description: "Memory type (ddr4, ddr5, lpddr4, lpddr5)"
      - name: storage
        type: list
        required: false
        description: "Local storage devices"
        properties:
          - name: size_gb
            type: number
            required: false
            description: "Storage size in GB"
          - name: type
            type: string
            required: false
            description: "Storage type (ssd, hdd, nvme)"

relation_types:
  connects:
    direction: symmetric
    description: "Physical connection"
    participants:
      source_kinds:
        - interface
      target_kinds:
        - interface
      min_participants: 2
      max_participants: 2

profiles: []
```
