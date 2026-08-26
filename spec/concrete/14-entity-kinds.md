# Entity Kinds

## Overview

Entity Kinds define the categories of objects that may exist in an infrastructure model.

Every Entity MUST define a kind.

The core specification defines the following Entity Kinds.

Implementations MAY introduce additional kinds through extensions.

---

## Common Properties

Every Entity shares the following common properties regardless of kind.

### Required

- id
- kind
- name

### Optional

- description
- status
- tags
- labels
- extensions

Individual Entity Kinds MAY define additional properties.

---

## Physical Infrastructure

### region

A geographic region where infrastructure is deployed.

#### Properties

| Property | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| address | string | no | - | Physical address |
| latitude | number | no | - | Geographic latitude |
| longitude | number | no | - | Geographic longitude |
| timezone | string | no | - | Timezone identifier |

#### Typical Ownership

- owned by: root (no owner specified)

#### Nestable Children

| Nest Key | Child Kind |
|----------|------------|
| racks | rack |
| clusters | cluster |
| availability_zones | availability_zone |

#### Typical Relations

- (none)

#### Example

```yaml
- id: region-ap-northeast-1
  kind: region
  name: Tokyo Datacenter 1
  status: active
  tags:
    - production
    - ap-northeast-1
  labels:
    region: asia-pacific
    tier: primary
```

---

### rack

A physical rack enclosure within a region.

#### Properties

| Property | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| height_units | integer | no | 42 | Rack height in rack units (U) |
| power_capacity_watts | integer | no | - | Total power capacity in watts |
| max_load_kg | number | no | - | Maximum weight capacity in kg |

#### Typical Ownership

- owned by: region

#### Nestable Children

| Nest Key | Child Kind |
|----------|------------|
| servers | server |
| switches | switch |
| routers | router |
| firewalls | firewall |

#### Typical Relations

- belongs_to → region

#### Example

```yaml
- id: rack-a01
  kind: rack
  name: Rack A01
  status: active
  labels:
    row: A
    zone: dc-1
```

---

### server

A physical or virtual compute host.

#### Properties

| Property | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| manufacturer | string | no | - | Hardware manufacturer |
| model | string | no | - | Hardware model |
| serial_number | string | no | - | Serial number |
| cpu | list[object] | no | - | CPU configurations |
| memory | list[object] | no | - | Memory modules |
| storage | list[object] | no | - | Local storage devices |
| platform | string | no | - | Virtualization platform (e.g., proxmox, vmware, kubernetes) |
| bios_version | string | no | - | BIOS/UEFI version |

##### cpu Properties

| Property | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| cores | integer | no | - | Number of CPU cores |
| architecture | string | no | - | CPU architecture (x86_64, arm64) |

##### memory Properties

| Property | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| size_gb | number | yes | - | Memory module size in GB |
| speed | integer | no | - | Memory speed in MHz |
| type | string | no | - | Memory type (ddr4, ddr5, lpddr4, lpddr5) |

##### storage Properties

| Property | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| size_gb | number | no | - | Storage size in GB |
| type | string | no | - | Storage type (ssd, hdd, nvme) |

#### Typical Ownership

- owned by: rack, region

#### Nestable Children

| Nest Key | Child Kind |
|----------|------------|
| networks | network |
| vms | vm |

#### Typical Relations

- belongs_to → rack
- belongs_to → region
- hosts → vm
- hosts → container

#### Example

```yaml
- id: srv-proxmox-01
  kind: server
  name: Proxmox Node 01
  status: active
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

**Note:** IP addresses are not direct properties of server. Use interface entities to assign IP addresses.

---

### cable

A physical cable connecting two or more interfaces.

#### Properties

| Property | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| cable_type | string | no | copper | Cable type (copper, fiber, dac) |
| length_meters | number | no | - | Cable length in meters |
| connector_a | string | no | - | Connector type at end A |
| connector_b | string | no | - | Connector type at end B |

#### Typical Relations

- connects → interface (symmetric)

#### Example

```yaml
- id: cable-001
  kind: cable
  name: Patch Cable A01-01 to Switch-01-Port24
  cable_type: cat6a
  length_meters: 3.0
