# Complete Example

[← README](README.md)

---

```yaml
objects:
  # Regions
  - id: region-ap-northeast-1
    kind: region
    name: Tokyo Datacenter 1
    attributes:
      status: active
      labels:
        region: ap-northeast-1

  # Racks
  - id: rack-a01
    kind: rack
    name: Rack A01
    attributes:
      owner: region-ap-northeast-1
      status: active
      labels:
        row: A
    spec:
      height_units: 42

  # Servers
  - id: srv-proxmox-01
    kind: server
    name: Proxmox Node 01
    attributes:
      owner: rack-a01
      status: active
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

  - id: srv-proxmox-02
    kind: server
    name: Proxmox Node 02
    attributes:
      owner: rack-a01
      status: active
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

  # Switches
  - id: sw-core-01
    kind: switch
    name: Core Switch 01
    attributes:
      owner: rack-a01
      status: active
    spec:
      manufacturer: cisco
      model: Catalyst 9300
      port_count: 48
      ports:
        - id: port1
          name: port1
          spec:
            type: ethernet
            speed_mbps: 10000
            mode: trunk

  # Interfaces
  - id: eno1
    kind: interface
    name: eno1
    attributes:
      owner: srv-proxmox-01
    spec:
      type: ethernet
      speed_mbps: 10000
      mac_address: "aa:bb:cc:dd:ee:f0"
      network: "@mgmt-network-01"
      ip_address: 10.0.0.10

  - id: eno2
    kind: interface
    name: eno2
    attributes:
      owner: srv-proxmox-01
    spec:
      type: ethernet
      speed_mbps: 10000
      mac_address: "aa:bb:cc:dd:ee:f1"

  # Network
  - id: mgmt-network-01
    kind: network
    name: Management Network
    spec:
      cidr: 10.0.0.0/24
      gateway: 10.0.0.1
      network_type: management

  # VMs
  - id: vm-web-01
    kind: vm
    name: Web Server 01
    attributes:
      owner: srv-proxmox-01
      status: active
    spec:
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

  # Applications
  - id: app-web-server
    kind: application
    name: Nginx Web Server
    attributes:
      owner: vm-web-01
      status: active
    spec:
      version: "1.24.0"
      port: 443
      protocol: https

  # Open Ports
  - id: port-443-nginx
    kind: open_port
    name: Nginx HTTPS
    attributes:
      owner: app-web-server
    spec:
      port: 443
      protocol: tcp
      state: listening
      address: 0.0.0.0
      process: nginx

  - id: port-5432-postgres
    kind: open_port
    name: PostgreSQL
    attributes:
      owner: vm-web-01
    spec:
      port: 5432
      protocol: tcp
      state: listening
      address: 10.0.2.10
      process: postgres

  # ACLs
  - id: acl-web-ingress
    kind: acl
    name: Web Server Ingress ACL
    attributes:
      owner: vm-web-01
      status: active
    spec:
      direction: inbound
      default_action: deny

  # ACL Rules
  - id: acl-rule-allow-https
    kind: acl_rule
    name: Allow HTTPS
    attributes:
      owner: acl-web-ingress
    spec:
      action: allow
      protocol: tcp
      source_address: 0.0.0.0/0
      destination_port: "443"
      enabled: true

  - id: acl-rule-allow-ssh
    kind: acl_rule
    name: Allow SSH from Management
    attributes:
      owner: acl-web-ingress
    spec:
      action: allow
      protocol: tcp
      source_address: 10.0.0.0/24
      destination_port: "22"
      enabled: true

  # Cluster
  - id: cluster-prod-01
    kind: cluster
    name: Production Cluster 01
    attributes:
      status: active
    spec:
      cluster_type: hyperconverged
      ha_enabled: true

  # Cables
  - id: cable-001
    kind: cable
    name: Patch Cable SRV01-SW01
    spec:
      cable_type: cat6a
      length_meters: 3.0

  # Connection Relations (connects)
  - id: rel-connects-srv-sw
    type: connects
    spec:
      connection_type: physical
      bandwidth_mbps: 10000
    participants:
      - srv-proxmox-01/eno1
      - sw-core-01/port1

  # Hosting Relations (hosts)
  - id: rel-hosts-server-vm
    type: hosts
    participants:
      source: srv-proxmox-01
      target: vm-web-01

  - id: rel-hosts-vm-app
    type: hosts
    participants:
      source: vm-web-01
      target: app-web-server

  # Membership Relations (belongs_to)
  - id: rel-belongsto-vm-cluster
    type: belongs_to
    participants:
      source: vm-web-01
      target: cluster-prod-01

  - id: rel-belongsto-intf-network
    type: belongs_to
    participants:
      source: vm-web-01/eth0
      target: mgmt-network-01

  # ACL Application Relations (applies_to)
  - id: rel-applies-web-acl
    type: applies_to
    participants:
      source: acl-web-ingress
      target: vm-web-01/eth0

  # Open Port Relations (belongs_to)
  - id: rel-belongsto-port-nginx
    type: belongs_to
    participants:
      source: port-443-nginx
      target: app-web-server

  - id: rel-belongsto-port-postgres
    type: belongs_to
    participants:
      source: port-5432-postgres
      target: vm-web-01

  # Port Listening Relations (listens_on)
  - id: rel-listens-nginx
    type: listens_on
    participants:
      source: port-443-nginx
      target: vm-web-01/eth0

  - id: rel-listens-postgres
    type: listens_on
    participants:
      source: port-5432-postgres
      target: vm-web-01/eth0
```

