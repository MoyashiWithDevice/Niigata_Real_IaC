package schema

import (
	"fmt"
	"math"
	"regexp"
	"sort"

	"IACForge/src/core"
)

// PropertyType represents a property type in the schema.
type PropertyType string

const (
	PropertyTypeString    PropertyType = "string"
	PropertyTypeInteger   PropertyType = "integer"
	PropertyTypeNumber    PropertyType = "number"
	PropertyTypeBoolean   PropertyType = "boolean"
	PropertyTypeList      PropertyType = "list"
	PropertyTypeMap       PropertyType = "map"
	PropertyTypeReference PropertyType = "reference"
)

// Constraint represents a validation constraint on a property.
type Constraint struct {
	Min         *float64 `yaml:"min,omitempty"`
	Max         *float64 `yaml:"max,omitempty"`
	MinLength   *int     `yaml:"min_length,omitempty"`
	MaxLength   *int     `yaml:"max_length,omitempty"`
	Pattern     *string  `yaml:"pattern,omitempty"`
	Enum        []string `yaml:"enum,omitempty"`
	UniqueItems *bool    `yaml:"unique_items,omitempty"`
}

// PropertyDefinition defines a property within an entity kind or relation type.
type PropertyDefinition struct {
	Name          string               `yaml:"name"`
	Type          PropertyType         `yaml:"type"`
	Required      bool                 `yaml:"required"`
	Default       interface{}          `yaml:"default,omitempty"`
	Description   string               `yaml:"description,omitempty"`
	Constraints   *Constraint          `yaml:"constraints,omitempty"`
	Properties    []PropertyDefinition `yaml:"properties,omitempty"`
	ValueProperty *PropertyDefinition  `yaml:"value_property,omitempty"`
}

// DirectionType represents the directionality of a relation type.
type DirectionType string

const (
	DirectionDirected  DirectionType = "directed"
	DirectionSymmetric DirectionType = "symmetric"
)

// ParticipantConstraints defines which entity kinds can participate in a relation.
type ParticipantConstraints struct {
	SourceKinds     []core.EntityKind `yaml:"source_kinds,omitempty"`
	TargetKinds     []core.EntityKind `yaml:"target_kinds,omitempty"`
	MinParticipants int               `yaml:"min_participants,omitempty"`
	MaxParticipants int               `yaml:"max_participants,omitempty"`
}

// NestingDefinition defines a nestable child relationship for an entity kind.
type NestingDefinition struct {
	NestKey            string                     `yaml:"nest_key"`
	ChildKind          core.EntityKind            `yaml:"child_kind"`
	ChildKeys          map[string]core.EntityKind `yaml:"child_keys,omitempty"`
	AutoRelationType   core.RelationType          `yaml:"auto_relation_type,omitempty"`
	AutoRelationSource string                     `yaml:"auto_relation_source,omitempty"` // "parent" or "child"
}

// EntityKindDefinition defines an entity kind in the schema.
type EntityKindDefinition struct {
	Description string               `yaml:"description,omitempty"`
	Properties  []PropertyDefinition `yaml:"properties,omitempty"`
	NestingDefs []NestingDefinition  `yaml:"nesting_defs,omitempty"`
}

// RelationTypeDefinition defines a relation type in the schema.
type RelationTypeDefinition struct {
	Direction    DirectionType           `yaml:"direction"`
	Description  string                  `yaml:"description,omitempty"`
	Participants *ParticipantConstraints `yaml:"participants,omitempty"`
	Properties   []PropertyDefinition    `yaml:"properties,omitempty"`
}

// SchemaVersion holds version information for a schema.
type SchemaVersion struct {
	SchemaVersion string `yaml:"schema_version"`
	SpecVersion   string `yaml:"spec_version"`
	Description   string `yaml:"description,omitempty"`
}

// Schema defines the complete structure of an infrastructure model.
type Schema struct {
	Version       SchemaVersion                                 `yaml:"schema"`
	EntityKinds   map[core.EntityKind]*EntityKindDefinition     `yaml:"entity_kinds"`
	RelationTypes map[core.RelationType]*RelationTypeDefinition `yaml:"relation_types"`
	NestingDefs   []NestingDefinition                           `yaml:"nesting_defs,omitempty"`
	Profiles      []*Profile                                    `yaml:"profiles,omitempty"`
}

