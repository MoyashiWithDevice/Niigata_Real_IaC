# Relation Types（リレーション種別）

## 概要

Relation Type は Entity 間の接続の意味論を定義する。

すべての Relation は type を定義しなければならない（MUST）。

本仕様は次の Relation Type を定義する。

実装は拡張を通じて追加の型を導入してよい（MAY）。

---

## 共通プロパティ

すべての Relation は type によらず次のプロパティを持つ。

### 必須

- id
- type
- participants

### 任意

- description
- status
- tags
- labels
- extensions

個々の Relation Type は追加のプロパティを定義してよい（MAY）。

---

## 方向性

Relation は方向性により分類される。

| Type | 説明 |
|------|------|
| directed | source と target の参加者を持つ |
| symmetric | すべての参加者が対等 |

---

## コア Relation Types

### located_in

Entity が area または terrain の中に位置することを表す。

| Property | Value |
|----------|-------|
| Direction | directed |
| Source | 位置する Entity |
| Target | area / terrain |
| Cardinality | N:N |

#### 参加者制約

| Role | 許可される Kind |
|------|----------------|
| source | ground, terrain, water_body, forest, tourism_spot, hot_spring, cultural_asset, event, species, population |
| target | area, terrain |

参加者数は 2（min 2 / max 2）。

#### プロパティ

| Property | Type | Required | Default | 説明 |
|----------|------|----------|---------|------|
| distance_km | number | no | - | 対象からの距離（km、0 以上） |

#### 例

```yaml
- id: rel-locatedin-spot-yahiko
  type: located_in
  participants:
    source: spot-yahiko-shrine
    target: terrain-mt-yahiko
  status: active

- id: rel-locatedin-toki-sado
  type: located_in
  participants:
    source: species-japanese-crested-ibis
    target: area-sado-city
```

---

### inhabits

種が生息地に生息することを表す。

| Property | Value |
|----------|-------|
| Direction | directed |
| Source | 種 / 個体群 |
| Target | 生息地 |
| Cardinality | N:N |

#### 参加者制約

| Role | 許可される Kind |
|------|----------------|
| source | species, population |
| target | ground, terrain, water_body, forest, area |

参加者数は 2（min 2 / max 2）。

#### プロパティ

| Property | Type | Required | Default | 説明 |
|----------|------|----------|---------|------|
| habitat_note | string | no | - | 生息利用に関する備考 |

#### 例

```yaml
- id: rel-inhabits-toki-forest
  type: inhabits
  participants:
    source: species-japanese-crested-ibis
    target: forest-sado-cedar
  habitat_note: 巣はスギ・ヒノキの高木に作られる。
```

---

### near

2 つの資源が地理的に近接していることを表す対称関係。

| Property | Value |
|----------|-------|
| Direction | symmetric |
| Participants | 対等な 2 参加者 |
| Cardinality | N:N |

#### 参加者制約

両参加者とも次の kind が許可される:

- tourism_spot, hot_spring, cultural_asset, event
- ground, terrain, water_body, forest
- area

参加者数は 2（min 2 / max 2）。

#### 制約

- near は対称である。参加者の順序は意味を持たない。
- 片方向のみの定義でも双方向の近接性を表す。

#### プロパティ

| Property | Type | Required | Default | 説明 |
|----------|------|----------|---------|------|
| walking_minutes | integer | no | - | 2 地点間の徒歩時間（分、0 以上） |

#### 例

```yaml
- id: rel-near-onsen-spot
  type: near
  participants:
    - onsen-tsukioka
    - spot-yahiko-shrine
  walking_minutes: 15
```

---

### depends_on

Entity 間の方向的依存を表す。例えば温泉が源泉や地形に依存する関係など。

| Property | Value |
|----------|-------|
| Direction | directed |
| Source | 依存する Entity |
| Target | 依存先の Entity |
| Cardinality | N:N |

#### 参加者制約

| Role | 許可される Kind |
|------|----------------|
| source | hot_spring, tourism_spot, cultural_asset, event |
| target | ground, terrain, water_body, forest, tourism_spot, hot_spring, cultural_asset, event |

参加者数は 2（min 2 / max 2）。

#### 制約

- A が B に depends_on する場合、A は B を必要とするが、B は A を必要としない。
- 依存関係は循環してよい（A → B → A は有効）。所有関係（ownership）の循環とは異なる。

#### プロパティ

| Property | Type | Required | Default | 説明 |
|----------|------|----------|---------|------|
| dependency_type | string | no | - | 依存の性質（source, landscape, access, ecosystem, event） |
| critical | boolean | no | false | 対象の喪失が依存元の価値を失わせるか |

#### 例

```yaml
- id: rel-depends-onsen-ground
  type: depends_on
  participants:
    source: onsen-tsukioka
    target: ground-yahiko-hill
  dependency_type: source
  critical: true

- id: rel-depends-event-landscape
  type: depends_on
  participants:
    source: event-nagaoka-hanabi
    target: waterbody-shinano-river
  dependency_type: landscape
  critical: true
```

