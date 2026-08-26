# AWS Relation Types

[← README](README.md)

---

AWS拡張（`iacforge.aws`）が定義するリレーションタイプのリファレンスです。

- 新規定義のリレーションはすべて `aws.` プレフィックスを持ち、directed（有向）です
- コアのリレーションタイプ（`belongs_to` 等）へのAWS Kind参加者追加は「Augmented Core Relations」を参照
- 正式な仕様は [spec/concrete/22-aws-extension.md](../../spec/concrete/22-aws-extension.md) を参照

## Relation Types一覧

| Type | Direction | Cardinality | Description |
|------|-----------|-------------|-------------|
| aws.associates | directed | 1:N | Network resource to VPC/subnet association |
| aws.attaches | directed | 1:N | Resource attachment |
| aws.launches | directed | 1:N | Auto Scaling/launch template launches EC2 |
| aws.routes | directed | 1:N | Route table owns route entries |
| aws.serves | directed | 1:N | Load balancer serves target group |
| aws.forwards | directed | 1:N | Listener forwards to target group |
| aws.registers | directed | 1:N | Target group registers compute target |
| aws.triggers | directed | 1:N | Rule/alarm triggers target |
| aws.subscribes | directed | 1:N | SNS subscription |
| aws.invokes | directed | 1:N | API Gateway invokes Lambda |
| aws.grants | directed | 1:N | IAM policy grants permissions |
| aws.assumes | directed | 1:N | IAM role assumption |

---

## AWS Relation Types

### aws.associates

ネットワークリソースとVPC/サブネットの関連付け。

| Property | Value |
|----------|-------|
| Direction | directed |
| Source | aws.security_group, aws.route_table, aws.network_acl, aws.elastic_ip |
| Target | aws.vpc, aws.subnet |
| Cardinality | 1:N |

```yaml
- id: rel-assoc-sg-vpc
  type: aws.associates
  participants:
    source: sg-web
    target: vpc-01
  attributes:
    status: active
```

---

### aws.attaches

リソースのアタッチメント。

| Property | Value |
|----------|-------|
| Direction | directed |
| Source | aws.ebs_volume, aws.internet_gateway, aws.elastic_ip, aws.efs, aws.vpc_peering_connection |
| Target | aws.ec2, aws.vpc, aws.transit_gateway, aws.nat_gateway |
| Cardinality | 1:N |

```yaml
- id: rel-attach-igw
  type: aws.attaches
  participants:
    source: igw-01
    target: vpc-01

- id: rel-attach-ebs
  type: aws.attaches
  participants:
    source: ebs-data-01
    target: ec2-web-01
```

---

### aws.launches

Auto Scalingグループ/起動テンプレートによるEC2起動。

| Property | Value |
|----------|-------|
| Direction | directed |
| Source | aws.auto_scaling_group, aws.launch_template |
| Target | aws.ec2 |
| Cardinality | 1:N |

```yaml
- id: rel-launch-asg
  type: aws.launches
  participants:
    source: asg-web
    target: ec2-web-01
```

---

### aws.routes

ルートテーブルによるルートエントリの所有。

| Property | Value |
|----------|-------|
| Direction | directed |
| Source | aws.route_table |
| Target | aws.route |
| Cardinality | 1:N |

```yaml
- id: rel-route-main
  type: aws.routes
  participants:
    source: rt-main
    target: route-public
```

---

### aws.serves

ロードバランサーからターゲットグループへのトラフィック配信。

| Property | Value |
|----------|-------|
| Direction | directed |
| Source | aws.load_balancer |
| Target | aws.target_group |
| Cardinality | 1:N |

```yaml
- id: rel-serve-alb
  type: aws.serves
  participants:
    source: alb-web
    target: tg-web
```

---

### aws.forwards

リスナーからターゲットグループへの転送。

| Property | Value |
|----------|-------|
| Direction | directed |
| Source | aws.listener |
| Target | aws.target_group |
| Cardinality | 1:N |

```yaml
- id: rel-forward-443
  type: aws.forwards
  participants:
    source: alb-web-443
    target: tg-web
```

---

### aws.registers

ターゲットグループによるコンピュートターゲットの登録。

| Property | Value |
|----------|-------|
| Direction | directed |
| Source | aws.target_group |
| Target | aws.ec2, aws.lambda_function |
| Cardinality | 1:N |

