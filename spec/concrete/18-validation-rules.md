# Validation Rules（検証ルール）

## 概要

Validation Rule は、Graph を Schema に対して評価する方法を定義する。

検証は決定的（deterministic）かつ副作用なしで行われる。

準拠実装はすべてのコア検証ルールをサポートしなければならない（MUST）。

---

## ルール構造

すべての Validation Rule は次のように定義される。

| Field | Type | Required | 説明 |
|-------|------|----------|------|
| id | string | yes | 一意なルール識別子 |
| name | string | yes | 人間が読める名前 |
| description | string | no | ルールの説明 |
| severity | enum | yes | Finding の重大度 |

---

## 重大度レベル

すべての Validation Finding は重大度レベルを持つ。

| Level | 説明 |
|-------|------|
| info | 参考情報 |
| warning | 潜在的な問題。検証はパスする |
| error | 修正が必須の違反。検証は失敗する |

実装は追加のレベルを定義してよい（MAY）。

---

## Finding 構造

検証は Finding を生成する。

| Field | Type | 説明 |
|-------|------|------|
| rule_id | string | ルール識別子 |
| severity | enum | Finding の重大度 |
| message | string | 人間が読めるメッセージ |
| object_id | string | 対象オブジェクトの識別子 |
| object_type | enum | オブジェクト型（entity/relation） |
| path | string | Graph 内のオブジェクトパス |

### Finding の例

```json
{
  "rule_id": "population-requires-species",
  "severity": "error",
  "message": "population entity \"pop-toki-01\" has no owner (must be owned by a species)",
  "object_id": "pop-toki-01",
  "object_type": "entity"
}
```

---

## コア検証ルール

### Graph 整合性ルール

| Rule ID | Severity | 説明 |
|---------|----------|------|
| unique-id | error | オブジェクト識別子は一意でなければならない（MUST） |
| valid-reference | error | 関係の参加者参照は既存オブジェクトを指さなければならない（MUST） |
| valid-owner | error | owner 識別子は既存 Entity を参照しなければならない（MUST） |
| single-owner | error | root 以外の Entity には owner が必須。root は 1 つのみ（root 権限を持つ kind を除く） |
| valid-property | warning | spec プロパティはスキーマ定義に適合しなければならない（型・enum・min/max・required）。未定義プロパティも報告される |

### Entity ルール

| Rule ID | Severity | 説明 |
|---------|----------|------|
| required-kind | error | Entity は kind を定義しなければならない（MUST） |
| required-name | error | Entity は name を定義しなければならない（MUST） |
| valid-kind | error | Entity kind は Schema で定義されていなければならない（MUST） |
| valid-status | warning | status は有効な列挙値であるべき（SHOULD）。Entity と Relation の両方に適用 |
| no-slash-in-id | error | Entity ID にスラッシュ `/` を含めてはならない（MUST NOT） |
| valid-nesting-parent | warning | 所有関係の親子は Schema で定義されたネストであるべき（SHOULD） |

### Relation ルール

| Rule ID | Severity | 説明 |
|---------|----------|------|
| required-type | error | Relation は type を定義しなければならない（MUST） |
| required-participants | error | Relation は最低 2 参加者を持たなければならない（MUST） |
| valid-type | error | Relation type は Schema で定義されていなければならない（MUST） |
| valid-direction | error | directed relation は source と target を持たなければならない（MUST） |
| valid-cardinality | error | 参加者数は min/max 制約を満たさなければならない（MUST） |
| valid-participant-kind | warning | 参加者の kind は型の制約で許可されているべき（SHOULD） |

### Ownership ルール

| Rule ID | Severity | 説明 |
|---------|----------|------|
| ownership-tree | error | 所有関係は単一ツリーを形成しなければならない（MUST）。全ルートが root 権限を持つ kind の場合は森を許容 |
| no-ownership-cycle | error | 所有関係にサイクルを含んではならない（MUST NOT） |
| root-entity | error | root Entity は 1 つでなければならない（root 権限を持つ kind を除く） |

### Reference ルール

| Rule ID | Severity | 説明 |
|---------|----------|------|
| dangling-reference | error | `@` 参照・owner・participants は既存オブジェクトに解決されなければならない（MUST） |
| invalid-path | error | パス表記は所有ツリーと一致しなければならない（末尾セグメントが ID、中間セグメントが実際の親） |

### 新潟ドメインルール

| Rule ID | Severity | 説明 |
|---------|----------|------|
| population-requires-species | error | population Entity は species を owner としなければならない（MUST） |
| positive-count | warning | population の count は 0 以上であるべき（SHOULD） |
| valid-niigata-coordinates | warning | area の緯度経度は新潟県の概略範囲内であるべき（SHOULD） |

