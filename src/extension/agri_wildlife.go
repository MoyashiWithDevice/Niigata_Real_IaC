package extension

import (
	"fmt"

	"github.com/bababa/Niigata_Real_IaC/src/core"
	"github.com/bababa/Niigata_Real_IaC/src/core/kinds"
	"github.com/bababa/Niigata_Real_IaC/src/core/types"
	"github.com/bababa/Niigata_Real_IaC/src/schema"
	"github.com/bababa/Niigata_Real_IaC/src/validation"
)

// Entity kinds contributed by the agri-wildlife builtin extension.
const (
	// KindWildlifeIncident records wildlife sighting / damage incidents
	// (e.g. bear encounters) observed in an area.
	KindWildlifeIncident core.EntityKind = "wildlife_incident"
	// KindCropHarvest records the annual harvest volume of a crop
	// (e.g. rice) produced in an area.
	KindCropHarvest core.EntityKind = "crop_harvest"
)

func init() {
	RegisterBuiltin(AgriWildlifeExtension())
}

// AgriWildlifeExtension returns the agri-wildlife builtin extension, which
// contributes the wildlife_incident and crop_harvest entity kinds, the
// agri-year-plausible validation rule, and area nesting for both kinds.
func AgriWildlifeExtension() *Extension {
	return &Extension{
		Manifest: &Manifest{
			ID:              "niigata.agri-wildlife",
			Name:            "Niigata Agriculture & Wildlife Extension",
			Version:         "0.1.0",
			Description:     "Wildlife incident (e.g. bear sightings) and crop harvest (e.g. rice yield) records owned by areas",
			Namespace:       "niigata.agri",
			ExtensionPoints: []string{string(ExtensionPointEntityKinds), string(ExtensionPointRelationTypes), string(ExtensionPointValidationRules), string(ExtensionPointRootKinds)},
		},
		EntityKinds:     agriWildlifeEntityKinds(),
		RelationTypes:   agriWildlifeRelationTypes(),
		ValidationRules: agriWildlifeValidationRules(),
		// Areas are natural roots for municipality-level records; granting
		// them root authority allows models that span multiple areas
		// (e.g. merged per-municipality files) to validate.
		RootKinds: []core.EntityKind{kinds.Area},
	}
}

// floatPtr returns a pointer to the given float64 value (constraint helper).
func floatPtr(v float64) *float64 { return &v }

// agriWildlifeGeoProperties mirrors the core geo coordinate properties so
// place-bearing extension kinds share the same bounds and semantics.
func agriWildlifeGeoProperties() []schema.PropertyDefinition {
	return []schema.PropertyDefinition{
		{Name: "latitude", Type: schema.PropertyTypeNumber, Required: false,
			Constraints: &schema.Constraint{Min: floatPtr(-90), Max: floatPtr(90)},
			Description: "Geographic latitude in decimal degrees"},
		{Name: "longitude", Type: schema.PropertyTypeNumber, Required: false,
			Constraints: &schema.Constraint{Min: floatPtr(-180), Max: floatPtr(180)},
			Description: "Geographic longitude in decimal degrees"},
	}
}

