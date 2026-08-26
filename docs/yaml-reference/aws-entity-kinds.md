# AWS Entity Kinds

[← README](README.md)

---

AWS拡張（`iacforge.aws`）が定義するエンティティKindのリファレンスです。

- すべてのKindは `aws.` プレフィックスを持ちます
- `aws.organization` はルート権限Kind（単一ルート所有権ツリーの例外）
- IDはケバブケースの短名、ARNは `labels.arn` に格納します
- 正式な仕様は [spec/concrete/22-aws-extension.md](../../spec/concrete/22-aws-extension.md) を参照

## Kind一覧

| Kind | Category | Description |
|------|----------|-------------|
| aws.organization | Organization | AWS organization |
| aws.account | Organization | AWS account |
| aws.iam_user | IAM | IAM user |
| aws.iam_group | IAM | IAM group |
| aws.iam_role | IAM | IAM role |
| aws.iam_policy | IAM | IAM policy document |
| aws.iam_instance_profile | IAM | IAM instance profile |
| aws.region | Network | AWS region |
| aws.availability_zone | Network | Availability zone |
| aws.vpc | Network | Virtual Private Cloud |
| aws.subnet | Network | Subnet |
| aws.route_table | Network | Route table |
| aws.route | Network | Route entry |
| aws.internet_gateway | Network | Internet gateway |
| aws.nat_gateway | Network | NAT gateway |
| aws.security_group | Network | Security group |
| aws.security_group_rule | Network | Security group rule |
| aws.elastic_ip | Network | Elastic IP |
| aws.network_acl | Network | Network ACL |
| aws.vpc_peering_connection | Network | VPC peering connection |
| aws.transit_gateway | Network | Transit gateway |
| aws.ec2 | Compute | EC2 instance |
| aws.ami | Compute | Amazon Machine Image |
| aws.key_pair | Compute | Key pair |
| aws.launch_template | Compute | Launch template |
| aws.auto_scaling_group | Compute | Auto Scaling group |
| aws.ebs_volume | Compute | EBS volume |
| aws.ebs_snapshot | Compute | EBS snapshot |
| aws.lambda_function | Compute | Lambda function |
| aws.s3_bucket | Storage | S3 bucket |
| aws.efs | Storage | Elastic File System |
| aws.rds | Database | RDS instance |
| aws.dynamodb_table | Database | DynamoDB table |
| aws.elasticache | Database | ElastiCache cluster |
| aws.load_balancer | Load Balancer | Elastic Load Balancer |
| aws.target_group | Load Balancer | Target group |
| aws.listener | Load Balancer | Load balancer listener |
| aws.sqs_queue | Integration | SQS queue |
| aws.sns_topic | Integration | SNS topic |
| aws.api_gateway | Integration | API Gateway |
| aws.cloudfront_distribution | Integration | CloudFront distribution |
| aws.eventbridge_rule | Integration | EventBridge rule |
| aws.cloudwatch_alarm | Monitoring | CloudWatch alarm |
| aws.cloudwatch_log_group | Monitoring | CloudWatch log group |
| aws.cloudwatch_dashboard | Monitoring | CloudWatch dashboard |
| aws.route53_zone | DNS | Route 53 hosted zone |
| aws.route53_record | DNS | Route 53 DNS record |

---

## Ownership Tree

