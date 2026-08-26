package aws

import (
	"sort"

	"IACForge/src/core"
	"IACForge/src/core/kinds"
	"IACForge/src/core/types"
	"IACForge/src/extension"
	"IACForge/src/schema"
)

func f64Ptr(v float64) *float64 { return &v }

func strProp(name string, required bool, desc string) schema.PropertyDefinition {
	return schema.PropertyDefinition{Name: name, Type: schema.PropertyTypeString, Required: required, Description: desc}
}

func strEnumProp(name string, required bool, desc string, values ...string) schema.PropertyDefinition {
	return schema.PropertyDefinition{
		Name:        name,
		Type:        schema.PropertyTypeString,
		Required:    required,
		Description: desc,
		Constraints: &schema.Constraint{Enum: values},
	}
}

func intProp(name string, required bool, desc string) schema.PropertyDefinition {
	return schema.PropertyDefinition{Name: name, Type: schema.PropertyTypeInteger, Required: required, Description: desc}
}

func intRangeProp(name string, required bool, min, max float64, desc string) schema.PropertyDefinition {
	return schema.PropertyDefinition{
		Name:        name,
		Type:        schema.PropertyTypeInteger,
		Required:    required,
		Description: desc,
		Constraints: &schema.Constraint{Min: f64Ptr(min), Max: f64Ptr(max)},
	}
}

func boolProp(name string, required bool, desc string) schema.PropertyDefinition {
	return schema.PropertyDefinition{Name: name, Type: schema.PropertyTypeBoolean, Required: required, Description: desc}
}

func boolPropDefault(name string, defaultVal bool, desc string) schema.PropertyDefinition {
	return schema.PropertyDefinition{Name: name, Type: schema.PropertyTypeBoolean, Default: defaultVal, Description: desc}
}

func numProp(name string, required bool, desc string) schema.PropertyDefinition {
	return schema.PropertyDefinition{Name: name, Type: schema.PropertyTypeNumber, Required: required, Description: desc}
}

func refProp(name, desc string) schema.PropertyDefinition {
	return schema.PropertyDefinition{Name: name, Type: schema.PropertyTypeReference, Description: desc}
}

func listProp(name, desc string) schema.PropertyDefinition {
	return schema.PropertyDefinition{Name: name, Type: schema.PropertyTypeList, Description: desc}
}

func mapProp(name string, required bool, desc string) schema.PropertyDefinition {
	return schema.PropertyDefinition{Name: name, Type: schema.PropertyTypeMap, Required: required, Description: desc}
}

func nest(nestKey string, child core.EntityKind) schema.NestingDefinition {
	return schema.NestingDefinition{
		NestKey:            nestKey,
		ChildKind:          child,
		AutoRelationType:   types.BelongsTo,
		AutoRelationSource: "child",
	}
}

