# YAML Syntax

## Overview

YAML is the canonical serialization format for the Object Model.

This specification defines the concrete YAML syntax.

All compliant implementations MUST support this syntax.

---

## Document Structure

A YAML document represents a single Graph.

The document root contains:

```yaml
objects:
  # Entities and Relations go here
```

---

## Entity Syntax

An Entity is defined with the following structure.

### Required Top-Level Properties

| Property | Type | Description |
|----------|------|-------------|
| id | string | Unique identifier |
| kind | string | Entity kind |
| name | string | Human-readable name |

### Attributes Section

The `attributes` sub-key contains optional properties common to all entities.

| Property | Type | Default | Description |
|----------|------|---------|-------------|
| owner | string | - | Parent Entity identifier for ownership |
| description | string | - | Documentation |
| status | enum | - | Lifecycle state |
| tags | list[string] | - | Labels |
| labels | map[string] | - | Key-value metadata |
| extensions | map[string] | - | Extension data |

### Spec Section

The `spec` sub-key contains kind-specific properties (soil_type, elevation_m, count, etc.).

### Basic Entity

```yaml
objects:
  - id: area-niigata-city
    kind: area
    name: 新潟市
```

### Entity with All Properties

```yaml
objects:
  - id: onsen-tsukioka
    kind: hot_spring
    name: 月岡温泉
    attributes:
      description: "硫黄泉の温泉地"
      status: active
      tags:
        - tourism
        - onsen
      labels:
        area: niigata-city
        season: all
      extensions:
        operator: tsukioka-onsen-kyodo
    spec:
      spring_quality: sulfur
      temperature_c: 78.0
      source_count: 1
```

---

## Nested Entity Definitions

Child entities can be defined inline within their parent entity's definition,
instead of as separate top-level objects. This provides a more concise and
hierarchical representation of the model.

### Nesting Rules

Each entity kind defines which child kinds can be nested and under which key.

A nesting rule with `any` as the parent kind means the child can be nested under any entity kind.

| Parent Kind | Nest Key | Child Kind |
|-------------|----------|------------|
| area | grounds | ground |
| area | terrains | terrain |
| area | water_bodies | water_body |
| area | forests | forest |
| area | tourism_spots | tourism_spot |
| area | species | species |
| area | cultural_assets | cultural_asset |
| terrain | grounds | ground |
| forest | grounds | ground |
| species | populations | population |
| tourism_spot | hot_springs | hot_spring |
| tourism_spot | events | event |
| hot_spring | events | event |

### Syntax

Nested children are defined as lists under the appropriate nest key.
The nest key can appear either inside the `spec` section or at the entity
definition level.

```yaml
objects:
  - id: area-sado-city
    kind: area
    name: 佐渡市
    spec:
      forests:
        - id: forest-sado-cedar
          name: 佐渡スギ林
          spec:
            grounds:
              - id: ground-forest-sado
                name: 林内地盤
                spec:
                  soil_type: volcanic_ash
                  elevation_m: 320
      species:
        - id: species-japanese-crested-ibis
          name: トキ
          spec:
            scientific_name: Nipponia nippon
            category: bird
            red_list_status: endangered
            populations:
              - id: pop-toki-sado-2025
                name: 個体群調査 2025
                spec:
                  count: 190
                  survey_method: visual_count
```

### Optional Fields in Nested Definitions

| Field | Required | Notes |
|-------|----------|-------|
| id | no | Scoped to parent if omitted (local reference only) |
| kind | optional | Inferred from the nest key |
| name | optional | Defaults to ID if omitted |
| spec | optional | Kind-specific properties |

### Species Populations

Populations are declared under the `populations` nest key of a `species`.

Each population entity is referenced using path notation:

```yaml
participants:
  source: pop-toki-sado-2025
  target: forest-sado-cedar
```

### Scoped IDs

When a nested entity omits its `id`, it receives a scoped ID that is only
referable within the parent's scope. The entity can be referenced using the
parent's path notation:

```yaml
participants:
  source: species-japanese-crested-ibis/pop-toki-sado-2025
```

Scoped entities cannot be referenced from outside their parent scope
using simple ID references.

### Ownership

Nested entities automatically receive their parent's ID as their `owner`.
The `owner` field should NOT be specified in nested definitions.

