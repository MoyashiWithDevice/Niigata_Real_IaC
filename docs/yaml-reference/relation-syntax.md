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

`connects`のような対称関係の場合：

```yaml
participants:
  - srv-proxmox-01/eno1
  - sw-core-01/port1
```

### マップ形式（有向関係）

`hosts`、`depends_on`のような有向関係の場合：

```yaml
participants:
  source: srv-proxmox-01
  target: vm-web-01
```

## 全プロパティを指定したRelation

```yaml
- id: rel-hosts-server-vm
  type: hosts
  participants:
    source: srv-proxmox-01
    target: vm-web-01
  attributes:
    description: "Server hosts VM"
    status: active
    tags:
      - hosting
    labels:
      source_type: server
      target_type: vm
    extensions:
      custom_key: custom_value
  spec: {}
```