```

---

### interface

A network interface (physical or virtual).

#### Properties

| Property | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| type | string | no | ethernet | Interface type (ethernet, fiber, wireless, virtual, bond, vlan, bridge, loopback) |
| mode | string | no | none | Interface mode (access, trunk, hybrid, none) |
| speed_mbps | integer | no | - | Interface speed in Mbps |
| mac_address | string | no | - | MAC address |
| ip_address | list[string] | no | - | IP addresses if configured |
| network | reference | no | - | Reference to the network this interface belongs to (e.g., `@mgmt-network`) |
| vlan_id | integer | no | - | VLAN identifier (1-4094) for a VLAN sub-interface |
| mtu | integer | no | 1500 | Maximum transmission unit |

> **Note:** An interface that carries IP addresses SHOULD reference a network, either via the `network` property or a `belongs_to` relation to a `network` entity. IP addresses are validated against the referenced network's `cidr`.

#### Typical Ownership

- owned by: server, switch, router, firewall, network, vm, container

#### Typical Relations

- connects → interface (symmetric, via cable)
- belongs_to → server, switch, router, firewall

#### Nestable Children

| Nest Key | Child Kind |
|----------|------------|
| vlans | vlan |
| cables | cable |
| interfaces | interface |

#### Examples

Physical interface:

```yaml
- id: eth0
  kind: interface
  name: Management Interface
  type: ethernet
  speed_mbps: 1000
  mac_address: "00:1a:2b:3c:4d:5e"
  network: "@mgmt-network"
  ip_address:
    - 10.0.1.10
```

VRRP virtual interface with physical members:

```yaml
- id: eth0-vrrp
  kind: interface
  name: VRRP Virtual Interface
  status: active
  spec:
    type: virtual
    ip_address:
      - 10.0.0.1
    interfaces:
      - id: eth0
        kind: interface
        name: eth0 - Primary
        status: active
        spec:
          type: ethernet
          ip_address:
            - 10.0.0.2
      - id: eth1
        kind: interface
        name: eth1 - Secondary
        status: standby
        spec:
          type: ethernet
          ip_address:
            - 10.0.0.3
```

LACP/teaming bond interface:

```yaml
- id: bond0
  kind: interface
  name: LAG Bundle
  spec:
    type: bond
    interfaces:
      - id: eth0
        kind: interface
        status: active
        spec:
          type: ethernet
          ip_address:
            - 192.168.1.1
      - id: eth1
        kind: interface
        status: standby
        spec:
          type: ethernet
          ip_address:
            - 192.168.1.2
```

Trunk port carrying multiple VLANs:

```yaml
- id: trunk-port1
  kind: interface
  name: Trunk Port to Access Switch
  spec:
    type: ethernet
    mode: trunk
    vlans:
      - id: trunk-port1-vlan10
        kind: vlan
        name: VLAN 10 - Management
        spec:
          vlan_id: 10
          tagged: false
          associated_network: "@mgmt-network"
      - id: trunk-port1-vlan100
        kind: vlan
        name: VLAN 100 - Production
        spec:
          vlan_id: 100
          tagged: true
          associated_network: "@prod-network"
      - id: trunk-port1-vlan200
        kind: vlan
        name: VLAN 200 - Storage
        spec:
          vlan_id: 200
          tagged: true
          associated_network: "@storage-network"
```

VLAN sub-interface (e.g., `vmbr.20`):

```yaml
- id: vmbr0
  kind: interface
  name: Linux Bridge vmbr0
  spec:
    type: bridge
    interfaces:
      - id: vmbr0.20
        kind: interface
        name: VLAN 20 on vmbr0
        spec:
          type: vlan
          vlan_id: 20
          ip_address:
            - 10.0.20.1/24
      - id: vmbr0.100
        kind: interface
        name: VLAN 100 on vmbr0
        spec:
          type: vlan
          vlan_id: 100
          ip_address:
            - 10.0.100.1/24
