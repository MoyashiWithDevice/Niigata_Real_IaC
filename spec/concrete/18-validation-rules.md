# Validation Rules

## Overview

Validation Rules define how a Graph is evaluated against the Schema.

Validation is deterministic and side-effect free.

Every compliant implementation MUST support the Core Validation Rules.

---

## Rule Structure

Every Validation Rule is defined with:

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| id | string | yes | Unique rule identifier |
| name | string | yes | Human-readable name |
| description | string | no | Rule description |
| severity | enum | yes | Finding severity level |
| scope | enum | yes | Rule evaluation scope |
| condition | string | yes | Rule condition expression |

---

## Severity Levels

Every Validation Finding has a severity level.

| Level | Description |
|-------|-------------|
| info | Informational finding |
| warning | Potential issue |
| error | Violation that must be fixed |

Implementations MAY define additional levels.

---

## Finding Structure

Validation produces Findings.

| Field | Type | Description |
|-------|------|-------------|
| rule_id | string | Rule identifier |
| severity | enum | Finding severity |
| message | string | Human-readable message |
| object_id | string | Related Object identifier |
| object_type | enum | Object type (entity/relation) |
| path | string | Object path in Graph |

### Finding Example

```json
{
  "rule_id": "unique-id",
  "severity": "error",
  "message": "Duplicate identifier: srv-01",
  "object_id": "srv-01",
  "object_type": "entity",
  "path": "/region-01/rack-01/srv-01"
}
```

---

## Core Validation Rules

### Graph Integrity Rules

| Rule ID | Severity | Description |
|---------|----------|-------------|
| unique-id | error | Object identifiers MUST be unique |
| valid-reference | error | References MUST point to existing Objects |
| valid-owner | error | Owner identifier MUST reference an existing Entity |
| single-owner | error | Every Entity except root has exactly one owner specified |

### Entity Rules

| Rule ID | Severity | Description |
|---------|----------|-------------|
| required-kind | error | Entity MUST define kind |
| required-name | error | Entity MUST define name |
| valid-kind | error | Entity kind MUST be defined in Schema |
| valid-status | warning | Status SHOULD be a valid enum value |
| valid-port-range | error | open_port port MUST be between 1 and 65535 |
| valid-acl-rule-parent | error | acl_rule MUST have owner referencing an acl |
| valid-ip-format | warning | interface ip_address values SHOULD be valid IP addresses or CIDR notation |
| ip-requires-network | warning | interface with IP addresses SHOULD reference a network (via `network` property or `belongs_to` relation) |
| network-reference-kind | warning | interface `network` reference SHOULD point to an entity of kind network |
| ip-in-cidr | warning | interface IP addresses SHOULD be within the CIDR of a referenced network |
| network-cidr-required | warning | network with member interfaces that have IP addresses SHOULD define a valid cidr |
| gateway-in-cidr | warning | network gateway SHOULD be within the network cidr |
| ip-unique-in-network | warning | IP addresses SHOULD be unique within a network |

### Relation Rules

| Rule ID | Severity | Description |
|---------|----------|-------------|
| required-type | error | Relation MUST define type |
| required-participants | error | Relation MUST define participants |
| valid-type | error | Relation type MUST be defined in Schema |
| valid-direction | error | Directed relations MUST have source and target |
| valid-cardinality | error | Participant count MUST satisfy constraints |
| valid-participant-kind | warning | Participant kind SHOULD be allowed by type |

### Ownership Rules

| Rule ID | Severity | Description |
|---------|----------|-------------|
| ownership-tree | error | Ownership MUST form exactly one tree |
| no-ownership-cycle | error | Ownership MUST NOT contain cycles |
| root-entity | error | Exactly one root Entity MUST exist |

### Reference Rules

| Rule ID | Severity | Description |
|---------|----------|-------------|
| dangling-reference | error | References MUST resolve to existing Objects |
| invalid-path | error | Path references MUST be valid |

---

## Rule Evaluation

### Scope

Rules define their evaluation scope.

| Scope | Description |
|-------|-------------|
| graph | Evaluates entire Graph |
| entity | Evaluates each Entity |
| relation | Evaluates each Relation |
| ownership | Evaluates ownership structure |

### Condition Expression