---

## AWSモデル例

AWS拡張（`aws.*` Kind）を使用したモデル例です。Kindの全定義は [AWS Entity Kinds](aws-entity-kinds.md) を参照してください。

```yaml
objects:
  # Organization / Account
  - id: org-01
    kind: aws.organization
    name: Example Org
    attributes:
      status: active

  - id: acct-01
    kind: aws.account
    name: Example Account
    attributes:
      owner: org-01
      status: active
    spec:
      account_id: "123456789012"
      alias: example-acct

  # Region / AZ / VPC
  - id: region-us-east-1
    kind: aws.region
    name: us-east-1
    attributes:
      owner: acct-01
    spec:
      region_code: us-east-1

  - id: az-us-east-1a
    kind: aws.availability_zone
    name: us-east-1a
    attributes:
      owner: region-us-east-1
    spec:
      zone_name: us-east-1a

  - id: vpc-01
    kind: aws.vpc
    name: Main VPC
    attributes:
      owner: region-us-east-1
    spec:
      cidr_block: 10.0.0.0/16

  - id: subnet-01
    kind: aws.subnet
    name: Public Subnet
    attributes:
      owner: vpc-01
    spec:
      cidr_block: 10.0.1.0/24

  # Compute / Storage
  - id: ec2-web-01
    kind: aws.ec2
    name: Web Server
    attributes:
      owner: subnet-01
      status: active
    spec:
      instance_type: t3.micro
      subnet: "@subnet-01"

  - id: s3-assets
    kind: aws.s3_bucket
    name: Assets Bucket
    attributes:
      owner: az-us-east-1a
    spec:
      versioning: true

  - id: lambda-processor
    kind: aws.lambda_function
    name: Processor
    attributes:
      owner: acct-01
    spec:
      runtime: python3.12

  - id: sns-events
    kind: aws.sns_topic
    name: Events Topic
    attributes:
      owner: acct-01
      status: active

  - id: sqs-jobs
    kind: aws.sqs_queue
    name: Job Queue
    attributes:
      owner: acct-01
      status: active

  # Relations
  - id: rel-ec2-subnet
    type: belongs_to
    participants:
      source: ec2-web-01
      target: subnet-01

  - id: rel-subscribe-sqs
    type: aws.subscribes
    participants:
      source: sns-events
      target: sqs-jobs
```