```text
aws.organization
└── aws.account
    ├── aws.region
    │   ├── aws.availability_zone
    │   │   ├── aws.s3_bucket
    │   │   ├── aws.elasticache
    │   │   └── aws.ebs_volume
    │   ├── aws.s3_bucket
    │   ├── aws.elasticache
    │   ├── aws.efs
    │   └── aws.vpc
    │       ├── aws.subnet
    │       │   ├── aws.ec2
    │       │   │   └── application (core kind)
    │       │   ├── aws.rds
    │       │   ├── aws.load_balancer
    │       │   │   └── aws.listener
    │       │   └── aws.nat_gateway
    │       ├── aws.security_group
    │       │   └── aws.security_group_rule
    │       ├── aws.route_table
    │       │   └── aws.route
    │       ├── aws.internet_gateway
    │       └── aws.network_acl
    ├── aws.iam_user / aws.iam_group / aws.iam_role / aws.iam_policy / aws.iam_instance_profile
    ├── aws.lambda_function / aws.sqs_queue / aws.sns_topic / aws.dynamodb_table / aws.api_gateway
    ├── aws.route53_zone (→ aws.route53_record)
    ├── aws.cloudfront_distribution / aws.eventbridge_rule
    ├── aws.cloudwatch_alarm / aws.cloudwatch_log_group / aws.cloudwatch_dashboard
    ├── aws.ami / aws.key_pair / aws.launch_template / aws.auto_scaling_group
    ├── aws.ebs_snapshot / aws.elastic_ip / aws.target_group / aws.efs
    └── aws.vpc_peering_connection / aws.transit_gateway
```

> Note: `aws.s3_bucket`, `aws.elasticache`, `aws.efs` はリージョン直下にも配置できます（レガシーのAZ/account配下も引き続き利用可）。

---

## Organization / Account

### aws.organization

AWS organization。ルート権限Kind（ownerを指定しない）。

| Property | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| org_root | string | no | - | Root organizational unit ID |
| org_units | list | no | - | Organizational units |

**Ownership:** Root (no owner specified)

**Nestable Children:** `accounts` (aws.account)

```yaml
- id: org-01
  kind: aws.organization
  name: Example Org
  attributes:
    status: active
  spec:
    org_units:
      - "ou-1"
```

---

### aws.account

AWSアカウント。

| Property | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| account_id | string | yes | - | AWS account ID (12 digits) |
| alias | string | no | - | Account alias |
| email | string | no | - | Account root email address |
| org_path | string | no | - | Organizational unit path |

**Ownership:** aws.organization

**Nestable Children:** `regions`, `iam_users`, `iam_groups`, `iam_roles`, `iam_policies`, `iam_instance_profiles`, `lambda_functions`, `sqs_queues`, `sns_topics`, `dynamodb_tables`, `api_gateways`, `route53_zones`, `cloudfront_distributions`, `eventbridge_rules`, `cloudwatch_alarms`, `cloudwatch_log_groups`, `cloudwatch_dashboards`, `amis`, `key_pairs`, `launch_templates`, `auto_scaling_groups`, `efs_filesystems`, `ebs_snapshots`, `elastic_ips`, `target_groups`, `vpc_peering_connections`, `transit_gateways`

```yaml
- id: acct-01
  kind: aws.account
  name: Example Account
  attributes:
    owner: org-01
    status: active
    labels:
      arn: arn:aws:organizations::123456789012:account/o-1234/123456789012
  spec:
    account_id: "123456789012"
    alias: example-acct
```

---

## IAM

### aws.iam_user

IAMユーザー。

| Property | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| arn | string | no | - | Amazon Resource Name |
| path | string | no | - | Path in the IAM hierarchy |

**Ownership:** aws.account

```yaml
- id: iam-user-deploy
  kind: aws.iam_user
  name: Deploy User
  attributes:
    owner: acct-01
  spec:
    path: /service/
```

---

### aws.iam_group

IAMグループ。

| Property | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| arn | string | no | - | Amazon Resource Name |
| path | string | no | - | Path in the IAM hierarchy |

**Ownership:** aws.account

```yaml
- id: iam-group-admins
  kind: aws.iam_group
  name: Admins
  attributes:
    owner: acct-01
```

---

### aws.iam_role

IAMロール。

| Property | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| arn | string | no | - | Amazon Resource Name |
| path | string | no | - | Path in the IAM hierarchy |
| assume_role_policy | map | no | - | Trust policy document |

**Ownership:** aws.account