Conditions use a declarative expression language.

#### Examples

```yaml
condition: "entity.id is unique"
condition: "relation.type is defined in schema"
condition: "entity.status in ['planned', 'active', 'maintenance', 'deprecated', 'offline', 'standby']"
condition: "relation.participants.count >= 2"
condition: "ownership.forms.tree"
```

### Expression Types

| Type | Description | Example |
|------|-------------|---------|
| equality | Compare values | `entity.kind == 'server'` |
| membership | Check set membership | `entity.status in allowed_statuses` |
| existence | Check existence | `relation.participants exist` |
| cardinality | Count check | `participants.count >= 2` |
| structural | Graph structure | `ownership.forms.tree` |

---

## Core Rule Definitions

### unique-id

```yaml
id: unique-id
name: Unique Identifier
description: "Object identifiers MUST be unique"
severity: error
scope: graph
condition: "all object.id are unique"
```

### valid-reference

```yaml
id: valid-reference
name: Valid Reference
description: "References MUST point to existing Objects"
severity: error
scope: relation
condition: "all relation.participants reference existing entities"
```

### single-owner

```yaml
id: single-owner
name: Single Owner
description: "Every Entity except root has exactly one owner specified"
severity: error
scope: ownership
condition: "all entities except root have exactly one owner specified"
```

### ownership-tree

```yaml
id: ownership-tree
name: Ownership Tree
description: "Ownership MUST form exactly one tree"
severity: error
scope: ownership
condition: "ownership.forms.single.tree"
```

### required-kind

```yaml
id: required-kind
name: Required Kind
description: "Entity MUST define kind"
severity: error
scope: entity
condition: "entity.kind is defined"
```

### required-name

```yaml
id: required-name
name: Required Name
description: "Entity MUST define name"
severity: error
scope: entity
condition: "entity.name is defined"
```

### required-type

```yaml
id: required-type
name: Required Type
description: "Relation MUST define type"
severity: error
scope: relation
condition: "relation.type is defined"
```

### required-participants

```yaml
id: required-participants
name: Required Participants
description: "Relation MUST define participants"
severity: error
scope: relation
condition: "relation.participants is defined and not empty"
```

### valid-kind

```yaml
id: valid-kind
name: Valid Kind
description: "Entity kind MUST be defined in Schema"
severity: error
scope: entity
condition: "entity.kind exists in schema.entity_kinds"
```

### valid-type

```yaml
id: valid-type
name: Valid Type
description: "Relation type MUST be defined in Schema"
severity: error
scope: relation
condition: "relation.type exists in schema.relation_types"
```

### valid-status

```yaml
id: valid-status
name: Valid Status
description: "Status SHOULD be a valid enum value"
severity: warning
scope: entity
condition: "entity.status in schema.status_enum"
```

### valid-cardinality

```yaml
id: valid-cardinality
name: Valid Cardinality
description: "Participant count MUST satisfy constraints"
severity: error
scope: relation
condition: "relation.participants.count between schema.min_participants and schema.max_participants"
```

### valid-participant-kind

```yaml
id: valid-participant-kind
name: Valid Participant Kind
description: "Participant kind SHOULD be allowed by type"
severity: warning
scope: relation
condition: "all relation.participants.kind in schema.relation_types[type].allowed_kinds"
```

### valid-port-range

```yaml
id: valid-port-range
name: Valid Port Range
description: "open_port port MUST be between 1 and 65535"
severity: error
scope: entity
condition: "entity.kind == 'open_port' implies 1 <= entity.port <= 65535"
```

### valid-acl-rule-parent

```yaml
id: valid-acl-rule-parent
name: Valid ACL Rule Parent
description: "acl_rule MUST have owner referencing an acl"
severity: error
scope: entity
condition: "entity.kind == 'acl_rule' implies entity.owner is defined and entity(entity.id == entity.owner).kind == 'acl'"
```

### valid-ip-format

```yaml
id: valid-ip-format
name: Valid IP Address Format
description: "interface ip_address values SHOULD be valid IP addresses or CIDR notation"
severity: warning
scope: entity
condition: "entity.kind == 'interface' implies all entity.ip_address parse as IP or CIDR"
```

### ip-requires-network