```

Loopback interface:

```yaml
- id: lo0
  kind: interface
  name: Loopback 0
  spec:
    type: loopback
    ip_address:
      - 10.255.255.1/32
```

---

### power_distribution

A power distribution unit (PDU) or power feed.

#### Properties

| Property | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| capacity_amps | integer | no | - | Total amperage capacity |
| voltage | number | no | 240 | Operating voltage |
| phases | integer | no | 1 | Number of phases |

#### Typical Relations

- belongs_to → rack
- connects → server
- connects → storage
- connects → switch

---

## Network

### network

A logical network or broadcast domain.

#### Properties

| Property | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| cidr | string | no | - | Network CIDR notation |
| gateway | string | no | - | Default gateway address |
| dns_servers | list[string] | no | - | DNS server addresses |
| vlan_id | integer | no | - | Associated VLAN ID |
| network_type | string | no | - | Network type (management, storage, vm, public) |

#### Typical Relations

- belongs_to → region
- belongs_to → cluster

#### Nestable Children

| Nest Key | Child Kind | Description |
|----------|------------|-------------|
| interfaces | interface | Network interfaces belonging to this network |

#### Interface Properties

When defining interfaces as nested children of a network, the following properties are available:

| Property | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| type | string | no | ethernet | Interface type (ethernet, fiber, wireless) |
| speed_mbps | integer | no | - | Interface speed in Mbps |
| mac_address | string | no | - | MAC address |
| ip_address | list[string] | no | - | IP addresses if configured |
| mtu | integer | no | 1500 | Maximum transmission unit |

#### Interface Nestable Children

| Nest Key | Child Kind |
|----------|------------|
| vlans | vlan |
| cables | cable |

#### Example

```yaml
- id: mgmt-network-01
  kind: network
  name: Management Network
  spec:
    cidr: 10.0.0.0/24
    gateway: 10.0.0.1
    network_type: management
    interfaces:
      - id: eth0
        spec:
          ip_address: 10.0.0.10
          type: ethernet
          speed_mbps: 10000
          mac_address: "aa:bb:cc:dd:ee:f0"
```

---

### vlan

A virtual LAN configuration.

#### Properties

| Property | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| vlan_id | integer | yes | - | VLAN identifier (1-4094) |
| tagged | boolean | no | false | Whether this VLAN carries tagged traffic on a trunk port |
| associated_network | string | no | - | Reference to parent network |

#### Typical Relations

- belongs_to → network
- belongs_to → region

---

### switch

A network switch.

#### Properties

| Property | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| manufacturer | string | no | - | Hardware manufacturer |
| model | string | no | - | Hardware model |
| serial_number | string | no | - | Serial number |
| port_count | integer | no | - | Total port count |
| managed | boolean | no | true | Whether switch is managed |
| stackable | boolean | no | false | Whether switch supports stacking |

#### Typical Ownership

- owned by: rack

#### Nestable Children

| Nest Key | Child Kind |
|----------|------------|
| ports | interface |
| interfaces | interface |

Each entry under `ports` is an `interface` entity representing a physical or logical port.

Switch ports are referenced using path notation: `sw-core-01/port1`.

#### Typical Relations

- belongs_to → rack
- connects → server (via cable)
- connects → switch (via cable)
- connects → router (via cable)

#### Example

```yaml
- id: sw-core-01
  kind: switch
  name: Core Switch 01
  spec:
    manufacturer: cisco
    model: Catalyst 9300
    port_count: 48
    managed: true
    ports:
      - id: port1
        name: Port 1 (Uplink)
        spec:
          type: ethernet
          speed_mbps: 10000
          mode: trunk
      - id: port2
        name: Port 2
        spec:
          type: ethernet
          speed_mbps: 1000
          mode: access