### Reference Syntax for Nested Entities

Entities with explicit IDs can be referenced by their ID:

```yaml
participants:
  source: pop-toki-sado-2025
  target: area-sado-city
```

Or by path notation:

```yaml
participants:
  source: area-sado-city/forest-sado-cedar/pop-toki-sado-2025
  target: area-sado-city
```

Scoped entities (without explicit IDs) must use path notation:

```yaml
participants:
  source: species-japanese-crested-ibis/pop-toki-sado-2025
```

### Mixed Definitions

Flat and nested definitions can be mixed in the same file:

```yaml
objects:
  # Flat definition
  - id: terrain-mt-yahiko
    kind: terrain
    name: 弥彦山
    attributes:
      owner: area-nishikanbara

  # Nested definition
  - id: spot-yahiko-shrine
    kind: tourism_spot
    name: 弥彦神社
    spec:
      hot_springs:
        - id: onsen-yahiko
          spec:
            spring_quality: simple
```

---

## Relation Syntax

A Relation is defined with the following structure.

### Required Top-Level Properties

| Property | Type | Description |
|----------|------|-------------|
| id | string | Unique identifier |
| type | string | Relation type |
| participants | list[string] or map | Entity references |

### Participant Formats

#### List Format (Symmetric Relations)

For symmetric relations like `near`:

```yaml
participants:
  - onsen-tsukioka
  - spot-yahiko-shrine
```

#### Map Format (Directed Relations)

For directed relations like `located_in`, `depends_on`:

```yaml
participants:
  source: spot-yahiko-shrine
  target: area-nishikanbara
```

### Attributes Section

The `attributes` sub-key contains optional properties common to all relations.

| Property | Type | Default | Description |
|----------|------|---------|-------------|
| description | string | - | Documentation |
| status | enum | - | Lifecycle state |
| tags | list[string] | - | Labels |
| labels | map[string] | - | Key-value metadata |
| extensions | map[string] | - | Extension data |

### Spec Section

The `spec` sub-key contains relation-type-specific properties (distance_km, walking_minutes, etc.).

### Relation with All Properties

```yaml
objects:
  - id: rel-near-onsen-spot
    type: near
    attributes:
      description: "温泉と神社の近接関係"
      status: active
      tags:
        - tourism
      labels:
        route: walking
    spec:
      walking_minutes: 15
    participants:
      - onsen-tsukioka
      - spot-yahiko-shrine
```

---

## Reference Syntax

References use Entity identifiers.

### Simple Reference

```yaml
source: spot-yahiko-shrine
target: onsen-tsukioka
```

### Qualified Reference

For unambiguous references:

```yaml
source: /area-nishikanbara/terrain-mt-yahiko/spot-yahiko-shrine
target: onsen-tsukioka
```

### Nested Entity Reference

Nested entities are referenced with path notation:

```yaml
participants:
  - species-japanese-crested-ibis/pop-toki-sado-2025
  - forest-sado-cedar
```

### Property Reference

Properties that reference other Objects use the `@` prefix to distinguish references from plain string values:

```yaml
spec:
  nearest_spot: "@spot-yahiko-shrine"
  related_events: ["@event-yahiko-fireworks"]
```

The `@` prefix indicates that the value is a reference to another Entity's ID.
At runtime, the parser converts `@`-prefixed strings to typed reference values.
When serializing back to YAML, the `@` prefix is automatically restored.

Reference properties are validated to ensure the referenced Entity exists in the graph.
A property value of `"@spot-yahiko-shrine"` references the Entity with ID `spot-yahiko-shrine`.
Path notation is also supported: `"@/area-nishikanbara/terrain-mt-yahiko/spot-yahiko-shrine"`.

---

## Complete Example

### Niigata Resource Model

