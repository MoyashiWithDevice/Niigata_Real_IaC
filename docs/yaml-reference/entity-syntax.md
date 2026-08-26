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
- id: region-ap-northeast-1
  kind: region
  name: Tokyo Datacenter 1
```

## 全プロパティを指定したEntity

```yaml
- id: srv-proxmox-01
  kind: server
  name: Proxmox Node 01
  attributes:
    description: "Primary Proxmox server"
    status: active
    tags:
      - production
      - compute
    labels:
      region: ap-northeast-1
      environment: production
    extensions:
      vendor: dell
      model: r740xd
  spec:
    platform: proxmox
    cpu:
      - cores: 16
        architecture: x86_64
      - cores: 16
        architecture: x86_64
    memory:
      - size_gb: 64
        speed: 3200
        type: ddr4
      - size_gb: 64
        speed: 3200
        type: ddr4
    storage:
      - size_gb: 500
        type: ssd
      - size_gb: 500
        type: ssd
```

---

## ネスト定義（Nested Entity Definition）

子エンティティを親エンティティの定義内で直接定義できます。

### ネスト可能ないたずら

| 親 Kind | ネストキー | 子 Kind |
|---------|-----------|---------|
| any | interfaces | interface |
| any | servers | server |
| any | switches | switch |
| any | routers | router |
| any | firewalls | firewall |
| any | networks | network |
| region | racks | rack |
| region | clusters | cluster |
| region | availability_zones | availability_zone |
| cluster | vms | vm |
| cluster | servers | server |
| server | vms | vm |
| switch | ports | interface |
| router | ports | interface |
| firewall | acls | acl |
| vm | applications | application |
| application | open_ports | open_port |
| acl | acl_rules | acl_rule |
| interface | vlans | vlan |
| interface | cables | cable |

### 基本構文

```yaml
objects:
  - id: srv-proxmox-01
    kind: server
    name: Proxmox Node 01
    spec:
      cpu:
        - cores: 16
          architecture: x86_64
      networks:
        - id: net-private
          name: private
          spec:
            cidr: 172.31.0.0/24
          interfaces:
            - id: eth1
              spec:
                ip_address: 172.31.0.15
                type: ethernet
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
  source: eth1
  target: sw-core-01/port1
```

パス表記でも参照できます：

```yaml
participants:
  source: srv-proxmox-01/net-private/eth1
  target: sw-core-01/port1
```

### フラットとネストの混在

同じファイル内でフラット定義とネスト定義を混在できます：

```yaml
objects:
  # フラット定義
  - id: rack-a01
    kind: rack
    name: Rack A01
    attributes:
      owner: region-ap-northeast-1

  # ネスト定義
  - id: srv-proxmox-01
    kind: server
    name: Proxmox Node 01
    spec:
      networks:
        - id: net-private
          interfaces:
            - id: eth1
```
