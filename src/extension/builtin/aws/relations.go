package aws

import (
	"IACForge/src/core"
	"IACForge/src/core/types"
	"IACForge/src/extension"
	"IACForge/src/schema"
)

// RelationTypeDefinitions returns the new relation type contributions of the
// AWS extension. All types are directed binary relations.
func RelationTypeDefinitions() []extension.RelationTypeContribution {
	return []extension.RelationTypeContribution{
		{
			Type: Associates,
			Definition: &schema.RelationTypeDefinition{
				Direction:   schema.DirectionDirected,
				Description: "Association between a network resource and a VPC or subnet",
				Participants: &schema.ParticipantConstraints{
					SourceKinds:     []core.EntityKind{SecurityGroup, RouteTable, NetworkACL, ElasticIP},
					TargetKinds:     []core.EntityKind{VPC, Subnet},
					MinParticipants: 2,
					MaxParticipants: 2,
				},
			},
		},
		{
			Type: Attaches,
			Definition: &schema.RelationTypeDefinition{
				Direction:   schema.DirectionDirected,
				Description: "Attachment of a resource (EBS volume, internet gateway, Elastic IP, EFS, peering) to a compute or network target",
				Participants: &schema.ParticipantConstraints{
					SourceKinds:     []core.EntityKind{EBSVolume, InternetGateway, ElasticIP, EFS, VPCPeeringConnection},
					TargetKinds:     []core.EntityKind{EC2, VPC, TransitGateway, NATGateway},
					MinParticipants: 2,
					MaxParticipants: 2,
				},
			},
		},
		{
			Type: Launches,
			Definition: &schema.RelationTypeDefinition{
				Direction:   schema.DirectionDirected,
				Description: "Launch of compute capacity by an Auto Scaling group or launch template",
				Participants: &schema.ParticipantConstraints{
					SourceKinds:     []core.EntityKind{AutoScalingGroup, LaunchTemplate},
					TargetKinds:     []core.EntityKind{EC2},
					MinParticipants: 2,
					MaxParticipants: 2,
				},
			},
		},
		// Route entries are also nestable under a route table (auto belongs_to).
		// This relation type expresses the routing relationship explicitly,
		// independent of ownership, and so remains usable in models that do not
		// use nesting.
		{
			Type: Routes,
			Definition: &schema.RelationTypeDefinition{
				Direction:   schema.DirectionDirected,
				Description: "Ownership of a route entry by a route table",
				Participants: &schema.ParticipantConstraints{
					SourceKinds:     []core.EntityKind{RouteTable},
					TargetKinds:     []core.EntityKind{Route},
					MinParticipants: 2,
					MaxParticipants: 2,
				},
			},
		},
		{
			Type: Serves,
			Definition: &schema.RelationTypeDefinition{
				Direction:   schema.DirectionDirected,
				Description: "Load balancer traffic distribution to a target group",
				Participants: &schema.ParticipantConstraints{
					SourceKinds:     []core.EntityKind{LoadBalancer},
					TargetKinds:     []core.EntityKind{TargetGroup},
					MinParticipants: 2,
					MaxParticipants: 2,
				},
			},
		},
		{
			Type: Forwards,
			Definition: &schema.RelationTypeDefinition{
				Direction:   schema.DirectionDirected,
				Description: "Forwarding of traffic from a listener to a target group",
				Participants: &schema.ParticipantConstraints{
					SourceKinds:     []core.EntityKind{Listener},
					TargetKinds:     []core.EntityKind{TargetGroup},
					MinParticipants: 2,
					MaxParticipants: 2,
				},
			},
		},
		{
			Type: Registers,
			Definition: &schema.RelationTypeDefinition{
				Direction:   schema.DirectionDirected,
				Description: "Registration of a compute target (EC2 instance or Lambda function) with a target group",
				Participants: &schema.ParticipantConstraints{
					SourceKinds:     []core.EntityKind{TargetGroup},
					TargetKinds:     []core.EntityKind{EC2, LambdaFunction},
					MinParticipants: 2,
					MaxParticipants: 2,
				},
			},
		},
		{
			Type: Triggers,
			Definition: &schema.RelationTypeDefinition{
				Direction:   schema.DirectionDirected,
				Description: "Triggering of a target by an EventBridge rule or CloudWatch alarm",
				Participants: &schema.ParticipantConstraints{
					SourceKinds:     []core.EntityKind{EventBridgeRule, CloudWatchAlarm},
					TargetKinds:     []core.EntityKind{LambdaFunction, SNSTopic, SQSQueue},
					MinParticipants: 2,
					MaxParticipants: 2,
				},
			},
		},
		{
			Type: Subscribes,
			Definition: &schema.RelationTypeDefinition{
				Direction:   schema.DirectionDirected,
				Description: "Subscription of a queue, function, or topic to an SNS topic",
				Participants: &schema.ParticipantConstraints{
					SourceKinds:     []core.EntityKind{SNSTopic},
					TargetKinds:     []core.EntityKind{SQSQueue, LambdaFunction, SNSTopic},
					MinParticipants: 2,
					MaxParticipants: 2,
				},
			},
		},
		{
			Type: Invokes,
			Definition: &schema.RelationTypeDefinition{
				Direction:   schema.DirectionDirected,
				Description: "Invocation of a Lambda function by an API Gateway",
				Participants: &schema.ParticipantConstraints{
					SourceKinds:     []core.EntityKind{APIGateway},
					TargetKinds:     []core.EntityKind{LambdaFunction},
					MinParticipants: 2,
					MaxParticipants: 2,
				},
			},
		},
		{
			Type: Grants,
			Definition: &schema.RelationTypeDefinition{
				Direction:   schema.DirectionDirected,
				Description: "Granting of permissions from an IAM policy to an IAM principal",
				Participants: &schema.ParticipantConstraints{
					SourceKinds:     []core.EntityKind{IAMPolicy},
					TargetKinds:     []core.EntityKind{IAMUser, IAMGroup, IAMRole},
					MinParticipants: 2,
					MaxParticipants: 2,
				},
			},
		},
		{
			Type: Assumes,
			Definition: &schema.RelationTypeDefinition{
				Direction:   schema.DirectionDirected,
				Description: "Role assumption by an IAM principal",
				Participants: &schema.ParticipantConstraints{
					SourceKinds:     []core.EntityKind{IAMRole, IAMUser},
					TargetKinds:     []core.EntityKind{IAMRole},
					MinParticipants: 2,
					MaxParticipants: 2,
				},
			},
		},
	}
}

