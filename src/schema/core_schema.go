package schema

import (
	"github.com/bababa/Niigata_Real_IaC/src/core"
	"github.com/bababa/Niigata_Real_IaC/src/core/kinds"
	"github.com/bababa/Niigata_Real_IaC/src/core/types"
)

func intPtr(v float64) *float64 { return &v }

// CoreSchema returns the default core schema with all entity kinds and relation types
// for Niigata nature and tourism resource management.
func CoreSchema() *Schema {
	s := NewSchema("1.0.0", "1.0")
	s.Version.Description = "Niigata Nature and Tourism Resource Schema"

	registerEntityKinds(s)
	registerRelationTypes(s)

	return s
}

// geoProperties returns the optional geographic coordinate properties shared by
// place-bearing entity kinds. Bounds follow WGS84 decimal degrees.
func geoProperties() []PropertyDefinition {
	return []PropertyDefinition{
		{Name: "latitude", Type: PropertyTypeNumber, Required: false,
			Constraints: &Constraint{Min: intPtr(-90), Max: intPtr(90)},
			Description: "Geographic latitude in decimal degrees"},
		{Name: "longitude", Type: PropertyTypeNumber, Required: false,
			Constraints: &Constraint{Min: intPtr(-180), Max: intPtr(180)},
			Description: "Geographic longitude in decimal degrees"},
	}
}

