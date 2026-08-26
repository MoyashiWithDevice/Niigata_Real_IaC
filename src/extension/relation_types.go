package extension

import (
	"fmt"

	"IACForge/src/core"
	"IACForge/src/schema"
)

// RelationTypesExtensionPoint manages relation type extensions.
type RelationTypesExtensionPoint struct {
	schema *schema.Schema
	types  map[core.RelationType]string // type -> extension ID that registered it
}

// NewRelationTypesExtensionPoint creates a new relation types extension point.
func NewRelationTypesExtensionPoint(s *schema.Schema) *RelationTypesExtensionPoint {
	return &RelationTypesExtensionPoint{
		schema: s,
		types:  make(map[core.RelationType]string),
	}
}

// Type returns the extension point type.
func (ep *RelationTypesExtensionPoint) Type() ExtensionPointType {
	return ExtensionPointRelationTypes
}

// Register registers all relation type contributions from the given extension.
func (ep *RelationTypesExtensionPoint) Register(ext *Extension) error {
	for _, contrib := range ext.RelationTypes {
		existing, ok := ep.schema.GetRelationTypeDef(contrib.Type)
		if !ok {
			// New relation type: register it as-is.
			ep.schema.AddRelationType(contrib.Type, contrib.Definition)
			ep.types[contrib.Type] = ext.Manifest.ID
			continue
		}

		if !contrib.Augment {
			return fmt.Errorf("%w: relation type %q already defined in schema", ErrCoreConflict, contrib.Type)
		}

		// Augment an existing definition by merging participant kinds.
		if contrib.Definition == nil || contrib.Definition.Participants == nil {
			return fmt.Errorf("%w: augmenting relation type %q requires participant kinds", ErrInvalidExtension, contrib.Type)
		}

		// Only participant kind augmentation is supported. Reject attempts to
		// change any other part of the definition instead of silently dropping it.
		if contrib.Definition.Direction != "" && existing.Direction != contrib.Definition.Direction {
			return fmt.Errorf("%w: augmenting relation type %q cannot change direction from %q to %q", ErrInvalidExtension, contrib.Type, existing.Direction, contrib.Definition.Direction)
		}
		if len(contrib.Definition.Properties) > 0 {
			return fmt.Errorf("%w: augmenting relation type %q cannot add properties", ErrInvalidExtension, contrib.Type)
		}
		if existing.Participants != nil {
			if contrib.Definition.Participants.MinParticipants != 0 && existing.Participants.MinParticipants != contrib.Definition.Participants.MinParticipants {
				return fmt.Errorf("%w: augmenting relation type %q cannot change min participants from %d to %d", ErrInvalidExtension, contrib.Type, existing.Participants.MinParticipants, contrib.Definition.Participants.MinParticipants)
			}
			if contrib.Definition.Participants.MaxParticipants != 0 && existing.Participants.MaxParticipants != contrib.Definition.Participants.MaxParticipants {
				return fmt.Errorf("%w: augmenting relation type %q cannot change max participants from %d to %d", ErrInvalidExtension, contrib.Type, existing.Participants.MaxParticipants, contrib.Definition.Participants.MaxParticipants)
			}
		}

		// Merge into a copy so the original definition is not mutated.
		merged := cloneRelationTypeDefinition(existing)
		if merged.Participants == nil {
			merged.Participants = &schema.ParticipantConstraints{}
		}
		merged.Participants.SourceKinds = mergeEntityKinds(merged.Participants.SourceKinds, contrib.Definition.Participants.SourceKinds)
		merged.Participants.TargetKinds = mergeEntityKinds(merged.Participants.TargetKinds, contrib.Definition.Participants.TargetKinds)
		ep.schema.AddRelationType(contrib.Type, merged)
		ep.types[contrib.Type] = ext.Manifest.ID
	}
	return nil
}

// mergeEntityKinds merges the additional kinds into the existing list, preserving
// order and removing duplicates.
func mergeEntityKinds(existing, additional []core.EntityKind) []core.EntityKind {
	if len(additional) == 0 {
		return existing
	}
	result := make([]core.EntityKind, 0, len(existing)+len(additional))
	seen := make(map[core.EntityKind]bool, len(existing)+len(additional))
	for _, k := range existing {
		if !seen[k] {
			seen[k] = true
			result = append(result, k)
		}
	}
	for _, k := range additional {
		if !seen[k] {
			seen[k] = true
			result = append(result, k)
		}
	}
	return result
}

// cloneRelationTypeDefinition returns a copy of a relation type definition.
// Slices are copied so mutating the copy does not affect the original.
func cloneRelationTypeDefinition(d *schema.RelationTypeDefinition) *schema.RelationTypeDefinition {
	if d == nil {
		return nil
	}
	c := *d
	if d.Participants != nil {
		p := *d.Participants
		p.SourceKinds = append([]core.EntityKind(nil), d.Participants.SourceKinds...)
		p.TargetKinds = append([]core.EntityKind(nil), d.Participants.TargetKinds...)
		c.Participants = &p
	}
	c.Properties = append([]schema.PropertyDefinition(nil), d.Properties...)
	return &c
}

// GetRelationTypesByExtension returns all relation types registered by a specific extension.
func (ep *RelationTypesExtensionPoint) GetRelationTypesByExtension(extensionID string) []core.RelationType {
	var result []core.RelationType
	for relType, extID := range ep.types {
		if extID == extensionID {
			result = append(result, relType)
		}
	}
	return result
}

// GetExtensionForType returns the extension ID that registered the given relation type.
func (ep *RelationTypesExtensionPoint) GetExtensionForType(relType core.RelationType) (string, bool) {
	extID, ok := ep.types[relType]
	return extID, ok
}
