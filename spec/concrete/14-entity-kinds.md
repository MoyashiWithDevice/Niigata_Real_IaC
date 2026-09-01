# Entity Kinds（エンティティ種別）

## 概要

Entity Kind は、モデル内に存在しうるオブジェクトのカテゴリを定義する。

すべての Entity は kind を定義しなければならない（MUST）。

本仕様（Niigata Nature and Tourism Resource Schema）は、新潟県の自然・観光資源管理のための次の Entity Kind を定義する。

実装は拡張（extension）を通じて追加の kind を導入してもよい（MAY）。

---

## 共通プロパティ

すべての Entity は kind によらず次の共通プロパティを持つ。

### 必須

- id
- kind
- name

### 任意

- description
- status
- tags
- labels
- extensions

個々の Entity Kind は追加のプロパティを定義してよい（MAY）。

---

## 地理・自然環境

### area

市町村・地区・島などの地理的領域。

#### プロパティ

| Property | Type | Required | Default | 説明 |
|----------|------|----------|---------|------|
| area_type | string | no | - | 行政区分（prefecture, city, town, village, district, island） |
| population | integer | no | - | 人口（0 以上） |
| address | string | no | - | 住所 |
| latitude | number | no | - | 緯度（10 進度数、-90〜90） |
| longitude | number | no | - | 経度（10 進度数、-180〜180） |
| timezone | string | no | - | タイムゾーン識別子 |

#### 典型的な所有関係

- owned by: root（owner 未指定）

#### ネスト可能な子

| Nest Key | Child Kind |
|----------|------------|
| grounds | ground |
| terrains | terrain |
| water_bodies | water_body |
| forests | forest |
| tourism_spots | tourism_spot |
| species | species |
| cultural_assets | cultural_asset |

ネストされた子は `belongs_to` 関係（子 → 親）を自動的に受け取る。

#### 典型的な Relation

- located_in → terrain
- near ↔ area / terrain / 自然資源 / 観光資源（対称）

#### 例

```yaml
- id: area-niigata-city
  kind: area
  name: 新潟市
  status: active
  area_type: city
  population: 800000
  latitude: 37.9161
  longitude: 139.0364
  timezone: Asia/Tokyo
  tags:
    - prefecture-capital
    - joetsu-shinkansen
```

---

### ground

特定地点の地盤・土壌状態。

#### プロパティ

| Property | Type | Required | Default | 説明 |
|----------|------|----------|---------|------|
| soil_type | string | no | - | 優占土壌型（loam, sand, clay, gravel, rock, volcanic_ash, peat） |
| elevation_m | number | no | - | 海抜高度（m、-10 以上） |
| slope_deg | number | no | - | 斜面角度（度、0〜90） |
| stability | string | no | - | 地盤安定度分類（stable, watch, unstable, landslide_prone, subsidence） |
| latitude | number | no | - | 緯度（10 進度数、-90〜90） |
| longitude | number | no | - | 経度（10 進度数、-180〜180） |

#### 典型的な所有関係

- owned by: area, terrain, forest

#### 典型的な Relation

- belongs_to → area / terrain / forest（ネスト時に自動生成）

#### 例

```yaml
- id: ground-yahiko-hill
  kind: ground
  name: 弥彦丘陵の地盤
  soil_type: loam
  elevation_m: 120.5
  slope_deg: 15.0
  stability: stable
```

---

### terrain

山・平野・海岸・谷などの地形。

#### プロパティ

| Property | Type | Required | Default | 説明 |
|----------|------|----------|---------|------|
| terrain_type | string | no | - | 地形タイプ（mountain, hill, plain, coast, valley, plateau, cave） |
| elevation_m | number | no | - | 最高標高（m、0 以上） |
| prominence_m | number | no | - | 地形突出度（m、0 以上） |
| latitude | number | no | - | 緯度（10 進度数、-90〜90） |
| longitude | number | no | - | 経度（10 進度数、-180〜180） |

#### 典型的な所有関係

- owned by: area

#### ネスト可能な子

| Nest Key | Child Kind |
|----------|------------|
| grounds | ground |

#### 典型的な Relation

- belongs_to → area（ネスト時に自動生成）
- located_in ← 自然資源 / 観光資源

#### 例

```yaml
- id: terrain-mt-yahiko
  kind: terrain
  name: 弥彦山
  status: active
  terrain_type: mountain
  elevation_m: 634
  prominence_m: 587
```

---

### water_body

河川・湖・海域・池・沼などの水域。

