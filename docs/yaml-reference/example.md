# Complete Example

[← README](README.md)

---

新潟の自然・観光資源モデルの例（佐渡島）です。動作する完全なモデルは `demo/niigata/model.yaml` にあります。

- Flat / Nested 混在定義（ネストは拡張 `belongs_to` を自動生成）
- 共通属性（`description`, `status`, `tags`, `labels`, `extensions`）を entity / relation 両方に記述
- リレーション種類（`located_in`, `inhabits`, `near`, `depends_on`, `belongs_to`, `flows_into`）の実演
- ネストされた子（例: `mt-kinpoku/ground-osado-hill`）も単純 ID で参照可能

```yaml
# Niigata nature and tourism resource model — Sado Island example.
# Demonstrates flat + nested (mixed) definitions, common attributes
# on entities and relations, auto-generated belongs_to relations, and
# the full set of core entity kinds and relation types.
schema_version: "1.0"

objects:
  # ==============================================================
  # Area (root) with nested children.
  # The nest keys (terrains, forests, species, tourism_spots) are
  # declared inline; each nested child receives its owner and an
  # auto-generated belongs_to relation automatically.
  # ==============================================================
  - id: sado-island
    kind: area
    name: Sado Island
    attributes:
      description: Island in the Sea of Japan, home of the reintroduced crested ibis.
      status: active
      labels:
        prefecture: niigata
        region: sado
      tags:
        - island
        - unesco-geopark
    spec:
      area_type: island
      population: 54897
      latitude: 38.04
      longitude: 138.35
      timezone: Asia/Tokyo

      terrains:
        - id: mt-kinpoku
          name: Mount Kinpoku
          attributes:
            description: Highest peak of Sado, at the heart of the Osado range.
            status: active
            tags:
              - mountain
              - hiking
          spec:
            terrain_type: mountain
            elevation_m: 1172
            prominence_m: 1092
            grounds:
              - id: ground-osado-hill
                name: Osado Hillside Ground
                attributes:
                  description: Volcanic hillside used as the recharge area for local hot springs.
                  status: maintenance
                  labels:
                    microregion: osado
                spec:
                  soil_type: volcanic_ash
                  elevation_m: 640
                  slope_deg: 28
                  stability: watch

        - id: coast-otoline
          name: Otago Coast Cliffs
          attributes:
            description: Eroded coastline along the southern shore of the island.
            status: active
          spec:
            terrain_type: coast
            elevation_m: 210

      forests:
        - id: forest-osado-beech
          name: Osado Beech Forest
          attributes:
            description: Old-growth beech forest on the Osado mountainside.
            status: active
            tags:
              - old-growth
              - forest
          spec:
            forest_type: beech
            area_ha: 1250
            grounds:
              - id: ground-beech-soil
                name: Beech Forest Soil Plot
                attributes:
                  description: Loam soil plot beneath the beech canopy.
                  status: active
                spec:
                  soil_type: loam
                  elevation_m: 720
                  slope_deg: 18
                  stability: stable

      species:
        - id: species-toki
          name: Crested Ibis (Toki)
          attributes:
            description: Bird that was extinct in the wild in Japan and later reintroduced on Sado.
            status: active
            tags:
              - symbolic-species
              - released-population
            labels:
              conservation: national
            extensions:
              reintroduction_program: sado-toki-center
          spec:
            scientific_name: Nipponia nippon
            category: bird
            red_list_status: endangered

        - id: species-medaka
          name: Sado Medaka
          attributes:
            description: Native rice-fish population preserved in irrigation channels.
            status: active
            tags:
              - endemic
            labels:
              conservation: prefectural
          spec:
            scientific_name: Oryzias latipes
            category: fish
            red_list_status: near_threatened

      tourism_spots:
        - id: spot-shukunegi
          name: Shukunegi Boat Village
          attributes:
            description: Edo-period shipwright village on Ogi Bay, with preserved timber storehouses.
            status: active
            labels:
              municipality: sado-city
            tags:
              - historic
              - village
          spec:
            spot_type: historic
            description: Edo-period shipwright village on Ogi Bay
            annual_visitors: 310000
            hot_springs:
              - id: onsen-ogi
                name: Ogi Onsen
                attributes:
                  description: Chloride spring feeding the town's public baths.
                  status: active
                  tags:
                    - onsen
                  extensions:
                    operator: ogi-onsen-kyodo
                spec:
                  spring_quality: chloride
                  temperature_c: 72.5
                  source_count: 3
            events:
              - id: event-earth-celebration
                name: Earth Celebration
                attributes:
                  description: World music festival tied to the taiko group Kodo.
                  status: planned
                  tags:
                    - festival
                    - music
                spec:
                  season: summer
                  held_month: 8
                  visitor_count: 25000

  # --------------------------------------------------------------
  # Flat entities (leaf top-level, explicit owner).
  # These stay flat to illustrate the "mixed definition" syntax.
  # --------------------------------------------------------------
  - id: toki-census-2025
    kind: population
    name: Toki Census 2025
    attributes:
      description: Spring visual count conducted across restored rice paddies.
      owner: species-toki
      status: active
      extensions:
        survey_org: sado-toki-center
    spec:
      count: 731
      survey_date: "2025-03-15"
      survey_method: visual_count

  - id: toki-census-2024
    kind: population
    name: Toki Census 2024
    attributes:
      description: Drone-assisted count from the previous year.
      owner: species-toki
      status: maintenance
    spec:
      count: 688
      survey_date: "2024-03-10"
      survey_method: drone

  - id: medaka-count-2025
    kind: population
    name: Medaka Count 2025
    attributes:
      description: Transect survey of the Kamo inlet irrigation network.
      owner: species-medaka
      status: active
    spec:
      count: 4200
      survey_date: "2025-06-01"
      survey_method: transect

  - id: lake-kamo
    kind: water_body
    name: Lake Kamo
    attributes:
      description: Shallow brackish lake fringed by reed beds and rice paddies.
      owner: sado-island
      status: active
      tags:
        - lake
        - wetland
    spec:
      water_type: lake
      max_depth_m: 8
      catchment_area_km2: 41

  - id: river-kamo-inlet
    kind: water_body
    name: Kamo Inlet Channel
    attributes:
      description: Channel carrying drainage from the Kamo basin into the lake.
      owner: sado-island
      status: active
    spec:
      water_type: river
      length_km: 1.6

  - id: asset-sado-goldmine
    kind: cultural_asset
    name: Sado Kinzan Gold Mine
    attributes:
      description: Historic gold mine inscribed as a World Heritage Site.
      owner: sado-island
      status: active
      labels:
        designation: world-heritage
      extensions:
        designation_authority: unesco
    spec:
      asset_type: historic_site
      designated_level: unesco
      designated_date: "2024-07-26"

  # --------------------------------------------------------------
  # Relations. Simple IDs reference entities (including nested ones).
  # --------------------------------------------------------------
  # The beech forest lies on the Osado mountainside.
  - id: rel-forest-locatedin
    type: located_in
    attributes:
      description: Beech forest occupies the slopes of Mount Kinpoku.
      status: active
    participants:
      source: forest-osado-beech
      target: mt-kinpoku

  # Toki inhabit the beech forest habitat.
  - id: rel-toki-inhabits
    type: inhabits
    attributes:
      description: Toki roost and forage in the forest edge and paddies.
      status: active
      tags:
        - roosting
    participants:
      source: species-toki
      target: forest-osado-beech
    spec:
      habitat_note: Roosts in tall trees; feeds in restored rice paddies

  # Medaka live in Lake Kamo.
  - id: rel-medaka-inhabits
    type: inhabits
    attributes:
      description: Medaka thrive in the lake's inflow channels.
      status: active
    participants:
      source: species-medaka
      target: lake-kamo

  # The inlet channel flows into Lake Kamo.
  - id: rel-flows-kamo
    type: flows_into
    attributes:
      description: Kamo inlet discharges into the lake.
      status: active
    participants:
      source: river-kamo-inlet
      target: lake-kamo

  # Shukunegi is close to the Otago coastline.
  # Symmetric relation uses the list participant format.
  - id: rel-shukunegi-near-coast
    type: near
    attributes:
      description: Short walk from the village to the cliff viewpoint.
      status: active
      tags:
        - access
      labels:
        route: walking
    participants:
      - spot-shukunegi
      - coast-otoline
    spec:
      walking_minutes: 25

  # Ogi Onsen depends on the hillside ground as its source area.
  - id: rel-onsen-depends-ground
    type: depends_on
    attributes:
      description: The onsen draws its mineral recharge from the hillside.
      status: active
      labels:
        dependency: source
    participants:
      source: onsen-ogi
      target: ground-osado-hill
    spec:
      dependency_type: source
      critical: true

  # The gold mine tour is a key attraction of the island's tourism.
  - id: rel-goldmine-belongs
    type: belongs_to
    attributes:
      description: The mine belongs to the island's cultural estate.
      status: active
    participants:
      source: asset-sado-goldmine
      target: sado-island

  # The festival depends on the scenic village landscape.
  - id: rel-event-depends-spot
    type: depends_on
    attributes:
      description: Earth Celebration relies on the village setting.
      status: planned
      labels:
        dependency: landscape
    participants:
      source: event-earth-celebration
      target: spot-shukunegi
    spec:
      dependency_type: landscape
      critical: false
```
