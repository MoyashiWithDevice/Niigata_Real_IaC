package kinds

import (
	"testing"

	"IACForge/src/core"
)

func TestIsValidStatus(t *testing.T) {
	tests := []struct {
		status core.Status
		valid  bool
	}{
		{core.StatusPlanned, true},
		{core.StatusActive, true},
		{core.StatusMaintenance, true},
		{core.StatusDeprecated, true},
		{core.StatusOffline, true},
		{core.StatusStandby, true},
		{"invalid", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(string(tt.status), func(t *testing.T) {
			if got := IsValidStatus(tt.status); got != tt.valid {
				t.Errorf("IsValidStatus(%s) = %v, want %v", tt.status, got, tt.valid)
			}
		})
	}
}

func TestKindValues(t *testing.T) {
	expected := map[core.EntityKind]string{
		Region:            "region",
		Rack:              "rack",
		Server:            "server",
		Interface:         "interface",
		Cable:             "cable",
		PowerDistribution: "power_distribution",
		Network:           "network",
		VLAN:              "vlan",
		Switch:            "switch",
		Router:            "router",
		Firewall:          "firewall",
		ACL:               "acl",
		ACLRule:           "acl_rule",
		VM:                "vm",
		Container:         "container",
		Application:       "application",
		OpenPort:          "open_port",
		Storage:           "storage",
		Volume:            "volume",
		Cluster:           "cluster",
		AvailabilityZone:  "availability_zone",
	}

	for kind, value := range expected {
		if string(kind) != value {
			t.Errorf("kind %v has wrong string value: got %s, want %s", kind, string(kind), value)
		}
	}
}