```yaml
- id: iam-role-lambda
  kind: aws.iam_role
  name: Lambda Execution Role
  attributes:
    owner: acct-01
  spec:
    assume_role_policy:
      Effect: Allow
      Principal:
        Service: lambda.amazonaws.com
```

---

### aws.iam_policy

IAMポリシードキュメント。

| Property | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| arn | string | no | - | Amazon Resource Name |
| path | string | no | - | Path in the IAM hierarchy |
| policy_document | map | yes | - | Policy statement document |

**Ownership:** aws.account

```yaml
- id: iam-policy-s3-read
  kind: aws.iam_policy
  name: S3 Read
  attributes:
    owner: acct-01
  spec:
    policy_document:
      Version: "2012-10-17"
      Statement:
        - Effect: Allow
          Action: s3:GetObject
          Resource: "*"
```

---

### aws.iam_instance_profile

IAMインスタンスプロファイル。

| Property | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| arn | string | no | - | Amazon Resource Name |
| path | string | no | - | Path in the IAM hierarchy |
| roles | list | no | - | Roles contained in the profile |

**Ownership:** aws.account

```yaml
- id: iam-profile-web
  kind: aws.iam_instance_profile
  name: Web Instance Profile
  attributes:
    owner: acct-01
  spec:
    roles:
      - iam-role-web
```

---

## Network

### aws.region

AWSリージョン。

| Property | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| region_code | string | yes | - | Region code (e.g. us-east-1) |
| partition | string | no | - | Partition (aws, aws-cn, aws-us-gov, aws-iso, aws-iso-b) |
| opt_in_status | string | no | - | Opt-in status (opted-in, not-opted-in, opt-in-not-required) |

**Ownership:** aws.account

**Nestable Children:** `availability_zones` (aws.availability_zone), `vpcs` (aws.vpc), `s3_buckets` (aws.s3_bucket), `elasticache_clusters` (aws.elasticache), `efs_filesystems` (aws.efs)

```yaml
- id: region-us-east-1
  kind: aws.region
  name: us-east-1
  attributes:
    owner: acct-01
  spec:
    region_code: us-east-1
    partition: aws
```

---

### aws.availability_zone

アベイラビリティゾーン。

| Property | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| zone_name | string | yes | - | AZ name (e.g. us-east-1a) |
| zone_id | string | no | - | AZ ID (e.g. use1-az1) |
| group_name | string | no | - | Local zone group name |

**Ownership:** aws.region

**Nestable Children:** `s3_buckets` (aws.s3_bucket), `elasticache_clusters` (aws.elasticache), `ebs_volumes` (aws.ebs_volume)

```yaml
- id: az-us-east-1a
  kind: aws.availability_zone
  name: us-east-1a
  attributes:
    owner: region-us-east-1
  spec:
    zone_name: us-east-1a
    zone_id: use1-az1
```

---

### aws.vpc

Virtual Private Cloud。

| Property | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| cidr_block | string | yes | - | IPv4 CIDR block |
| tenancy | string | no | - | Tenancy (default, dedicated, host) |
| dns_support | boolean | no | true | DNS resolution support |
| flow_logs | boolean | no | false | VPC flow logs |

**Ownership:** aws.region

**Nestable Children:** `subnets` (aws.subnet), `security_groups` (aws.security_group), `route_tables` (aws.route_table), `internet_gateways` (aws.internet_gateway), `network_acls` (aws.network_acl)

```yaml
- id: vpc-01
  kind: aws.vpc
  name: Main VPC
  attributes:
    owner: region-us-east-1
    status: active
  spec:
    cidr_block: 10.0.0.0/16
    dns_support: true
```

---

### aws.subnet

サブネット。

| Property | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| cidr_block | string | yes | - | IPv4 CIDR block |
| map_public_ip | boolean | no | false | Public IP by default |
| default_for_az | boolean | no | false | Default subnet for the AZ |

**Ownership:** aws.vpc

