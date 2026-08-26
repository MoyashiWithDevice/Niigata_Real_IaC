package aws

import "IACForge/src/core"

// Relation types introduced by the AWS extension.
const (
	Associates core.RelationType = "aws.associates"
	Attaches   core.RelationType = "aws.attaches"
	Launches   core.RelationType = "aws.launches"
	Routes     core.RelationType = "aws.routes"
	Serves     core.RelationType = "aws.serves"
	Forwards   core.RelationType = "aws.forwards"
	Triggers   core.RelationType = "aws.triggers"
	Subscribes core.RelationType = "aws.subscribes"
	Invokes    core.RelationType = "aws.invokes"
	Grants     core.RelationType = "aws.grants"
	Assumes    core.RelationType = "aws.assumes"
	Registers  core.RelationType = "aws.registers"
)