```yaml
- id: rel-register-ec2
  type: aws.registers
  participants:
    source: tg-web
    target: ec2-web-01

- id: rel-register-lambda
  type: aws.registers
  participants:
    source: tg-api
    target: lambda-processor
```

---

### aws.triggers

EventBridgeルール/CloudWatchアラームによるターゲット起動。

| Property | Value |
|----------|-------|
| Direction | directed |
| Source | aws.eventbridge_rule, aws.cloudwatch_alarm |
| Target | aws.lambda_function, aws.sns_topic, aws.sqs_queue |
| Cardinality | 1:N |

```yaml
- id: rel-trigger-lambda
  type: aws.triggers
  participants:
    source: eb-daily
    target: lambda-processor
```

---

### aws.subscribes

SNSトピックへのサブスクリプション。

| Property | Value |
|----------|-------|
| Direction | directed |
| Source | aws.sns_topic |
| Target | aws.sqs_queue, aws.lambda_function, aws.sns_topic |
| Cardinality | 1:N |

```yaml
- id: rel-subscribe-sqs
  type: aws.subscribes
  participants:
    source: sns-events
    target: sqs-jobs
```

---

### aws.invokes

API GatewayによるLambda関数の呼び出し。

| Property | Value |
|----------|-------|
| Direction | directed |
| Source | aws.api_gateway |
| Target | aws.lambda_function |
| Cardinality | 1:N |

```yaml
- id: rel-invoke-lambda
  type: aws.invokes
  participants:
    source: api-orders
    target: lambda-processor
```

---

### aws.grants

IAMポリシーからIAMプリンシパルへの権限付与。

| Property | Value |
|----------|-------|
| Direction | directed |
| Source | aws.iam_policy |
| Target | aws.iam_user, aws.iam_group, aws.iam_role |
| Cardinality | 1:N |

```yaml
- id: rel-grant-policy
  type: aws.grants
  participants:
    source: iam-policy-s3-read
    target: iam-role-lambda
```

---

### aws.assumes

IAMプリンシパルによるロール引き受け。

| Property | Value |
|----------|-------|
| Direction | directed |
| Source | aws.iam_role, aws.iam_user |
| Target | aws.iam_role |
| Cardinality | 1:N |

```yaml
- id: rel-assume-role
  type: aws.assumes
  participants:
    source: iam-user-deploy
    target: iam-role-lambda
```

---

## Augmented Core Relations

AWS拡張はコアのリレーションタイプの参加者Kindを追加します（方向・意味は不変）。

| Relation Type | Added Source Kinds | Added Target Kinds |
|---------------|--------------------|--------------------|
| belongs_to | All `aws.*` kinds | All `aws.*` kinds |
| depends_on | aws.ec2, aws.lambda_function, aws.rds, aws.load_balancer, aws.api_gateway, aws.auto_scaling_group, aws.sqs_queue | aws.ec2, aws.rds, aws.dynamodb_table, aws.s3_bucket, aws.sqs_queue, aws.lambda_function, aws.elasticache, aws.efs, aws.api_gateway, aws.cloudwatch_log_group |
| hosts | aws.ec2, aws.lambda_function | aws.ec2, aws.lambda_function |
| monitors | aws.cloudwatch_alarm | aws.ec2, aws.rds, aws.load_balancer, aws.lambda_function, aws.dynamodb_table, aws.s3_bucket, aws.elasticache, aws.efs, aws.sqs_queue, aws.api_gateway |
| backs_up | aws.rds, aws.ebs_volume, aws.dynamodb_table, aws.efs, aws.s3_bucket | aws.ebs_snapshot, aws.s3_bucket, aws.efs |
| mounted_on | aws.ebs_volume, aws.efs | aws.ec2 |

```yaml
# belongs_to (aws kind participants)
- id: rel-belongsto-ec2-subnet
  type: belongs_to
  participants:
    source: ec2-web-01
    target: subnet-public-01

# monitors
- id: rel-monitor-cpu
  type: monitors
  participants:
    source: cw-cpu-alarm
    target: ec2-web-01

# backs_up
- id: rel-backup-ebs
  type: backs_up
  participants:
    source: ebs-data-01
    target: snap-data-01
```