```

---

### router

A network router.

#### Properties

| Property | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| manufacturer | string | no | - | Hardware manufacturer |
| model | string | no | - | Hardware model |
| serial_number | string | no | - | Serial number |

#### Typical Ownership

- owned by: rack

#### Nestable Children

| Nest Key | Child Kind |
|----------|------------|
| ports | interface |
| interfaces | interface |

Each entry under `ports` is an `interface` entity representing a physical or logical port.

Router ports are referenced using path notation: `rt-core-01/ge0/0`.

#### Typical Relations

- belongs_to → rack
- connects → switch (via cable)
- connects → firewall (via cable)

#### Example

```yaml
- id: rt-core-01
  kind: router
  name: Core Router 01
  spec:
    manufacturer: mikrotik
    model: CCR1036
    ports:
      - id: ge0/0
        name: GigabitEthernet0/0
        spec:
          type: ethernet
          speed_mbps: 1000
```

---

### firewall

A network firewall.

#### Properties

| Property | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| manufacturer | string | no | - | Hardware manufacturer |
| model | string | no | - | Hardware model |
| serial_number | string | no | - | Serial number |
| throughput_gbps | number | no | - | Maximum throughput in Gbps |

#### Typical Ownership

- owned by: rack

#### Nestable Children

| Nest Key | Child Kind |
|----------|------------|
| interfaces | interface |
| acls | acl |

#### Typical Relations

- belongs_to → rack
- connects → router (via cable)

---

### acl

An Access Control List containing ordered rules for filtering network traffic.

An ACL is a container entity that holds `acl_rule` children in evaluation order.

Rules are evaluated top-to-bottom; the first matching rule determines the action.

#### Properties

| Property | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| default_action | string | no | deny | Default action when no rule matches (allow, deny) |
| direction | string | no | - | Traffic direction this ACL applies to (inbound, outbound, both) |
| protocol | string | no | any | Protocol filter (tcp, udp, icmp, any) |

#### Typical Ownership

- owned by: firewall, interface, server, vm, container

#### Nestable Children

| Nest Key | Child Kind |
|----------|------------|
| acl_rules | acl_rule |

#### Typical Relations

- belongs_to → firewall
- belongs_to → interface
- belongs_to → server
- belongs_to → vm
- belongs_to → container
- applies_to → interface (via applies_to)
- applies_to → firewall (via applies_to)

#### Example

```yaml
- id: acl-web-ingress
  kind: acl
  name: Web Server Ingress ACL
  status: active
  direction: inbound
  default_action: deny
  labels:
    environment: production
    tier: web
```

---

### acl_rule

A single rule within an Access Control List.

ACL rules are evaluated in order within their parent ACL.

The first matching rule determines whether traffic is allowed or denied.

#### Properties

| Property | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| action | string | yes | - | Rule action (allow, deny) |
| protocol | string | no | any | Protocol (tcp, udp, icmp, any) |
| source_address | string | no | any | Source IP address or CIDR |
| source_port | string | no | any | Source port or range (e.g., "80", "1024-65535") |
| destination_address | string | no | any | Destination IP address or CIDR |
| destination_port | string | no | any | Destination port or range (e.g., "443", "8080-8090") |
| enabled | boolean | no | true | Whether this rule is active |

#### Typical Ownership

- owned by: acl

#### Typical Relations

- (none)

#### Example

```yaml
- id: acl-rule-allow-https
  kind: acl_rule
  name: Allow HTTPS
  action: allow
  protocol: tcp
  source_address: 0.0.0.0/0
  destination_port: "443"
  enabled: true

- id: acl-rule-allow-ssh
  kind: acl_rule
  name: Allow SSH from Management
  action: allow
  protocol: tcp
  source_address: 10.0.0.0/24
  destination_port: "22"
  enabled: true
