# Relation Syntax

[← README](README.md)

---

## 必須プロパティ

| Property | Type | Description |
|----------|------|-------------|
| id | string | ユニーク識別子 |
| type | string | Relation type |
| participants | list[string] or map | Entity参照 |

## 任意の共通プロパティ（`attributes:` 以下）

| Property | Type | Default | Description |
|----------|------|---------|-------------|
| description | string | - | ドキュメント |
| status | enum | - | ライフサイクル状態 |
| tags | list[string] | - | ラベル |
| labels | map[string] | - | メタデータ |
| extensions | map[string] | - | 拡張データ |

## Participantフォーマット

### リスト形式（対称関係）

`near`のような対称関係の場合：

```yaml
participants:
  - spot-shukunegi
  - coast-otoline
```

### マップ形式（有向関係）

`located_in`、`depends_on`のような有向関係の場合：

```yaml
participants:
  source: species-toki
  target: forest-osado-beech
```

## 全プロパティを指定したRelation

```yaml
- id: rel-toki-inhabits
  type: inhabits
  participants:
    source: species-toki
    target: forest-osado-beech
  attributes:
    description: "Toki inhabit the beech forest"
    status: active
    tags:
      - habitat
    labels:
      source_type: species
      target_type: forest
    extensions:
      custom_key: custom_value
  spec:
    habitat_note: Roosts in tall trees
```