#### プロパティ

| Property | Type | Required | Default | 説明 |
|----------|------|----------|---------|------|
| water_type | string | no | - | 水域タイプ（river, lake, sea, pond, marsh, waterfall） |
| length_km | number | no | - | 長さ（km、0 以上。河川向け） |
| max_depth_m | number | no | - | 最大深度（m、0 以上） |
| catchment_area_km2 | number | no | - | 流域面積（km²、0 以上） |
| latitude | number | no | - | 緯度（10 進度数、-90〜90） |
| longitude | number | no | - | 経度（10 進度数、-180〜180） |

#### 典型的な所有関係

- owned by: area

#### 典型的な Relation

- flows_into → water_body
- inhabits ← species / population

#### 例

```yaml
- id: waterbody-shinano-river
  kind: water_body
  name: 信濃川
  status: active
  water_type: river
  length_km: 367
  catchment_area_km2: 11900
```

---

### forest

森林・林地。

#### プロパティ

| Property | Type | Required | Default | 説明 |
|----------|------|----------|---------|------|
| forest_type | string | no | - | 優占林型（natural, beech, cedar_plantation, pine, bamboo, mixed） |
| area_ha | number | no | - | 面積（ha、0 以上） |
| latitude | number | no | - | 緯度（10 進度数、-90〜90） |
| longitude | number | no | - | 経度（10 進度数、-180〜180） |

#### 典型的な所有関係

- owned by: area

#### ネスト可能な子

| Nest Key | Child Kind |
|----------|------------|
| grounds | ground |

#### 典型的な Relation

- belongs_to → area（ネスト時に自動生成）
- inhabits ← species / population

#### 例

```yaml
- id: forest-myoko-beech
  kind: forest
  name: 妙高ブナ林
  status: active
  forest_type: beech
  area_ha: 250.0
```

---

## 生態系

### species

地域に生息する動植物種。

#### プロパティ

| Property | Type | Required | Default | 説明 |
|----------|------|----------|---------|------|
| scientific_name | string | no | - | 学名（ラテン語名） |
| category | string | no | - | 生物分類（mammal, bird, reptile, amphibian, fish, insect, plant, other） |
| red_list_status | string | no | - | レッドリスト評価（extinct, extinct_in_wild, critically_endangered, endangered, vulnerable, near_threatened, least_concern, data_deficient） |

#### 典型的な所有関係

- owned by: area

#### ネスト可能な子

| Nest Key | Child Kind |
|----------|------------|
| populations | population |

**注意:** population は必ず species を owner としなければならない（MUST）。この制約は検証ルール `population-requires-species` により強制される。

#### 典型的な Relation

- inhabits → water_body / forest / terrain / ground / area

#### 例

```yaml
- id: species-japanese-crested-ibis
  kind: species
  name: トキ
  status: active
  scientific_name: Nipponia nippon
  category: bird
  red_list_status: endangered
```

---

### population

調査由来の種別個体群記録。

#### プロパティ

| Property | Type | Required | Default | 説明 |
|----------|------|----------|---------|------|
| count | integer | yes | - | 観測または推定された個体数（0 以上） |
| survey_date | string | no | - | 調査日（YYYY-MM-DD） |
| survey_method | string | no | - | 調査手法（visual_count, transect, drone, camera_trap, interview, estimate） |

#### 典型的な所有関係

- owned by: species（必須）

#### 典型的な Relation

- inhabits → 生息地
- located_in → area / terrain

#### 例

```yaml
- id: pop-toki-sado-2025
  kind: population
  name: トキ個体群調査 2025
  count: 190
  survey_date: "2025-11-15"
  survey_method: visual_count
```

---

## 観光・文化

### tourism_spot

景勝地・公園・神社仏閣・博物館などの観光スポット。

#### プロパティ

| Property | Type | Required | Default | 説明 |
|----------|------|----------|---------|------|
| spot_type | string | no | - | スポット種別（scenic, viewpoint, park, historic, shrine_temple, museum, market, ski_resort, beach） |
| description | string | no | - | 簡単な説明 |
| annual_visitors | integer | no | - | 年間来場者数（0 以上） |
| latitude | number | no | - | 緯度（10 進度数、-90〜90） |
| longitude | number | no | - | 経度（10 進度数、-180〜180） |

#### 典型的な所有関係

- owned by: area

#### ネスト可能な子

| Nest Key | Child Kind |
|----------|------------|
| hot_springs | hot_spring |
| events | event |

