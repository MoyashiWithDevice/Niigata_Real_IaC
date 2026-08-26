# Relation Types

[← README](README.md)

---

## Relation Types一覧

| Type | Direction | Cardinality | Description |
|------|-----------|-------------|-------------|
| connects | symmetric | N:N | Physical/logical connection |
| hosts | directed | 1:N | Execution hosting |
| depends_on | directed | N:N | Dependency |
| belongs_to | directed | N:N | Logical membership |
| replicates_to | directed | 1:N | Data replication |
| backs_up | directed | 1:N | Backup relationship |
| monitors | directed | 1:N | Monitoring |
| managed_by | directed | N:1 | Management |
| mounted_on | directed | N:1 | Storage mounting |
| applies_to | directed | N:N | ACL application |
| listens_on | directed | N:1 | Port listening |

> **AWS拡張:** AWSのRelation Type（`aws.subscribes`, `aws.grants` 等）とコアタイプへのAWS参加者追加は [AWS Relation Types](aws-relation-types.md) を参照してください。

---

## Core Relation Types

### connects

Represents a physical or logical connection between Entities.

| Property | Value |
|----------|-------|
| Direction | symmetric |
| Participants | Two or more Entities |
| Cardinality | N:N |

**Additional Properties:**

| Property | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| connection_type | string | no | - | Type of connection (physical, logical, virtual) |
| bandwidth_mbps | integer | no | - | Connection bandwidth in Mbps |

```yaml
- id: rel-connects-srv-sw
  type: connects
  participants:
    - srv-proxmox-01/eno1
    - sw-core-01/port1
  attributes:
    status: active
  spec:
    connection_type: physical
    bandwidth_mbps: 10000
```

---

### hosts

Represents an execution or hosting relationship.

| Property | Value |
|----------|-------|
| Direction | directed |
| Source | Hosting Entity |
| Target | Hosted Entity |
| Cardinality | 1:N |

```yaml
- id: rel-hosts-server-vm
  type: hosts
  participants:
    source: srv-proxmox-01
    target: vm-web-01
  attributes:
    status: active

- id: rel-hosts-vm-app
  type: hosts
  participants:
    source: vm-web-01
    target: app-web-server
  attributes:
    status: active
```

---

### depends_on

Represents a directional dependency between Entities.

| Property | Value |
|----------|-------|
| Direction | directed |
| Source | Dependent Entity |
| Target | Dependency Entity |
| Cardinality | N:N |

**Additional Properties:**

| Property | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| dependency_type | string | no | - | Type of dependency (runtime, build, network, storage) |
| critical | boolean | no | false | Whether failure causes cascading failure |

```yaml
- id: rel-depends-app-db
  type: depends_on
  participants:
    source: app-web-server
    target: app-database
  attributes:
    status: active
  spec:
    dependency_type: runtime
    critical: true
```

---

### belongs_to

Represents logical membership or association.

| Property | Value |
|----------|-------|
| Direction | directed |
| Source | Member Entity |
| Target | Group Entity |
| Cardinality | N:N |

```yaml
- id: rel-belongsto-vm-cluster
  type: belongs_to
  participants:
    source: vm-web-01
    target: cluster-prod-01
  attributes:
    status: active

- id: rel-belongsto-intf-network
  type: belongs_to
  participants:
    source: vm-web-01/eth0
    target: mgmt-network-01
  attributes:
    status: active
```

---

## Extended Relation Types

### replicates_to

Represents data replication between Entities.

| Property | Value |
|----------|-------|
| Direction | directed |
| Source | Primary Entity |
| Target | Replica Entity |
| Cardinality | 1:N |

**Additional Properties:**

| Property | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| replication_type | string | no | synchronous | Replication type (synchronous, asynchronous) |
| lag_seconds | number | no | - | Replication lag in seconds |

```yaml
- id: rel-replicates-db
  type: replicates_to
  participants:
    source: app-database-primary
    target: app-database-replica
  spec:
    replication_type: asynchronous
    lag_seconds: 0.5
```

---

### backs_up

Represents a backup relationship.

| Property | Value |
|----------|-------|
| Direction | directed |
| Source | Source Entity |
| Target | Backup Entity |
| Cardinality | 1:N |

**Additional Properties:**

| Property | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| backup_type | string | no | - | Backup type (full, incremental, differential) |
| schedule | string | no | - | Backup schedule (cron expression) |
| retention_days | integer | no | - | Backup retention in days |

```yaml
- id: rel-backup-vm
  type: backs_up
  participants:
    source: vm-web-01
    target: vol-web-backup
  spec:
    backup_type: incremental
    schedule: "0 2 * * *"
    retention_days: 30
```

---

### monitors

Represents a monitoring relationship.

| Property | Value |
|----------|-------|
| Direction | directed |
| Source | Monitoring Entity |
| Target | Monitored Entity |
| Cardinality | 1:N |

**Additional Properties:**

| Property | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| monitor_type | string | no | - | Monitor type (agent, agentless, snmp, api) |
| interval_seconds | integer | no | 60 | Monitoring interval in seconds |

```yaml
- id: rel-monitor-prometheus
  type: monitors
  participants:
    source: app-prometheus
    target: srv-proxmox-01
  spec:
    monitor_type: snmp
    interval_seconds: 30
```

---

### managed_by

Represents a management relationship.

| Property | Value |
|----------|-------|
| Direction | directed |
| Source | Managed Entity |
| Target | Management Entity |
| Cardinality | N:1 |

**Additional Properties:**

| Property | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| management_type | string | no | - | Management type (configuration, orchestration, monitoring) |

```yaml
- id: rel-managed-vm
  type: managed_by
  participants:
    source: vm-web-01
    target: app-ansible
  spec:
    management_type: configuration
```

---

### mounted_on

Represents a mounting relationship (storage).

| Property | Value |
|----------|-------|
| Direction | directed |
| Source | Volume Entity |
| Target | Compute Entity |
| Cardinality | N:1 |

**Additional Properties:**

| Property | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| mount_point | string | no | - | Mount point path |
| filesystem | string | no | - | Filesystem type |
| options | string | no | - | Mount options |

```yaml
- id: rel-mounted-data
  type: mounted_on
  participants:
    source: vol-web-data
    target: vm-web-01
  spec:
    mount_point: /data
    filesystem: ext4
    options: "rw,noatime"
```

---

### applies_to

Represents that an ACL is applied to a network target.

| Property | Value |
|----------|-------|
| Direction | directed |
| Source | ACL Entity |
| Target | Network Target Entity |
| Cardinality | N:N |

```yaml
- id: rel-applies-web-acl
  type: applies_to
  participants:
    source: acl-web-ingress
    target: vm-web-01/eth0
  attributes:
    status: active
```

---

### listens_on

Represents that an open port is listening on a network interface or address.

| Property | Value |
|----------|-------|
| Direction | directed |
| Source | Open Port Entity |
| Target | Interface or Host Entity |
| Cardinality | N:1 |

```yaml
- id: rel-listens-nginx
  type: listens_on
  participants:
    source: port-443-nginx
    target: vm-web-01/eth0
  attributes:
    status: active
```
