/*
 * IACForge Reference - search index
 * Client-side search data. Kept dependency-free and static.
 * Each entry: { path, title, section, description, keywords }
 */

window.IACFORGE_SEARCH_INDEX = [
  {
    path: "index.html",
    title: "Reference",
    section: "Overview",
    description: "IACForgeのYAMLファイル作成のための完全版リファレンスのホーム。",
    keywords: "home overview top index reference yaml ホーム 概要 はじめに",
  },
  {
    path: "document-structure.html",
    title: "Document Structure",
    section: "YAML Reference",
    description: "YAMLドキュメントの基本構造、コメント、記述順序。",
    keywords: "document structure yaml graph objects comment order ドキュメント 構造 コメント 順序",
  },
  {
    path: "entity-syntax.html",
    title: "Entity Syntax",
    section: "YAML Reference",
    description: "Entity共通プロパティ、必須プロパティ、ステータス値、ネスト定義。",
    keywords: "entity syntax id kind name attributes owner status planned active maintenance deprecated offline standby tags labels extensions nested nest エンティティ 構文 必須 ステータス ネスト",
  },
  {
    path: "entity-kinds.html",
    title: "Entity Kinds",
    section: "YAML Reference",
    description: "全Entity種類の定義とプロパティ。region, rack, server, vm, network, storageなど。",
    keywords: "entity kinds region rack server interface cable power_distribution network vlan switch router firewall acl acl_rule vm container application open_port storage volume cluster availability_zone mode trunk access hybrid tagged port_count ports ip_address cidr kind 種類 定義",
  },
  {
    path: "relation-syntax.html",
    title: "Relation Syntax",
    section: "YAML Reference",
    description: "Relation共通構文、participantフォーマット（リスト・マップ形式）。",
    keywords: "relation syntax id type participants source target symmetric directed attributes リレーション 構文 参加者",
  },
  {
    path: "relation-types.html",
    title: "Relation Types",
    section: "YAML Reference",
    description: "全Relation種類の定義とプロパティ。connects, hosts, depends_on, belongs_toなど。",
    keywords: "relation types connects hosts depends_on belongs_to replicates_to backs_up monitors managed_by mounted_on applies_to listens_on 種類",
  },
  {
    path: "references.html",
    title: "References",
    section: "YAML Reference",
    description: "参照構文（シンプル、修飾、インターフェース、パス参照）と参照ルール。",
    keywords: "references reference path qualified interface simple nested participant source target 参照 パス 修飾",
  },
  {
    path: "validation.html",
    title: "Validation",
    section: "YAML Reference",
    description: "検証ルール、命名規則、Graph制約（Ownership, Reference, Nesting, Cardinality, Network整合性）。",
    keywords: "validation rules constraints ownership tree no_cycles unique_id cardinality naming kebab-case ip-requires-network ip-in-cidr network-reference-kind network-cidr-required gateway-in-cidr ip-unique-in-network valid-ip-format 検証 ルール 制約 命名規則 カーディナリティ ネットワーク 整合性",
  },
  {
    path: "example.html",
    title: "Complete Example",
    section: "YAML Reference",
    description: "完全なインフラモデルの例。regions, racks, servers, VMs, ACLs, relations。",
    keywords: "example complete full model yaml infrastructure regions racks servers vms acl relations 例 完全 サンプル",
  },  {
    path: "aws-entity-kinds.html",
    title: "AWS Entity Kinds",
    section: "YAML Reference",
    description: "AWS拡張（iacforge.aws）が定義する全47種類のEntity Kind定義とプロパティ。",
    keywords: "aws entity kinds organization account iam region availability_zone vpc subnet route_table route internet_gateway nat_gateway security_group security_group_rule elastic_ip network_acl vpc_peering_connection transit_gateway ec2 ami key_pair launch_template auto_scaling_group ebs_volume ebs_snapshot lambda_function s3_bucket efs rds dynamodb_table elasticache load_balancer target_group listener sqs_queue sns_topic api_gateway cloudfront_distribution eventbridge_rule cloudwatch_alarm cloudwatch_log_group cloudwatch_dashboard route53_zone route53_record AWS 種類 定義",
  },
  {
    path: "aws-relation-types.html",
    title: "AWS Relation Types",
    section: "YAML Reference",
    description: "AWS拡張（iacforge.aws）が定義するRelation TypeとAugmented Core Relations。",
    keywords: "aws relation types associates attaches launches routes serves forwards triggers subscribes invokes grants assumes belongs_to depends_on hosts monitors backs_up mounted_on AWS リレーション",
  },
];
