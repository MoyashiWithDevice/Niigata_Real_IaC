# Projection Model

> **Implementation status:** The `src/projection/` implementation has been
> removed. Projection functionality is planned to be integrated into the Query
> model in a future release rather than maintained as a standalone package.
> This document describes the design intent; refer to the Query model
> (`spec/19-query-model.md`) for the currently supported behavior.

## Overview

A Projection transforms one Graph into another Graph.

A Projection never modifies the source Graph.

The output Graph represents the same resource knowledge from a different perspective.

Projections are composable and deterministic.

Every compliant implementation MUST support the Core Projection Model.

---

## Projection Structure

Every Projection is defined with:

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| id | string | yes | Unique projection identifier |
| name | string | yes | Human-readable name |
| description | string | no | Projection description |
| input | InputClause | yes | Input selection |
| operations | list[Operation] | yes | Transformation operations |
| output | OutputClause | no | Output configuration |

---

## Input Clause

The Input clause defines what enters the Projection.

### Input Types

| Type | Description |
|------|-------------|
| graph | Entire Graph |
| query | Query results |
| projection | Another Projection's output |

### Input Examples

```yaml
input:
  type: graph

input:
  type: query
  query_id: active-spots

input:
  type: projection
  projection_id: nature-topology
```

---

## Operations

A Projection performs one or more operations.

### Operation Types

| Operation | Description |
|-----------|-------------|
| select | Choose subset of Objects |
| filter | Remove Objects matching conditions |
| traverse | Follow Relations to discover Objects |
| aggregate | Combine multiple Objects into one |
| expand | Replace abstract Object with detailed Objects |
| annotate | Attach computed metadata |
| group | Organize Objects into collections |
| flatten | Simplify hierarchical structure |
| enrich | Add computed properties |
| transform | Transform Object properties |

---

## Select Operation

Choose a subset of Objects.

### Structure

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| type | enum | yes | "select" |
| entities | list[EntitySelector] | no | Entity selection criteria |
| relations | list[RelationSelector] | no | Relation selection criteria |

### Entity Selector

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| kind | string | no | Entity kind filter |
| where | WhereClause | no | Filtering conditions |

### Relation Selector

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| type | string | no | Relation type filter |
| where | WhereClause | no | Filtering conditions |

### Examples

```yaml
operations:
  - type: select
    entities:
      - kind: tourism_spot
      - kind: hot_spring
        where:
          status: active
    relations:
      - type: near
```

---

## Filter Operation

Remove Objects matching specific conditions.

### Structure

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| type | enum | yes | "filter" |
| action | enum | yes | "include" or "exclude" |
| target | enum | yes | "entities" or "relations" |
| where | WhereClause | yes | Filtering conditions |

### Examples

```yaml
# Include only active tourism spots
operations:
  - type: filter
    action: include
    target: entities
    where:
      kind: tourism_spot
      status: active

# Exclude events
operations:
  - type: filter
    action: exclude
    target: entities
    where:
      kind: event
```

---

## Traverse Operation

Follow Relations to discover additional Objects.

### Structure

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| type | enum | yes | "traverse" |
| direction | enum | yes | "forward", "backward", or "both" |
| relation_type | string | no | Relation type to follow |
| owner | boolean | no | Follow owner property for ownership traversal |
| depth | integer | no | Maximum traversal depth |
| include_origin | boolean | no | Include source Objects |

### Examples

```yaml
# Get all habitats of selected species (inhabits relations)
operations:
  - type: traverse
    direction: forward
    relation_type: inhabits
    depth: 2
    include_origin: true

# Get all ancestors of selected populations (via owner property)
operations:
  - type: traverse
    direction: backward
    owner: true
    depth: 3
    include_origin: false
```

---

## Aggregate Operation

Combine multiple Objects into a single derived Object.

### Structure

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| type | enum | yes | "aggregate" |
| source_selector | Selector | yes | Objects to aggregate |
| target_kind | string | yes | Kind of derived Object |
| group_by | list[string] | no | Properties to group by |
| aggregations | list[Aggregation] | no | Aggregation functions |

### Aggregation

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| property | string | yes | Property to aggregate |
| function | enum | yes | Aggregation function |
| target_property | string | yes | Property name in derived Object |

### Aggregation Functions

| Function | Description |
|----------|-------------|
| count | Count Objects |
| sum | Sum numeric values |
| avg | Average numeric values |
| min | Minimum value |
| max | Maximum value |
| list | Collect values into list |
| first | First value |
| last | Last value |

### Examples

```yaml
# Create area summary from populations
operations:
  - type: aggregate
    source_selector:
      kind: population
    target_kind: area_summary
    group_by:
      - labels.area
    aggregations:
      - property: count
        function: sum
        target_property: total_individuals
      - property: id
        function: count
        target_property: survey_count
      - property: count
        function: max
        target_property: max_count
```

---

## Expand Operation

Replace an abstract Object with more detailed Objects.

### Structure

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| type | enum | yes | "expand" |
| source_selector | Selector | yes | Objects to expand |
| expansion | ExpansionConfig | yes | Expansion configuration |

### Expansion Config

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| target_kind | string | yes | Kind of expanded Objects |
| property_mapping | map | no | Property mapping |
| owner | string | no | Owner identifier for expanded Objects |

### Examples

```yaml
# Expand tourism spots into their events
operations:
  - type: expand
    source_selector:
      kind: tourism_spot
    expansion:
      target_kind: event
      property_mapping:
        name: event_name
        season: season
      owner: tourism_spot
```

---

## Annotate Operation

Attach computed metadata to Objects.

### Structure

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| type | enum | yes | "annotate" |
| target_selector | Selector | yes | Objects to annotate |
| annotations | list[Annotation] | yes | Annotations to attach |