**Nestable Children:** `ec2_instances` (aws.ec2), `rds_instances` (aws.rds), `load_balancers` (aws.load_balancer), `nat_gateways` (aws.nat_gateway)

```yaml
- id: subnet-public-01
  kind: aws.subnet
  name: Public Subnet 01
  attributes:
    owner: vpc-01
  spec:
    cidr_block: 10.0.1.0/24
    map_public_ip: true
```

---

### aws.route_table

ルートテーブル。

| Property | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| (none) | - | - | - | Uses only common properties |

**Ownership:** aws.vpc

**Nestable Children:** `routes` (aws.route)

```yaml
- id: rt-main
  kind: aws.route_table
  name: Main Route Table
  attributes:
    owner: vpc-01
  spec:
    routes:
      - id: route-public
        spec:
          destination_cidr: 0.0.0.0/0
          gateway_id: "@igw-01"
```

---

### aws.route

ルートエントリ。

| Property | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| destination_cidr | string | no | - | Destination IPv4 CIDR |
| gateway_id | reference | no | - | Reference to the internet gateway or virtual private gateway |
| nat_gateway_id | reference | no | - | Reference to the NAT gateway |
| transit_gateway_id | reference | no | - | Reference to the transit gateway |
| vpc_peering_connection_id | reference | no | - | Reference to the VPC peering connection |

**Ownership:** aws.route_table

```yaml
- id: route-public
  kind: aws.route
  name: Default Route
  attributes:
    owner: rt-main
  spec:
    destination_cidr: 0.0.0.0/0
    gateway_id: "@igw-01"
```

---

### aws.internet_gateway

インターネットゲートウェイ。

| Property | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| (none) | - | - | - | Uses only common properties |

**Ownership:** aws.vpc

```yaml
- id: igw-01
  kind: aws.internet_gateway
  name: Internet GW
  attributes:
    owner: vpc-01
```

---

### aws.nat_gateway

NATゲートウェイ。

| Property | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| connectivity_type | string | no | - | Connectivity type (public, private) |

**Ownership:** aws.subnet

**Relations:** `aws.attaches` ← aws.elastic_ip

```yaml
- id: nat-01
  kind: aws.nat_gateway
  name: NAT GW 01
  attributes:
    owner: subnet-public-01
  spec:
    connectivity_type: public
```

---

### aws.security_group

セキュリティグループ。

| Property | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| description | string | no | - | Description |
| vpc_id | reference | no | - | Reference to the VPC |

**Ownership:** aws.vpc

**Nestable Children:** `security_group_rules` (aws.security_group_rule)

```yaml
- id: sg-web
  kind: aws.security_group
  name: Web SG
  attributes:
    owner: vpc-01
  spec:
    description: Web server security group
    vpc_id: "@vpc-01"
```

---

### aws.security_group_rule

セキュリティグループルール。

| Property | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| type | string | no | - | Direction (ingress, egress) |
| from_port | integer | no | - | Start port (0-65535) |
| to_port | integer | no | - | End port (0-65535) |
| protocol | string | no | - | Protocol (tcp, udp, icmp, icmpv6, -1) |
| source_security_group | reference | no | - | Source SG reference |
| source_cidr | string | no | - | Source IPv4 CIDR |
| destination_cidr | string | no | - | Destination IPv4 CIDR |

**Ownership:** aws.security_group

```yaml
- id: sg-web-https
  kind: aws.security_group_rule
  name: Allow HTTPS
  attributes:
    owner: sg-web
  spec:
    type: ingress
    protocol: tcp
    from_port: 443
    to_port: 443
    source_cidr: 0.0.0.0/0
```

---

### aws.elastic_ip

Elastic IPアドレス。

| Property | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| domain | string | no | - | Domain (vpc, standard) |
| public_ip | string | no | - | Public IPv4 address |
| association | reference | no | - | Associated entity reference |

**Ownership:** aws.account

```yaml
- id: eip-01
  kind: aws.elastic_ip
  name: Web EIP
  attributes:
    owner: acct-01
  spec:
    domain: vpc
```

