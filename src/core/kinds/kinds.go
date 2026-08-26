package kinds

import "IACForge/src/core"

const (
	Region            core.EntityKind = "region"
	Rack              core.EntityKind = "rack"
	Server            core.EntityKind = "server"
	Interface         core.EntityKind = "interface"
	Cable             core.EntityKind = "cable"
	PowerDistribution core.EntityKind = "power_distribution"
	Network           core.EntityKind = "network"
	VLAN              core.EntityKind = "vlan"
	Switch            core.EntityKind = "switch"
	Router            core.EntityKind = "router"
	Firewall          core.EntityKind = "firewall"
	ACL               core.EntityKind = "acl"
	ACLRule           core.EntityKind = "acl_rule"
	VM                core.EntityKind = "vm"
	Container         core.EntityKind = "container"
	Application       core.EntityKind = "application"
	OpenPort          core.EntityKind = "open_port"
	Storage           core.EntityKind = "storage"
	Volume            core.EntityKind = "volume"
	Cluster           core.EntityKind = "cluster"
	AvailabilityZone  core.EntityKind = "availability_zone"
)

func IsValidStatus(s core.Status) bool {
	validStatuses := []core.Status{
		core.StatusPlanned,
		core.StatusActive,
		core.StatusMaintenance,
		core.StatusDeprecated,
		core.StatusOffline,
		core.StatusStandby,
	}
	for _, status := range validStatuses {
		if status == s {
			return true
		}
	}
	return false
}
