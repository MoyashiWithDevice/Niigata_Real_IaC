# Entity Kinds

[← README](README.md)

---

## Entity Kinds一覧

| Kind | Category | Description |
|------|----------|-------------|
| area | Geography | Geographic area such as a city, town, village, district, or island |
| ground | Nature | Ground or soil condition at a specific site |
| terrain | Nature | Landform such as a mountain, plain, coast, or valley |
| water_body | Nature | River, lake, sea area, pond, or marsh |
| forest | Nature | Forest or wooded area |
| species | Nature | Animal or plant species living in the area |
| population | Survey | Population record of a species from a survey |
| tourism_spot | Tourism | Tourist attraction such as a scenic spot, park, shrine, or museum |
| hot_spring | Tourism | Hot spring source or bath facility |
| cultural_asset | Culture | Historic site or cultural property |
| event | Tourism | Festival or recurring local event |

---

## ネスト定義（Nesting）

親Entityの定義内で直接子Entityを定義できます。ネストされた子は自動的に`belongs_to`リレーション（子 → 親）を生成し、親を`owner`として受け取ります。

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

---

## 地域（Area）

### area

都市・町村・地区・島などの地理的領域です。Graphのルート（所有者なし）になることを想定しています。

| Property | Type | Required | Constraints | Description |
|----------|------|----------|-------------|-------------|
| area_type | string | no | enum: `prefecture`, `city`, `town`, `village`, `district`, `island` | 行政区分タイプ |
| population | integer | no | min: 0 | 住民数 |
| address | string | no | - | 物理住所 |
| latitude | number | no | min: -90 / max: 90 | 緯度（10進度数） |
| longitude | number | no | min: -180 / max: 180 | 経度（10進度数） |
| timezone | string | no | - | タイムゾーン識別子 |

**Ownership:** ルート想定（ownerなし）

**ネスト可能な子:** `grounds` (ground), `terrains` (terrain), `water_bodies` (water_body), `forests` (forest), `tourism_spots` (tourism_spot), `species` (species), `cultural_assets` (cultural_asset)

```yaml
- id: sado-island
  kind: area
  name: Sado Island
  attributes:
    status: active
    labels:
      prefecture: niigata
  spec:
    area_type: island
    population: 54897
    latitude: 38.04
    longitude: 138.35
```

