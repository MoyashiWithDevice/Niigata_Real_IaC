# Query Model

## Overview

The Query Model defines how Objects are selected from a Graph.

Queries never modify the Graph.

Queries produce deterministic results.

Every compliant implementation MUST support the Core Query Model.

---

## Query Structure

Every Query is defined with:

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| id | string | no | Query identifier |
| select | SelectClause | yes | What to select |
| where | WhereClause | no | Filtering conditions |
| traverse | TraverseClause | no | Traversal operations |
| project | ProjectClause | no | Projection options |
| limit | integer | no | Maximum results |
| offset | integer | no | Result offset |

---

## Select Clause

The Select clause defines what Objects to select.

### Entity Selection

```yaml
select:
  entities:
    - kind: server
    - kind: vm
      where:
        status: active
```

### Relation Selection

```yaml
select:
  relations:
    - type: connects
      where:
        connection_type: physical
```

### Combined Selection

```yaml
select:
  entities:
    - kind: server
    - kind: vm
  relations:
    - type: hosts
```

---

## Where Clause

The Where clause filters Objects based on conditions.

### Condition Operators

| Operator | Description | Example |
|----------|-------------|---------|
| eq | Equals | `status: active` |
| ne | Not equals | `status: {ne: offline}` |
| in | In list | `kind: [server, vm]` |
| nin | Not in list | `kind: {nin: [cable]}` |
| gt | Greater than | `cpu.cores: {gt: 8}` |
| ge | Greater than or equal | `memory.size_gb: {gte: 16}` |
| lt | Less than | `storage.size_gb: {lt: 1000}` |
| le | Less than or equal | `storage.size_gb: {lte: 2000}` |
| contains | String contains | `name: {contains: web}` |
| starts_with | String starts with | `name: {starts_with: srv}` |
| ends_with | String ends with | `name: {ends_with: 01}` |
| matches | Regex match | `ip_address: {matches: "10\\.0\\..*"}` |
| defined | Property exists | `description: {defined: true}` |
| undefined | Property does not exist | `description: {defined: false}` |

### Condition Examples

```yaml
where:
  status: active
  kind: server
  cpu.cores: {ge: 16}
  tags: {contains: production}
  labels:
    region: ap-northeast-1
```

### Logical Operators

| Operator | Description | Example |
|----------|-------------|---------|
| and | Logical AND | `{and: [condition1, condition2]}` |
| or | Logical OR | `{or: [condition1, condition2]}` |
| not | Logical NOT | `{not: condition}` |

### Logical Operator Examples

```yaml
where:
  or:
    - kind: server
    - kind: vm
  and:
    - status: active
    - tags: {contains: production}
  not:
    kind: cable
```

---

## Traverse Clause

The Traverse clause defines how to navigate the Graph.

### Traversal Types

| Type | Description | Direction |
|------|-------------|-----------|
| ownership | Follow owner property | parent → child |
| reverse_ownership | Reverse ownership | child → parent |
| relations | Follow specific relations | configurable |
| incoming | Follow incoming relations | target → source |
| outgoing | Follow outgoing relations | source → target |

### Traversal Operations

| Operation | Description | Example |
|-----------|-------------|---------|
| children | Get child Entities | `ownership.children` |
| parent | Get parent Entity | `ownership.parent` |
| ancestors | Get all ancestors | `ownership.ancestors` |
| descendants | Get all descendants | `ownership.descendants` |
| related | Get related Entities | `relations.related` |
| sources | Get source Entities | `incoming.sources` |
| targets | Get target Entities | `outgoing.targets` |

### Traversal Examples

```yaml
# Get all children of a region
traverse:
  from: region-ap-northeast-1
  operation: children

# Get all descendants of a region
traverse:
  from: region-ap-northeast-1
  operation: descendants

# Get all VMs hosted by a server
traverse:
  from: srv-proxmox-01
  operation: outgoing
  relation_type: hosts

# Get all servers hosting a VM
traverse:
  from: vm-web-01
  operation: incoming
  relation_type: hosts
```

### Traversal Depth

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| depth | integer | 1 | Maximum traversal depth |
| max_depth | integer | unlimited | Maximum depth limit |

```yaml
traverse:
  from: region-ap-northeast-1
  operation: descendants
  depth: 3
```

---

## Project Clause

The Project clause defines how results are presented.

### Projection Types

