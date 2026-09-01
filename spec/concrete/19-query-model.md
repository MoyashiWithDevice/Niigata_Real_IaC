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
    - kind: tourism_spot
    - kind: hot_spring
      where:
        status: active
```

### Relation Selection

```yaml
select:
  relations:
    - type: near
      where:
        walking_minutes: {le: 10}
```

### Combined Selection

```yaml
select:
  entities:
    - kind: species
    - kind: population
  relations:
    - type: inhabits
```

---

## Where Clause

The Where clause filters Objects based on conditions.

### Condition Operators

| Operator | Description | Example |
|----------|-------------|---------|
| eq | Equals | `status: active` |
| ne | Not equals | `status: {ne: offline}` |
| in | In list | `kind: [species, population]` |
| nin | Not in list | `kind: {nin: [event]}` |
| gt | Greater than | `population: {gt: 10000}` |
| ge | Greater than or equal | `count: {ge: 16}` |
| lt | Less than | `area_ha: {lt: 1000}` |
| le | Less than or equal | `elevation_m: {le: 2000}` |
| contains | String contains | `name: {contains: 温泉}` |
| starts_with | String starts with | `name: {starts_with: spot}` |
| ends_with | String ends with | `name: {ends_with: 01}` |
| matches | Regex match | `scientific_name: {matches: "Nipponia.*"}` |
| defined | Property exists | `description: {defined: true}` |
| undefined | Property does not exist | `survey_date: {defined: false}` |

### Condition Examples

```yaml
where:
  status: active
  kind: tourism_spot
  annual_visitors: {ge: 100000}
  tags: {contains: tourism}
  labels:
    area: niigata-city
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
    - kind: forest
    - kind: water_body
  and:
    - status: active
    - tags: {contains: conservation}
  not:
    kind: event
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
# Get all children of an area
traverse:
  from: area-niigata-city
  operation: children

# Get all descendants of an area
traverse:
  from: area-niigata-city
  operation: descendants

# Get all populations surveyed for a species
traverse:
  from: species-japanese-crested-ibis
  operation: outgoing
  relation_type: belongs_to

# Get the species a population belongs to
traverse:
  from: pop-toki-2025
  operation: incoming
  relation_type: belongs_to
```

### Traversal Depth

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| depth | integer | 1 | Maximum traversal depth |
| max_depth | integer | unlimited | Maximum depth limit |

```yaml
traverse:
  from: area-niigata-city
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
    - count
    - survey_date

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
| sum | Sum numeric property | `{sum: count}` |
| avg | Average numeric property | `{avg: count}` |
| min | Minimum value | `{min: elevation_m}` |
| max | Maximum value | `{max: elevation_m}` |
| group_by | Group results | `{group_by: kind}` |

```yaml
project:
  type: summary
  aggregation:
    count: true
    group_by: kind
    sum: annual_visitors
    avg: count
```

---

## Query Composition

Queries MAY be combined.

The output of one query MAY become the input of another.

### Composition Examples

```yaml
# Chain queries
queries:
  - id: active-spots
    select:
      entities:
        - kind: tourism_spot
    where:
      status: active

  - id: springs-of-active-spots
    select:
      entities:
        - kind: hot_spring
    traverse:
      from: query.active-spots
      operation: descendants
```

### Named Queries

```yaml
queries:
  - id: endangered-species
    select:
      entities:
        - kind: species
    where:
      red_list_status: {in: [critically_endangered, endangered, vulnerable]}

  - id: populations-of-endangered
    select:
      entities:
        - kind: population
    traverse:
      from: query.endangered-species
      operation: descendants

  - id: habitats-of-endangered
    select:
      entities:
        - kind: forest
        - kind: water_body
    traverse:
      from: query.endangered-species
      operation: outgoing
      relation_type: inhabits
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
  "query_id": "active-spots",
  "results": [
    {
      "id": "spot-yahiko-shrine",
      "type": "entity",
      "path": "/area-niigata-city/spot-yahiko-shrine",
      "object": {
        "id": "spot-yahiko-shrine",
        "kind": "tourism_spot",
        "name": "弥彦神社",
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

### Select All Tourism Spots

```yaml
select:
  entities:
    - kind: tourism_spot
```

### Select Active Hot Springs with Filters

```yaml
select:
  entities:
    - kind: hot_spring
where:
  status: active
  temperature_c: {ge: 40}
  source_count: {ge: 1}
```

### Select All Habitat Relations

```yaml
select:
  relations:
    - type: inhabits
```

### Traverse to Find Populations

```yaml
select:
  entities:
    - kind: population
traverse:
  from: species-japanese-crested-ibis
  operation: descendants
```

### Complex Query

```yaml
id: complex-query
select:
  entities:
    - kind: area
    - kind: tourism_spot
    - kind: cultural_asset
  relations:
    - type: located_in
    - type: depends_on
where:
  or:
    - kind: area
    - and:
      - kind: tourism_spot
      - status: active
traverse:
  from: area-niigata-city
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