---

### belongs_to

論理的な所属・帰属を表す。

| Property | Value |
|----------|-------|
| Direction | directed |
| Source | 所属する Entity（メンバー） |
| Target | 所属先の Entity（グループ） |
| Cardinality | N:N |

#### 参加者制約

| Role | 許可される Kind |
|------|----------------|
| source | ground, terrain, water_body, forest, tourism_spot, hot_spring, cultural_asset, event, species, population |
| target | area, ground, terrain, water_body, forest, tourism_spot, hot_spring, cultural_asset, event |

参加者数は 2（min 2 / max 2）。

#### 制約

- 所属は所有（ownership）を意味しない。ネスト定義から自動生成される belongs_to も、owner ツリーとは独立した Relation である。
- Entity は複数のグループに属してよい。
- ネストキーで子を定義すると、子 → 親の belongs_to 関係が自動生成される。

#### プロパティ

| Property | Type | Required | Default | 説明 |
|----------|------|----------|---------|------|
| (none) | - | - | - | 共通 relation プロパティのみ使用 |

#### 例

```yaml
- id: rel-belongsto-forest-area
  type: belongs_to
  participants:
    source: forest-myoko-beech
    target: area-myoko-city
  status: active

- id: rel-belongsto-population-species
  type: belongs_to
  participants:
    source: pop-toki-sado-2025
    target: species-japanese-crested-ibis
  status: active
```

---

### flows_into

河川などの水の流れの行き先を表す。

| Property | Value |
|----------|-------|
| Direction | directed |
| Source | 流れる水域 |
| Target | 流れ先の水域 |
| Cardinality | N:N |

#### 参加者制約

| Role | 許可される Kind |
|------|----------------|
| source | water_body |
| target | water_body |

参加者数は 2（min 2 / max 2）。

#### 制約

- 水系の合流・注入手順を表す。サイクルを持ってはならない（SHOULD NOT）。

#### プロパティ

| Property | Type | Required | Default | 説明 |
|----------|-------|----------|---------|------|
| (none) | - | - | - | 共通 relation プロパティのみ使用 |

#### 例

```yaml
- id: rel-flows-shinano-sea
  type: flows_into
  participants:
    source: waterbody-shinano-river
    target: waterbody-sea-of-japan
```

---

## 参加者制約

### 許可される参加者 Kind

各 Relation Type はどの Entity Kind が参加できるかを定義してよい（MAY）。

#### コア型の制約一覧

| Relation Type | Direction | Source Kinds | Target Kinds |
|---------------|-----------|--------------|--------------|
| located_in | directed | ground, terrain, water_body, forest, tourism_spot, hot_spring, cultural_asset, event, species, population | area, terrain |
| inhabits | directed | species, population | ground, terrain, water_body, forest, area |
| near | symmetric | tourism_spot, hot_spring, cultural_asset, event, ground, terrain, water_body, forest, area | （同左） |
| depends_on | directed | hot_spring, tourism_spot, cultural_asset, event | ground, terrain, water_body, forest, tourism_spot, hot_spring, cultural_asset, event |
| belongs_to | directed | ground, terrain, water_body, forest, tourism_spot, hot_spring, cultural_asset, event, species, population | area, ground, terrain, water_body, forest, tourism_spot, hot_spring, cultural_asset, event |
| flows_into | directed | water_body | water_body |

すべてのコア型は min_participants = 2 / max_participants = 2 である。

### 拡張 Relation Types

拡張は `namespace.type` 形式で追加の relation type を定義してよい（MAY）。

例:

- `sado.docks_at` - 佐渡航路と埠頭の係留関係
- `ecotour.guided_by` - エコツアーとガイドの関係

### カーディナリティ

カーディナリティは参加者間にいくつのリレーションが存在しうるかを定義する。

| Type | Cardinality | 説明 |
|------|-------------|------|
| located_in | N:N | 多対多の位置関係 |
| inhabits | N:N | 一種が複数生息地、一生息地に多種 |
| near | N:N | 多対多の近接関係 |
| depends_on | N:N | 多対多の依存関係 |
| belongs_to | N:N | 多メンバー、多グループ |
| flows_into | N:N | 多対多の流路 |

---

## Status 値

すべての Relation は status を持ってよい（MAY）。

コア仕様が定義する status は次の通り。

| Status | 説明 |
|--------|------|
| planned | 計画されているが未活性 |
| active | 有効 |
| maintenance | 保全中 |
| deprecated | 廃止予定 |
| offline | 無効 |
| standby | 待機状態 |

---

## 同一性

Relation はその識別子によって一意に識別される。

participants を変更すると既存の Relation を変更したことになる。

識別子を変更すると別の Relation を作成したことになる。