---

## コアルール定義

### unique-id

```yaml
id: unique-id
name: Unique Identifier
description: "オブジェクト識別子は一意でなければならない"
severity: error
condition: "all object.id are unique"
```

### valid-reference

```yaml
id: valid-reference
name: Valid Reference
description: "Relation の参加者は既存オブジェクトを参照しなければならない"
severity: error
condition: "all relation.participants reference existing objects"
```

### valid-owner

```yaml
id: valid-owner
name: Valid Owner
description: "owner は既存 Entity を参照しなければならない"
severity: error
condition: "entity.owner is undefined or references an existing entity"
```

### single-owner

```yaml
id: single-owner
name: Single Owner
description: "root 以外の Entity には owner が必須。root は 1 つのみ"
severity: error
condition: "exactly one root unless all roots have root authority"
```

### valid-property

```yaml
id: valid-property
name: Valid Property
description: "spec プロパティはスキーマ定義に適合しなければならない"
severity: warning
condition: "entity.spec conforms to schema property definitions and contains no undefined properties"
```

### ownership-tree

```yaml
id: ownership-tree
name: Ownership Tree
description: "所有関係は単一ツリーを形成しなければならない"
severity: error
condition: "ownership.forms.single.tree and tree reaches every entity"
```

### no-ownership-cycle

```yaml
id: no-ownership-cycle
name: No Ownership Cycle
description: "所有関係にサイクルを含んではならない"
severity: error
condition: "owner chain never repeats an entity"
```

### root-entity

```yaml
id: root-entity
name: Root Entity
description: "root Entity は 1 つでなければならない"
severity: error
condition: "count(entities without owner) == 1 unless all roots are root-authorized kinds"
```

### required-kind

```yaml
id: required-kind
name: Required Kind
description: "Entity は kind を定義しなければならない"
severity: error
condition: "entity.kind is defined"
```

### required-name

```yaml
id: required-name
name: Required Name
description: "Entity は name を定義しなければならない"
severity: error
condition: "entity.name is defined"
```

### no-slash-in-id

```yaml
id: no-slash-in-id
name: No Slash in Entity ID
description: "Entity ID にスラッシュを含めてはならない"
severity: error
condition: "'/' not in entity.id"
```

### valid-nesting-parent

```yaml
id: valid-nesting-parent
name: Valid Nesting Parent
description: "親子のネストは Schema で定義されているべき"
severity: warning
condition: "nesting(parent.kind, entity.kind) is defined in schema"
```

### required-type

```yaml
id: required-type
name: Required Type
description: "Relation は type を定義しなければならない"
severity: error
condition: "relation.type is defined"
```

### required-participants

```yaml
id: required-participants
name: Required Participants
description: "Relation は最低 2 参加者を持たなければならない"
severity: error
condition: "relation.participants.count >= 2"
```

### valid-kind

```yaml
id: valid-kind
name: Valid Kind
description: "Entity kind は Schema で定義されていなければならない"
severity: error
condition: "entity.kind exists in schema.entity_kinds"
```

### valid-type

```yaml
id: valid-type
name: Valid Type
description: "Relation type は Schema で定義されていなければならない"
severity: error
condition: "relation.type exists in schema.relation_types"
```

### valid-status

```yaml
id: valid-status
name: Valid Status
description: "status は有効な列挙値であるべき"
severity: warning
condition: "object.status in ['planned', 'active', 'maintenance', 'deprecated', 'offline', 'standby']"
```

### valid-direction

```yaml
id: valid-direction
name: Valid Direction
description: "directed relation は source と target を持たなければならない"
severity: error
condition: "schema.direction == 'directed' implies relation.source is defined and relation.target is defined"
```

### valid-cardinality

```yaml
id: valid-cardinality
name: Valid Cardinality
description: "参加者数は制約を満たさなければならない"
severity: error
condition: "relation.participants.count between schema.min_participants and schema.max_participants"
```

### valid-participant-kind

```yaml
id: valid-participant-kind
name: Valid Participant Kind
description: "参加者の kind は型の制約で許可されているべき"
severity: warning
condition: "all relation.participants.kind in schema.relation_types[type].allowed_kinds"
```

### dangling-reference

```yaml
id: dangling-reference
name: Dangling Reference
description: "@ 参照・owner・participants は既存オブジェクトに解決されなければならない"
severity: error
condition: "all @-references resolve to existing objects"
```

### invalid-path

```yaml
id: invalid-path
name: Invalid Path
description: "パス表記は所有ツリーと一致しなければならない"
severity: error
condition: "each path segment is an existing entity owned by the previous segment"
```

---

## 新潟ドメインルール定義

### population-requires-species

population Entity は必ず species を owner としなければならない。

違反例:

```yaml
objects:
  - id: pop-orphan
    kind: population
    name: 孤立した個体群記録
    count: 12
    # owner 未指定 → error:
    # population entity "pop-orphan" has no owner (must be owned by a species)

  - id: pop-wrong-parent
    kind: population
    name: 親が種ではない個体群記録
    attributes:
      owner: forest-myoko-beech   # species ではない → error
    count: 5
```

正しい例:

```yaml
objects:
  - id: species-japanese-serow
    kind: species
    name: ニホンカモシカ
    category: mammal

  - id: pop-serow-2025
    kind: population
    name: カモシカ生息数調査 2025
    attributes:
      owner: species-japanese-serow   # species を owner とする
    spec:
      count: 42
      survey_date: "2025-10-01"
      survey_method: transect
```

### positive-count

population の count は 0 以上であるべき。負の値は warning となる。

違反例:

```yaml
- id: pop-negative
  kind: population
  name: 負の個体数
  attributes:
    owner: species-japanese-serow
  spec:
    count: -3
    # warning: population entity "pop-negative" has negative count -3
```

count が非数値の場合も warning となる。count 自体の欠如は `valid-property`（required プロパティ）により報告される。

### valid-niigata-coordinates

area の latitude / longitude は新潟県（佐渡島を含む）の概略範囲内であるべき。

範囲:

| 項目 | 下限 | 上限 |
|------|------|------|
| latitude | 36.6 | 38.7 |
| longitude | 137.9 | 139.9 |

違反例:

```yaml
- id: area-out-of-bounds
  kind: area
  name: 範囲外の地点
  spec:
    latitude: 35.6812    # 東京。warning: outside Niigata bounds (36.6-38.7)
    longitude: 139.7671
```

latitude / longitude が両方とも未指定の場合は何も報告しない。片方のみ指定された場合は指定された側のみ検査する。非数値の場合も warning となる。

---

## カスタムルール

実装はカスタムルールを定義してよい（MAY）。

カスタムルールはルール構造に従わなければならない（MUST）。

カスタムルールはコアルールの意味論を再定義してはならない（MUST NOT）。

### カスタムルールの例

```yaml
id: custom-onsen-requires-source
name: Onsen Requires Source
description: "温泉は源泉への依存を持つべき"
severity: warning
scope: entity
condition: "entity.kind == 'hot_spring' implies exists relation type == 'depends_on' where source == entity.id"
```

---

## Validation Profiles

Validation Rule は Profile にグループ化できる。

Profile はユーザー定義である。

コア Schema は特定の profile を定義しない。

profile に rules が指定された場合、リストに含まれないルールは評価から除外される。

### Profile 構造

| Field | Type | Required | 説明 |
|-------|------|----------|------|
| name | string | yes | Profile 名 |
| description | string | no | Profile の説明 |
| rules | list[string] | no | 含める Rule ID |
| required_kinds | list[string] | no | 存在が必須の Entity Kind |
| required_relations | list[string] | no | 存在が必須の Relation Type |

`required_kinds` / `required_relations` が満たされない場合、それぞれ `profile-required-kind` / `profile-required-relation`（error）が報告される。

### Profile の例

```yaml
- name: nature-survey-profile
  description: 自然環境調査データ用プロファイル
  rules:
    - unique-id
    - valid-reference
    - valid-owner
    - single-owner
    - ownership-tree
    - required-kind
    - required-name
    - required-type
    - required-participants
    - population-requires-species
    - positive-count
  required_kinds:
    - area
    - species
  required_relations:
    - inhabits
```

---

## 検証実行

### 入力

- 検証対象の Graph
- Schema（任意）
- Profile（任意）

### 処理

1. Graph を読み込む
2. Schema を読み込む（指定された場合）
3. Profile を読み込む（指定された場合）
4. 適用されるルールを選択する
5. Graph に対してルールを実行する
6. Finding を収集する
7. Result を返す

### 出力

| Field | Type | 説明 |
|-------|------|------|
| findings | list[Finding] | 検証結果のリスト |
| passed | boolean | エラーが 0 件の場合 true |
| summary | map | サマリー統計 |

### Summary 構造

| Field | Type | 説明 |
|-------|------|------|
| total_rules | integer | 評価されたルール数 |
| total_findings | integer | Finding 合計数 |
| errors | integer | エラー数 |
| warnings | integer | 警告数 |
| infos | integer | 情報数 |

---

## 決定性

検証は決定的である。

同じ Graph に対する同じ検証は、常に同一の findings を生成しなければならない（MUST）。

---

## 拡張性

実装はカスタム Validation Rule を導入してよい（MAY）。

カスタム Rule はコア Schema の意味論を再定義してはならない（MUST NOT）。