```yaml
objects:
  # 地域
  - id: area-niigata-city
    kind: area
    name: 新潟市
    attributes:
      status: active
      labels:
        area_type: city
    spec:
      area_type: city
      population: 800000
      latitude: 37.9161
      longitude: 139.0364
      timezone: Asia/Tokyo

  # 水域
  - id: waterbody-shinano-river
    kind: water_body
    name: 信濃川
    attributes:
      owner: area-niigata-city
      status: active
    spec:
      water_type: river
      length_km: 367
      catchment_area_km2: 11900

  - id: waterbody-sea-of-japan
    kind: water_body
    name: 日本海
    attributes:
      owner: area-niigata-city
      status: active
    spec:
      water_type: sea
      max_depth_m: 3796

  # 観光スポット（温泉をネスト）
  - id: spot-yahiko-shrine
    kind: tourism_spot
    name: 弥彦神社
    attributes:
      status: active
    spec:
      spot_type: shrine_temple
      description: 越後一宮。弥彦山の麓に鎮座する。
      annual_visitors: 2000000
      hot_springs:
        - id: onsen-yahiko
          name: 弥彦温泉
          spec:
            spring_quality: simple
            temperature_c: 42.0
            source_count: 1

  # 種と個体群調査
  - id: species-japanese-crested-ibis
    kind: species
    name: トキ
    attributes:
      owner: area-niigata-city
      status: active
    spec:
      scientific_name: Nipponia nippon
      category: bird
      red_list_status: endangered
      populations:
        - id: pop-toki-2025
          name: トキ個体群調査 2025
          spec:
            count: 190
            survey_date: "2025-11-15"
            survey_method: visual_count

  # 文化財
  - id: cultural-asset-sado-gold-mine
    kind: cultural_asset
    name: 佐渡金山
    attributes:
      owner: area-niigata-city
      status: active
    spec:
      asset_type: archaeological
      designated_level: unesco
      designated_date: "2024-07-26"

  # イベント（温泉配下）
  - id: onsen-yahiko-matsuri
    kind: event
    name: 弥彦温泉まつり
    attributes:
      owner: spot-yahiko-shrine
      status: planned
    spec:
      season: autumn
      held_month: 10
      visitor_count: 20000

  # 近接関係（near・対称）
  - id: rel-near-onsen-spot
    type: near
    spec:
      walking_minutes: 15
    participants:
      - onsen-yahiko
      - spot-yahiko-shrine

  - id: rel-near-culture-water
    type: near
    participants:
      - cultural-asset-sado-gold-mine
      - waterbody-sea-of-japan

  # 位置関係（located_in・有向）
  - id: rel-locatedin-spot-area
    type: located_in
    participants:
      source: spot-yahiko-shrine
      target: area-niigata-city

  # 生息関係（inhabits・有向）
  - id: rel-inhabits-toki-river
    type: inhabits
    participants:
      source: pop-toki-2025
      target: waterbody-shinano-river
    spec:
      habitat_note: 稲田での採餌が観察される。

  # 所属関係（belongs_to）
  - id: rel-belongsto-population-species
    type: belongs_to
    participants:
      source: pop-toki-2025
      target: species-japanese-crested-ibis

  # 流路関係（flows_into・有向）
  - id: rel-flows-shinano-sea
    type: flows_into
    participants:
      source: waterbody-shinano-river
      target: waterbody-sea-of-japan
```

---

## Validation

### Required Fields

| Object | Required Fields |
|--------|-----------------|
| Entity | id, kind, name |
| Relation | id, type, participants |

### Ownership Validation

- Root Entity MUST NOT specify owner
- Non-root Entity MUST specify exactly one owner
- Owner identifier MUST reference an existing Entity
- Ownership MUST form exactly one tree

### Reference Validation

- References MUST point to existing Objects
- Nested entity references use path notation (entity/child)
- Unknown references are validation errors

### Identifier Rules

- IDs MUST be unique within their scope
- IDs SHOULD be descriptive and stable
- IDs SHOULD follow naming conventions (kebab-case recommended)

---

## Naming Conventions

### Recommended Patterns

| Pattern | Example |
|---------|---------|
| kebab-case | `spot-yahiko-shrine` |
| snake_case | `niigata_city` |
| camelCase | `niigataCity` |

### Kind Naming

- Use lowercase
- Use singular form
- Examples: `area`, `species`, `tourism_spot`

### Relation Type Naming

- Use snake_case
- Use verbs or verb phrases
- Examples: `located_in`, `inhabits`, `depends_on`

---

## Comments

Comments are preserved during round-trip conversion.

```yaml
# 地域情報
objects:
  - id: area-niigata-city
    kind: area
    name: 新潟市
    # 県庁所在地
    attributes:
      status: active
```