// NewSchema creates a new empty Schema with initialized maps.
func NewSchema(schemaVersion, specVersion string) *Schema {
	return &Schema{
		Version: SchemaVersion{
			SchemaVersion: schemaVersion,
			SpecVersion:   specVersion,
		},
		EntityKinds:   make(map[core.EntityKind]*EntityKindDefinition),
		RelationTypes: make(map[core.RelationType]*RelationTypeDefinition),
		NestingDefs:   make([]NestingDefinition, 0),
	}
}

// AddEntityKind registers an entity kind definition in the schema.
func (s *Schema) AddEntityKind(kind core.EntityKind, def *EntityKindDefinition) {
	s.EntityKinds[kind] = def
}

// AddRelationType registers a relation type definition in the schema.
func (s *Schema) AddRelationType(relType core.RelationType, def *RelationTypeDefinition) {
	s.RelationTypes[relType] = def
}

// HasEntityKind checks if the schema defines the given entity kind.
func (s *Schema) HasEntityKind(kind core.EntityKind) bool {
	_, ok := s.EntityKinds[kind]
	return ok
}

// HasRelationType checks if the schema defines the given relation type.
func (s *Schema) HasRelationType(relType core.RelationType) bool {
	_, ok := s.RelationTypes[relType]
	return ok
}

// GetEntityKindDef returns the definition for the given entity kind.
func (s *Schema) GetEntityKindDef(kind core.EntityKind) (*EntityKindDefinition, bool) {
	def, ok := s.EntityKinds[kind]
	return def, ok
}

// GetRelationTypeDef returns the definition for the given relation type.
func (s *Schema) GetRelationTypeDef(relType core.RelationType) (*RelationTypeDefinition, bool) {
	def, ok := s.RelationTypes[relType]
	return def, ok
}

// GetNestingDefs returns the nesting definitions for the given entity kind.
// Global nesting defs are merged with per-kind nesting defs.
// Per-kind defs take precedence over global defs when NestKey conflicts.
func (s *Schema) GetNestingDefs(kind core.EntityKind) []NestingDefinition {
	def, ok := s.EntityKinds[kind]
	if !ok {
		return s.NestingDefs
	}

	if len(s.NestingDefs) == 0 {
		return def.NestingDefs
	}
	if len(def.NestingDefs) == 0 {
		return s.NestingDefs
	}

	// Build merged list: per-kind overrides global by NestKey.
	// Per-kind defs are returned first so they take precedence for
	// consumers that select the first match (e.g., serializer nest key).
	result := make([]NestingDefinition, 0, len(s.NestingDefs)+len(def.NestingDefs))
	result = append(result, def.NestingDefs...)
	for _, nd := range s.NestingDefs {
		overridden := false
		for _, pnd := range def.NestingDefs {
			if pnd.NestKey == nd.NestKey {
				overridden = true
				break
			}
		}
		if !overridden {
			result = append(result, nd)
		}
	}
	return result
}

// FindNestingByNestKey finds the nesting definition for a given parent kind and nest key.
// Global nesting defs are checked first, then per-kind defs.
func (s *Schema) FindNestingByNestKey(parentKind core.EntityKind, nestKey string) (*NestingDefinition, bool) {
	defs := s.GetNestingDefs(parentKind)
	for i := range defs {
		if defs[i].NestKey == nestKey {
			return &defs[i], true
		}
	}
	return nil, false
}

// FindNestingByChildKind finds the nesting definition for a given parent kind and child kind.
// Global nesting defs are checked first, then per-kind defs.
func (s *Schema) FindNestingByChildKind(parentKind core.EntityKind, childKind core.EntityKind) (*NestingDefinition, bool) {
	defs := s.GetNestingDefs(parentKind)
	for i := range defs {
		if defs[i].ChildKind == childKind {
			return &defs[i], true
		}
	}
	return nil, false
}