---

### aws.network_acl

ネットワークACL。

| Property | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| default_action | string | no | - | Default action (allow, deny) |

**Ownership:** aws.vpc

```yaml
- id: nacl-public
  kind: aws.network_acl
  name: Public NACL
  attributes:
    owner: vpc-01
  spec:
    default_action: deny
```

---

### aws.vpc_peering_connection

VPCピアリング接続。

| Property | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| auto_accept | boolean | no | false | Auto accept |
| peer_vpc_id | string | no | - | Peer VPC ID |
| peer_region | string | no | - | Peer region |

**Ownership:** aws.account

```yaml
- id: peering-01
  kind: aws.vpc_peering_connection
  name: Peering VPC-01/VPC-02
  attributes:
    owner: acct-01
  spec:
    peer_vpc_id: vpc-02
```

---

### aws.transit_gateway

トランジットゲートウェイ。

| Property | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| amazon_side_asn | integer | no | - | Private ASN |
| dns_support | string | no | - | DNS support (enable, disable) |
| vpn_ecmp_support | string | no | - | VPN ECMP support (enable, disable) |

**Ownership:** aws.account

```yaml
- id: tgw-01
  kind: aws.transit_gateway
  name: Transit GW
  attributes:
    owner: acct-01
  spec:
    amazon_side_asn: 64512
```

---

## Compute

### aws.ec2

EC2インスタンス。

| Property | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| instance_type | string | yes | - | Instance type (e.g. t3.micro) |
| ami | reference | no | - | AMI reference |
| key_pair | reference | no | - | Key pair reference |
| subnet | reference | no | - | Subnet reference |
| security_groups | list | no | - | Security group references |
| user_data | string | no | - | User data script |
| iam_instance_profile | reference | no | - | Instance profile reference |

**Ownership:** aws.subnet

**Nestable Children:** `applications` (application)

```yaml
- id: ec2-web-01
  kind: aws.ec2
  name: Web Server 01
  attributes:
    owner: subnet-public-01
    status: active
  spec:
    instance_type: t3.micro
    ami: "@ami-ubuntu-2404"
    subnet: "@subnet-public-01"
    security_groups:
      - "@sg-web"
```

---

### aws.ami

Amazon Machine Image。

| Property | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| image_id | string | no | - | AMI ID |
| architecture | string | no | - | Architecture (x86_64, arm64) |
| virtualization_type | string | no | - | Virtualization (hvm, paravirtual) |

**Ownership:** aws.account

```yaml
- id: ami-ubuntu-2404
  kind: aws.ami
  name: Ubuntu 24.04
  attributes:
    owner: acct-01
  spec:
    image_id: ami-0123456789abcdef0
    architecture: x86_64
```

---

### aws.key_pair

キーペア。

| Property | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| key_type | string | no | - | Key type (rsa, ed25519) |
| fingerprint | string | no | - | SHA-1 fingerprint |
| public_key | string | no | - | Public key material |

**Ownership:** aws.account

```yaml
- id: kp-web
  kind: aws.key_pair
  name: Web Key Pair
  attributes:
    owner: acct-01
  spec:
    key_type: rsa
```

---

### aws.launch_template

起動テンプレート。

| Property | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| image_id | string | no | - | AMI image ID |
| instance_type | string | no | - | Instance type |
| key_name | string | no | - | Key pair name |

**Ownership:** aws.account

```yaml
- id: lt-web
  kind: aws.launch_template
  name: Web LT
  attributes:
    owner: acct-01
  spec:
    image_id: ami-0123456789abcdef0
    instance_type: t3.micro
```

---

### aws.auto_scaling_group

Auto Scalingグループ。

| Property | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| min_size | integer | no | - | Minimum instances (0-5000) |
| max_size | integer | no | - | Maximum instances (0-5000) |
| desired_capacity | integer | no | - | Desired instances (0-5000) |
| launch_template | reference | no | - | Launch template reference |
| vpc_zone_identifier | list | no | - | Subnet references |

