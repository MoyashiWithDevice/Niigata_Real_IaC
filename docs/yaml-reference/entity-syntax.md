# Entity Syntax

[← README](README.md)

---

## 必須プロパティ

| Property | Type | Description |
|----------|------|-------------|
| id | string | ユニーク識別子（kebab-case推奨） |
| kind | string | Entity kind（小文字・単数形） |
| name | string | 人間が読める名前 |

## 任意の共通プロパティ（`attributes:` 配下）

必須プロパティ以外の共通プロパティは、`attributes:` サブキーの配下に配置します。

| Property | Type | Default | Description |
|----------|------|---------|-------------|
| owner | string | - | 親Entity ID（所有権階層） |
| description | string | - | ドキュメント（Markdown対応） |
| status | enum | - | ライフサイクル状態 |
| tags | list[string] | - | グループ用ラベル |
| labels | map[string] | - | マシン可読メタデータ |
| extensions | map[string] | - | 拡張データ |

## ステータス値

| Value | Description |
|-------|-------------|
| planned | 未デプロイ |
| active | 稼働中 |
| maintenance | メンテナンス中 |
| deprecated | 削除予定 |
| offline | オフライン |
| standby | スタンバイ（冗長メンバー等） |

## 基本的なEntity

```yaml
- id: sado-island
  kind: area
  name: Sado Island
```

## 全プロパティを指定したEntity

```yaml
- id: mt-kinpoku
  kind: terrain
  name: Mount Kinpoku
  attributes:
    description: "Sado's highest peak"
    status: active
    tags:
      - mountain
      - hiking
    labels:
      municipality: sado-city
      prefecture: niigata
    extensions:
      trailhead: ono-game
  spec:
    terrain_type: mountain
    elevation_m: 1172
    prominence_m: 1092
```

---

## ネスト定義（Nested Entity Definition）

子エンティティを親エンティティの定義内で直接定義できます。

### ネスト可能な組み合わせ

| 親 Kind | ネストキー | 子 Kind |
|---------|-----------|---------|
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

### 基本構文

```yaml
objects:
  - id: sado-island
    kind: area
    name: Sado Island
    spec:
      area_type: island
      species:
        - id: species-toki
          name: Crested Ibis (Toki)
          spec:
            scientific_name: Nipponia nippon
            category: bird
            populations:
              - id: toki-census-2025
                spec:
                  count: 731
                  survey_date: "2025-03-15"
                  survey_method: visual_count
```

### 省略可能なフィールド

| フィールド | 必須 | 備考 |
|-----------|------|------|
| id | 任意 | 省略時は自動生成（`_{parentID}-{childKind}`） |
| kind | 任意 | ネストキーから自動推測 |
| name | 任意 | 省略時はIDが使用される |
| spec | 任意 | Kind固有プロパティ |

### ID自動生成

ネストされたエンティティが`id`を省略した場合、内部用スコープIDが自動生成されます：
`_{parent-id}-{child-kind}`

このIDは内部処理用で、ユーザーはパス表記（`parent/child`）で参照します。

注意: 同一Kindの子を複数`id`なしで定義すると同じ自動IDになり衝突します。
`id`を省略するのは親配下に1つの子だけの場合に限定してください。

### 所有権

ネストされたエンティティは自動的に親のIDを`owner`として受け取ります。ネスト定義で`owner`フィールドを指定しないでください。

### 参照構文

ネストされたエンティティはIDで参照できます：

```yaml
participants:
  source: species-toki
  target: forest-osado-beech
```

パス表記でも参照できます：

```yaml
participants:
  source: sado-island/species-toki/toki-census-2025
  target: forest-osado-beech
```

### フラットとネストの混在

同じファイル内でフラット定義とネスト定義を混在できます：

```yaml
objects:
  # フラット定義
  - id: forest-osado-beech
    kind: forest
    name: Osado Beech Forest
    attributes:
      owner: sado-island

  # ネスト定義
  - id: spot-shukunegi
    kind: tourism_spot
    name: Shukunegi Boat Village
    attributes:
      owner: sado-island
    spec:
      spot_type: historic
      events:
        - id: event-earth-celebration
```