func agriWildlifeEntityKinds() []EntityKindContribution {
	return []EntityKindContribution{
		{
			Kind: KindWildlifeIncident,
			Definition: &schema.EntityKindDefinition{
				Description: "Wildlife incident record such as a bear sighting or damage report in an area",
				Properties: append(
					[]schema.PropertyDefinition{
						{Name: "count", Type: schema.PropertyTypeInteger, Required: true,
							Constraints: &schema.Constraint{Min: floatPtr(0)},
							Description: "Number of incidents in the period"},
						{Name: "incident_type", Type: schema.PropertyTypeString, Required: false,
							Constraints: &schema.Constraint{Enum: []string{
								"sighting", "crop_damage", "livestock_damage",
								"injury", "property_damage", "other",
							}},
							Description: "Most severe category of the incident"},
						{Name: "year", Type: schema.PropertyTypeInteger, Required: false,
							Constraints: &schema.Constraint{Min: floatPtr(1900), Max: floatPtr(2100)},
							Description: "Calendar year the incidents were recorded"},
						{Name: "date", Type: schema.PropertyTypeString, Required: false,
							Description: "First incident date (YYYY-MM-DD)"},
						{Name: "location", Type: schema.PropertyTypeString, Required: false,
							Description: "Free-form location description (district, landmark)"},
					},
					agriWildlifeGeoProperties()...),
			},
			ParentKinds: []core.EntityKind{kinds.Area},
			NestKey:     "wildlife_incidents",
		},
		{
			Kind: KindCropHarvest,
			Definition: &schema.EntityKindDefinition{
				Description: "Annual harvest record of a crop produced in an area",
				Properties: append(
					[]schema.PropertyDefinition{
						{Name: "crop_type", Type: schema.PropertyTypeString, Required: true,
							Constraints: &schema.Constraint{Enum: []string{
								"rice", "hay", "vegetable", "fruit", "flower", "other",
							}},
							Description: "Type of crop harvested"},
						{Name: "harvest_t", Type: schema.PropertyTypeNumber, Required: false,
							Constraints: &schema.Constraint{Min: floatPtr(0)},
							Description: "Harvest volume in metric tons"},
						{Name: "year", Type: schema.PropertyTypeInteger, Required: false,
							Constraints: &schema.Constraint{Min: floatPtr(1900), Max: floatPtr(2100)},
							Description: "Calendar year of the harvest"},
						{Name: "area_ha", Type: schema.PropertyTypeNumber, Required: false,
							Constraints: &schema.Constraint{Min: floatPtr(0)},
							Description: "Harvested area in hectares"},
					},
					agriWildlifeGeoProperties()...),
			},
			ParentKinds: []core.EntityKind{kinds.Area},
			NestKey:     "crop_harvests",
		},
	}
}

func agriWildlifeValidationRules() []ValidationRuleContribution {
	return []ValidationRuleContribution{
		{
			Rule: &validation.Rule{
				ID:       "agri-year-plausible",
				Name:     "Plausible Agri Record Year",
				Severity: validation.SeverityWarning,
			},
			Fn: ruleAgriYearPlausible,
		},
	}
}

// agriWildlifeRelationTypes augments the core belongs_to type so the nested
// auto-relations between areas and the contributed record kinds pass the
// valid-participant-kind rule.
func agriWildlifeRelationTypes() []RelationTypeContribution {
	return []RelationTypeContribution{
		{
			Type: types.BelongsTo,
			Definition: &schema.RelationTypeDefinition{
				Participants: &schema.ParticipantConstraints{
					SourceKinds: []core.EntityKind{KindWildlifeIncident, KindCropHarvest},
				},
			},
			Augment: true,
		},
	}
}

// ruleAgriYearPlausible warns when a wildlife_incident or crop_harvest record
// declares a year property outside the plausible reporting window.
func ruleAgriYearPlausible(ctx *validation.Context) []validation.Finding {
	g := ctx.Graph.(*core.Graph)
	var findings []validation.Finding

	for _, e := range g.Entities() {
		if e.Kind != KindWildlifeIncident && e.Kind != KindCropHarvest {
			continue
		}
		year, ok := e.Properties["year"]
		if !ok {
			continue
		}
		y, ok := toInt(year)
		if !ok {
			findings = append(findings, validation.Finding{
				Severity:   validation.SeverityWarning,
				Message:    fmt.Sprintf("%s %q property \"year\" is not an integer", e.Kind, e.ID),
				ObjectID:   e.ID,
				ObjectType: validation.ObjectTypeEntity,
			})
			continue
		}
		if y < 1900 || y > 2100 {
			findings = append(findings, validation.Finding{
				Severity:   validation.SeverityWarning,
				Message:    fmt.Sprintf("%s %q year %d is outside the plausible range (1900-2100)", e.Kind, e.ID, y),
				ObjectID:   e.ID,
				ObjectType: validation.ObjectTypeEntity,
			})
		}
	}
	return findings
}

// toInt converts common YAML integer representations to int64.
func toInt(v interface{}) (int64, bool) {
	switch n := v.(type) {
	case int:
		return int64(n), true
	case int64:
		return n, true
	case float64:
		return int64(n), true
	default:
		return 0, false
	}
}
