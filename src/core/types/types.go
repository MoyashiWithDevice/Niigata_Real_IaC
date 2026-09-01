package types

import "github.com/bababa/Niigata_Real_IaC/src/core"

// Relation types for Niigata nature and tourism resource management.
const (
	LocatedIn core.RelationType = "located_in"
	Inhabits  core.RelationType = "inhabits"
	Near      core.RelationType = "near"
	DependsOn core.RelationType = "depends_on"
	BelongsTo core.RelationType = "belongs_to"
	FlowsInto core.RelationType = "flows_into"
)