```

---

## Compute

### vm

A virtual machine.

#### Properties

| Property | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| cpu | list[object] | no | - | Virtual CPU configurations |
| memory | list[object] | no | - | Memory modules |
| storage | list[object] | no | - | Virtual disk configurations |
| os | string | no | - | Operating system |
| os_version | string | no | - | Operating system version |

##### cpu Properties

| Property | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| cores | integer | no | - | Number of virtual CPU cores |
| architecture | string | no | - | CPU architecture (x86_64, arm64) |

##### memory Properties

| Property | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| size_gb | number | yes | - | Memory module size in GB |
| speed | integer | no | - | Memory speed in MHz |
| type | string | no | - | Memory type (ddr4, ddr5, lpddr4, lpddr5) |

##### storage Properties

| Property | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| size_gb | number | no | - | Disk size in GB |
| type | string | no | - | Disk type (ssd, hdd, nvme) |

#### Typical Ownership

- owned by: server

#### Nestable Children

| Nest Key | Child Kind |
|----------|------------|
| networks | network |
| applications | application |

#### Typical Relations

- belongs_to → server
- belongs_to → cluster
- hosts → application

#### Example

```yaml
- id: vm-web-01
  kind: vm
  name: Web Server 01
  cpu:
    - cores: 4
      architecture: x86_64
  memory:
    - size_gb: 8
      speed: 3200
      type: ddr4
  storage:
    - size_gb: 100
      type: ssd
  os: ubuntu
  os_version: "22.04"
```

**Note:** IP addresses are not direct properties of vm. Use interface entities to assign IP addresses.

---

### container

A containerized workload.

#### Properties

| Property | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| image | string | no | - | Container image |
| image_tag | string | no | latest | Image tag |
| cpu_limit | string | no | - | CPU limit (e.g., "2.0") |
| memory_limit | string | no | - | Memory limit (e.g., "512Mi") |
| ports | list[integer] | no | - | Exposed ports |

#### Typical Relations

- belongs_to → vm
- belongs_to → server (direct)
- hosts → application

---

### application

A software application or service.

#### Properties

| Property | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| version | string | no | - | Application version |
| port | integer | no | - | Primary listening port |
| protocol | string | no | - | Network protocol (http, https, tcp, udp) |
| url | string | no | - | Application URL if applicable |

#### Typical Relations

- belongs_to → vm
- belongs_to → container
- depends_on → vm
- depends_on → application

#### Nestable Children

| Nest Key | Child Kind |
|----------|------------|
| open_ports | open_port |

#### Example

```yaml
- id: app-web-server
  kind: application
  name: Nginx Web Server
  version: "1.24.0"
  port: 443
  protocol: https
```

---

### open_port

A listening or open network port on a host, VM, container, or application.

Represents a discovered or declared port that is accepting connections.

#### Properties

| Property | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| port | integer | yes | - | Port number (1-65535) |
| protocol | string | yes | - | Transport protocol (tcp, udp) |
| state | string | no | listening | Port state (listening, established, closed) |
| address | string | no | 0.0.0.0 | Listening IP address |
| process | string | no | - | Process or service name using this port |
| pid | integer | no | - | Process ID if known |

#### Typical Relations

- belongs_to → server
- belongs_to → vm
- belongs_to → container
- belongs_to → application
- listens_on → interface (via listens_on)

#### Example

```yaml
- id: port-443-nginx
  kind: open_port
  name: Nginx HTTPS
  port: 443
  protocol: tcp
  state: listening
  address: 0.0.0.0
  process: nginx

- id: port-5432-postgres
  kind: open_port
  name: PostgreSQL
  port: 5432
  protocol: tcp
  state: listening
  address: 10.0.2.10
  process: postgres