**Ownership:** aws.account

```yaml
- id: asg-web
  kind: aws.auto_scaling_group
  name: Web ASG
  attributes:
    owner: acct-01
  spec:
    min_size: 1
    max_size: 4
    desired_capacity: 2
    launch_template: "@lt-web"
    vpc_zone_identifier:
      - "@subnet-public-01"
```

---

### aws.ebs_volume

EBSボリューム。

| Property | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| volume_type | string | no | - | Volume type (gp2, gp3, io1, io2, st1, sc1, standard) |
| size_gb | integer | no | - | Size in GiB |
| iops | integer | no | - | Provisioned IOPS |
| encrypted | boolean | no | - | Encrypted |
| kms_key_id | string | no | - | KMS key ID |

**Ownership:** aws.availability_zone

```yaml
- id: ebs-data-01
  kind: aws.ebs_volume
  name: Data Volume
  attributes:
    owner: az-us-east-1a
  spec:
    volume_type: gp3
    size_gb: 100
    encrypted: true
```

---

### aws.ebs_snapshot

EBSスナップショット。

| Property | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| volume_id | string | no | - | Source volume ID |
| size_gb | integer | no | - | Size in GiB |
| encrypted | boolean | no | - | Encrypted |
| state | string | no | - | State (pending, completed, error) |

**Ownership:** aws.account

```yaml
- id: snap-data-01
  kind: aws.ebs_snapshot
  name: Data Snapshot
  attributes:
    owner: acct-01
  spec:
    volume_id: ebs-data-01
    state: completed
```

---

### aws.lambda_function

Lambda関数。

| Property | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| runtime | string | yes | - | Runtime (e.g. python3.12) |
| handler | string | no | - | Function handler |
| role | reference | no | - | IAM role reference |
| memory_size | integer | no | - | Memory in MB (128-10240) |
| vpc_config | map | no | - | VPC configuration |

**Ownership:** aws.account

```yaml
- id: lambda-processor
  kind: aws.lambda_function
  name: Processor
  attributes:
    owner: acct-01
  spec:
    runtime: python3.12
    handler: index.handler
    role: "@iam-role-lambda"
    memory_size: 512
```

---

## Storage / Database

### aws.s3_bucket

S3バケット。

| Property | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| versioning | boolean | no | false | Bucket versioning |
| encryption | boolean | no | false | Server-side encryption |
| lifecycle_rules | list | no | - | Lifecycle rules |

**Ownership:** aws.availability_zone

```yaml
- id: s3-assets
  kind: aws.s3_bucket
  name: Assets Bucket
  attributes:
    owner: az-us-east-1a
    labels:
      arn: arn:aws:s3:::assets-bucket
  spec:
    versioning: true
    encryption: true
```

---

### aws.efs

Elastic File System。

| Property | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| encrypted | boolean | no | false | Encrypted |
| performance_mode | string | no | - | Performance mode (generalPurpose, maxIO) |
| throughput_mode | string | no | - | Throughput mode (bursting, provisioned, elastic) |

**Ownership:** aws.account

```yaml
- id: efs-data
  kind: aws.efs
  name: Data EFS
  attributes:
    owner: acct-01
  spec:
    performance_mode: generalPurpose
```

---

### aws.rds

RDSインスタンス。

| Property | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| engine | string | no | - | Engine (aurora, mysql, postgres, ...) |
| engine_version | string | no | - | Engine version |
| instance_class | string | yes | - | Instance class (e.g. db.t3.micro) |
| multi_az | boolean | no | false | Multi-AZ |
| storage_gb | integer | no | - | Storage in GiB |

**Ownership:** aws.subnet

```yaml
- id: rds-main
  kind: aws.rds
  name: Main Database
  attributes:
    owner: subnet-private-01
  spec:
    engine: postgres
    instance_class: db.t3.micro
    multi_az: true
```

