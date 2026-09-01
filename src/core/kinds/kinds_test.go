package kinds

import (
	"testing"

	"github.com/bababa/Niigata_Real_IaC/src/core"
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
		Area:          "area",
		Ground:        "ground",
		Terrain:       "terrain",
		WaterBody:     "water_body",
		Forest:        "forest",
		Species:       "species",
		Population:    "population",
		TourismSpot:   "tourism_spot",
		HotSpring:     "hot_spring",
		CulturalAsset: "cultural_asset",
		Event:         "event",
	}

	for kind, value := range expected {
		if string(kind) != value {
			t.Errorf("kind %v has wrong string value: got %s, want %s", kind, string(kind), value)
		}
	}
}
