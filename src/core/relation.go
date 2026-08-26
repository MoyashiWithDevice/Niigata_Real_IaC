package core

import (
	"errors"
	"fmt"
	"strings"
)

var (
	ErrRelationMissingID           = errors.New("relation missing required property: id")
	ErrRelationMissingType         = errors.New("relation missing required property: type")
	ErrRelationMissingParticipants = errors.New("relation missing required property: participants")
	ErrRelationInvalidDirection    = errors.New("relation has invalid direction for type")
)

type Direction string

const (
	DirectionDirected  Direction = "directed"
	DirectionSymmetric Direction = "symmetric"
)

type RelationType string

type Participants struct {
	Source string   `yaml:"source,omitempty"`
	Target string   `yaml:"target,omitempty"`
	List   []string `yaml:"list,omitempty"`
}

func (p *Participants) AllIDs() []string {
	if len(p.List) > 0 {
		result := make([]string, len(p.List))
		copy(result, p.List)
		return result
	}
	var ids []string
	if p.Source != "" {
		ids = append(ids, p.Source)
	}
	if p.Target != "" {
		ids = append(ids, p.Target)
	}
	return ids
}

func (p *Participants) Count() int {
	if len(p.List) > 0 {
		return len(p.List)
	}
	count := 0
	if p.Source != "" {
		count++
	}
	if p.Target != "" {
		count++
	}
	return count
}

type Relation struct {
	ID           string                 `yaml:"id"`
	Type         RelationType           `yaml:"type"`
	Participants Participants           `yaml:"participants"`
	Direction    Direction              `yaml:"direction,omitempty"`
	Description  string                 `yaml:"description,omitempty"`
	Status       Status                 `yaml:"status,omitempty"`
	Tags         []string               `yaml:"tags,omitempty"`
	Labels       map[string]string      `yaml:"labels,omitempty"`
	Extensions   map[string]interface{} `yaml:"extensions,omitempty"`
	Properties   map[string]interface{} `yaml:"properties,omitempty"`
}

func NewRelation(id string, relType RelationType, direction Direction) *Relation {
	return &Relation{
		ID:         id,
		Type:       relType,
		Direction:  direction,
		Properties: make(map[string]interface{}),
	}
}

func NewDirectedRelation(id string, relType RelationType, source, target string) *Relation {
	return &Relation{
		ID:        id,
		Type:      relType,
		Direction: DirectionDirected,
		Participants: Participants{
			Source: source,
			Target: target,
		},
		Properties: make(map[string]interface{}),
	}
}

func NewSymmetricRelation(id string, relType RelationType, participants []string) *Relation {
	return &Relation{
		ID:        id,
		Type:      relType,
		Direction: DirectionSymmetric,
		Participants: Participants{
			List: participants,
		},
		Properties: make(map[string]interface{}),
	}
}

func (r *Relation) SetStatus(status Status) {
	r.Status = status
}

func (r *Relation) AddTag(tag string) {
	r.Tags = append(r.Tags, tag)
}

func (r *Relation) HasTag(tag string) bool {
	for _, t := range r.Tags {
		if t == tag {
			return true
		}
	}
	return false
}

func (r *Relation) SetLabel(key, value string) {
	if r.Labels == nil {
		r.Labels = make(map[string]string)
	}
	r.Labels[key] = value
}

func (r *Relation) GetLabel(key string) (string, bool) {
	if r.Labels == nil {
		return "", false
	}
	v, ok := r.Labels[key]
	return v, ok
}

func (r *Relation) SetProperty(key string, value interface{}) {
	if r.Properties == nil {
		r.Properties = make(map[string]interface{})
	}
	r.Properties[key] = value
}

func (r *Relation) GetProperty(key string) (interface{}, bool) {
	if r.Properties == nil {
		return nil, false
	}
	v, ok := r.Properties[key]
	return v, ok
}

func (r *Relation) IsDirected() bool {
	return r.Direction == DirectionDirected
}

func (r *Relation) IsSymmetric() bool {
	return r.Direction == DirectionSymmetric
}

func (r *Relation) Source() string {
	return r.Participants.Source
}

func (r *Relation) Target() string {
	return r.Participants.Target
}

func (r *Relation) ParticipantIDs() []string {
	return r.Participants.AllIDs()
}

func (r *Relation) Validate() error {
	if r.ID == "" {
		return ErrRelationMissingID
	}
	if r.Type == "" {
		return ErrRelationMissingType
	}
	if r.Participants.Count() < 2 {
		return ErrRelationMissingParticipants
	}
	if r.Direction == "" {
		r.Direction = DirectionDirected
	}
	return nil
}

func (r *Relation) String() string {
	var parts []string
	parts = append(parts, fmt.Sprintf("id=%s", r.ID))
	parts = append(parts, fmt.Sprintf("type=%s", r.Type))
	parts = append(parts, fmt.Sprintf("direction=%s", r.Direction))
	if r.Participants.Source != "" || r.Participants.Target != "" {
		parts = append(parts, fmt.Sprintf("source=%s", r.Participants.Source))
		parts = append(parts, fmt.Sprintf("target=%s", r.Participants.Target))
	} else if len(r.Participants.List) > 0 {
		parts = append(parts, fmt.Sprintf("participants=%v", r.Participants.List))
	}
	if r.Status != "" {
		parts = append(parts, fmt.Sprintf("status=%s", r.Status))
	}
	return fmt.Sprintf("Relation{%s}", strings.Join(parts, ", "))
}