// KindDefinitions returns the entity kind contributions of the AWS extension.
// Kinds are sorted by name for deterministic registration order.
func KindDefinitions() []extension.EntityKindContribution {
	defs := map[core.EntityKind]*schema.EntityKindDefinition{
		Organization: {
			Description: "AWS organization containing accounts and organizational units",
			Properties: []schema.PropertyDefinition{
				strProp("org_root", false, "Root organizational unit ID"),
				listProp("org_units", "Organizational units within the organization"),
			},
			NestingDefs: []schema.NestingDefinition{
				nest("accounts", Account),
			},
		},
		Account: {
			Description: "AWS account within an organization",
			Properties: []schema.PropertyDefinition{
				strProp("account_id", true, "AWS account ID (12 digits)"),
				strProp("alias", false, "Account alias used in the sign-in URL"),
				strProp("email", false, "Account root email address"),
				strProp("org_path", false, "Organizational unit path within the organization"),
			},
			NestingDefs: []schema.NestingDefinition{
				nest("regions", Region),
				nest("iam_users", IAMUser),
				nest("iam_groups", IAMGroup),
				nest("iam_roles", IAMRole),
				nest("iam_policies", IAMPolicy),
				nest("iam_instance_profiles", IAMInstanceProfile),
				nest("lambda_functions", LambdaFunction),
				nest("sqs_queues", SQSQueue),
				nest("sns_topics", SNSTopic),
				nest("dynamodb_tables", DynamoDBTable),
				nest("api_gateways", APIGateway),
				nest("route53_zones", Route53Zone),
				nest("cloudfront_distributions", CloudFrontDistribution),
				nest("eventbridge_rules", EventBridgeRule),
				nest("cloudwatch_alarms", CloudWatchAlarm),
				nest("cloudwatch_log_groups", CloudWatchLogGroup),
				nest("cloudwatch_dashboards", CloudWatchDashboard),
				nest("amis", AMI),
				nest("key_pairs", KeyPair),
				nest("launch_templates", LaunchTemplate),
				nest("auto_scaling_groups", AutoScalingGroup),
				nest("efs_filesystems", EFS),
				nest("ebs_snapshots", EBSSnapshot),
				nest("elastic_ips", ElasticIP),
				nest("target_groups", TargetGroup),
				nest("vpc_peering_connections", VPCPeeringConnection),
				nest("transit_gateways", TransitGateway),
			},
		},
		IAMUser: {
			Description: "IAM user",
			Properties: []schema.PropertyDefinition{
				strProp("arn", false, "Amazon Resource Name of the user"),
				strProp("path", false, "Path to the user in the IAM hierarchy"),
			},
		},
		IAMGroup: {
			Description: "IAM group",
			Properties: []schema.PropertyDefinition{
				strProp("arn", false, "Amazon Resource Name of the group"),
				strProp("path", false, "Path to the group in the IAM hierarchy"),
			},
		},
		IAMRole: {
			Description: "IAM role",
			Properties: []schema.PropertyDefinition{
				strProp("arn", false, "Amazon Resource Name of the role"),
				strProp("path", false, "Path to the role in the IAM hierarchy"),
				mapProp("assume_role_policy", false, "Trust policy document granting role assumption"),
			},
		},
		IAMPolicy: {
			Description: "IAM policy document",
			Properties: []schema.PropertyDefinition{
				strProp("arn", false, "Amazon Resource Name of the policy"),
				strProp("path", false, "Path to the policy in the IAM hierarchy"),
				mapProp("policy_document", true, "Policy statement document"),
			},
		},
		IAMInstanceProfile: {
			Description: "IAM instance profile attached to compute resources",
			Properties: []schema.PropertyDefinition{
				strProp("arn", false, "Amazon Resource Name of the instance profile"),
				strProp("path", false, "Path to the instance profile in the IAM hierarchy"),
				listProp("roles", "Roles contained in the instance profile"),
			},
		},
		Region: {
			Description: "AWS region",
			Properties: []schema.PropertyDefinition{
				strProp("region_code", true, "Region code (e.g. us-east-1)"),
				strEnumProp("partition", false, "AWS partition", "aws", "aws-cn", "aws-us-gov", "aws-iso", "aws-iso-b"),
				strEnumProp("opt_in_status", false, "Region opt-in status", "opted-in", "not-opted-in", "opt-in-not-required"),
			},
			NestingDefs: []schema.NestingDefinition{
				nest("availability_zones", AvailabilityZone),
				nest("vpcs", VPC),
				nest("s3_buckets", S3Bucket),
				nest("elasticache_clusters", ElastiCache),
				nest("efs_filesystems", EFS),
			},
		},
		AvailabilityZone: {
			Description: "AWS availability zone within a region",
			Properties: []schema.PropertyDefinition{
				strProp("zone_name", true, "Availability zone name (e.g. us-east-1a)"),
				strProp("zone_id", false, "Availability zone ID (e.g. use1-az1)"),
				strProp("group_name", false, "Local zone or wavelength zone group name"),
			},
			NestingDefs: []schema.NestingDefinition{
				nest("s3_buckets", S3Bucket),
				nest("elasticache_clusters", ElastiCache),
				nest("ebs_volumes", EBSVolume),
			},
		},
		VPC: {
			Description: "AWS Virtual Private Cloud",
			Properties: []schema.PropertyDefinition{
				strProp("cidr_block", true, "IPv4 CIDR block of the VPC"),
				strEnumProp("tenancy", false, "Instance tenancy", "default", "dedicated", "host"),
				boolPropDefault("dns_support", true, "Whether DNS resolution is supported"),
				boolPropDefault("flow_logs", false, "Whether VPC flow logs are enabled"),
			},
			NestingDefs: []schema.NestingDefinition{
				nest("subnets", Subnet),
				nest("security_groups", SecurityGroup),
				nest("route_tables", RouteTable),
				nest("internet_gateways", InternetGateway),
				nest("network_acls", NetworkACL),
			},
		},
		Subnet: {
			Description: "AWS subnet within a VPC",
			Properties: []schema.PropertyDefinition{
				strProp("cidr_block", true, "IPv4 CIDR block of the subnet"),
				boolPropDefault("map_public_ip", false, "Whether instances receive public IPs by default"),
				boolPropDefault("default_for_az", false, "Whether this is the default subnet for the availability zone"),
			},
			NestingDefs: []schema.NestingDefinition{
				nest("ec2_instances", EC2),
				nest("rds_instances", RDS),
				nest("load_balancers", LoadBalancer),
				nest("nat_gateways", NATGateway),
			},
		},
		RouteTable: {
			Description: "AWS route table associated with a VPC",
			NestingDefs: []schema.NestingDefinition{
				nest("routes", Route),
			},
		},
		Route: {
			Description: "Route entry within a route table",
			Properties: []schema.PropertyDefinition{
				strProp("destination_cidr", false, "Destination IPv4 CIDR block"),
				refProp("gateway_id", "Reference to the internet gateway or virtual private gateway"),
				refProp("nat_gateway_id", "Reference to the NAT gateway"),
				refProp("transit_gateway_id", "Reference to the transit gateway"),
				refProp("vpc_peering_connection_id", "Reference to the VPC peering connection"),
			},
		},
		InternetGateway: {
			Description: "AWS internet gateway attached to a VPC",
		},
		NATGateway: {
			Description: "AWS NAT gateway",
			Properties: []schema.PropertyDefinition{
				strEnumProp("connectivity_type", false, "Connectivity type", "public", "private"),
			},
		},
		SecurityGroup: {
			Description: "AWS security group",
			Properties: []schema.PropertyDefinition{
				strProp("description", false, "Description of the security group"),
				refProp("vpc_id", "Reference to the VPC this security group belongs to"),
			},
			NestingDefs: []schema.NestingDefinition{
				nest("security_group_rules", SecurityGroupRule),
			},
		},
		SecurityGroupRule: {
			Description: "Rule within a security group",
			Properties: []schema.PropertyDefinition{
				strEnumProp("type", false, "Rule direction", "ingress", "egress"),
				intRangeProp("from_port", false, 0, 65535, "Start of the port range"),
				intRangeProp("to_port", false, 0, 65535, "End of the port range"),
				strEnumProp("protocol", false, "IP protocol", "tcp", "udp", "icmp", "icmpv6", "-1"),
				refProp("source_security_group", "Source security group reference"),
				strProp("source_cidr", false, "Source IPv4 CIDR block"),
				strProp("destination_cidr", false, "Destination IPv4 CIDR block"),
			},
		},
		ElasticIP: {
			Description: "AWS Elastic IP address",
			Properties: []schema.PropertyDefinition{
				strEnumProp("domain", false, "Address domain", "vpc", "standard"),
				strProp("public_ip", false, "Public IPv4 address"),
				refProp("association", "Entity this Elastic IP is associated with"),
			},
		},
		NetworkACL: {
			Description: "AWS network access control list",
			Properties: []schema.PropertyDefinition{
				strEnumProp("default_action", false, "Default action for traffic", "allow", "deny"),
			},
		},
		VPCPeeringConnection: {
			Description: "AWS VPC peering connection",
			Properties: []schema.PropertyDefinition{
				boolPropDefault("auto_accept", false, "Whether the peering connection is accepted automatically"),
				strProp("peer_vpc_id", false, "Peer VPC ID"),
				strProp("peer_region", false, "Region of the peer VPC"),
			},
		},
		TransitGateway: {
			Description: "AWS transit gateway",
			Properties: []schema.PropertyDefinition{
				intProp("amazon_side_asn", false, "Private ASN of the transit gateway"),
				strEnumProp("dns_support", false, "DNS support", "enable", "disable"),
				strEnumProp("vpn_ecmp_support", false, "VPN ECMP support", "enable", "disable"),
			},
		},
		EC2: {
			Description: "AWS EC2 instance",
			Properties: []schema.PropertyDefinition{
				strProp("instance_type", true, "Instance type (e.g. t3.micro)"),
				refProp("ami", "Reference to the AMI used to launch the instance"),
				refProp("key_pair", "Reference to the key pair used for SSH access"),
				refProp("subnet", "Reference to the subnet the instance runs in"),
				listProp("security_groups", "References to the security groups attached to the instance"),
				strProp("user_data", false, "User data script passed at launch"),
				refProp("iam_instance_profile", "Reference to the IAM instance profile"),
			},
			NestingDefs: []schema.NestingDefinition{
				nest("applications", kinds.Application),
			},
		},
		AMI: {
			Description: "AWS Amazon Machine Image",
			Properties: []schema.PropertyDefinition{
				strProp("image_id", false, "AMI image ID (e.g. ami-12345678)"),
				strEnumProp("architecture", false, "CPU architecture", "x86_64", "arm64"),
				strEnumProp("virtualization_type", false, "Virtualization type", "hvm", "paravirtual"),
			},
		},
		KeyPair: {
			Description: "AWS key pair",
			Properties: []schema.PropertyDefinition{
				strEnumProp("key_type", false, "Key type", "rsa", "ed25519"),
				strProp("fingerprint", false, "SHA-1 fingerprint of the public key"),
				strProp("public_key", false, "Public key material"),
			},
		},
		LaunchTemplate: {
			Description: "AWS EC2 launch template",
			Properties: []schema.PropertyDefinition{
				strProp("image_id", false, "AMI image ID"),
				strProp("instance_type", false, "Instance type"),
				strProp("key_name", false, "Key pair name"),
			},
		},
		AutoScalingGroup: {
			Description: "AWS Auto Scaling group",
			Properties: []schema.PropertyDefinition{
				intRangeProp("min_size", false, 0, 5000, "Minimum number of instances"),
				intRangeProp("max_size", false, 0, 5000, "Maximum number of instances"),
				intRangeProp("desired_capacity", false, 0, 5000, "Desired number of instances"),
				refProp("launch_template", "Reference to the launch template"),
				listProp("vpc_zone_identifier", "References to the subnets across which instances are distributed"),
			},
		},
		EBSVolume: {
			Description: "AWS Elastic Block Store volume",
			Properties: []schema.PropertyDefinition{
				strEnumProp("volume_type", false, "Volume type", "gp2", "gp3", "io1", "io2", "st1", "sc1", "standard"),
				intProp("size_gb", false, "Volume size in GiB"),
				intProp("iops", false, "Provisioned IOPS"),
				boolProp("encrypted", false, "Whether the volume is encrypted"),
				strProp("kms_key_id", false, "KMS key ID used for encryption"),
			},
		},
		EBSSnapshot: {
			Description: "AWS EBS snapshot",
			Properties: []schema.PropertyDefinition{
				strProp("volume_id", false, "Source volume ID"),
				intProp("size_gb", false, "Snapshot size in GiB"),
				boolProp("encrypted", false, "Whether the snapshot is encrypted"),
				strEnumProp("state", false, "Snapshot state", "pending", "completed", "error"),
			},
		},
		LambdaFunction: {
			Description: "AWS Lambda function",
			Properties: []schema.PropertyDefinition{
				strProp("runtime", true, "Runtime identifier (e.g. python3.12, nodejs20.x)"),
				strProp("handler", false, "Function handler"),
				refProp("role", "Reference to the IAM role"),
				intRangeProp("memory_size", false, 128, 10240, "Memory allocated in MB"),
				mapProp("vpc_config", false, "VPC configuration (subnet and security group IDs)"),
			},
		},
		S3Bucket: {
			Description: "AWS Simple Storage Service bucket",
			Properties: []schema.PropertyDefinition{
				boolPropDefault("versioning", false, "Whether bucket versioning is enabled"),
				boolPropDefault("encryption", false, "Whether default server-side encryption is enabled"),
				listProp("lifecycle_rules", "Lifecycle rule configurations"),
			},
		},
		EFS: {
			Description: "AWS Elastic File System",
			Properties: []schema.PropertyDefinition{
				boolPropDefault("encrypted", false, "Whether the file system is encrypted"),
				strEnumProp("performance_mode", false, "Performance mode", "generalPurpose", "maxIO"),
				strEnumProp("throughput_mode", false, "Throughput mode", "bursting", "provisioned", "elastic"),
			},
		},
		RDS: {
			Description: "AWS Relational Database Service instance",
			Properties: []schema.PropertyDefinition{
				strEnumProp("engine", false, "Database engine", "aurora", "aurora-mysql", "aurora-postgresql", "mysql", "postgres", "mariadb", "oracle-se2", "sqlserver-ee"),
				strProp("engine_version", false, "Engine version"),
				strProp("instance_class", true, "Instance class (e.g. db.t3.micro)"),
				boolPropDefault("multi_az", false, "Whether Multi-AZ deployment is enabled"),
				intProp("storage_gb", false, "Allocated storage in GiB"),
			},
		},
		DynamoDBTable: {
			Description: "AWS DynamoDB table",
			Properties: []schema.PropertyDefinition{
				strEnumProp("billing_mode", false, "Billing mode", "PROVISIONED", "ON_DEMAND"),
				strEnumProp("table_class", false, "Table class", "STANDARD", "STANDARD_INFREQUENT_ACCESS"),
				boolPropDefault("stream_enabled", false, "Whether DynamoDB Streams is enabled"),
				strProp("partition_key", false, "Partition key name"),
				strProp("sort_key", false, "Sort key name"),
			},
		},
		ElastiCache: {
			Description: "AWS ElastiCache cluster",
			Properties: []schema.PropertyDefinition{
				strEnumProp("engine", false, "Cache engine", "redis", "memcached"),
				strProp("node_type", false, "Cache node type (e.g. cache.t3.micro)"),
				intProp("num_cache_nodes", false, "Number of cache nodes"),
				boolPropDefault("cluster_mode", false, "Whether cluster mode is enabled"),
			},
		},
		LoadBalancer: {
			Description: "AWS Elastic Load Balancer",
			Properties: []schema.PropertyDefinition{
				strEnumProp("type", false, "Load balancer type", "application", "network", "gateway", "classic"),
				strEnumProp("scheme", false, "Load balancer scheme", "internet-facing", "internal"),
				strProp("dns_name", false, "DNS name of the load balancer"),
				listProp("subnets", "References to the subnets the load balancer spans"),
			},
			NestingDefs: []schema.NestingDefinition{
				nest("listeners", Listener),
			},
		},
		TargetGroup: {
			Description: "AWS target group for a load balancer",
			Properties: []schema.PropertyDefinition{
				strEnumProp("protocol", false, "Protocol", "HTTP", "HTTPS", "TCP", "TCP_UDP", "UDP", "TLS"),
				intRangeProp("port", false, 1, 65535, "Traffic port"),
				strEnumProp("target_type", false, "Target type", "instance", "ip", "lambda", "alb"),
				strProp("health_check_path", false, "Health check path"),
			},
		},
		Listener: {
			Description: "AWS load balancer listener",
			Properties: []schema.PropertyDefinition{
				strEnumProp("protocol", false, "Listener protocol", "HTTP", "HTTPS", "TCP", "TLS"),
				intRangeProp("port", false, 1, 65535, "Listener port"),
				strProp("default_action", false, "Default action (e.g. forward, redirect)"),
				strProp("certificate_arn", false, "SSL/TLS certificate ARN"),
			},
		},
		SQSQueue: {
			Description: "AWS Simple Queue Service queue",
			Properties: []schema.PropertyDefinition{
				boolPropDefault("fifo", false, "Whether the queue is FIFO"),
				intProp("visibility_timeout", false, "Visibility timeout in seconds"),
				intProp("delay_seconds", false, "Message delay in seconds"),
				intProp("message_retention_seconds", false, "Message retention period in seconds"),
			},
		},
		SNSTopic: {
			Description: "AWS Simple Notification Service topic",
			Properties: []schema.PropertyDefinition{
				boolPropDefault("fifo", false, "Whether the topic is FIFO"),
				strProp("display_name", false, "Display name for email notifications"),
			},
		},
		APIGateway: {
			Description: "AWS API Gateway",
			Properties: []schema.PropertyDefinition{
				strEnumProp("endpoint_type", false, "Endpoint type", "REGIONAL", "EDGE", "PRIVATE"),
				strEnumProp("protocol_type", false, "API protocol", "REST", "HTTP", "WEBSOCKET"),
			},
		},
		CloudFrontDistribution: {
			Description: "AWS CloudFront content delivery network distribution",
			Properties: []schema.PropertyDefinition{
				strProp("origin", false, "Origin domain name"),
				boolPropDefault("enabled", true, "Whether the distribution is enabled"),
				strEnumProp("price_class", false, "Price class", "PriceClass_100", "PriceClass_200", "PriceClass_All"),
				strProp("domain_name", false, "CloudFront domain name"),
			},
		},
		EventBridgeRule: {
			Description: "AWS EventBridge rule",
			Properties: []schema.PropertyDefinition{
				strProp("event_pattern", false, "Event pattern JSON"),
				strProp("schedule_expression", false, "Cron or rate schedule expression"),
				strEnumProp("state", false, "Rule state", "ENABLED", "DISABLED"),
			},
		},
		CloudWatchAlarm: {
			Description: "AWS CloudWatch alarm",
			Properties: []schema.PropertyDefinition{
				strProp("metric_name", false, "Metric name"),
				strProp("namespace", false, "Metric namespace"),
				numProp("threshold", false, "Alarm threshold value"),
				strEnumProp("comparison_operator", false, "Comparison operator", "GreaterThanOrEqualToThreshold", "GreaterThanThreshold", "LessThanThreshold", "LessThanOrEqualToThreshold"),
				intProp("period_seconds", false, "Evaluation period in seconds"),
			},
		},
		CloudWatchLogGroup: {
			Description: "AWS CloudWatch log group",
			Properties: []schema.PropertyDefinition{
				intRangeProp("retention_days", false, 1, 3653, "Log retention period in days"),
			},
		},
		CloudWatchDashboard: {
			Description: "AWS CloudWatch dashboard",
			Properties: []schema.PropertyDefinition{
				mapProp("body", false, "Dashboard body widget configuration"),
			},
		},
		Route53Zone: {
			Description: "AWS Route 53 hosted zone",
			Properties: []schema.PropertyDefinition{
				boolPropDefault("private", false, "Whether the hosted zone is private"),
				strProp("comment", false, "Zone comment"),
				refProp("vpc_id", "Reference to the VPC a private zone is associated with"),
			},
			NestingDefs: []schema.NestingDefinition{
				nest("records", Route53Record),
			},
		},
		Route53Record: {
			Description: "AWS Route 53 DNS record",
			Properties: []schema.PropertyDefinition{
				strProp("record_name", true, "Record name"),
				strEnumProp("type", false, "Record type", "A", "AAAA", "CNAME", "MX", "NS", "PTR", "SOA", "SRV", "TXT"),
				intProp("ttl", false, "Time to live in seconds"),
				strProp("alias_target", false, "Alias target resource"),
				listProp("records", "Resource records"),
			},
		},
	}

	kinds := AllKinds()
	contribs := make([]extension.EntityKindContribution, 0, len(kinds))
	for _, kind := range kinds {
		def, ok := defs[kind]
		if !ok {
			panic("aws: missing definition for kind " + string(kind))
		}
		contribs = append(contribs, extension.EntityKindContribution{Kind: kind, Definition: def})
	}
	sort.Slice(contribs, func(i, j int) bool { return contribs[i].Kind < contribs[j].Kind })
	return contribs
}