---

### aws.dynamodb_table

DynamoDBテーブル。

| Property | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| billing_mode | string | no | - | Billing mode (PROVISIONED, ON_DEMAND) |
| table_class | string | no | - | Table class (STANDARD, STANDARD_INFREQUENT_ACCESS) |
| stream_enabled | boolean | no | false | DynamoDB Streams |
| partition_key | string | no | - | Partition key name |
| sort_key | string | no | - | Sort key name |

**Ownership:** aws.account

```yaml
- id: dynamo-sessions
  kind: aws.dynamodb_table
  name: Sessions
  attributes:
    owner: acct-01
  spec:
    billing_mode: ON_DEMAND
    partition_key: session_id
```

---

### aws.elasticache

ElastiCacheクラスター。

| Property | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| engine | string | no | - | Engine (redis, memcached) |
| node_type | string | no | - | Node type |
| num_cache_nodes | integer | no | - | Number of nodes |
| cluster_mode | boolean | no | false | Cluster mode |

**Ownership:** aws.availability_zone

```yaml
- id: redis-cache
  kind: aws.elasticache
  name: Redis Cache
  attributes:
    owner: az-us-east-1a
  spec:
    engine: redis
    node_type: cache.t3.micro
```

---

## Load Balancer

### aws.load_balancer

Elastic Load Balancer。

| Property | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| type | string | no | - | Type (application, network, gateway, classic) |
| scheme | string | no | - | Scheme (internet-facing, internal) |
| dns_name | string | no | - | DNS name |
| subnets | list | no | - | Subnet references |

**Ownership:** aws.subnet

**Nestable Children:** `listeners` (aws.listener)

```yaml
- id: alb-web
  kind: aws.load_balancer
  name: Web ALB
  attributes:
    owner: subnet-public-01
  spec:
    type: application
    scheme: internet-facing
    subnets:
      - "@subnet-public-01"
```

---

### aws.target_group

ターゲットグループ。

| Property | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| protocol | string | no | - | Protocol (HTTP, HTTPS, TCP, ...) |
| port | integer | no | - | Traffic port (1-65535) |
| target_type | string | no | - | Target type (instance, ip, lambda, alb) |
| health_check_path | string | no | - | Health check path |

**Ownership:** aws.account

```yaml
- id: tg-web
  kind: aws.target_group
  name: Web TG
  attributes:
    owner: acct-01
  spec:
    protocol: HTTP
    port: 80
    target_type: instance
```

---

### aws.listener

ロードバランサーリスナー。

| Property | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| protocol | string | no | - | Protocol (HTTP, HTTPS, TCP, TLS) |
| port | integer | no | - | Listener port (1-65535) |
| default_action | string | no | - | Default action (forward, redirect) |
| certificate_arn | string | no | - | SSL/TLS certificate ARN |

**Ownership:** aws.load_balancer

```yaml
- id: alb-web-443
  kind: aws.listener
  name: HTTPS Listener
  attributes:
    owner: alb-web
  spec:
    protocol: HTTPS
    port: 443
    default_action: forward
```

---

## Integration

### aws.sqs_queue

SQSキュー。

| Property | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| fifo | boolean | no | false | FIFO queue |
| visibility_timeout | integer | no | - | Visibility timeout (s) |
| delay_seconds | integer | no | - | Delay (s) |
| message_retention_seconds | integer | no | - | Message retention (s) |

**Ownership:** aws.account

```yaml
- id: sqs-jobs
  kind: aws.sqs_queue
  name: Jobs Queue
  attributes:
    owner: acct-01
  spec:
    visibility_timeout: 30
```

---

### aws.sns_topic

SNSトピック。

| Property | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| fifo | boolean | no | false | FIFO topic |
| display_name | string | no | - | Display name |

**Ownership:** aws.account

```yaml
- id: sns-events
  kind: aws.sns_topic
  name: Events Topic
  attributes:
    owner: acct-01
```