```yaml
id: ip-requires-network
name: IP Requires Network
description: "interface with IP addresses SHOULD reference a network"
severity: warning
scope: entity
condition: "entity.kind == 'interface' and entity.ip_address is defined and not empty implies entity.network is defined or exists relation type == 'belongs_to' where source == entity.id and target.kind == 'network'"
```

### network-reference-kind

```yaml
id: network-reference-kind
name: Network Reference Kind
description: "interface network reference SHOULD point to an entity of kind network"
severity: warning
scope: entity
condition: "entity.kind == 'interface' and entity.network is defined implies entity(entity.id == entity.network).kind == 'network'"
```

### ip-in-cidr

```yaml
id: ip-in-cidr
name: IP Within Network CIDR
description: "interface IP addresses SHOULD be within the CIDR of a referenced network"
severity: warning
scope: entity
condition: "entity.kind == 'interface' implies all entity.ip_address are within the cidr of a referenced network"
```

### network-cidr-required

```yaml
id: network-cidr-required
name: Network CIDR Required
description: "network with member interfaces that have IP addresses SHOULD define a valid cidr"
severity: warning
scope: entity
condition: "entity.kind == 'network' and exists interface with ip_address belonging to entity implies entity.cidr is defined and valid"
```

### gateway-in-cidr

```yaml
id: gateway-in-cidr
name: Gateway Within Network CIDR
description: "network gateway SHOULD be within the network cidr"
severity: warning
scope: entity
condition: "entity.kind == 'network' and entity.gateway is defined implies entity.gateway is within entity.cidr"
```

### ip-unique-in-network

```yaml
id: ip-unique-in-network
name: Unique IP Within Network
description: "IP addresses SHOULD be unique within a network"
severity: warning
scope: graph
condition: "for each network, all member interface ip_address are unique"
```

---

## Custom Rules

Implementations MAY define custom rules.

Custom rules MUST follow the rule structure.

Custom rules MUST NOT redefine core rule semantics.

### Custom Rule Example

```yaml
id: custom-required-ip
name: Required IP Address
description: "Servers MUST have IP address"
severity: error
scope: entity
condition: "entity.kind == 'server' implies entity.ip_address is defined"
```

```yaml
id: custom-ssh-restricted
name: SSH Restricted
description: "SSH access SHOULD be restricted to management network"
severity: warning
scope: relation
condition: "relation.type == 'applies_to' and relation.source.kind == 'acl' implies acl_rule(action='allow', destination_port='22').source_address in ['10.0.0.0/24']"
```

---

## Validation Profiles

Validation Rules are grouped into Profiles.

Profiles are user-defined.

The Core Schema does not define specific profiles.

### Profile Structure

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| name | string | yes | Profile name |
| description | string | no | Profile description |
| rules | list[string] | yes | Rule IDs to include |
| required_kinds | list[string] | no | Entity Kinds that must exist |
| required_relations | list[string] | no | Relation Types that must exist |

### Profile Example

```yaml
- name: my-profile
  description: Custom validation profile
  rules:
    - unique-id
    - valid-reference
    - single-owner
    - ownership-tree
    - required-kind
    - required-name
    - required-type
    - required-participants
    - valid-port-range
    - valid-acl-rule-parent
  required_kinds:
    - region
    - rack
    - server
```

---

## Validation Execution

### Input

- Graph to validate
- Schema (optional)
- Profile (optional)

### Process

1. Load Graph
2. Load Schema (if provided)
3. Load Profile (if provided)
4. Select applicable rules
5. Execute rules against Graph
6. Collect Findings
7. Return Results

### Output

| Field | Type | Description |
|-------|------|-------------|
| findings | list[Finding] | List of validation findings |
| passed | boolean | Whether validation passed |
| summary | map | Summary statistics |

### Summary Structure

| Field | Type | Description |
|-------|------|-------------|
| total_rules | integer | Number of rules evaluated |
| total_findings | integer | Total findings count |
| errors | integer | Error count |
| warnings | integer | Warning count |
| infos | integer | Info count |

---

## Determinism

Validation is deterministic.

Running the same validation against the same Graph MUST always produce identical findings.

---

## Extensibility

Implementations MAY introduce custom Validation Rules.

Custom Rules MUST NOT redefine the semantics of the Core Schema.
