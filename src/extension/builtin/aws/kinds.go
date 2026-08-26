package aws

import "IACForge/src/core"

// Organization / Account
const (
	Organization core.EntityKind = "aws.organization"
	Account      core.EntityKind = "aws.account"
)

// IAM
const (
	IAMUser            core.EntityKind = "aws.iam_user"
	IAMGroup           core.EntityKind = "aws.iam_group"
	IAMRole            core.EntityKind = "aws.iam_role"
	IAMPolicy          core.EntityKind = "aws.iam_policy"
	IAMInstanceProfile core.EntityKind = "aws.iam_instance_profile"
)

// Network
const (
	Region               core.EntityKind = "aws.region"
	AvailabilityZone     core.EntityKind = "aws.availability_zone"
	VPC                  core.EntityKind = "aws.vpc"
	Subnet               core.EntityKind = "aws.subnet"
	RouteTable           core.EntityKind = "aws.route_table"
	Route                core.EntityKind = "aws.route"
	InternetGateway      core.EntityKind = "aws.internet_gateway"
	NATGateway           core.EntityKind = "aws.nat_gateway"
	SecurityGroup        core.EntityKind = "aws.security_group"
	SecurityGroupRule    core.EntityKind = "aws.security_group_rule"
	ElasticIP            core.EntityKind = "aws.elastic_ip"
	NetworkACL           core.EntityKind = "aws.network_acl"
	VPCPeeringConnection core.EntityKind = "aws.vpc_peering_connection"
	TransitGateway       core.EntityKind = "aws.transit_gateway"
)

// Compute
const (
	EC2              core.EntityKind = "aws.ec2"
	AMI              core.EntityKind = "aws.ami"
	KeyPair          core.EntityKind = "aws.key_pair"
	LaunchTemplate   core.EntityKind = "aws.launch_template"
	AutoScalingGroup core.EntityKind = "aws.auto_scaling_group"
	EBSVolume        core.EntityKind = "aws.ebs_volume"
	EBSSnapshot      core.EntityKind = "aws.ebs_snapshot"
	LambdaFunction   core.EntityKind = "aws.lambda_function"
)

// Storage / DB
const (
	S3Bucket      core.EntityKind = "aws.s3_bucket"
	EFS           core.EntityKind = "aws.efs"
	RDS           core.EntityKind = "aws.rds"
	DynamoDBTable core.EntityKind = "aws.dynamodb_table"
	ElastiCache   core.EntityKind = "aws.elasticache"
)

// Load Balancer
const (
	LoadBalancer core.EntityKind = "aws.load_balancer"
	TargetGroup  core.EntityKind = "aws.target_group"
	Listener     core.EntityKind = "aws.listener"
)

// Integration
const (
	SQSQueue               core.EntityKind = "aws.sqs_queue"
	SNSTopic               core.EntityKind = "aws.sns_topic"
	APIGateway             core.EntityKind = "aws.api_gateway"
	CloudFrontDistribution core.EntityKind = "aws.cloudfront_distribution"
	EventBridgeRule        core.EntityKind = "aws.eventbridge_rule"
)

// Monitoring
const (
	CloudWatchAlarm     core.EntityKind = "aws.cloudwatch_alarm"
	CloudWatchLogGroup  core.EntityKind = "aws.cloudwatch_log_group"
	CloudWatchDashboard core.EntityKind = "aws.cloudwatch_dashboard"
)

// DNS
const (
	Route53Zone   core.EntityKind = "aws.route53_zone"
	Route53Record core.EntityKind = "aws.route53_record"
)

// AllKinds returns every entity kind defined by the AWS extension.
func AllKinds() []core.EntityKind {
	return []core.EntityKind{
		Organization, Account,
		IAMUser, IAMGroup, IAMRole, IAMPolicy, IAMInstanceProfile,
		Region, AvailabilityZone, VPC, Subnet, RouteTable, Route,
		InternetGateway, NATGateway, SecurityGroup, SecurityGroupRule,
		ElasticIP, NetworkACL, VPCPeeringConnection, TransitGateway,
		EC2, AMI, KeyPair, LaunchTemplate, AutoScalingGroup, EBSVolume,
		EBSSnapshot, LambdaFunction,
		S3Bucket, EFS, RDS, DynamoDBTable, ElastiCache,
		LoadBalancer, TargetGroup, Listener,
		SQSQueue, SNSTopic, APIGateway, CloudFrontDistribution, EventBridgeRule,
		CloudWatchAlarm, CloudWatchLogGroup, CloudWatchDashboard,
		Route53Zone, Route53Record,
	}
}
