# AWS Extension

## Overview

The AWS Extension adds vendor-specific Entity Kinds and Relation Types for modeling AWS infrastructure.

The core specification remains vendor-neutral. All AWS-specific concepts are defined by this extension under the `aws` namespace.

The extension is a **built-in extension** of IACForge. It is registered in-process (see `src/extension/builtin/aws/`) and is available without building a `.so` plugin.

### Manifest

| Field | Value |
|-------|-------|
| ID | `iacforge.aws` |
| Name | AWS |
| Version | `1.0.0` |
| Namespace | `aws` |
| Extension Points | `entity_kinds`, `relation_types`, `root_kinds` |

### Root Kind

The extension grants **root authority** to `aws.organization`, relaxing the exactly-one-root ownership rule. This allows modeling multiple AWS accounts under a single root organization.

### ID / ARN Conventions

- Entity `id` uses the short name in kebab-case (e.g. `srv-web-01`).
- The full ARN is stored in the `labels.arn` label (e.g. `labels: { arn: arn:aws:ec2:us-east-1:123456789012:instance/i-0123456789abcdef0 }`).
- `@`-prefixed property references resolve against existing entity IDs, not ARNs.

---

## Ownership Tree

```
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

Nested entities receive the parent as their `owner`, and an auto-generated `belongs_to` relation (member → parent) is created.

> Note: `aws.s3_bucket`, `aws.elasticache`, and `aws.efs` are regional resources in AWS and may be nested directly under `aws.region`. The legacy `aws.availability_zone` (`s3_buckets`, `elasticache_clusters`) and `aws.account` (`efs_filesystems`) nest keys remain available for backward compatibility.

---

## Entity Kinds

### Organization / Account

#### aws.organization

The root of an AWS organization. Owned by no entity (root kind).

| Property | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| org_root | string | no | - | Root organizational unit ID |
| org_units | list | no | - | Organizational units within the organization |

**Typical Ownership:** root (no owner specified)

**Nestable Children:**

| Nest Key | Child Kind |
|----------|------------|
| accounts | aws.account |

**Typical Relations:** (none)

---

#### aws.account

An AWS account within an organization.

| Property | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| account_id | string | yes | - | AWS account ID (12 digits) |
| alias | string | no | - | Account alias used in the sign-in URL |
| email | string | no | - | Account root email address |
| org_path | string | no | - | Organizational unit path within the organization |

**Typical Ownership:** aws.organization

**Nestable Children:**

| Nest Key | Child Kind |
|----------|------------|
| regions | aws.region |
| iam_users | aws.iam_user |
| iam_groups | aws.iam_group |
| iam_roles | aws.iam_role |
| iam_policies | aws.iam_policy |
| iam_instance_profiles | aws.iam_instance_profile |
| lambda_functions | aws.lambda_function |
| sqs_queues | aws.sqs_queue |
| sns_topics | aws.sns_topic |
| dynamodb_tables | aws.dynamodb_table |
| api_gateways | aws.api_gateway |
| route53_zones | aws.route53_zone |
| cloudfront_distributions | aws.cloudfront_distribution |
| eventbridge_rules | aws.eventbridge_rule |
| cloudwatch_alarms | aws.cloudwatch_alarm |
| cloudwatch_log_groups | aws.cloudwatch_log_group |
| cloudwatch_dashboards | aws.cloudwatch_dashboard |
| amis | aws.ami |
| key_pairs | aws.key_pair |
| launch_templates | aws.launch_template |
| auto_scaling_groups | aws.auto_scaling_group |
| efs_filesystems | aws.efs |
| ebs_snapshots | aws.ebs_snapshot |
| elastic_ips | aws.elastic_ip |
| target_groups | aws.target_group |
| vpc_peering_connections | aws.vpc_peering_connection |
| transit_gateways | aws.transit_gateway |

**Typical Relations:** (none)

---

### IAM

#### aws.iam_user

An IAM user.

| Property | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| arn | string | no | - | Amazon Resource Name of the user |
| path | string | no | - | Path to the user in the IAM hierarchy |

**Typical Ownership:** aws.account

**Typical Relations:** `aws.grants` ← aws.iam_policy, `aws.assumes` → aws.iam_role

---

#### aws.iam_group

An IAM group.

| Property | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| arn | string | no | - | Amazon Resource Name of the group |
| path | string | no | - | Path to the group in the IAM hierarchy |

**Typical Ownership:** aws.account

**Typical Relations:** `aws.grants` ← aws.iam_policy

---

#### aws.iam_role

An IAM role.

| Property | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| arn | string | no | - | Amazon Resource Name of the role |
| path | string | no | - | Path to the role in the IAM hierarchy |
| assume_role_policy | map | no | - | Trust policy document granting role assumption |

**Typical Ownership:** aws.account

**Typical Relations:** `aws.grants` ← aws.iam_policy, `aws.assumes` → aws.iam_role, `aws.assumes` ← aws.iam_role / aws.iam_user

---

#### aws.iam_policy

An IAM policy document.

| Property | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| arn | string | no | - | Amazon Resource Name of the policy |
| path | string | no | - | Path to the policy in the IAM hierarchy |
| policy_document | map | yes | - | Policy statement document |

**Typical Ownership:** aws.account

**Typical Relations:** `aws.grants` → aws.iam_user, aws.iam_group, aws.iam_role

---

#### aws.iam_instance_profile

An IAM instance profile attached to compute resources.

| Property | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| arn | string | no | - | Amazon Resource Name of the instance profile |
| path | string | no | - | Path to the instance profile in the IAM hierarchy |
| roles | list | no | - | Roles contained in the instance profile |

**Typical Ownership:** aws.account

**Typical Relations:** (none)

---

### Network

#### aws.region

An AWS region.

| Property | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| region_code | string | yes | - | Region code (e.g. us-east-1) |
| partition | string | no | - | AWS partition (aws, aws-cn, aws-us-gov, aws-iso, aws-iso-b) |
| opt_in_status | string | no | - | Region opt-in status (opted-in, not-opted-in, opt-in-not-required) |

**Typical Ownership:** aws.account

**Nestable Children:**

| Nest Key | Child Kind |
|----------|------------|
| availability_zones | aws.availability_zone |
| vpcs | aws.vpc |
| s3_buckets | aws.s3_bucket |
| elasticache_clusters | aws.elasticache |
| efs_filesystems | aws.efs |

**Typical Relations:** (none)

---

#### aws.availability_zone

An availability zone within a region.

| Property | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| zone_name | string | yes | - | Availability zone name (e.g. us-east-1a) |
| zone_id | string | no | - | Availability zone ID (e.g. use1-az1) |
| group_name | string | no | - | Local zone or wavelength zone group name |

**Typical Ownership:** aws.region

**Nestable Children:**

| Nest Key | Child Kind |
|----------|------------|
| s3_buckets | aws.s3_bucket |
| elasticache_clusters | aws.elasticache |
| ebs_volumes | aws.ebs_volume |

**Typical Relations:** (none)

---

#### aws.vpc

An AWS Virtual Private Cloud.

| Property | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| cidr_block | string | yes | - | IPv4 CIDR block of the VPC |
| tenancy | string | no | - | Instance tenancy (default, dedicated, host) |
| dns_support | boolean | no | true | Whether DNS resolution is supported |
| flow_logs | boolean | no | false | Whether VPC flow logs are enabled |

**Typical Ownership:** aws.region

**Nestable Children:**

| Nest Key | Child Kind |
|----------|------------|
| subnets | aws.subnet |
| security_groups | aws.security_group |
| route_tables | aws.route_table |
| internet_gateways | aws.internet_gateway |
| network_acls | aws.network_acl |

**Typical Relations:** `aws.associates` ← aws.security_group, aws.route_table, aws.network_acl, aws.elastic_ip; `aws.attaches` ← aws.internet_gateway, aws.elastic_ip, aws.vpc_peering_connection

---

#### aws.subnet

An AWS subnet within a VPC.

| Property | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| cidr_block | string | yes | - | IPv4 CIDR block of the subnet |
| map_public_ip | boolean | no | false | Whether instances receive public IPs by default |
| default_for_az | boolean | no | false | Whether this is the default subnet for the availability zone |

**Typical Ownership:** aws.vpc

**Nestable Children:**

| Nest Key | Child Kind |
|----------|------------|
| ec2_instances | aws.ec2 |
| rds_instances | aws.rds |
| load_balancers | aws.load_balancer |
| nat_gateways | aws.nat_gateway |

**Typical Relations:** `belongs_to` → aws.vpc, `aws.associates` ← aws.route_table, aws.network_acl

---

#### aws.route_table

An AWS route table associated with a VPC.

| Property | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| (none) | - | - | - | Uses only common properties |

**Typical Ownership:** aws.vpc

**Nestable Children:**

| Nest Key | Child Kind |
|----------|------------|
| routes | aws.route |

**Typical Relations:** `aws.associates` → aws.vpc / aws.subnet, `aws.routes` → aws.route

---

#### aws.route

A route entry within a route table.

| Property | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| destination_cidr | string | no | - | Destination IPv4 CIDR block |
| gateway_id | reference | no | - | Reference to the internet gateway or virtual private gateway |
| nat_gateway_id | reference | no | - | Reference to the NAT gateway |
| transit_gateway_id | reference | no | - | Reference to the transit gateway |
| vpc_peering_connection_id | reference | no | - | Reference to the VPC peering connection |

**Typical Ownership:** aws.route_table

**Typical Relations:** `aws.routes` ← aws.route_table

---

#### aws.internet_gateway

An AWS internet gateway attached to a VPC.

| Property | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| (none) | - | - | - | Uses only common properties |

**Typical Ownership:** aws.vpc

**Typical Relations:** `aws.attaches` → aws.vpc

---

#### aws.nat_gateway

An AWS NAT gateway.

| Property | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| connectivity_type | string | no | - | Connectivity type (public, private) |

**Typical Ownership:** aws.subnet

**Typical Relations:** `aws.attaches` ← aws.elastic_ip

---

#### aws.security_group

An AWS security group.

| Property | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| description | string | no | - | Description of the security group |
| vpc_id | reference | no | - | Reference to the VPC this security group belongs to |

**Typical Ownership:** aws.vpc

**Nestable Children:**

| Nest Key | Child Kind |
|----------|------------|
| security_group_rules | aws.security_group_rule |

**Typical Relations:** `aws.associates` → aws.vpc, `belongs_to` → aws.vpc

---

#### aws.security_group_rule

A rule within a security group.

| Property | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| type | string | no | - | Rule direction (ingress, egress) |
| from_port | integer | no | - | Start of the port range (0-65535) |
| to_port | integer | no | - | End of the port range (0-65535) |
| protocol | string | no | - | IP protocol (tcp, udp, icmp, icmpv6, -1) |
| source_security_group | reference | no | - | Source security group reference |
| source_cidr | string | no | - | Source IPv4 CIDR block |
| destination_cidr | string | no | - | Destination IPv4 CIDR block |

**Typical Ownership:** aws.security_group

**Typical Relations:** (none)

---

#### aws.elastic_ip

An AWS Elastic IP address.

| Property | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| domain | string | no | - | Address domain (vpc, standard) |
| public_ip | string | no | - | Public IPv4 address |
| association | reference | no | - | Entity this Elastic IP is associated with |

**Typical Ownership:** aws.account

**Typical Relations:** `aws.associates` → aws.vpc, `aws.attaches` → aws.ec2, aws.nat_gateway

---

#### aws.network_acl

An AWS network access control list.

| Property | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| default_action | string | no | - | Default action for traffic (allow, deny) |

**Typical Ownership:** aws.vpc

**Typical Relations:** `aws.associates` → aws.vpc / aws.subnet

---

#### aws.vpc_peering_connection

An AWS VPC peering connection.

| Property | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| auto_accept | boolean | no | false | Whether the peering connection is accepted automatically |
| peer_vpc_id | string | no | - | Peer VPC ID |
| peer_region | string | no | - | Region of the peer VPC |

**Typical Ownership:** aws.account

**Typical Relations:** `aws.attaches` → aws.vpc / aws.transit_gateway

---

#### aws.transit_gateway

An AWS transit gateway.

| Property | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| amazon_side_asn | integer | no | - | Private ASN of the transit gateway |
| dns_support | string | no | - | DNS support (enable, disable) |
| vpn_ecmp_support | string | no | - | VPN ECMP support (enable, disable) |

**Typical Ownership:** aws.account

**Typical Relations:** `aws.attaches` ← aws.vpc_peering_connection

---

### Compute

#### aws.ec2

An AWS EC2 instance.

| Property | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| instance_type | string | yes | - | Instance type (e.g. t3.micro) |
| ami | reference | no | - | Reference to the AMI used to launch the instance |
| key_pair | reference | no | - | Reference to the key pair used for SSH access |
| subnet | reference | no | - | Reference to the subnet the instance runs in |
| security_groups | list | no | - | References to the security groups attached to the instance |
| user_data | string | no | - | User data script passed at launch |
| iam_instance_profile | reference | no | - | Reference to the IAM instance profile |

**Typical Ownership:** aws.subnet

**Nestable Children:**

| Nest Key | Child Kind |
|----------|------------|
| applications | application (core kind) |

**Typical Relations:** `belongs_to` → aws.subnet, `aws.attaches` ← aws.ebs_volume, `aws.launches` ← aws.auto_scaling_group / aws.launch_template, `aws.registers` ← aws.target_group, `monitors` ← aws.cloudwatch_alarm, `mounted_on` ← aws.ebs_volume / aws.efs

---

#### aws.ami

An AWS Amazon Machine Image.

| Property | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| image_id | string | no | - | AMI image ID (e.g. ami-12345678) |
| architecture | string | no | - | CPU architecture (x86_64, arm64) |
| virtualization_type | string | no | - | Virtualization type (hvm, paravirtual) |

**Typical Ownership:** aws.account

**Typical Relations:** (none)

---

#### aws.key_pair

An AWS key pair.

| Property | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| key_type | string | no | - | Key type (rsa, ed25519) |
| fingerprint | string | no | - | SHA-1 fingerprint of the public key |
| public_key | string | no | - | Public key material |

**Typical Ownership:** aws.account

**Typical Relations:** (none)

---

#### aws.launch_template

An AWS EC2 launch template.

| Property | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| image_id | string | no | - | AMI image ID |
| instance_type | string | no | - | Instance type |
| key_name | string | no | - | Key pair name |

**Typical Ownership:** aws.account

**Typical Relations:** `aws.launches` → aws.ec2

---

#### aws.auto_scaling_group

An AWS Auto Scaling group.

| Property | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| min_size | integer | no | - | Minimum number of instances (0-5000) |
| max_size | integer | no | - | Maximum number of instances (0-5000) |
| desired_capacity | integer | no | - | Desired number of instances (0-5000) |
| launch_template | reference | no | - | Reference to the launch template |
| vpc_zone_identifier | list | no | - | References to the subnets across which instances are distributed |

**Typical Ownership:** aws.account

**Typical Relations:** `aws.launches` → aws.ec2

---

#### aws.ebs_volume

An AWS Elastic Block Store volume.

| Property | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| volume_type | string | no | - | Volume type (gp2, gp3, io1, io2, st1, sc1, standard) |
| size_gb | integer | no | - | Volume size in GiB |
| iops | integer | no | - | Provisioned IOPS |
| encrypted | boolean | no | - | Whether the volume is encrypted |
| kms_key_id | string | no | - | KMS key ID used for encryption |

**Typical Ownership:** aws.availability_zone

**Typical Relations:** `aws.attaches` → aws.ec2, `mounted_on` → aws.ec2, `backs_up` → aws.ebs_snapshot

---

#### aws.ebs_snapshot

An AWS EBS snapshot.

| Property | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| volume_id | string | no | - | Source volume ID |
| size_gb | integer | no | - | Snapshot size in GiB |
| encrypted | boolean | no | - | Whether the snapshot is encrypted |
| state | string | no | - | Snapshot state (pending, completed, error) |

**Typical Ownership:** aws.account

**Typical Relations:** `backs_up` ← aws.ebs_volume

---

#### aws.lambda_function

An AWS Lambda function.

| Property | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| runtime | string | yes | - | Runtime identifier (e.g. python3.12, nodejs20.x) |
| handler | string | no | - | Function handler |
| role | reference | no | - | Reference to the IAM role |
| memory_size | integer | no | - | Memory allocated in MB (128-10240) |
| vpc_config | map | no | - | VPC configuration (subnet and security group IDs) |

**Typical Ownership:** aws.account

**Typical Relations:** `aws.triggers` ← aws.eventbridge_rule / aws.cloudwatch_alarm, `aws.subscribes` ← aws.sns_topic, `aws.invokes` ← aws.api_gateway

---

### Storage / DB

#### aws.s3_bucket

An AWS Simple Storage Service bucket.

| Property | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| versioning | boolean | no | false | Whether bucket versioning is enabled |
| encryption | boolean | no | false | Whether default server-side encryption is enabled |
| lifecycle_rules | list | no | - | Lifecycle rule configurations |

**Typical Ownership:** aws.availability_zone

**Typical Relations:** `depends_on` ← aws.ec2, `backs_up` → aws.efs, `monitors` ← aws.cloudwatch_alarm

---

#### aws.efs

An AWS Elastic File System.

| Property | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| encrypted | boolean | no | false | Whether the file system is encrypted |
| performance_mode | string | no | - | Performance mode (generalPurpose, maxIO) |
| throughput_mode | string | no | - | Throughput mode (bursting, provisioned, elastic) |

**Typical Ownership:** aws.account

**Typical Relations:** `mounted_on` → aws.ec2, `backs_up` → aws.ebs_snapshot, `backs_up` → aws.s3_bucket

---

#### aws.rds

An AWS Relational Database Service instance.

| Property | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| engine | string | no | - | Database engine (aurora, aurora-mysql, aurora-postgresql, mysql, postgres, mariadb, oracle-se2, sqlserver-ee) |
| engine_version | string | no | - | Engine version |
| instance_class | string | yes | - | Instance class (e.g. db.t3.micro) |
| multi_az | boolean | no | false | Whether Multi-AZ deployment is enabled |
| storage_gb | integer | no | - | Allocated storage in GiB |

**Typical Ownership:** aws.subnet

**Typical Relations:** `belongs_to` → aws.subnet, `depends_on` → aws.rds / aws.dynamodb_table / aws.s3_bucket, `backs_up` → aws.ebs_snapshot, `monitors` ← aws.cloudwatch_alarm

---

#### aws.dynamodb_table

An AWS DynamoDB table.

| Property | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| billing_mode | string | no | - | Billing mode (PROVISIONED, ON_DEMAND) |
| table_class | string | no | - | Table class (STANDARD, STANDARD_INFREQUENT_ACCESS) |
| stream_enabled | boolean | no | false | Whether DynamoDB Streams is enabled |
| partition_key | string | no | - | Partition key name |
| sort_key | string | no | - | Sort key name |

**Typical Ownership:** aws.account

**Typical Relations:** `depends_on` ← aws.ec2 / aws.lambda_function, `monitors` ← aws.cloudwatch_alarm

---

#### aws.elasticache

An AWS ElastiCache cluster.

| Property | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| engine | string | no | - | Cache engine (redis, memcached) |
| node_type | string | no | - | Cache node type (e.g. cache.t3.micro) |
| num_cache_nodes | integer | no | - | Number of cache nodes |
| cluster_mode | boolean | no | false | Whether cluster mode is enabled |

**Typical Ownership:** aws.availability_zone

**Typical Relations:** `depends_on` ← aws.ec2, `monitors` ← aws.cloudwatch_alarm

---

### Load Balancer

#### aws.load_balancer

An AWS Elastic Load Balancer.

| Property | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| type | string | no | - | Load balancer type (application, network, gateway, classic) |
| scheme | string | no | - | Load balancer scheme (internet-facing, internal) |
| dns_name | string | no | - | DNS name of the load balancer |
| subnets | list | no | - | References to the subnets the load balancer spans |

**Typical Ownership:** aws.subnet

**Nestable Children:**

| Nest Key | Child Kind |
|----------|------------|
| listeners | aws.listener |

**Typical Relations:** `belongs_to` → aws.subnet, `aws.serves` → aws.target_group, `monitors` ← aws.cloudwatch_alarm

---

#### aws.target_group

An AWS target group for a load balancer.

| Property | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| protocol | string | no | - | Protocol (HTTP, HTTPS, TCP, TCP_UDP, UDP, TLS) |
| port | integer | no | - | Traffic port (1-65535) |
| target_type | string | no | - | Target type (instance, ip, lambda, alb) |
| health_check_path | string | no | - | Health check path |

**Typical Ownership:** aws.account

**Typical Relations:** `aws.serves` ← aws.load_balancer, `aws.forwards` ← aws.listener, `aws.registers` → aws.ec2, aws.lambda_function

---

#### aws.listener

An AWS load balancer listener.

| Property | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| protocol | string | no | - | Listener protocol (HTTP, HTTPS, TCP, TLS) |
| port | integer | no | - | Listener port (1-65535) |
| default_action | string | no | - | Default action (e.g. forward, redirect) |
| certificate_arn | string | no | - | SSL/TLS certificate ARN |

**Typical Ownership:** aws.load_balancer

**Typical Relations:** `aws.forwards` → aws.target_group

---

### Integration

#### aws.sqs_queue

An AWS Simple Queue Service queue.

| Property | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| fifo | boolean | no | false | Whether the queue is FIFO |
| visibility_timeout | integer | no | - | Visibility timeout in seconds |
| delay_seconds | integer | no | - | Message delay in seconds |
| message_retention_seconds | integer | no | - | Message retention period in seconds |

**Typical Ownership:** aws.account

**Typical Relations:** `depends_on` ← aws.lambda_function, `aws.subscribes` ← aws.sns_topic, `aws.triggers` ← aws.eventbridge_rule, `monitors` ← aws.cloudwatch_alarm

---

#### aws.sns_topic

An AWS Simple Notification Service topic.

| Property | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| fifo | boolean | no | false | Whether the topic is FIFO |
| display_name | string | no | - | Display name for email notifications |

**Typical Ownership:** aws.account

**Typical Relations:** `aws.subscribes` → aws.sqs_queue, aws.lambda_function, aws.sns_topic; `aws.triggers` ← aws.eventbridge_rule

---

#### aws.api_gateway

An AWS API Gateway.

| Property | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| endpoint_type | string | no | - | Endpoint type (REGIONAL, EDGE, PRIVATE) |
| protocol_type | string | no | - | API protocol (REST, HTTP, WEBSOCKET) |

**Typical Ownership:** aws.account

**Typical Relations:** `aws.invokes` → aws.lambda_function, `depends_on` → aws.lambda_function, `monitors` ← aws.cloudwatch_alarm

---

#### aws.cloudfront_distribution

An AWS CloudFront content delivery network distribution.

| Property | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| origin | string | no | - | Origin domain name |
| enabled | boolean | no | true | Whether the distribution is enabled |
| price_class | string | no | - | Price class (PriceClass_100, PriceClass_200, PriceClass_All) |
| domain_name | string | no | - | CloudFront domain name |

**Typical Ownership:** aws.account

**Typical Relations:** (none)

---

#### aws.eventbridge_rule

An AWS EventBridge rule.

| Property | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| event_pattern | string | no | - | Event pattern JSON |
| schedule_expression | string | no | - | Cron or rate schedule expression |
| state | string | no | - | Rule state (ENABLED, DISABLED) |

**Typical Ownership:** aws.account

**Typical Relations:** `aws.triggers` → aws.lambda_function, aws.sns_topic, aws.sqs_queue

---

### Monitoring

#### aws.cloudwatch_alarm

An AWS CloudWatch alarm.

| Property | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| metric_name | string | no | - | Metric name |
| namespace | string | no | - | Metric namespace |
| threshold | number | no | - | Alarm threshold value |
| comparison_operator | string | no | - | Comparison operator (GreaterThanOrEqualToThreshold, GreaterThanThreshold, LessThanThreshold, LessThanOrEqualToThreshold) |
| period_seconds | integer | no | - | Evaluation period in seconds |

**Typical Ownership:** aws.account

**Typical Relations:** `monitors` → aws.ec2, aws.rds, aws.load_balancer, aws.lambda_function, aws.dynamodb_table, aws.s3_bucket, aws.elasticache, aws.efs, aws.sqs_queue, aws.api_gateway; `aws.triggers` → aws.lambda_function, aws.sns_topic, aws.sqs_queue

---

#### aws.cloudwatch_log_group

An AWS CloudWatch log group.

| Property | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| retention_days | integer | no | - | Log retention period in days (1-3653) |

**Typical Ownership:** aws.account

**Typical Relations:** `depends_on` ← aws.ec2 / aws.lambda_function

---

#### aws.cloudwatch_dashboard

An AWS CloudWatch dashboard.

| Property | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| body | map | no | - | Dashboard body widget configuration |

**Typical Ownership:** aws.account

**Typical Relations:** (none)

---

### DNS

#### aws.route53_zone

An AWS Route 53 hosted zone.

| Property | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| private | boolean | no | false | Whether the hosted zone is private |
| comment | string | no | - | Zone comment |
| vpc_id | reference | no | - | Reference to the VPC a private zone is associated with |

**Typical Ownership:** aws.account

**Nestable Children:**

| Nest Key | Child Kind |
|----------|------------|
| records | aws.route53_record |

**Typical Relations:** (none)

---

#### aws.route53_record

An AWS Route 53 DNS record.

| Property | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| record_name | string | yes | - | Record name |
| type | string | no | - | Record type (A, AAAA, CNAME, MX, NS, PTR, SOA, SRV, TXT) |
| ttl | integer | no | - | Time to live in seconds |
| alias_target | string | no | - | Alias target resource |
| records | list | no | - | Resource records |

**Typical Ownership:** aws.route53_zone

**Typical Relations:** (none)

---

## Relation Types

### New Relation Types

All new relation types are directed binary relations (cardinality 1:N).

| Type | Direction | Source Kinds | Target Kinds | Description |
|------|-----------|--------------|--------------|-------------|
| aws.associates | directed | aws.security_group, aws.route_table, aws.network_acl, aws.elastic_ip | aws.vpc, aws.subnet | Association between a network resource and a VPC or subnet |
| aws.attaches | directed | aws.ebs_volume, aws.internet_gateway, aws.elastic_ip, aws.efs, aws.vpc_peering_connection | aws.ec2, aws.vpc, aws.transit_gateway, aws.nat_gateway | Attachment of a resource to a compute or network target |
| aws.launches | directed | aws.auto_scaling_group, aws.launch_template | aws.ec2 | Launch of compute capacity |
| aws.routes | directed | aws.route_table | aws.route | Ownership of a route entry by a route table |
| aws.serves | directed | aws.load_balancer | aws.target_group | Load balancer traffic distribution to a target group |
| aws.forwards | directed | aws.listener | aws.target_group | Forwarding of traffic from a listener to a target group |
| aws.registers | directed | aws.target_group | aws.ec2, aws.lambda_function | Registration of a compute target with a target group |
| aws.triggers | directed | aws.eventbridge_rule, aws.cloudwatch_alarm | aws.lambda_function, aws.sns_topic, aws.sqs_queue | Triggering of a target by a rule or alarm |
| aws.subscribes | directed | aws.sns_topic | aws.sqs_queue, aws.lambda_function, aws.sns_topic | Subscription of a queue, function, or topic to an SNS topic |
| aws.invokes | directed | aws.api_gateway | aws.lambda_function | Invocation of a Lambda function by an API Gateway |
| aws.grants | directed | aws.iam_policy | aws.iam_user, aws.iam_group, aws.iam_role | Granting of permissions from an IAM policy to an IAM principal |
| aws.assumes | directed | aws.iam_role, aws.iam_user | aws.iam_role | Role assumption by an IAM principal |

#### Properties

New relation types use only the common relation properties (id, type, participants, description, status, tags, labels, extensions).

#### Example

```yaml
- id: rel-assoc-sg-vpc
  type: aws.associates
  participants:
    source: sg-01
    target: vpc-01
  status: active

