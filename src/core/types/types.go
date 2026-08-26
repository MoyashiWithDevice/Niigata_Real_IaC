package types

import "IACForge/src/core"

const (
	Connects     core.RelationType = "connects"
	Hosts        core.RelationType = "hosts"
	DependsOn    core.RelationType = "depends_on"
	BelongsTo    core.RelationType = "belongs_to"
	ReplicatesTo core.RelationType = "replicates_to"
	BacksUp      core.RelationType = "backs_up"
	Monitors     core.RelationType = "monitors"
	ManagedBy    core.RelationType = "managed_by"
	MountedOn    core.RelationType = "mounted_on"
	AppliesTo    core.RelationType = "applies_to"
	ListensOn    core.RelationType = "listens_on"
)
