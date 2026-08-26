package schema

import (
	"IACForge/src/core"
	"IACForge/src/core/kinds"
	"IACForge/src/core/types"
)

func intPtr(v float64) *float64 { return &v }

// CoreSchema returns the default core schema with all entity kinds and relation types.
func CoreSchema() *Schema {
	s := NewSchema("1.0.0", "1.0")
	s.Version.Description = "Core Infrastructure Schema"

	registerEntityKinds(s)
	registerRelationTypes(s)

	return s
}

func registerEntityKinds(s *Schema) {
	// Global nesting definitions shared by all entity kinds
	s.NestingDefs = []NestingDefinition{
		{NestKey: "interfaces", ChildKind: kinds.Interface, AutoRelationType: types.BelongsTo, AutoRelationSource: "child"},
		{NestKey: "servers", ChildKind: kinds.Server, AutoRelationType: types.BelongsTo, AutoRelationSource: "child"},
		{NestKey: "switches", ChildKind: kinds.Switch, AutoRelationType: types.BelongsTo, AutoRelationSource: "child"},
		{NestKey: "routers", ChildKind: kinds.Router, AutoRelationType: types.BelongsTo, AutoRelationSource: "child"},
		{NestKey: "firewalls", ChildKind: kinds.Firewall, AutoRelationType: types.BelongsTo, AutoRelationSource: "child"},
		{NestKey: "networks", ChildKind: kinds.Network, AutoRelationType: types.BelongsTo, AutoRelationSource: "child"},
	}

	s.AddEntityKind(kinds.Region, &EntityKindDefinition{
		Description: "Geographic region where infrastructure is deployed",
		Properties: []PropertyDefinition{
			{Name: "address", Type: PropertyTypeString, Required: false, Description: "Physical address"},
			{Name: "latitude", Type: PropertyTypeNumber, Required: false, Description: "Geographic latitude"},
			{Name: "longitude", Type: PropertyTypeNumber, Required: false, Description: "Geographic longitude"},
			{Name: "timezone", Type: PropertyTypeString, Required: false, Description: "Timezone identifier"},
		},
		NestingDefs: []NestingDefinition{
			{NestKey: "racks", ChildKind: kinds.Rack, AutoRelationType: types.BelongsTo, AutoRelationSource: "child"},
			{NestKey: "clusters", ChildKind: kinds.Cluster, AutoRelationType: types.BelongsTo, AutoRelationSource: "child"},
			{NestKey: "availability_zones", ChildKind: kinds.AvailabilityZone, AutoRelationType: types.BelongsTo, AutoRelationSource: "child"},
		},
	})

	s.AddEntityKind(kinds.Rack, &EntityKindDefinition{
		Description: "Physical rack enclosure within a region",
		Properties: []PropertyDefinition{
			{Name: "height_units", Type: PropertyTypeInteger, Required: false, Default: 42, Description: "Rack height in rack units (U)"},
			{Name: "power_capacity_watts", Type: PropertyTypeInteger, Required: false, Description: "Total power capacity in watts"},
			{Name: "max_load_kg", Type: PropertyTypeNumber, Required: false, Description: "Maximum weight capacity in kg"},
		},
	})

	s.AddEntityKind(kinds.Server, &EntityKindDefinition{
		Description: "Physical or virtual compute host",
		Properties: []PropertyDefinition{
			{Name: "manufacturer", Type: PropertyTypeString, Required: false, Description: "Hardware manufacturer"},
			{Name: "model", Type: PropertyTypeString, Required: false, Description: "Hardware model"},
			{Name: "serial_number", Type: PropertyTypeString, Required: false, Description: "Serial number"},
			{Name: "cpu", Type: PropertyTypeList, Required: false, Description: "CPU configurations", Properties: []PropertyDefinition{
				{Name: "cores", Type: PropertyTypeInteger, Required: false, Constraints: &Constraint{Min: intPtr(1), Max: intPtr(1024)}, Description: "Number of CPU cores"},
				{Name: "architecture", Type: PropertyTypeString, Required: false, Constraints: &Constraint{Enum: []string{"x86_64", "arm64"}}, Description: "CPU architecture (x86_64, arm64)"},
			}},
			{Name: "memory", Type: PropertyTypeList, Required: false, Description: "Memory modules", Properties: []PropertyDefinition{
				{Name: "size_gb", Type: PropertyTypeNumber, Required: true, Description: "Memory module size in GB"},
				{Name: "speed", Type: PropertyTypeInteger, Required: false, Description: "Memory speed in MHz"},
				{Name: "type", Type: PropertyTypeString, Required: false, Description: "Memory type (ddr4, ddr5, lpddr4, lpddr5)"},
			}},
			{Name: "storage", Type: PropertyTypeList, Required: false, Description: "Local storage devices", Properties: []PropertyDefinition{
				{Name: "size_gb", Type: PropertyTypeNumber, Required: false, Description: "Storage size in GB"},
				{Name: "type", Type: PropertyTypeString, Required: false, Description: "Storage type (ssd, hdd, nvme)"},
			}},
			{Name: "platform", Type: PropertyTypeString, Required: false, Constraints: &Constraint{Enum: []string{"proxmox", "vmware", "kubernetes", "baremetal"}}, Description: "Virtualization platform"},
			{Name: "bios_version", Type: PropertyTypeString, Required: false, Description: "BIOS/UEFI version"},
		},
		NestingDefs: []NestingDefinition{
			{NestKey: "vms", ChildKind: kinds.VM, AutoRelationType: types.Hosts, AutoRelationSource: "parent"},
			{NestKey: "containers", ChildKind: kinds.Container, AutoRelationType: types.Hosts, AutoRelationSource: "parent"},
		},
	})

	s.AddEntityKind(kinds.Interface, &EntityKindDefinition{
		Description: "Network interface on a device",
		Properties: []PropertyDefinition{
			{Name: "type", Type: PropertyTypeString, Required: false, Default: "ethernet", Constraints: &Constraint{Enum: []string{"ethernet", "fiber", "wireless", "virtual", "bond", "vlan", "bridge", "loopback"}}, Description: "Interface type (ethernet, fiber, wireless, virtual, bond, vlan, bridge, loopback)"},
			{Name: "mode", Type: PropertyTypeString, Required: false, Default: "none", Constraints: &Constraint{Enum: []string{"access", "trunk", "hybrid", "none"}}, Description: "Interface mode (access, trunk, hybrid, none)"},
			{Name: "speed_mbps", Type: PropertyTypeInteger, Required: false, Description: "Interface speed in Mbps"},
			{Name: "mac_address", Type: PropertyTypeString, Required: false, Description: "MAC address"},
			{Name: "ip_address", Type: PropertyTypeList, Required: false, Description: "IP addresses if configured"},
			{Name: "network", Type: PropertyTypeReference, Required: false, Description: "Reference to the network this interface belongs to"},
			{Name: "vlan_id", Type: PropertyTypeInteger, Required: false, Constraints: &Constraint{Min: intPtr(1), Max: intPtr(4094)}, Description: "VLAN identifier for a VLAN sub-interface (e.g., vmbr.20)"},
			{Name: "mtu", Type: PropertyTypeInteger, Required: false, Default: 1500, Description: "Maximum transmission unit"},
		},
		NestingDefs: []NestingDefinition{
			{NestKey: "vlans", ChildKind: kinds.VLAN, AutoRelationType: types.BelongsTo, AutoRelationSource: "child"},
			{NestKey: "cables", ChildKind: kinds.Cable},
		},
	})

	s.AddEntityKind(kinds.Cable, &EntityKindDefinition{
		Description: "Physical cable connecting two or more interfaces",
		Properties: []PropertyDefinition{
			{Name: "cable_type", Type: PropertyTypeString, Required: false, Default: "copper", Description: "Cable type (copper, fiber, dac)"},
			{Name: "length_meters", Type: PropertyTypeNumber, Required: false, Description: "Cable length in meters"},
			{Name: "connector_a", Type: PropertyTypeString, Required: false, Description: "Connector type at end A"},
			{Name: "connector_b", Type: PropertyTypeString, Required: false, Description: "Connector type at end B"},
		},
	})

	s.AddEntityKind(kinds.PowerDistribution, &EntityKindDefinition{
		Description: "Power distribution unit (PDU) or power feed",
		Properties: []PropertyDefinition{
			{Name: "capacity_amps", Type: PropertyTypeInteger, Required: false, Description: "Total amperage capacity"},
			{Name: "voltage", Type: PropertyTypeNumber, Required: false, Default: float64(240), Description: "Operating voltage"},
			{Name: "phases", Type: PropertyTypeInteger, Required: false, Default: 1, Description: "Number of phases"},
		},
	})

	s.AddEntityKind(kinds.Network, &EntityKindDefinition{
		Description: "Logical network or broadcast domain",
		Properties: []PropertyDefinition{
			{Name: "cidr", Type: PropertyTypeString, Required: false, Description: "Network CIDR notation"},
			{Name: "gateway", Type: PropertyTypeString, Required: false, Description: "Default gateway address"},
			{Name: "dns_servers", Type: PropertyTypeList, Required: false, Description: "DNS server addresses"},
			{Name: "vlan_id", Type: PropertyTypeInteger, Required: false, Description: "Associated VLAN ID"},
			{Name: "network_type", Type: PropertyTypeString, Required: false, Constraints: &Constraint{Enum: []string{"management", "storage", "vm", "public"}}, Description: "Network type (management, storage, vm, public)"},
		},
	})

	s.AddEntityKind(kinds.VLAN, &EntityKindDefinition{
		Description: "Virtual LAN configuration",
		Properties: []PropertyDefinition{
			{Name: "vlan_id", Type: PropertyTypeInteger, Required: true, Constraints: &Constraint{Min: intPtr(1), Max: intPtr(4094)}, Description: "VLAN identifier (1-4094)"},
			{Name: "tagged", Type: PropertyTypeBoolean, Required: false, Default: false, Description: "Whether this VLAN carries tagged traffic on a trunk port"},
			{Name: "associated_network", Type: PropertyTypeReference, Required: false, Description: "Reference to parent network"},
		},
	})

	s.AddEntityKind(kinds.Switch, &EntityKindDefinition{
		Description: "Network switch",
		Properties: []PropertyDefinition{
			{Name: "manufacturer", Type: PropertyTypeString, Required: false, Description: "Hardware manufacturer"},
			{Name: "model", Type: PropertyTypeString, Required: false, Description: "Hardware model"},
			{Name: "serial_number", Type: PropertyTypeString, Required: false, Description: "Serial number"},
			{Name: "port_count", Type: PropertyTypeInteger, Required: false, Description: "Total port count"},
			{Name: "managed", Type: PropertyTypeBoolean, Required: false, Default: true, Description: "Whether switch is managed"},
			{Name: "stackable", Type: PropertyTypeBoolean, Required: false, Default: false, Description: "Whether switch supports stacking"},
		},
		NestingDefs: []NestingDefinition{
			{NestKey: "ports", ChildKind: kinds.Interface, AutoRelationType: types.BelongsTo, AutoRelationSource: "child"},
		},
	})

	s.AddEntityKind(kinds.Router, &EntityKindDefinition{
		Description: "Network router",
		Properties: []PropertyDefinition{
			{Name: "manufacturer", Type: PropertyTypeString, Required: false, Description: "Hardware manufacturer"},
			{Name: "model", Type: PropertyTypeString, Required: false, Description: "Hardware model"},
			{Name: "serial_number", Type: PropertyTypeString, Required: false, Description: "Serial number"},
		},
		NestingDefs: []NestingDefinition{
			{NestKey: "ports", ChildKind: kinds.Interface, AutoRelationType: types.BelongsTo, AutoRelationSource: "child"},
		},
	})

	s.AddEntityKind(kinds.Firewall, &EntityKindDefinition{
		Description: "Network firewall",
		Properties: []PropertyDefinition{
			{Name: "manufacturer", Type: PropertyTypeString, Required: false, Description: "Hardware manufacturer"},
			{Name: "model", Type: PropertyTypeString, Required: false, Description: "Hardware model"},
			{Name: "serial_number", Type: PropertyTypeString, Required: false, Description: "Serial number"},
			{Name: "throughput_gbps", Type: PropertyTypeNumber, Required: false, Description: "Maximum throughput in Gbps"},
		},
		NestingDefs: []NestingDefinition{
			{NestKey: "acls", ChildKind: kinds.ACL, AutoRelationType: types.BelongsTo, AutoRelationSource: "child"},
		},
	})

	s.AddEntityKind(kinds.ACL, &EntityKindDefinition{
		Description: "Access Control List containing ordered rules",
		Properties: []PropertyDefinition{
			{Name: "default_action", Type: PropertyTypeString, Required: false, Default: "deny", Constraints: &Constraint{Enum: []string{"allow", "deny"}}, Description: "Default action when no rule matches"},
			{Name: "direction", Type: PropertyTypeString, Required: false, Constraints: &Constraint{Enum: []string{"inbound", "outbound", "both"}}, Description: "Traffic direction"},
			{Name: "protocol", Type: PropertyTypeString, Required: false, Default: "any", Constraints: &Constraint{Enum: []string{"tcp", "udp", "icmp", "any"}}, Description: "Protocol filter (tcp, udp, icmp, any)"},
		},
		NestingDefs: []NestingDefinition{
			{NestKey: "acl_rules", ChildKind: kinds.ACLRule, AutoRelationType: types.BelongsTo, AutoRelationSource: "child"},
		},
	})

	s.AddEntityKind(kinds.ACLRule, &EntityKindDefinition{
		Description: "Single rule within an Access Control List",
		Properties: []PropertyDefinition{
			{Name: "action", Type: PropertyTypeString, Required: true, Constraints: &Constraint{Enum: []string{"allow", "deny"}}, Description: "Rule action (allow, deny)"},
			{Name: "protocol", Type: PropertyTypeString, Required: false, Default: "any", Description: "Protocol (tcp, udp, icmp, any)"},
			{Name: "source_address", Type: PropertyTypeString, Required: false, Default: "any", Description: "Source IP address or CIDR"},
			{Name: "source_port", Type: PropertyTypeString, Required: false, Default: "any", Description: "Source port or range"},
			{Name: "destination_address", Type: PropertyTypeString, Required: false, Default: "any", Description: "Destination IP address or CIDR"},
			{Name: "destination_port", Type: PropertyTypeString, Required: false, Default: "any", Description: "Destination port or range"},
			{Name: "enabled", Type: PropertyTypeBoolean, Required: false, Default: true, Description: "Whether this rule is active"},
		},
	})

	s.AddEntityKind(kinds.VM, &EntityKindDefinition{
		Description: "Virtual machine",
		Properties: []PropertyDefinition{
			{Name: "cpu", Type: PropertyTypeList, Required: false, Description: "Virtual CPU configurations", Properties: []PropertyDefinition{
				{Name: "cores", Type: PropertyTypeInteger, Required: false, Description: "Number of virtual CPU cores"},
				{Name: "architecture", Type: PropertyTypeString, Required: false, Constraints: &Constraint{Enum: []string{"x86_64", "arm64"}}, Description: "CPU architecture (x86_64, arm64)"},
			}},
			{Name: "memory", Type: PropertyTypeList, Required: false, Description: "Memory modules", Properties: []PropertyDefinition{
				{Name: "size_gb", Type: PropertyTypeNumber, Required: true, Description: "Memory module size in GB"},
				{Name: "speed", Type: PropertyTypeInteger, Required: false, Description: "Memory speed in MHz"},
				{Name: "type", Type: PropertyTypeString, Required: false, Description: "Memory type (ddr4, ddr5, lpddr4, lpddr5)"},
			}},
			{Name: "storage", Type: PropertyTypeList, Required: false, Description: "Virtual disk configurations", Properties: []PropertyDefinition{
				{Name: "size_gb", Type: PropertyTypeNumber, Required: false, Description: "Disk size in GB"},
				{Name: "type", Type: PropertyTypeString, Required: false, Description: "Disk type (ssd, hdd, nvme)"},
			}},
			{Name: "os", Type: PropertyTypeString, Required: false, Description: "Operating system"},
			{Name: "os_version", Type: PropertyTypeString, Required: false, Description: "Operating system version"},
		},
		NestingDefs: []NestingDefinition{
			{NestKey: "applications", ChildKind: kinds.Application, AutoRelationType: types.Hosts, AutoRelationSource: "parent"},
			{NestKey: "containers", ChildKind: kinds.Container, AutoRelationType: types.Hosts, AutoRelationSource: "parent"},
		},
	})

	s.AddEntityKind(kinds.Container, &EntityKindDefinition{
		Description: "Containerized workload",
		Properties: []PropertyDefinition{
			{Name: "image", Type: PropertyTypeString, Required: false, Description: "Container image"},
			{Name: "image_tag", Type: PropertyTypeString, Required: false, Default: "latest", Description: "Image tag"},
			{Name: "cpu_limit", Type: PropertyTypeString, Required: false, Description: "CPU limit"},
			{Name: "memory_limit", Type: PropertyTypeString, Required: false, Description: "Memory limit"},
			{Name: "ports", Type: PropertyTypeList, Required: false, Description: "Exposed ports"},
		},
		NestingDefs: []NestingDefinition{
			{NestKey: "applications", ChildKind: kinds.Application, AutoRelationType: types.Hosts, AutoRelationSource: "parent"},
		},
	})

	s.AddEntityKind(kinds.Application, &EntityKindDefinition{
		Description: "Software application or service",
		Properties: []PropertyDefinition{
			{Name: "version", Type: PropertyTypeString, Required: false, Description: "Application version"},
			{Name: "port", Type: PropertyTypeInteger, Required: false, Description: "Primary listening port"},
			{Name: "protocol", Type: PropertyTypeString, Required: false, Constraints: &Constraint{Enum: []string{"http", "https", "tcp", "udp"}}, Description: "Network protocol (http, https, tcp, udp)"},
			{Name: "url", Type: PropertyTypeString, Required: false, Description: "Application URL if applicable"},
		},
		NestingDefs: []NestingDefinition{
			{NestKey: "containers", ChildKind: kinds.Container, AutoRelationType: types.Hosts, AutoRelationSource: "parent"},
			{NestKey: "open_ports", ChildKind: kinds.OpenPort, AutoRelationType: types.BelongsTo, AutoRelationSource: "child"},
		},
	})

	s.AddEntityKind(kinds.OpenPort, &EntityKindDefinition{
		Description: "Listening or open network port",
		Properties: []PropertyDefinition{
			{Name: "port", Type: PropertyTypeInteger, Required: true, Constraints: &Constraint{Min: intPtr(1), Max: intPtr(65535)}, Description: "Port number (1-65535)"},
			{Name: "protocol", Type: PropertyTypeString, Required: true, Constraints: &Constraint{Enum: []string{"tcp", "udp"}}, Description: "Transport protocol (tcp, udp)"},
			{Name: "state", Type: PropertyTypeString, Required: false, Default: "listening", Constraints: &Constraint{Enum: []string{"listening", "established", "closed"}}, Description: "Port state (listening, established, closed)"},
			{Name: "address", Type: PropertyTypeString, Required: false, Default: "0.0.0.0", Description: "Listening IP address"},
			{Name: "process", Type: PropertyTypeString, Required: false, Description: "Process or service name using this port"},
			{Name: "pid", Type: PropertyTypeInteger, Required: false, Description: "Process ID if known"},
		},
	})

	s.AddEntityKind(kinds.Storage, &EntityKindDefinition{
		Description: "Storage system or array",
		Properties: []PropertyDefinition{
			{Name: "manufacturer", Type: PropertyTypeString, Required: false, Description: "Hardware manufacturer"},
			{Name: "model", Type: PropertyTypeString, Required: false, Description: "Hardware model"},
			{Name: "total_capacity_gb", Type: PropertyTypeNumber, Required: false, Description: "Total raw capacity in GB"},
			{Name: "usable_capacity_gb", Type: PropertyTypeNumber, Required: false, Description: "Usable capacity after redundancy"},
			{Name: "raid_level", Type: PropertyTypeString, Required: false, Description: "RAID level if applicable"},
			{Name: "protocol", Type: PropertyTypeString, Required: false, Constraints: &Constraint{Enum: []string{"nfs", "iscsi", "fc", "local"}}, Description: "Storage protocol (nfs, iscsi, fc, local)"},
		},
	})

	s.AddEntityKind(kinds.Volume, &EntityKindDefinition{
		Description: "Logical storage volume",
		Properties: []PropertyDefinition{
			{Name: "capacity_gb", Type: PropertyTypeNumber, Required: false, Description: "Volume capacity in GB"},
			{Name: "filesystem", Type: PropertyTypeString, Required: false, Description: "Filesystem type if mounted"},
			{Name: "mount_point", Type: PropertyTypeString, Required: false, Description: "Mount point if applicable"},
			{Name: "thin_provisioned", Type: PropertyTypeBoolean, Required: false, Default: false, Description: "Whether volume is thin provisioned"},
		},
	})

	s.AddEntityKind(kinds.Cluster, &EntityKindDefinition{
		Description: "Logical grouping of compute resources",
		Properties: []PropertyDefinition{
			{Name: "cluster_type", Type: PropertyTypeString, Required: false, Constraints: &Constraint{Enum: []string{"compute", "storage", "hyperconverged"}}, Description: "Cluster type (compute, storage, hyperconverged)"},
			{Name: "ha_enabled", Type: PropertyTypeBoolean, Required: false, Default: false, Description: "Whether HA is enabled"},
			{Name: "drs_enabled", Type: PropertyTypeBoolean, Required: false, Default: false, Description: "Whether DRS is enabled"},
		},
		NestingDefs: []NestingDefinition{
			{NestKey: "vms", ChildKind: kinds.VM, AutoRelationType: types.BelongsTo, AutoRelationSource: "child"},
			{NestKey: "servers", ChildKind: kinds.Server, AutoRelationType: types.BelongsTo, AutoRelationSource: "child"},
		},
	})

	s.AddEntityKind(kinds.AvailabilityZone, &EntityKindDefinition{
		Description: "Logical availability zone within a region",
		Properties: []PropertyDefinition{
			{Name: "state", Type: PropertyTypeString, Required: false, Default: "available", Constraints: &Constraint{Enum: []string{"available", "impaired", "unavailable"}}, Description: "Availability zone state (available, impaired, unavailable)"},
		},
	})
}