- id: rel-subscribe-sqs
  type: aws.subscribes
  participants:
    source: sns-01
    target: sqs-01
```

---

### Augmented Core Relation Types

The extension augments the following core relation types by adding AWS participant kinds. Only participant kinds are merged; direction and semantics are unchanged.

| Relation Type | Added Source Kinds | Added Target Kinds |
|---------------|--------------------|--------------------|
| belongs_to | All AWS kinds | All AWS kinds |
| depends_on | aws.ec2, aws.lambda_function, aws.rds, aws.load_balancer, aws.api_gateway, aws.auto_scaling_group, aws.sqs_queue | aws.ec2, aws.rds, aws.dynamodb_table, aws.s3_bucket, aws.sqs_queue, aws.lambda_function, aws.elasticache, aws.efs, aws.api_gateway, aws.cloudwatch_log_group |
| hosts | aws.ec2, aws.lambda_function | aws.ec2, aws.lambda_function |
| monitors | aws.cloudwatch_alarm | aws.ec2, aws.rds, aws.load_balancer, aws.lambda_function, aws.dynamodb_table, aws.s3_bucket, aws.elasticache, aws.efs, aws.sqs_queue, aws.api_gateway |
| backs_up | aws.rds, aws.ebs_volume, aws.dynamodb_table, aws.efs, aws.s3_bucket | aws.ebs_snapshot, aws.s3_bucket, aws.efs |
| mounted_on | aws.ebs_volume, aws.efs | aws.ec2 |

---

## Validation

- `valid-participant-kind` (Warning): all relation participants must be allowed by the relation type definition. Directed relations with explicit source/target participants are checked positionally (source against source kinds, target against target kinds); other shapes are checked against the union. AWS kinds added via `Augment` are validated against the merged constraints.
- `valid-property` (Warning): entity and relation `spec` properties are validated against the extension kind definitions (type, enum, required, min/max, unknown properties).
- Root kind rule: `aws.organization` is allowed as a root, so multiple accounts can exist under a single organization.

---

## Example Model

See `src/extension/builtin/aws/testdata/aws-example.yaml` for a complete round-trip example covering account → region → vpc → subnet → ec2 + s3 + rds + lambda + sqs.

```yaml
objects:
  - id: org-01
    kind: aws.organization
    name: Example Org

  - id: acct-01
    kind: aws.account
    name: Example Account
    attributes:
      owner: org-01
    spec:
      account_id: "123456789012"

  - id: region-01
    kind: aws.region
    name: us-east-1
    attributes:
      owner: acct-01
    spec:
      region_code: us-east-1

  - id: vpc-01
    kind: aws.vpc
    name: Main VPC
    attributes:
      owner: region-01
    spec:
      cidr_block: 10.0.0.0/16

  - id: subnet-01
    kind: aws.subnet
    name: Public Subnet
    attributes:
      owner: vpc-01
    spec:
      cidr_block: 10.0.1.0/24

  - id: ec2-01
    kind: aws.ec2
    name: Web Server
    attributes:
      owner: subnet-01
    spec:
      instance_type: t3.micro
      subnet: "@subnet-01"
```