```

---

## Storage

### storage

A storage system or array.

#### Properties

| Property | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| manufacturer | string | no | - | Hardware manufacturer |
| model | string | no | - | Hardware model |
| total_capacity_gb | number | no | - | Total raw capacity in GB |
| usable_capacity_gb | number | no | - | Usable capacity after redundancy |
| raid_level | string | no | - | RAID level if applicable |
| protocol | string | no | - | Storage protocol (nfs, iscsi, fc, local) |

#### Typical Ownership

- owned by: rack

#### Typical Relations

- belongs_to → rack
- hosts → vm (for boot storage)

---

### volume

A logical storage volume.

#### Properties

| Property | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| capacity_gb | number | no | - | Volume capacity in GB |
| filesystem | string | no | - | Filesystem type if mounted |
| mount_point | string | no | - | Mount point if applicable |
| thin_provisioned | boolean | no | false | Whether volume is thin provisioned |

#### Typical Relations

- belongs_to → storage
- belongs_to → server (local disks)
- hosts → vm

---

## Logical

### cluster

A logical grouping of compute resources.

#### Properties

| Property | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| cluster_type | string | no | - | Cluster type (compute, storage, hyperconverged) |
| ha_enabled | boolean | no | false | Whether HA is enabled |
| drs_enabled | boolean | no | false | Whether DRS is enabled |

#### Typical Ownership

- owned by: region

#### Nestable Children

| Nest Key | Child Kind | Description |
|----------|------------|-------------|
| vms | vm | VM nodes that compose the cluster |
| servers | server | Bare-metal nodes that compose the cluster |

Nested nodes receive the cluster as their `owner` and an auto-generated `belongs_to` relation (member → cluster) is created.

#### Typical Relations

- belongs_to → region
- belongs_to → network
- belongs_to ← server (nodes)
- belongs_to ← vm (nodes)

#### Example

```yaml
- id: cluster-prod-01
  kind: cluster
  name: Production Cluster 01
  cluster_type: hyperconverged
  ha_enabled: true
  drs_enabled: true
```

Kubernetes cluster with nested node machines:

```yaml
- id: k8s-prod
  kind: cluster
  name: Production Kubernetes Cluster
  attributes:
    owner: region-ap-northeast-1
  spec:
    cluster_type: compute
    ha_enabled: true
    vms:
      - id: vm-k8s-node-01
        name: K8s Node 01
        spec:
          cpu:
            - cores: 4
          memory:
            - size_gb: 16
      - id: vm-k8s-node-02
        name: K8s Node 02
        spec:
          cpu:
            - cores: 4
          memory:
            - size_gb: 16
    servers:
      - id: srv-k8s-node-01
        name: K8s Bare-metal Node 01
```

---

### availability_zone

A logical availability zone within a region.

#### Properties

| Property | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| state | string | no | available | Availability zone state (available, impaired, unavailable) |

#### Typical Ownership

- owned by: region

#### Typical Relations

- belongs_to → region

---

## Vendor Kinds

Vendor-specific Entity Kinds MUST NOT replace core kinds.

Vendors MAY introduce additional kinds through extensions.

### Extension Naming Convention

Extension kinds MUST use namespace prefixes.

Examples:

- `proxmox.vm` - Proxmox-specific VM extension
- `kubernetes.pod` - Kubernetes pod
- `aws.vpc` - AWS Virtual Private Cloud
- `network.switch` - Extended switch properties

### Built-in AWS Extension

The AWS extension (`iacforge.aws`) defines vendor kinds under the `aws` namespace (e.g. `aws.vpc`, `aws.ec2`, `aws.s3_bucket`).

The extension also defines new relation types (e.g. `aws.subscribes`, `aws.grants`) and augments core relation participant constraints.

See [AWS Extension](22-aws-extension.md) for the full kind definitions, ownership tree, and relation types.

---

## Status Values

Every Entity MAY have a status.

The core specification defines the following statuses:

| Status | Description |
|--------|-------------|
| planned | Entity is planned but not yet deployed |
| active | Entity is operational |
| maintenance | Entity is under maintenance |
| deprecated | Entity is scheduled for removal |
| offline | Entity is not operational |
| standby | Entity is in standby state (e.g. redundant member) |

Implementations MAY introduce additional statuses.

---

## Equality

Two Entities are considered different if their identifiers differ.

Changing properties does not create a new Entity.

Changing an identifier creates a different Entity.