// ValidateProperty validates a property value against its definition.
func (s *Schema) ValidateProperty(propDef *PropertyDefinition, value interface{}) error {
	if value == nil {
		if propDef.Required {
			return fmt.Errorf("property %q is required but not set", propDef.Name)
		}
		return nil
	}

	if propDef.Type == PropertyTypeList {
		list, ok := value.([]interface{})
		if !ok {
			return fmt.Errorf("property %q: expected list value, got %T", propDef.Name, value)
		}
		if len(propDef.Properties) > 0 {
			allowedKeys := make(map[string]bool, len(propDef.Properties))
			names := make([]string, 0, len(propDef.Properties))
			for _, subProp := range propDef.Properties {
				allowedKeys[subProp.Name] = true
				names = append(names, subProp.Name)
			}
			sort.Strings(names)
			for i, item := range list {
				itemMap, ok := item.(map[string]interface{})
				if !ok {
					return fmt.Errorf("property %q[%d]: expected map value, got %T", propDef.Name, i, item)
				}
				for _, subProp := range propDef.Properties {
					subVal, exists := itemMap[subProp.Name]
					if subVal == nil && !exists {
						if subProp.Required {
							return fmt.Errorf("property %q[%d].%s: required but not set", propDef.Name, i, subProp.Name)
						}
						continue
					}
					if err := s.ValidateProperty(&subProp, subVal); err != nil {
						return fmt.Errorf("property %q[%d]: %w", propDef.Name, i, err)
					}
				}
				for key := range itemMap {
					if !allowedKeys[key] {
						return fmt.Errorf("property %q[%d]: unknown key %q (allowed keys: %v)", propDef.Name, i, key, names)
					}
				}
			}
		}
	}

	// Validate map values against an optional element schema (B4).
	if propDef.Type == PropertyTypeMap && propDef.ValueProperty != nil {
		values, ok := mapValues(value)
		if !ok {
			return fmt.Errorf("property %q: expected map value, got %T", propDef.Name, value)
		}
		keys := make([]string, 0, len(values))
		for k := range values {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			if err := s.ValidateProperty(propDef.ValueProperty, values[k]); err != nil {
				return fmt.Errorf("property %q[%s]: %w", propDef.Name, k, err)
			}
		}
	}

	// Generic type validation (list and reference types are handled separately).
	if err := validatePropertyType(propDef, value); err != nil {
		return err
	}

	if propDef.Constraints != nil {
		if err := validateConstraints(propDef, value); err != nil {
			return fmt.Errorf("property %q: %w", propDef.Name, err)
		}
	}

	// Type-specific validation for reference properties
	if propDef.Type == PropertyTypeReference {
		if !core.IsReferenceValue(value) {
			if str, ok := value.(string); ok && len(str) > 0 && str[0] == '@' {
				// @ prefix present but not yet converted to ReferenceValue (e.g., raw input)
				return nil
			}
			return fmt.Errorf("property %q: expected a reference (use @ prefix), got %T", propDef.Name, value)
		}
	}

	return nil
}

func validatePropertyType(propDef *PropertyDefinition, value interface{}) error {
	switch propDef.Type {
	case PropertyTypeString:
		switch value.(type) {
		case string, core.ReferenceValue:
			// ReferenceValue is accepted because the parser converts any
			// @-prefixed string into a reference regardless of the declared type.
		default:
			return fmt.Errorf("property %q: expected string value, got %T", propDef.Name, value)
		}
	case PropertyTypeInteger:
		f, ok := numericValue(value)
		if !ok {
			return fmt.Errorf("property %q: expected integer value, got %T", propDef.Name, value)
		}
		if f != math.Trunc(f) {
			return fmt.Errorf("property %q: expected integer value, got %T (%v)", propDef.Name, value, f)
		}
	case PropertyTypeNumber:
		if _, ok := numericValue(value); !ok {
			return fmt.Errorf("property %q: expected numeric value, got %T", propDef.Name, value)
		}
	case PropertyTypeBoolean:
		if _, ok := value.(bool); !ok {
			return fmt.Errorf("property %q: expected boolean value, got %T", propDef.Name, value)
		}
	case PropertyTypeMap:
		switch value.(type) {
		case map[string]interface{}, map[string]string:
		default:
			return fmt.Errorf("property %q: expected map value, got %T", propDef.Name, value)
		}
	case PropertyTypeList:
		// List values are validated in ValidateProperty itself.
	case PropertyTypeReference:
		// Reference values are validated in ValidateProperty itself.
	}
	return nil
}

