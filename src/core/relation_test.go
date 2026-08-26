package core

import (
	"testing"
)

func TestNewDirectedRelation(t *testing.T) {
	r := NewDirectedRelation("rel-01", "hosts", "srv-01", "vm-01")
	if r.ID != "rel-01" {
		t.Errorf("expected ID rel-01, got %s", r.ID)
	}
	if r.Type != "hosts" {
		t.Errorf("expected Type hosts, got %s", r.Type)
	}
	if r.Direction != DirectionDirected {
		t.Errorf("expected Direction directed, got %s", r.Direction)
	}
	if r.Source() != "srv-01" {
		t.Errorf("expected Source srv-01, got %s", r.Source())
	}
	if r.Target() != "vm-01" {
		t.Errorf("expected Target vm-01, got %s", r.Target())
	}
	if !r.IsDirected() {
		t.Error("expected relation to be directed")
	}
	if r.IsSymmetric() {
		t.Error("expected relation not to be symmetric")
	}
}

func TestNewSymmetricRelation(t *testing.T) {
	r := NewSymmetricRelation("rel-02", "connects", []string{"intf-01", "intf-02"})
	if r.ID != "rel-02" {
		t.Errorf("expected ID rel-02, got %s", r.ID)
	}
	if r.Type != "connects" {
		t.Errorf("expected Type connects, got %s", r.Type)
	}
	if r.Direction != DirectionSymmetric {
		t.Errorf("expected Direction symmetric, got %s", r.Direction)
	}
	if r.IsDirected() {
		t.Error("expected relation not to be directed")
	}
	if !r.IsSymmetric() {
		t.Error("expected relation to be symmetric")
	}
	ids := r.ParticipantIDs()
	if len(ids) != 2 || ids[0] != "intf-01" || ids[1] != "intf-02" {
		t.Errorf("expected participants [intf-01 intf-02], got %v", ids)
	}
}

func TestNewRelation(t *testing.T) {
	r := NewRelation("rel-03", "depends_on", DirectionDirected)
	if r.ID != "rel-03" {
		t.Errorf("expected ID rel-03, got %s", r.ID)
	}
	if r.Type != "depends_on" {
		t.Errorf("expected Type depends_on, got %s", r.Type)
	}
	if r.Direction != DirectionDirected {
		t.Errorf("expected Direction directed, got %s", r.Direction)
	}
	if r.Properties == nil {
		t.Error("expected Properties to be initialized")
	}
}

func TestRelationParticipantsAllIDs(t *testing.T) {
	p := &Participants{Source: "a", Target: "b"}
	ids := p.AllIDs()
	if len(ids) != 2 || ids[0] != "a" || ids[1] != "b" {
		t.Errorf("expected [a b], got %v", ids)
	}

	p2 := &Participants{List: []string{"x", "y", "z"}}
	ids2 := p2.AllIDs()
	if len(ids2) != 3 {
		t.Errorf("expected 3 participants, got %d", len(ids2))
	}
}

func TestRelationParticipantsCount(t *testing.T) {
	p := &Participants{Source: "a", Target: "b"}
	if p.Count() != 2 {
		t.Errorf("expected count 2, got %d", p.Count())
	}

	p2 := &Participants{List: []string{"x", "y"}}
	if p2.Count() != 2 {
		t.Errorf("expected count 2, got %d", p2.Count())
	}

	p3 := &Participants{}
	if p3.Count() != 0 {
		t.Errorf("expected count 0, got %d", p3.Count())
	}
}

func TestRelationSetStatus(t *testing.T) {
	r := NewRelation("rel-01", "hosts", DirectionDirected)
	r.SetStatus(StatusActive)
	if r.Status != StatusActive {
		t.Errorf("expected status active, got %s", r.Status)
	}
}

func TestRelationTags(t *testing.T) {
	r := NewRelation("rel-01", "hosts", DirectionDirected)
	r.AddTag("critical")
	r.AddTag("network")

	if len(r.Tags) != 2 {
		t.Errorf("expected 2 tags, got %d", len(r.Tags))
	}
	if !r.HasTag("critical") {
		t.Error("expected to have tag critical")
	}
	if r.HasTag("nonexistent") {
		t.Error("expected not to have tag nonexistent")
	}
}

func TestRelationLabels(t *testing.T) {
	r := NewRelation("rel-01", "hosts", DirectionDirected)
	r.SetLabel("env", "prod")
	r.SetLabel("tier", "web")

	if v, ok := r.GetLabel("env"); !ok || v != "prod" {
		t.Errorf("expected label env=prod, got %s", v)
	}
	if _, ok := r.GetLabel("nonexistent"); ok {
		t.Error("expected no label nonexistent")
	}
}

func TestRelationProperties(t *testing.T) {
	r := NewRelation("rel-01", "depends_on", DirectionDirected)
	r.SetProperty("dependency_type", "runtime")
	r.SetProperty("critical", true)

	if v, ok := r.GetProperty("dependency_type"); !ok || v != "runtime" {
		t.Errorf("expected property dependency_type=runtime, got %v", v)
	}
	if v, ok := r.GetProperty("critical"); !ok || v != true {
		t.Errorf("expected property critical=true, got %v", v)
	}
}

func TestRelationValidate(t *testing.T) {
	tests := []struct {
		name     string
		relation *Relation
		wantErr  error
	}{
		{
			name:     "valid directed relation",
			relation: NewDirectedRelation("rel-01", "hosts", "srv-01", "vm-01"),
			wantErr:  nil,
		},
		{
			name:     "valid symmetric relation",
			relation: NewSymmetricRelation("rel-02", "connects", []string{"intf-01", "intf-02"}),
			wantErr:  nil,
		},
		{
			name:     "missing ID",
			relation: NewDirectedRelation("", "hosts", "srv-01", "vm-01"),
			wantErr:  ErrRelationMissingID,
		},
		{
			name:     "missing type",
			relation: NewDirectedRelation("rel-01", "", "srv-01", "vm-01"),
			wantErr:  ErrRelationMissingType,
		},
		{
			name:     "missing participants",
			relation: NewRelation("rel-01", "hosts", DirectionDirected),
			wantErr:  ErrRelationMissingParticipants,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.relation.Validate()
			if err != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRelationString(t *testing.T) {
	r := NewDirectedRelation("rel-01", "hosts", "srv-01", "vm-01")
	r.SetStatus(StatusActive)
	s := r.String()
	if s == "" {
		t.Error("expected non-empty string representation")
	}
}

func TestDirectionConstants(t *testing.T) {
	if DirectionDirected != "directed" {
		t.Errorf("expected directed, got %s", DirectionDirected)
	}
	if DirectionSymmetric != "symmetric" {
		t.Errorf("expected symmetric, got %s", DirectionSymmetric)
	}
}