func registerEntityKinds(s *Schema) {
	s.AddEntityKind(kinds.Area, &EntityKindDefinition{
		Description: "Geographic area such as a city, town, village, district, or island",
		Properties: []PropertyDefinition{
			{Name: "area_type", Type: PropertyTypeString, Required: false,
				Constraints: &Constraint{Enum: []string{"prefecture", "city", "town", "village", "district", "island"}},
				Description: "Administrative type of the area"},
			{Name: "population", Type: PropertyTypeInteger, Required: false,
				Constraints: &Constraint{Min: intPtr(0)},
				Description: "Number of residents"},
			{Name: "address", Type: PropertyTypeString, Required: false, Description: "Physical address"},
			{Name: "latitude", Type: PropertyTypeNumber, Required: false,
				Constraints: &Constraint{Min: intPtr(-90), Max: intPtr(90)},
				Description: "Geographic latitude in decimal degrees"},
			{Name: "longitude", Type: PropertyTypeNumber, Required: false,
				Constraints: &Constraint{Min: intPtr(-180), Max: intPtr(180)},
				Description: "Geographic longitude in decimal degrees"},
			{Name: "timezone", Type: PropertyTypeString, Required: false, Description: "Timezone identifier"},
		},
		NestingDefs: []NestingDefinition{
			{NestKey: "grounds", ChildKind: kinds.Ground, AutoRelationType: types.BelongsTo, AutoRelationSource: "child"},
			{NestKey: "terrains", ChildKind: kinds.Terrain, AutoRelationType: types.BelongsTo, AutoRelationSource: "child"},
			{NestKey: "water_bodies", ChildKind: kinds.WaterBody, AutoRelationType: types.BelongsTo, AutoRelationSource: "child"},
			{NestKey: "forests", ChildKind: kinds.Forest, AutoRelationType: types.BelongsTo, AutoRelationSource: "child"},
			{NestKey: "tourism_spots", ChildKind: kinds.TourismSpot, AutoRelationType: types.BelongsTo, AutoRelationSource: "child"},
			{NestKey: "species", ChildKind: kinds.Species, AutoRelationType: types.BelongsTo, AutoRelationSource: "child"},
			{NestKey: "cultural_assets", ChildKind: kinds.CulturalAsset, AutoRelationType: types.BelongsTo, AutoRelationSource: "child"},
		},
	})

	s.AddEntityKind(kinds.Ground, &EntityKindDefinition{
		Description: "Ground or soil condition at a specific site",
		Properties: append(
			[]PropertyDefinition{
				{Name: "soil_type", Type: PropertyTypeString, Required: false,
					Constraints: &Constraint{Enum: []string{"loam", "sand", "clay", "gravel", "rock", "volcanic_ash", "peat"}},
					Description: "Dominant soil type"},
				{Name: "elevation_m", Type: PropertyTypeNumber, Required: false,
					Constraints: &Constraint{Min: intPtr(-10)}, Description: "Elevation above sea level in meters"},
				{Name: "slope_deg", Type: PropertyTypeNumber, Required: false,
					Constraints: &Constraint{Min: intPtr(0), Max: intPtr(90)}, Description: "Slope angle in degrees"},
				{Name: "stability", Type: PropertyTypeString, Required: false,
					Constraints: &Constraint{Enum: []string{"stable", "watch", "unstable", "landslide_prone", "subsidence"}},
					Description: "Ground stability classification"},
			},
			geoProperties()...),
	})

	s.AddEntityKind(kinds.Terrain, &EntityKindDefinition{
		Description: "Landform such as a mountain, plain, coast, or valley",
		Properties: append(
			[]PropertyDefinition{
				{Name: "terrain_type", Type: PropertyTypeString, Required: false,
					Constraints: &Constraint{Enum: []string{"mountain", "hill", "plain", "coast", "valley", "plateau", "cave"}},
					Description: "Type of landform"},
				{Name: "elevation_m", Type: PropertyTypeNumber, Required: false,
					Constraints: &Constraint{Min: intPtr(0)}, Description: "Highest elevation in meters"},
				{Name: "prominence_m", Type: PropertyTypeNumber, Required: false,
					Constraints: &Constraint{Min: intPtr(0)}, Description: "Topographic prominence in meters"},
			},
			geoProperties()...),
		NestingDefs: []NestingDefinition{
			{NestKey: "grounds", ChildKind: kinds.Ground, AutoRelationType: types.BelongsTo, AutoRelationSource: "child"},
		},
	})

	s.AddEntityKind(kinds.WaterBody, &EntityKindDefinition{
		Description: "River, lake, sea area, pond, or marsh",
		Properties: append(
			[]PropertyDefinition{
				{Name: "water_type", Type: PropertyTypeString, Required: false,
					Constraints: &Constraint{Enum: []string{"river", "lake", "sea", "pond", "marsh", "waterfall"}},
					Description: "Type of water body"},
				{Name: "length_km", Type: PropertyTypeNumber, Required: false,
					Constraints: &Constraint{Min: intPtr(0)}, Description: "Length in kilometers (rivers)"},
				{Name: "max_depth_m", Type: PropertyTypeNumber, Required: false,
					Constraints: &Constraint{Min: intPtr(0)}, Description: "Maximum depth in meters"},
				{Name: "catchment_area_km2", Type: PropertyTypeNumber, Required: false,
					Constraints: &Constraint{Min: intPtr(0)}, Description: "Catchment area in square kilometers"},
			},
			geoProperties()...),
	})

	s.AddEntityKind(kinds.Forest, &EntityKindDefinition{
		Description: "Forest or wooded area",
		Properties: append(
			[]PropertyDefinition{
				{Name: "forest_type", Type: PropertyTypeString, Required: false,
					Constraints: &Constraint{Enum: []string{"natural", "beech", "cedar_plantation", "pine", "bamboo", "mixed"}},
					Description: "Dominant forest type"},
				{Name: "area_ha", Type: PropertyTypeNumber, Required: false,
					Constraints: &Constraint{Min: intPtr(0)}, Description: "Area in hectares"},
			},
			geoProperties()...),
		NestingDefs: []NestingDefinition{
			{NestKey: "grounds", ChildKind: kinds.Ground, AutoRelationType: types.BelongsTo, AutoRelationSource: "child"},
		},
	})

	s.AddEntityKind(kinds.Species, &EntityKindDefinition{
		Description: "Animal or plant species living in the area",
		Properties: []PropertyDefinition{
			{Name: "scientific_name", Type: PropertyTypeString, Required: false, Description: "Scientific (Latin) name"},
			{Name: "category", Type: PropertyTypeString, Required: false,
				Constraints: &Constraint{Enum: []string{"mammal", "bird", "reptile", "amphibian", "fish", "insect", "plant", "other"}},
				Description: "Biological category"},
			{Name: "red_list_status", Type: PropertyTypeString, Required: false,
				Constraints: &Constraint{Enum: []string{"extinct", "extinct_in_wild", "critically_endangered", "endangered", "vulnerable", "near_threatened", "least_concern", "data_deficient"}},
				Description: "Red List conservation status"},
		},
		NestingDefs: []NestingDefinition{
			{NestKey: "populations", ChildKind: kinds.Population, AutoRelationType: types.BelongsTo, AutoRelationSource: "child"},
		},
	})

	s.AddEntityKind(kinds.Population, &EntityKindDefinition{
		Description: "Population record of a species from a survey",
		Properties: []PropertyDefinition{
			{Name: "count", Type: PropertyTypeInteger, Required: true,
				Constraints: &Constraint{Min: intPtr(0)}, Description: "Number of individuals observed or estimated"},
			{Name: "survey_date", Type: PropertyTypeString, Required: false, Description: "Survey date (YYYY-MM-DD)"},
			{Name: "survey_method", Type: PropertyTypeString, Required: false,
				Constraints: &Constraint{Enum: []string{"visual_count", "transect", "drone", "camera_trap", "interview", "estimate"}},
				Description: "How the count was performed"},
		},
	})

	s.AddEntityKind(kinds.TourismSpot, &EntityKindDefinition{
		Description: "Tourist attraction such as a scenic spot, park, shrine, or museum",
		Properties: append(
			[]PropertyDefinition{
				{Name: "spot_type", Type: PropertyTypeString, Required: false,
					Constraints: &Constraint{Enum: []string{"scenic", "viewpoint", "park", "historic", "shrine_temple", "museum", "market", "ski_resort", "beach"}},
					Description: "Type of tourist spot"},
				{Name: "description", Type: PropertyTypeString, Required: false, Description: "Short description"},
				{Name: "annual_visitors", Type: PropertyTypeInteger, Required: false,
					Constraints: &Constraint{Min: intPtr(0)}, Description: "Annual number of visitors"},
			},
			geoProperties()...),
		NestingDefs: []NestingDefinition{
			{NestKey: "hot_springs", ChildKind: kinds.HotSpring, AutoRelationType: types.BelongsTo, AutoRelationSource: "child"},
			{NestKey: "events", ChildKind: kinds.Event, AutoRelationType: types.BelongsTo, AutoRelationSource: "child"},
		},
	})

	s.AddEntityKind(kinds.HotSpring, &EntityKindDefinition{
		Description: "Hot spring source or bath facility",
		Properties: append(
			[]PropertyDefinition{
				{Name: "spring_quality", Type: PropertyTypeString, Required: false,
					Constraints: &Constraint{Enum: []string{"sulfur", "chloride", "simple", "carbonated", "iron", "alum", "sulfate"}},
					Description: "Chemical quality of the spring"},
				{Name: "temperature_c", Type: PropertyTypeNumber, Required: false,
					Description: "Source temperature in degrees Celsius"},
				{Name: "source_count", Type: PropertyTypeInteger, Required: false,
					Constraints: &Constraint{Min: intPtr(0)}, Description: "Number of spring sources"},
			},
			geoProperties()...),
		NestingDefs: []NestingDefinition{
			{NestKey: "events", ChildKind: kinds.Event, AutoRelationType: types.BelongsTo, AutoRelationSource: "child"},
		},
	})

	s.AddEntityKind(kinds.CulturalAsset, &EntityKindDefinition{
		Description: "Historic site or cultural property",
		Properties: append(
			[]PropertyDefinition{
				{Name: "asset_type", Type: PropertyTypeString, Required: false,
					Constraints: &Constraint{Enum: []string{"historic_site", "treasure", "building", "monument", "archaeological", "folk_property"}},
					Description: "Type of cultural asset"},
				{Name: "designated_level", Type: PropertyTypeString, Required: false,
					Constraints: &Constraint{Enum: []string{"national", "prefectural", "municipal", "unesco"}},
					Description: "Designation level"},
				{Name: "designated_date", Type: PropertyTypeString, Required: false, Description: "Designation date (YYYY-MM-DD)"},
			},
			geoProperties()...),
	})

	s.AddEntityKind(kinds.Event, &EntityKindDefinition{
		Description: "Festival or recurring local event",
		Properties: []PropertyDefinition{
			{Name: "season", Type: PropertyTypeString, Required: false,
				Constraints: &Constraint{Enum: []string{"spring", "summer", "autumn", "winter", "all"}},
				Description: "Season when the event is held"},
			{Name: "held_month", Type: PropertyTypeInteger, Required: false,
				Constraints: &Constraint{Min: intPtr(1), Max: intPtr(12)}, Description: "Month when the event is held (1-12)"},
			{Name: "visitor_count", Type: PropertyTypeInteger, Required: false,
				Constraints: &Constraint{Min: intPtr(0)}, Description: "Number of visitors per holding"},
		},
	})
}