#### 典型的な Relation

- located_in → area / terrain
- near ↔ 他の観光資源 / 自然的資源 / area（対称）
- depends_on → 自然環境

#### 例

```yaml
- id: spot-yahiko-shrine
  kind: tourism_spot
  name: 弥彦神社
  status: active
  spot_type: shrine_temple
  description: 越後一宮。弥彦山の麓に鎮座する。
  annual_visitors: 2000000
```

---

### hot_spring

温泉源または入浴施設。

#### プロパティ

| Property | Type | Required | Default | 説明 |
|----------|------|----------|---------|------|
| spring_quality | string | no | - | 泉質（sulfur, chloride, simple, carbonated, iron, alum, sulfate） |
| temperature_c | number | no | - | 源泉温度（摂氏） |
| source_count | integer | no | - | 湧出数（0 以上） |
| latitude | number | no | - | 緯度（10 進度数、-90〜90） |
| longitude | number | no | - | 経度（10 進度数、-180〜180） |

#### 典型的な所有関係

- owned by: tourism_spot

#### ネスト可能な子

| Nest Key | Child Kind |
|----------|------------|
| events | event |

#### 典型的な Relation

- depends_on → ground / terrain（源泉・地盤への依存）

#### 例

```yaml
- id: onsen-tsukioka
  kind: hot_spring
  name: 月岡温泉
  status: active
  spring_quality: sulfur
  temperature_c: 78.0
  source_count: 1
```

---

### cultural_asset

史跡・文化財。

#### プロパティ

| Property | Type | Required | Default | 説明 |
|----------|------|----------|---------|------|
| asset_type | string | no | - | 文化財種別（historic_site, treasure, building, monument, archaeological, folk_property） |
| designated_level | string | no | - | 指定階層（national, prefectural, municipal, unesco） |
| designated_date | string | no | - | 指定日（YYYY-MM-DD） |
| latitude | number | no | - | 緯度（10 進度数、-90〜90） |
| longitude | number | no | - | 経度（10 進度数、-180〜180） |

#### 典型的な所有関係

- owned by: area

#### 典型的な Relation

- located_in → area / terrain
- near ↔ 観光資源

#### 例

```yaml
- id: cultural-asset-iwamuro-tanada
  kind: cultural_asset
  name: 山北棚田群
  status: active
  asset_type: historic_site
  designated_level: national
```

---

### event

祭り・恒例行事などのイベント。

#### プロパティ

| Property | Type | Required | Default | 説明 |
|----------|------|----------|---------|------|
| season | string | no | - | 開催季節（spring, summer, autumn, winter, all） |
| held_month | integer | no | - | 開催月（1〜12） |
| visitor_count | integer | no | - | 1 回あたり来場者数（0 以上） |

#### 典型的な所有関係

- owned by: tourism_spot, hot_spring

#### 典型的な Relation

- belongs_to → tourism_spot / hot_spring（ネスト時に自動生成）
- near ↔ 観光資源

#### 例

```yaml
- id: event-nagaoka-hanabi
  kind: event
  name: 長岡まつり大花火大会
  status: active
  season: summer
  held_month: 8
  visitor_count: 1000000
```

---

## ベンダー Kind

ベンダー固有の Entity Kind はコア kind を置き換えてはならない（MUST NOT）。

ベンダーは拡張を通じて追加の kind を導入してよい（MAY）。

### 拡張の命名規約

拡張 kind は名前空間プレフィックスを使用しなければならない（MUST）。

形式: `<vendor>.<kind>`

例:

- `sado.ferry` - 佐渡航路固有の拡張
- `nagaoka.fireworks` - 長岡花火固有の拡張
- `ecotour.tour` - エコツアー事業者固有の拡張
- `geo.survey_point` - 測量ポイントの拡張プロパティを持つ kind

---

## Status 値

すべての Entity は status を持ってよい（MAY）。

コア仕様が定義する status は次の通り。

| Status | 説明 |
|--------|------|
| planned | 計画済みだが未整備 |
| active | 利用可能・運用中 |
| maintenance | 保全・メンテナンス中 |
| deprecated | 提供終了が予定されている |
| offline | 利用不可 |
| standby | 待機状態（例：冗長構成の待機メンバー） |

実装は追加の status を導入してよい（MAY）。

---

## 同一性

2 つの Entity は、識別子が異なる場合に異なるものとみなされる。

プロパティの変更は新しい Entity を作らない。

識別子の変更は異なる Entity の作成を意味する。
