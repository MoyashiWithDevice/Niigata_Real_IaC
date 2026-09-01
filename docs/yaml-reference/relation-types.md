# Relation Types

[← README](README.md)

---

## Relation Types一覧

| Type | Direction | Cardinality | Description |
|------|-----------|-------------|-------------|
| located_in | directed | 2 participants | Entityがareaまたはterrain内に位置する |
| inhabits | directed | 2 participants | 種が生息地に住む |
| near | symmetric | 2 participants | 2つの資源が地理的に近い |
| depends_on | directed | 2 participants | 方向性のある依存関係 |
| belongs_to | directed | 2 participants | 論理的な所属・帰属 |
| flows_into | directed | 2 participants | 河川などの流れの行き先 |

---

## Core Relation Types

### located_in

Entityがareaまたはterrainの中に位置することを表します。

| Property | Value |
|----------|-------|
| Direction | directed |
| Source | nature/tourism資源, species, population |
| Target | area, terrain |

**Additional Properties:**

| Property | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| distance_km | number | no | - | ターゲットからの距離（km） |

```yaml
- id: rel-forest-locatedin
  type: located_in
  participants:
    source: forest-osado-beech
    target: mt-kinpoku
```

---

### inhabits

種（species / population）が生息地に住むことを表します。

| Property | Value |
|----------|-------|
| Direction | directed |
| Source | species, population |
| Target | ground, terrain, water_body, forest, area |

**Additional Properties:**

| Property | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| habitat_note | string | no | - | 生息地利用に関するメモ |

```yaml
- id: rel-toki-inhabits
  type: inhabits
  participants:
    source: species-toki
    target: forest-osado-beech
  spec:
    habitat_note: Roosts in tall trees; feeds in restored rice paddies
```

---

### near

2つの資源が地理的に近いことを表します。対称関係のため、participantsはリスト形式で指定します。

| Property | Value |
|----------|-------|
| Direction | symmetric |
| Participants | tourism_spot, hot_spring, cultural_asset, event, ground, terrain, water_body, forest, area のうち2つ |

**Additional Properties:**

| Property | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| walking_minutes | integer | no | - | 徒歩での移動時間（分） |

```yaml
- id: rel-shukunegi-near-coast
  type: near
  participants:
    - spot-shukunegi
    - coast-otoline
  spec:
    walking_minutes: 25
```

---

### depends_on

方向性のある依存関係を表します。例えば温泉が源泉の地盤に依存する、イベントが景観に依存する、などです。

| Property | Value |
|----------|-------|
| Direction | directed |
| Source | hot_spring, tourism_spot, cultural_asset, event |
| Target | ground, terrain, water_body, forest, tourism_spot, hot_spring, cultural_asset, event |

**Additional Properties:**

| Property | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| dependency_type | string | no | - | 依存の性質 (enum: `source`, `landscape`, `access`, `ecosystem`, `event`) |
| critical | boolean | no | false | ターゲットを失うとソースの価値が失われるか |

```yaml
- id: rel-onsen-depends-ground
  type: depends_on
  participants:
    source: onsen-ogi
    target: ground-osado-hill
  spec:
    dependency_type: source
    critical: true
```

---

### belongs_to

論理的な所属・関連を表します。ネスト定義からは自動生成されます（子 → 親）。

| Property | Value |
|----------|-------|
| Direction | directed |
| Source | ground, terrain, water_body, forest, tourism_spot, hot_spring, cultural_asset, event, species, population |
| Target | area および nature/tourism資源Kind |

```yaml
- id: rel-goldmine-belongs
  type: belongs_to
  participants:
    source: asset-sado-goldmine
    target: sado-island
```

---

### flows_into

河川などの水の流れの行き先を表します。

| Property | Value |
|----------|-------|
| Direction | directed |
| Source | water_body |
| Target | water_body |

```yaml
- id: rel-flows-kamo
  type: flows_into
  participants:
    source: river-kamo-inlet
    target: lake-kamo
```