---

### aws.api_gateway

API Gateway。

| Property | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| endpoint_type | string | no | - | Endpoint type (REGIONAL, EDGE, PRIVATE) |
| protocol_type | string | no | - | Protocol (REST, HTTP, WEBSOCKET) |

**Ownership:** aws.account

```yaml
- id: api-orders
  kind: aws.api_gateway
  name: Orders API
  attributes:
    owner: acct-01
  spec:
    protocol_type: REST
```

---

### aws.cloudfront_distribution

CloudFrontディストリビューション。

| Property | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| origin | string | no | - | Origin domain name |
| enabled | boolean | no | true | Distribution enabled |
| price_class | string | no | - | Price class |
| domain_name | string | no | - | CloudFront domain name |

**Ownership:** aws.account

```yaml
- id: cf-web
  kind: aws.cloudfront_distribution
  name: Web CDN
  attributes:
    owner: acct-01
  spec:
    origin: d123.cloudfront.net
```

---

### aws.eventbridge_rule

EventBridgeルール。

| Property | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| event_pattern | string | no | - | Event pattern JSON |
| schedule_expression | string | no | - | Cron/rate expression |
| state | string | no | - | State (ENABLED, DISABLED) |

**Ownership:** aws.account

```yaml
- id: eb-daily
  kind: aws.eventbridge_rule
  name: Daily Job
  attributes:
    owner: acct-01
  spec:
    schedule_expression: "cron(0 2 * * ? *)"
```

---

## Monitoring

### aws.cloudwatch_alarm

CloudWatchアラーム。

| Property | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| metric_name | string | no | - | Metric name |
| namespace | string | no | - | Metric namespace |
| threshold | number | no | - | Threshold |
| comparison_operator | string | no | - | Comparison operator |
| period_seconds | integer | no | - | Evaluation period (s) |

**Ownership:** aws.account

```yaml
- id: cw-cpu-alarm
  kind: aws.cloudwatch_alarm
  name: CPU Alarm
  attributes:
    owner: acct-01
  spec:
    metric_name: CPUUtilization
    namespace: AWS/EC2
    threshold: 80
    period_seconds: 60
```

---

### aws.cloudwatch_log_group

CloudWatchロググループ。

| Property | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| retention_days | integer | no | - | Retention days (1-3653) |

**Ownership:** aws.account

```yaml
- id: cw-log-app
  kind: aws.cloudwatch_log_group
  name: /aws/app
  attributes:
    owner: acct-01
  spec:
    retention_days: 30
```

---

### aws.cloudwatch_dashboard

CloudWatchダッシュボード。

| Property | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| body | map | no | - | Widget configuration |

**Ownership:** aws.account

```yaml
- id: cw-dashboard-overview
  kind: aws.cloudwatch_dashboard
  name: Overview
  attributes:
    owner: acct-01
```

---

## DNS

### aws.route53_zone

Route 53ホストゾーン。

| Property | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| private | boolean | no | false | Private zone |
| comment | string | no | - | Zone comment |
| vpc_id | reference | no | - | VPC reference (private zone) |

**Ownership:** aws.account

**Nestable Children:** `records` (aws.route53_record)

```yaml
- id: zone-example
  kind: aws.route53_zone
  name: example.com
  attributes:
    owner: acct-01
  spec:
    private: false
```

---

### aws.route53_record

Route 53 DNSレコード。

| Property | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| record_name | string | yes | - | Record name |
| type | string | no | - | Record type (A, AAAA, CNAME, MX, NS, PTR, SOA, SRV, TXT) |
| ttl | integer | no | - | TTL (s) |
| alias_target | string | no | - | Alias target |
| records | list | no | - | Resource records |

**Ownership:** aws.route53_zone

```yaml
- id: record-www
  kind: aws.route53_record
  name: www.example.com
  attributes:
    owner: zone-example
  spec:
    type: A
    ttl: 300
    records:
      - 10.0.1.10
```