// AugmentDefinitions returns contributions that add AWS participant kinds to
// existing core relation types. Only participant kinds are merged; no other
// definition field is changed.
func AugmentDefinitions() []extension.RelationTypeContribution {
	allKinds := AllKinds()
	return []extension.RelationTypeContribution{
		{
			Type:    types.BelongsTo,
			Augment: true,
			Definition: &schema.RelationTypeDefinition{
				Participants: &schema.ParticipantConstraints{
					SourceKinds: allKinds,
					TargetKinds: allKinds,
				},
			},
		},
		{
			Type:    types.DependsOn,
			Augment: true,
			Definition: &schema.RelationTypeDefinition{
				Participants: &schema.ParticipantConstraints{
					SourceKinds: []core.EntityKind{EC2, LambdaFunction, RDS, LoadBalancer, APIGateway, AutoScalingGroup, SQSQueue},
					TargetKinds: []core.EntityKind{EC2, RDS, DynamoDBTable, S3Bucket, SQSQueue, LambdaFunction, ElastiCache, EFS, APIGateway, CloudWatchLogGroup},
				},
			},
		},
		{
			Type:    types.Hosts,
			Augment: true,
			Definition: &schema.RelationTypeDefinition{
				Participants: &schema.ParticipantConstraints{
					SourceKinds: []core.EntityKind{EC2, LambdaFunction},
					TargetKinds: []core.EntityKind{EC2, LambdaFunction},
				},
			},
		},
		{
			Type:    types.Monitors,
			Augment: true,
			Definition: &schema.RelationTypeDefinition{
				Participants: &schema.ParticipantConstraints{
					SourceKinds: []core.EntityKind{CloudWatchAlarm},
					TargetKinds: []core.EntityKind{EC2, RDS, LoadBalancer, LambdaFunction, DynamoDBTable, S3Bucket, ElastiCache, EFS, SQSQueue, APIGateway},
				},
			},
		},
		{
			Type:    types.BacksUp,
			Augment: true,
			Definition: &schema.RelationTypeDefinition{
				Participants: &schema.ParticipantConstraints{
					SourceKinds: []core.EntityKind{RDS, EBSVolume, DynamoDBTable, EFS, S3Bucket},
					TargetKinds: []core.EntityKind{EBSSnapshot, S3Bucket, EFS},
				},
			},
		},
		{
			Type:    types.MountedOn,
			Augment: true,
			Definition: &schema.RelationTypeDefinition{
				Participants: &schema.ParticipantConstraints{
					SourceKinds: []core.EntityKind{EBSVolume, EFS},
					TargetKinds: []core.EntityKind{EC2},
				},
			},
		},
	}
}