> **Note:** `latitude` / `longitude` は新潟県の範囲（緯度36.6–38.7、経度137.9–139.9、佐渡島を含む概算）から外れると [valid-niigata-coordinates](validation.md#niigataドメインルール) の警告が出ます。

---

## 自然資源

### ground

特定地点の地盤・土壌状態です。area / terrain / forest の子としてネストできます。

| Property | Type | Required | Constraints | Description |
|----------|------|----------|-------------|-------------|
| soil_type | string | no | enum: `loam`, `sand`, `clay`, `gravel`, `rock`, `volcanic_ash`, `peat` | 優勢な土壌タイプ |
| elevation_m | number | no | min: -10 | 海抜高度（m） |
| slope_deg | number | no | min: 0 / max: 90 | 斜面角度（度） |
| stability | string | no | enum: `stable`, `watch`, `unstable`, `landslide_prone`, `subsidence` | 地盤安定度分類 |

```yaml
- id: ground-osado-hill
  kind: ground
  name: Osado Hillside Ground
  attributes:
    owner: mt-kinpoku
  spec:
    soil_type: volcanic_ash
    elevation_m: 640
    slope_deg: 28
    stability: watch
```

---

### terrain

山・平野・海岸・谷などの地形です。

| Property | Type | Required | Constraints | Description |
|----------|------|----------|-------------|-------------|
| terrain_type | string | no | enum: `mountain`, `hill`, `plain`, `coast`, `valley`, `plateau`, `cave` | 地形タイプ |
| elevation_m | number | no | min: 0 | 最高標高（m） |
| prominence_m | number | no | min: 0 | 地形突出度（m） |

**Ownership:** area

**ネスト可能な子:** `grounds` (ground)

```yaml
- id: mt-kinpoku
  kind: terrain
  name: Mount Kinpoku
  attributes:
    owner: sado-island
  spec:
    terrain_type: mountain
    elevation_m: 1172
    prominence_m: 1092
```

---

### water_body

河川・湖・海域・池・沼などの水域です。

| Property | Type | Required | Constraints | Description |
|----------|------|----------|-------------|-------------|
| water_type | string | no | enum: `river`, `lake`, `sea`, `pond`, `marsh`, `waterfall` | 水域タイプ |
| length_km | number | no | min: 0 | 長さ（km、河川用） |
| max_depth_m | number | no | min: 0 | 最大深度（m） |
| catchment_area_km2 | number | no | min: 0 | 集水面積（km²） |

**Ownership:** area

```yaml
- id: lake-kamo
  kind: water_body
  name: Lake Kamo
  attributes:
    owner: sado-island
  spec:
    water_type: lake
    max_depth_m: 8
    catchment_area_km2: 41
```

---

### forest

森林・林地です。

| Property | Type | Required | Constraints | Description |
|----------|------|----------|-------------|-------------|
| forest_type | string | no | enum: `natural`, `beech`, `cedar_plantation`, `pine`, `bamboo`, `mixed` | 優勢な森林タイプ |
| area_ha | number | no | min: 0 | 面積（ha） |

**Ownership:** area

**ネスト可能な子:** `grounds` (ground)

```yaml
- id: forest-osado-beech
  kind: forest
  name: Osado Beech Forest
  attributes:
    owner: sado-island
  spec:
    forest_type: beech
    area_ha: 1250
```

---

### species

地域に生息する動植物の種です。

| Property | Type | Required | Constraints | Description |
|----------|------|----------|-------------|-------------|
| scientific_name | string | no | - | 学名（ラテン語名） |
| category | string | no | enum: `mammal`, `bird`, `reptile`, `amphibian`, `fish`, `insect`, `plant`, `other` | 生物カテゴリ |
| red_list_status | string | no | enum: `extinct`, `extinct_in_wild`, `critically_endangered`, `endangered`, `vulnerable`, `near_threatened`, `least_concern`, `data_deficient` | レッドリスト保全状況 |

**Ownership:** area

**ネスト可能な子:** `populations` (population)

> **Note:** population は必ず species を親とする必要があります（[population-requires-species](validation.md#niigataドメインルール) ルール、error）。

```yaml
- id: species-toki
  kind: species
  name: Crested Ibis (Toki)
  attributes:
    owner: sado-island
  spec:
    scientific_name: Nipponia nippon
    category: bird
    red_list_status: endangered
```

---

## 調査記録

### population

調査による種の個体群記録です。必ず species の子として定義します。

| Property | Type | Required | Constraints | Description |
|----------|------|----------|-------------|-------------|
| count | integer | **yes** | min: 0 | 観察または推定された個体数 |
| survey_date | string | no | - | 調査日（YYYY-MM-DD） |
| survey_method | string | no | enum: `visual_count`, `transect`, `drone`, `camera_trap`, `interview`, `estimate` | 計数方法 |

**Ownership:** species

```yaml
- id: toki-census-2025
  kind: population
  name: Toki Census 2025
  attributes:
    owner: species-toki
  spec:
    count: 731
    survey_date: "2025-03-15"
    survey_method: visual_count
```

---

## 観光資源

### tourism_spot

景勝地・公園・神社仏閣・博物館などの観光スポットです。

| Property | Type | Required | Constraints | Description |
|----------|------|----------|-------------|-------------|
| spot_type | string | no | enum: `scenic`, `viewpoint`, `park`, `historic`, `shrine_temple`, `museum`, `market`, `ski_resort`, `beach` | スポットタイプ |
| description | string | no | - | 簡単な説明 |
| annual_visitors | integer | no | min: 0 | 年間来場者数 |

**Ownership:** area

**ネスト可能な子:** `hot_springs` (hot_spring), `events` (event)

```yaml
- id: spot-shukunegi
  kind: tourism_spot
  name: Shukunegi Boat Village
  attributes:
    owner: sado-island
  spec:
    spot_type: historic
    description: Edo-period shipwright village on Ogi Bay
    annual_visitors: 310000
```

---

### hot_spring

温泉源または入浴施設です。

| Property | Type | Required | Constraints | Description |
|----------|------|----------|-------------|-------------|
| spring_quality | string | no | enum: `sulfur`, `chloride`, `simple`, `carbonated`, `iron`, `alum`, `sulfate` | 泉質 |
| temperature_c | number | no | - | 源泉温度（℃） |
| source_count | integer | no | min: 0 | 湧出源の数 |

**Ownership:** tourism_spot

**ネスト可能な子:** `events` (event)

```yaml
- id: onsen-ogi
  kind: hot_spring
  name: Ogi Onsen
  attributes:
    owner: spot-shukunegi
  spec:
    spring_quality: chloride
    temperature_c: 72.5
    source_count: 3
```

---

## 文化財

### cultural_asset

史跡・文化財です。

| Property | Type | Required | Constraints | Description |
|----------|------|----------|-------------|-------------|
| asset_type | string | no | enum: `historic_site`, `treasure`, `building`, `monument`, `archaeological`, `folk_property` | 文化財タイプ |
| designated_level | string | no | enum: `national`, `prefectural`, `municipal`, `unesco` | 指定階層 |
| designated_date | string | no | - | 指定日（YYYY-MM-DD） |

**Ownership:** area

```yaml
- id: asset-sado-goldmine
  kind: cultural_asset
  name: Sado Kinzan Gold Mine
  attributes:
    owner: sado-island
  spec:
    asset_type: historic_site
    designated_level: unesco
    designated_date: "2024-07-26"
```

---

### event

祭りや恒例の地域イベントです。

| Property | Type | Required | Constraints | Description |
|----------|------|----------|-------------|-------------|
| season | string | no | enum: `spring`, `summer`, `autumn`, `winter`, `all` | 開催季節 |
| held_month | integer | no | min: 1 / max: 12 | 開催月（1–12） |
| visitor_count | integer | no | min: 0 | 1回あたりの来場者数 |

**Ownership:** tourism_spot, hot_spring

```yaml
- id: event-earth-celebration
  kind: event
  name: Earth Celebration
  attributes:
    owner: spot-shukunegi
  spec:
    season: summer
    held_month: 8
    visitor_count: 25000
```