| Type | Description |
|------|-------------|
| objects | Return full Objects |
| properties | Return specific properties |
| paths | Return Object paths |
| ids | Return Object identifiers |
| summary | Return summary information |

### Projection Examples

```yaml
# Return full objects
project:
  type: objects

# Return specific properties
project:
  type: properties
  properties:
    - name
    - status
    - cpu.cores
    - memory.size_gb

# Return paths
project:
  type: paths

# Return identifiers
project:
  type: ids

# Return summary
project:
  type: summary
  group_by: kind
```

### Property Selection

```yaml
project:
  type: properties
  properties:
    - name: name
    - name: status
    - name: labels
      transform: to_json
```

### Aggregation

| Function | Description | Example |
|----------|-------------|---------|
| count | Count Objects | `{count: true}` |
| sum | Sum numeric property | `{sum: cpu.cores}` |
| avg | Average numeric property | `{avg: memory.size_gb}` |
| min | Minimum value | `{min: storage.size_gb}` |
| max | Maximum value | `{max: storage.size_gb}` |
| group_by | Group results | `{group_by: kind}` |

```yaml
project:
  type: summary
  aggregation:
    count: true
    group_by: kind
    sum: cpu.cores
    avg: memory.size_gb
```

---

## Query Composition

Queries MAY be combined.

The output of one query MAY become the input of another.

### Composition Examples

```yaml
# Chain queries
queries:
  - id: active-servers
    select:
      entities:
        - kind: server
    where:
      status: active

  - id: vms-on-active-servers
    select:
      entities:
        - kind: vm
    traverse:
      from: query.active-servers
      operation: outgoing
      relation_type: hosts
```

### Named Queries

```yaml
queries:
  - id: prod-servers
    select:
      entities:
        - kind: server
    where:
      status: active
      tags: {contains: production}

  - id: prod-vms
    select:
      entities:
        - kind: vm
    traverse:
      from: query.prod-servers
      operation: outgoing
      relation_type: hosts

  - id: prod-apps
    select:
      entities:
        - kind: application
    traverse:
      from: query.prod-vms
      operation: outgoing
      relation_type: hosts
```

---

## Result Structure

### Result Object

| Field | Type | Description |
|-------|------|-------------|
| query_id | string | Query identifier |
| results | list[Object] | Query results |
| count | integer | Number of results |
| truncated | boolean | Whether results were truncated |
| metadata | map | Additional information |

### Result Object Structure

| Field | Type | Description |
|-------|------|-------------|
| id | string | Object identifier |
| type | enum | Object type (entity/relation) |
| path | string | Object path |
| object | Object | Full Object data |

### Result Example

```json
{
  "query_id": "active-servers",
  "results": [
    {
      "id": "srv-proxmox-01",
      "type": "entity",
      "path": "/region-ap-northeast-1/rack-a01/srv-proxmox-01",
      "object": {
        "id": "srv-proxmox-01",
        "kind": "server",
        "name": "Proxmox Node 01",
        "status": "active"
      }
    }
  ],
  "count": 1,
  "truncated": false
}
```

---

## Query Examples

### Select All Servers

```yaml
select:
  entities:
    - kind: server
```

### Select Active VMs with Filters

```yaml
select:
  entities:
    - kind: vm
where:
  status: active
  cpu.cores: {ge: 4}
  memory.size_gb: {ge: 8}
```

### Select All Connections

```yaml
select:
  relations:
    - type: connects
```

### Traverse to Find VMs

```yaml
select:
  entities:
    - kind: vm
traverse:
  from: srv-proxmox-01
  operation: descendants
```

### Complex Query

```yaml
id: complex-query
select:
  entities:
    - kind: server
    - kind: vm
    - kind: application
  relations:
    - type: hosts
    - type: depends_on
where:
  or:
    - kind: server
    - and:
      - kind: vm
      - status: active
traverse:
  from: region-ap-northeast-1
  operation: descendants
  depth: 4
project:
  type: properties
  properties:
    - name
    - kind
    - status
    - path
limit: 100
```

---

## Determinism

Executing the same query on the same Graph MUST always produce the same result.

---

## Side Effects

Queries never change the Graph.

Queries are pure operations.

---

## Serialization

Implementations MAY provide query mechanisms:

- CLI
- API
- GraphQL
- DSL

All query mechanisms operate on the same Query Model.
