package types

import (
	"testing"

	"IACForge/src/core"
)

func TestRelationTypeValues(t *testing.T) {
	expected := map[core.RelationType]string{
		Connects:     "connects",
		Hosts:        "hosts",
		DependsOn:    "depends_on",
		BelongsTo:    "belongs_to",
		ReplicatesTo: "replicates_to",
		BacksUp:      "backs_up",
		Monitors:     "monitors",
		ManagedBy:    "managed_by",
		MountedOn:    "mounted_on",
		AppliesTo:    "applies_to",
		ListensOn:    "listens_on",
	}

	for typ, value := range expected {
		if string(typ) != value {
			t.Errorf("type %v has wrong string value: got %s, want %s", typ, string(typ), value)
		}
	}
}