### Annotation

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| property | string | yes | Property name |
| value | any | no | Static value |
| expression | string | no | Computed value expression |
| source_property | string | no | Source property for computation |

### Annotation Functions

| Function | Description |
|----------|-------------|
| count | Count related Objects |
| sum | Sum related property |
| concat | Concatenate strings |
| format | Format string |
| timestamp | Current timestamp |
| hash | Generate hash |

### Examples

```yaml
# Add population counts to species
operations:
  - type: annotate
    target_selector:
      kind: species
    annotations:
      - property: survey_count
        expression: "count(children.populations)"
      - property: latest_count
        expression: "sum(children.populations.count)"
      - property: annotation_timestamp
        function: timestamp
```

---

## Group Operation

Organize Objects into logical collections.

### Structure

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| type | enum | yes | "group" |
| source_selector | Selector | yes | Objects to group |
| group_kind | string | yes | Kind of group Object |
| group_by | list[string] | yes | Properties to group by |

### Examples

```yaml
# Group events by season
operations:
  - type: group
    source_selector:
      kind: event
    group_kind: event_group
    group_by:
      - season

# Group hot springs by spring quality
operations:
  - type: group
    source_selector:
      kind: hot_spring
    group_kind: onsen_group
    group_by:
      - spring_quality
```

---

## Flatten Operation

Simplify hierarchical structure.

### Structure

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| type | enum | yes | "flatten" |
| target_selector | Selector | yes | Objects to flatten |
| preserve_relations | boolean | no | Keep existing Relations |

### Examples

```yaml
# Flatten hierarchy to flat list
operations:
  - type: flatten
    target_selector:
      kind: area
    preserve_relations: true
```

---

## Enrich Operation

Add computed properties to Objects.

### Structure

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| type | enum | yes | "enrich" |
| target_selector | Selector | yes | Objects to enrich |
| properties | list[ComputedProperty] | yes | Properties to add |

### Computed Property

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| name | string | yes | Property name |
| type | string | yes | Property type |
| expression | string | yes | Computation expression |

### Examples

```yaml
# Add visitor metrics
operations:
  - type: enrich
    target_selector:
      kind: area
    properties:
      - name: visitor_density
        type: number
        expression: "(annual_visitors / population) * 100"
      - name: spot_per_km2
        type: number
        expression: "count(children.tourism_spots) / area_km2"
```

---

## Transform Operation

Transform Object properties.

### Structure

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| type | enum | yes | "transform" |
| target_selector | Selector | yes | Objects to transform |
| transformations | list[Transformation] | yes | Transformations to apply |

### Transformation

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| property | string | yes | Property to transform |
| operation | enum | yes | Transformation operation |
| value | any | no | Operation value |

### Transformation Operations

| Operation | Description |
|-----------|-------------|
| rename | Rename property |
| cast | Change property type |
| set | Set property value |
| remove | Remove property |
| default | Set default if null |

### Examples

```yaml
# Transform properties
operations:
  - type: transform
    target_selector:
      kind: population
    transformations:
      - property: name
        operation: rename
        value: survey_name
      - property: survey_date
        operation: cast
        value: string
      - property: status
        operation: default
        value: unknown
```

---

## Derived Objects

Projections MAY create derived Objects.

### Derived Object Structure

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| id | string | yes | New unique identifier |
| kind | string | yes | Derived Object kind |
| name | string | yes | Human-readable name |
| provenance | Provenance | yes | Source information |

### Provenance

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| source_ids | list[string] | yes | Source Object identifiers |
| projection_id | string | yes | Projection identifier |
| timestamp | string | yes | Creation timestamp |
| operation | string | yes | Operation that created it |

### Derived Object Example

```yaml
- id: area-summary-niigata-city
  kind: area_summary
  name: 新潟市 個体群サマリー
  provenance:
    source_ids:
      - pop-toki-2025
      - pop-serow-2025
      - pop-kamoshika-2025
    projection_id: nature-topology
    timestamp: "2025-12-01T10:30:00Z"
    operation: aggregate
  total_individuals: 240
  survey_count: 3
  max_count: 190
```

---

## Output Clause

The Output clause configures Projection output.

### Output Options

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| format | enum | graph | Output format |
| include_provenance | boolean | true | Include provenance info |
| include_derived | boolean | true | Include derived Objects |

### Output Formats

| Format | Description |
|--------|-------------|
| graph | Full Graph output |
| summary | Summary only |
| filtered | Filtered subset |

---

## Composition

Projections MAY be chained.

### Chain Example

```yaml
projections:
  - id: nature-topology
    name: Nature Topology
    input:
      type: graph
    operations:
      - type: select
        entities:
          - kind: area
          - kind: terrain
          - kind: forest
          - kind: water_body
        relations:
          - type: belongs_to

  - id: habitat-topology
    name: Habitat Topology
    input:
      type: projection
      projection_id: nature-topology
    operations:
      - type: select
        entities:
          - kind: species
          - kind: population
        relations:
          - type: inhabits
          - type: belongs_to

  - id: documentation
    name: Documentation View
    input:
      type: projection
      projection_id: habitat-topology
    operations:
      - type: annotate
        target_selector:
          kind: species
        annotations:
          - property: description
            expression: "format('%s (%s)', name, scientific_name)"
```

---

## Determinism

A Projection MUST be deterministic.

Applying the same Projection to the same Graph MUST always produce the same output Graph.

---

## Side Effects

Projections are pure operations.

A Projection MUST NOT modify:

- the source Graph
- external systems
- persistent storage

---

## Extensibility

Implementations MAY define additional Projection operations.

Additional operations MUST preserve semantic equivalence with the source Graph.
