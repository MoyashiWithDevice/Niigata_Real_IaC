# Validation

[← README](README.md)

---

## 必須フィールド

| Object | Required Fields |
|--------|-----------------|
| Entity | id, kind, name |
| Relation | id, type, participants |

## 識別子ルール

- IDはそのスコープ内でユニークである必要があります
- IDは記述的で安定していることが望ましいです
- IDは命名規則（kebab-case推奨）に従うことが望ましいです
- IDにスラッシュ（`/`）を含めることはできません

---

## 命名規則

| 対象 | パターン | 例 |
|------|----------|-----|
| ID | kebab-case (推奨) | `sado-island` |
| Kind | 小文字・単数形 | `area`, `species` |
| Relation Type | snake_case | `located_in`, `depends_on` |
| Property | snake_case | `soil_type`, `survey_date` |

---

## Graph制約

### Ownership制約

| Constraint | Description |
|------------|-------------|
| single_owner | Root以外のEntityは正確に1つのownerを持つ |
| tree_structure | Ownershipは正確に1つのツリーを形成 |
| no_cycles | Ownershipはサイクルを含まない |
| owner_exists | Owner識別子は既存のEntityを参照する |

> **拡張ルート権限:** 拡張によってルート権限を持つKindが許可されます。拡張が登録したKind（例: 組織を表す拡張Kind）には `AddAllowedRootKind` でルート権限を付与でき、複数ルートの共存が可能になります。

### Reference制約

| Constraint | Description |
|------------|-------------|
| valid_reference | Referencesは既存のObjectsを指す |
| unique_id | Object識別子はユニークである |
| relation_exists | Relationsは既存のObjectsを参照する |

### Nesting制約

| Constraint | Description |
|------------|-------------|
| no-slash-in-id | IDにスラッシュを含めることはできません |
| valid-nesting-parent | 親子関係がスキーマのネスト定義と一致すること（例: population の親は species） |

### Cardinality制約

すべてのコアRelation Typeは参加者数2（source/target）で検証されます。詳細は [Relation Types](relation-types.md) を参照してください。

---

## 検証ルール一覧

検証エンジンが実行するコアルールの一覧です。

### Graph整合性

| Rule ID | Severity | Description |
|---------|----------|-------------|
| unique-id | error | Entity/RelationのIDが重複していないこと |
| valid-reference | error | Relationのparticipantsが存在するオブジェクトを参照すること |
| valid-owner | error | Entityのownerが存在するEntityを指すこと |
| single-owner | error | ルートは1つのみ（ルート権限を持つKindを除く） |
| valid-property | warning | spec/propertiesがスキーマ定義（型・enum・required・min/max）に適合すること。未定義プロパティも報告される |

### Entity

| Rule ID | Severity | Description |
|---------|----------|-------------|
| required-kind | error | Entityにkindがあること |
| required-name | error | Entityにnameがあること |
| valid-kind | error | kindがスキーマに定義されていること |
| valid-status | warning | statusが有効な値であること |
| no-slash-in-id | error | IDにスラッシュを含めないこと |
| valid-nesting-parent | warning | 親子ネストがスキーマのネスト定義と一致すること |

### Relation

| Rule ID | Severity | Description |
|---------|----------|-------------|
| required-type | error | Relationにtypeがあること |
| required-participants | error | Relationが最低2つのparticipantを持つこと |
| valid-type | error | typeがスキーマに定義されていること |
| valid-direction | error | 有向Relationがsource/targetを持つこと |
| valid-cardinality | error | 参加者数がスキーマ定義のmin/max内であること |
| valid-participant-kind | warning | participantのkindがそのRelation Typeで許可されていること |

### Ownership

| Rule ID | Severity | Description |
|---------|----------|-------------|
| ownership-tree | error | 所有権が単一の接続された木を形成すること |
| no-ownership-cycle | error | 所有権チェーンにサイクルがないこと |
| root-entity | error | グラフにルートEntityが存在し、複数ないこと |

### Reference

| Rule ID | Severity | Description |
|---------|----------|-------------|
| dangling-reference | error | owner / participants / `@`プロパティ参照が既存オブジェクトを指すこと |
| invalid-path | error | パス参照が正しい所有権チェーンと一致すること |

### Profile

プロファイルで必須Kind/Relationを宣言した場合に評価されます。

| Rule ID | Severity | Description |
|---------|----------|-------------|
| profile-required-kind | error | プロファイルが要求するkindのEntityが少なくとも1つあること |
| profile-required-relation | error | プロファイルが要求するtypeのRelationが少なくとも1つあること |

---

## Niigataドメインルール

新潟の自然・観光資源モデル向けのドメイン固有ルールです。

| Rule ID | Severity | Description |
|---------|----------|-------------|
| population-requires-species | error | population Entityは必ず species をownerとすること |
| positive-count | warning | population のcountは0以上であること |
| valid-niigata-coordinates | warning | area の座標が新潟県の概算範囲（緯度36.6–38.7、経度137.9–139.9、佐渡島を含む）内にあること |

### 例: population は species を親に持つ

```yaml
# OK: species 配下にネスト
- id: species-toki
  kind: species
  name: Crested Ibis (Toki)
  attributes:
    owner: sado-island
  spec:
    scientific_name: Nipponia nippon
    category: bird
    red_list_status: endangered
    populations:
      - id: toki-census-2025
        name: Toki Census 2025
        spec:
          count: 731
          survey_date: "2025-03-15"
          survey_method: visual_count
```

```yaml
# NG: area 直下の population → population-requires-species (error)
- id: orphan-count
  kind: population
  name: Orphan Count
  attributes:
    owner: sado-island   # 親がspeciesではない
  spec:
    count: -5            # negative-count → positive-count (warning)
```

### 例: 新潟県外の座標

```yaml
# 東京の座標を指定すると警告が出る
- id: tokyo-bay
  kind: area
  name: Tokyo Bay
  spec:
    area_type: district
    latitude: 35.65      # 新潟範囲(36.6-38.7)外 → valid-niigata-coordinates (warning)
    longitude: 139.75
```