func registerRelationTypes(s *Schema) {
	natureKinds := []core.EntityKind{kinds.Ground, kinds.Terrain, kinds.WaterBody, kinds.Forest}
	tourismKinds := []core.EntityKind{kinds.TourismSpot, kinds.HotSpring, kinds.CulturalAsset, kinds.Event}

	s.AddRelationType(types.LocatedIn, &RelationTypeDefinition{
		Direction:   DirectionDirected,
		Description: "Entity is located within an area or terrain",
		Participants: &ParticipantConstraints{
			SourceKinds:     append(append([]core.EntityKind{}, natureKinds...), append(tourismKinds, kinds.Species, kinds.Population)...),
			TargetKinds:     []core.EntityKind{kinds.Area, kinds.Terrain},
			MinParticipants: 2,
			MaxParticipants: 2,
		},
		Properties: []PropertyDefinition{
			{Name: "distance_km", Type: PropertyTypeNumber, Required: false,
				Constraints: &Constraint{Min: intPtr(0)}, Description: "Distance from the target in kilometers"},
		},
	})

	s.AddRelationType(types.Inhabits, &RelationTypeDefinition{
		Direction:   DirectionDirected,
		Description: "Species lives in a habitat",
		Participants: &ParticipantConstraints{
			SourceKinds:     []core.EntityKind{kinds.Species, kinds.Population},
			TargetKinds:     append(append([]core.EntityKind{}, natureKinds...), kinds.Area),
			MinParticipants: 2,
			MaxParticipants: 2,
		},
		Properties: []PropertyDefinition{
			{Name: "habitat_note", Type: PropertyTypeString, Required: false, Description: "Notes about the habitat use"},
		},
	})

	s.AddRelationType(types.Near, &RelationTypeDefinition{
		Direction:   DirectionSymmetric,
		Description: "Two resources are geographically close to each other",
		Participants: &ParticipantConstraints{
			SourceKinds:     append(append([]core.EntityKind{}, tourismKinds...), append(natureKinds, kinds.Area)...),
			TargetKinds:     append(append([]core.EntityKind{}, tourismKinds...), append(natureKinds, kinds.Area)...),
			MinParticipants: 2,
			MaxParticipants: 2,
		},
		Properties: []PropertyDefinition{
			{Name: "walking_minutes", Type: PropertyTypeInteger, Required: false,
				Constraints: &Constraint{Min: intPtr(0)}, Description: "Walking time between the two in minutes"},
		},
	})

	s.AddRelationType(types.DependsOn, &RelationTypeDefinition{
		Direction:   DirectionDirected,
		Description: "Directional dependency, e.g. a hot spring depends on its ground source",
		Participants: &ParticipantConstraints{
			SourceKinds:     append([]core.EntityKind{kinds.HotSpring}, tourismKinds...),
			TargetKinds:     append(append([]core.EntityKind{}, natureKinds...), tourismKinds...),
			MinParticipants: 2,
			MaxParticipants: 2,
		},
		Properties: []PropertyDefinition{
			{Name: "dependency_type", Type: PropertyTypeString, Required: false,
				Constraints: &Constraint{Enum: []string{"source", "landscape", "access", "ecosystem", "event"}},
				Description: "Nature of the dependency"},
			{Name: "critical", Type: PropertyTypeBoolean, Required: false, Default: false,
				Description: "Whether loss of the target destroys the source's value"},
		},
	})

	s.AddRelationType(types.BelongsTo, &RelationTypeDefinition{
		Direction:   DirectionDirected,
		Description: "Logical membership or association",
		Participants: &ParticipantConstraints{
			SourceKinds:     append(append([]core.EntityKind{}, natureKinds...), append(tourismKinds, kinds.Species, kinds.Population)...),
			TargetKinds:     append([]core.EntityKind{kinds.Area}, append(natureKinds, tourismKinds...)...),
			MinParticipants: 2,
			MaxParticipants: 2,
		},
	})

	s.AddRelationType(types.FlowsInto, &RelationTypeDefinition{
		Direction:   DirectionDirected,
		Description: "River or water flow destination",
		Participants: &ParticipantConstraints{
			SourceKinds:     []core.EntityKind{kinds.WaterBody},
			TargetKinds:     []core.EntityKind{kinds.WaterBody},
			MinParticipants: 2,
			MaxParticipants: 2,
		},
	})
}
