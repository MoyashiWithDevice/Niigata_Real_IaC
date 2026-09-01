package kinds

import "github.com/bababa/Niigata_Real_IaC/src/core"

// Entity kinds for Niigata nature and tourism resource management.
const (
	Area         core.EntityKind = "area"
	Ground       core.EntityKind = "ground"
	Terrain      core.EntityKind = "terrain"
	WaterBody    core.EntityKind = "water_body"
	Forest       core.EntityKind = "forest"
	Species      core.EntityKind = "species"
	Population   core.EntityKind = "population"
	TourismSpot  core.EntityKind = "tourism_spot"
	HotSpring    core.EntityKind = "hot_spring"
	CulturalAsset core.EntityKind = "cultural_asset"
	Event        core.EntityKind = "event"
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