func registerRelationTypes(s *Schema) {
	s.AddRelationType(types.Connects, &RelationTypeDefinition{
		Direction:   DirectionSymmetric,
		Description: "Physical or logical connection between entities",
		Participants: &ParticipantConstraints{
			SourceKinds:     []core.EntityKind{kinds.Interface},
			TargetKinds:     []core.EntityKind{kinds.Interface},
			MinParticipants: 2,
			MaxParticipants: 0, // 0 means unlimited
		},
		Properties: []PropertyDefinition{
			{Name: "connection_type", Type: PropertyTypeString, Required: false, Description: "Type of connection (physical, logical, virtual)"},
			{Name: "bandwidth_mbps", Type: PropertyTypeInteger, Required: false, Description: "Connection bandwidth in Mbps"},
		},
	})

	s.AddRelationType(types.Hosts, &RelationTypeDefinition{
		Direction:   DirectionDirected,
		Description: "Execution hosting relationship",
		Participants: &ParticipantConstraints{
			SourceKinds:     []core.EntityKind{kinds.Server, kinds.VM, kinds.Container, kinds.Application},
			TargetKinds:     []core.EntityKind{kinds.VM, kinds.Container, kinds.Application},
			MinParticipants: 2,
			MaxParticipants: 2,
		},
	})

	s.AddRelationType(types.DependsOn, &RelationTypeDefinition{
		Direction:   DirectionDirected,
		Description: "Directional dependency between entities",
		Participants: &ParticipantConstraints{
			SourceKinds:     []core.EntityKind{kinds.VM, kinds.Container, kinds.Application},
			TargetKinds:     []core.EntityKind{kinds.VM, kinds.Container, kinds.Application, kinds.Storage, kinds.Network},
			MinParticipants: 2,
			MaxParticipants: 2,
		},
		Properties: []PropertyDefinition{
			{Name: "dependency_type", Type: PropertyTypeString, Required: false, Description: "Type of dependency (runtime, build, network, storage)"},
			{Name: "critical", Type: PropertyTypeBoolean, Required: false, Default: false, Description: "Whether failure causes cascading failure"},
		},
	})

	s.AddRelationType(types.BelongsTo, &RelationTypeDefinition{
		Direction:   DirectionDirected,
		Description: "Logical membership or association",
		Participants: &ParticipantConstraints{
			SourceKinds: []core.EntityKind{
				kinds.VM, kinds.Container, kinds.Interface, kinds.Server, kinds.Switch,
				kinds.Router, kinds.Firewall, kinds.Storage, kinds.ACL, kinds.ACLRule, kinds.OpenPort, kinds.AvailabilityZone,
			},
			TargetKinds: []core.EntityKind{
				kinds.Cluster, kinds.Network, kinds.Region, kinds.Firewall, kinds.Interface,
				kinds.Server, kinds.VM, kinds.Container, kinds.Application,
			},
			MinParticipants: 2,
			MaxParticipants: 2,
		},
	})

	s.AddRelationType(types.ReplicatesTo, &RelationTypeDefinition{
		Direction:   DirectionDirected,
		Description: "Data replication between entities",
		Participants: &ParticipantConstraints{
			SourceKinds:     []core.EntityKind{kinds.VM, kinds.Container, kinds.Application, kinds.Storage},
			TargetKinds:     []core.EntityKind{kinds.VM, kinds.Container, kinds.Application, kinds.Storage},
			MinParticipants: 2,
			MaxParticipants: 2,
		},
		Properties: []PropertyDefinition{
			{Name: "replication_type", Type: PropertyTypeString, Required: false, Default: "synchronous", Description: "Replication type (synchronous, asynchronous)"},
			{Name: "lag_seconds", Type: PropertyTypeNumber, Required: false, Description: "Replication lag in seconds"},
		},
	})

	s.AddRelationType(types.BacksUp, &RelationTypeDefinition{
		Direction:   DirectionDirected,
		Description: "Backup relationship",
		Participants: &ParticipantConstraints{
			SourceKinds:     []core.EntityKind{kinds.VM, kinds.Container, kinds.Application, kinds.Storage, kinds.Volume},
			TargetKinds:     []core.EntityKind{kinds.Volume, kinds.Storage},
			MinParticipants: 2,
			MaxParticipants: 2,
		},
		Properties: []PropertyDefinition{
			{Name: "backup_type", Type: PropertyTypeString, Required: false, Description: "Backup type (full, incremental, differential)"},
			{Name: "schedule", Type: PropertyTypeString, Required: false, Description: "Backup schedule (cron expression)"},
			{Name: "retention_days", Type: PropertyTypeInteger, Required: false, Description: "Backup retention in days"},
		},
	})

	s.AddRelationType(types.Monitors, &RelationTypeDefinition{
		Direction:   DirectionDirected,
		Description: "Monitoring relationship",
		Participants: &ParticipantConstraints{
			SourceKinds:     []core.EntityKind{kinds.Server, kinds.VM, kinds.Container, kinds.Application},
			TargetKinds:     []core.EntityKind{kinds.Server, kinds.VM, kinds.Container, kinds.Application, kinds.Storage},
			MinParticipants: 2,
			MaxParticipants: 2,
		},
		Properties: []PropertyDefinition{
			{Name: "monitor_type", Type: PropertyTypeString, Required: false, Description: "Monitor type (agent, agentless, snmp, api)"},
			{Name: "interval_seconds", Type: PropertyTypeInteger, Required: false, Default: 60, Description: "Monitoring interval in seconds"},
		},
	})

	s.AddRelationType(types.ManagedBy, &RelationTypeDefinition{
		Direction:   DirectionDirected,
		Description: "Management relationship",
		Participants: &ParticipantConstraints{
			SourceKinds: []core.EntityKind{
				kinds.Server, kinds.VM, kinds.Container, kinds.Application,
				kinds.Switch, kinds.Router, kinds.Firewall, kinds.Storage,
			},
			TargetKinds: []core.EntityKind{
				kinds.Server, kinds.VM, kinds.Container, kinds.Application,
			},
			MinParticipants: 2,
			MaxParticipants: 2,
		},
		Properties: []PropertyDefinition{
			{Name: "management_type", Type: PropertyTypeString, Required: false, Description: "Management type (configuration, orchestration, monitoring)"},
		},
	})

	s.AddRelationType(types.MountedOn, &RelationTypeDefinition{
		Direction:   DirectionDirected,
		Description: "Storage mounting relationship",
		Participants: &ParticipantConstraints{
			SourceKinds:     []core.EntityKind{kinds.Volume},
			TargetKinds:     []core.EntityKind{kinds.Server, kinds.VM, kinds.Container, kinds.Storage},
			MinParticipants: 2,
			MaxParticipants: 2,
		},
		Properties: []PropertyDefinition{
			{Name: "mount_point", Type: PropertyTypeString, Required: false, Description: "Mount point path"},
			{Name: "filesystem", Type: PropertyTypeString, Required: false, Description: "Filesystem type"},
			{Name: "options", Type: PropertyTypeString, Required: false, Description: "Mount options"},
		},
	})

	s.AddRelationType(types.AppliesTo, &RelationTypeDefinition{
		Direction:   DirectionDirected,
		Description: "ACL applied to a network target",
		Participants: &ParticipantConstraints{
			SourceKinds:     []core.EntityKind{kinds.ACL},
			TargetKinds:     []core.EntityKind{kinds.Interface, kinds.Firewall, kinds.Server, kinds.VM, kinds.Container},
			MinParticipants: 2,
			MaxParticipants: 2,
		},
	})

	s.AddRelationType(types.ListensOn, &RelationTypeDefinition{
		Direction:   DirectionDirected,
		Description: "Open port listening on a network interface or host",
		Participants: &ParticipantConstraints{
			SourceKinds:     []core.EntityKind{kinds.OpenPort},
			TargetKinds:     []core.EntityKind{kinds.Interface, kinds.Server, kinds.VM, kinds.Container},
			MinParticipants: 2,
			MaxParticipants: 2,
		},
	})
}