// numericValue coerces a value to float64 if it is a numeric type.
func numericValue(v interface{}) (float64, bool) {
	switch n := v.(type) {
	case int:
		return float64(n), true
	case int8:
		return float64(n), true
	case int16:
		return float64(n), true
	case int32:
		return float64(n), true
	case int64:
		return float64(n), true
	case uint:
		return float64(n), true
	case uint8:
		return float64(n), true
	case uint16:
		return float64(n), true
	case uint32:
		return float64(n), true
	case uint64:
		return float64(n), true
	case float32:
		return float64(n), true
	case float64:
		return n, true
	default:
		return 0, false
	}
}

// mapValues coerces a map value to map[string]interface{} for element validation.
func mapValues(value interface{}) (map[string]interface{}, bool) {
	switch m := value.(type) {
	case map[string]interface{}:
		return m, true
	case map[string]string:
		out := make(map[string]interface{}, len(m))
		for k, v := range m {
			out[k] = v
		}
		return out, true
	default:
		return nil, false
	}
}

func validateConstraints(propDef *PropertyDefinition, value interface{}) error {
	c := propDef.Constraints

	switch propDef.Type {
	case PropertyTypeInteger, PropertyTypeNumber:
		return validateNumericConstraints(c, value)
	case PropertyTypeString:
		return validateStringConstraints(c, value)
	case PropertyTypeList:
		return validateListConstraints(c, value)
	case PropertyTypeReference:
		return validateReferenceConstraints(c, value)
	}

	return nil
}

func validateNumericConstraints(c *Constraint, value interface{}) error {
	fval, ok := numericValue(value)
	if !ok {
		return fmt.Errorf("expected numeric value, got %T", value)
	}

	if c.Min != nil && fval < *c.Min {
		return fmt.Errorf("value %v is less than minimum %v", fval, *c.Min)
	}
	if c.Max != nil && fval > *c.Max {
		return fmt.Errorf("value %v is greater than maximum %v", fval, *c.Max)
	}

	return nil
}

func validateListConstraints(c *Constraint, value interface{}) error {
	list, ok := value.([]interface{})
	if !ok {
		return fmt.Errorf("expected list value, got %T", value)
	}

	if c.UniqueItems != nil && *c.UniqueItems {
		seen := make(map[interface{}]bool)
		for _, item := range list {
			if seen[item] {
				return fmt.Errorf("list contains duplicate item: %v", item)
			}
			seen[item] = true
		}
	}

	return nil
}

func validateStringConstraints(c *Constraint, value interface{}) error {
	switch v := value.(type) {
	case core.ReferenceValue:
		value = v.String()
	case string:
	default:
		return fmt.Errorf("expected string value, got %T", value)
	}

	sval, ok := value.(string)
	if !ok {
		return fmt.Errorf("expected string value, got %T", value)
	}

	if c.MinLength != nil && len(sval) < *c.MinLength {
		return fmt.Errorf("string length %d is less than minimum %d", len(sval), *c.MinLength)
	}
	if c.MaxLength != nil && len(sval) > *c.MaxLength {
		return fmt.Errorf("string length %d is greater than maximum %d", len(sval), *c.MaxLength)
	}
	if c.Pattern != nil {
		re, err := regexp.Compile(*c.Pattern)
		if err != nil {
			return fmt.Errorf("invalid pattern %q: %v", *c.Pattern, err)
		}
		if !re.MatchString(sval) {
			return fmt.Errorf("value %q does not match pattern %q", sval, *c.Pattern)
		}
	}
	if len(c.Enum) > 0 {
		found := false
		for _, allowed := range c.Enum {
			if sval == allowed {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("value %q is not in allowed enum values %v", sval, c.Enum)
		}
	}

	return nil
}

func validateReferenceConstraints(c *Constraint, value interface{}) error {
	// Basic type check: must be a ReferenceValue or a string with @ prefix
	if _, ok := value.(core.ReferenceValue); !ok {
		if str, ok := value.(string); ok && len(str) > 0 && str[0] == '@' {
			// Raw @ prefix string (not yet converted) is acceptable
			return nil
		}
		return fmt.Errorf("expected a reference value, got %T", value)
	}
	return nil
}
