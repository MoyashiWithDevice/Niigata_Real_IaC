package types

import (
	"testing"

	"github.com/bababa/Niigata_Real_IaC/src/core"
)

func TestRelationTypeValues(t *testing.T) {
	expected := map[core.RelationType]string{
		LocatedIn: "located_in",
		Inhabits:  "inhabits",
		Near:      "near",
		DependsOn: "depends_on",
		BelongsTo: "belongs_to",
		FlowsInto: "flows_into",
	}

	for typ, value := range expected {
		if string(typ) != value {
			t.Errorf("type %v has wrong string value: got %s, want %s", typ, string(typ), value)
		}
	}
}
