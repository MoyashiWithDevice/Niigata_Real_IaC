package schema

// Profile defines a subset of the Core Schema for specific use cases.
type Profile struct {
	Name               string              `yaml:"name"`
	Description        string              `yaml:"description,omitempty"`
	Rules              []string            `yaml:"rules,omitempty"`
	RequiredKinds      []string            `yaml:"required_kinds,omitempty"`
	RequiredRelations  []string            `yaml:"required_relations,omitempty"`
	RequiredProperties map[string][]string `yaml:"required_properties,omitempty"`
}

// NewProfile creates a new validation profile.
func NewProfile(name string) *Profile {
	return &Profile{
		Name:               name,
		RequiredProperties: make(map[string][]string),
	}
}

// AddRequiredKind adds a required entity kind to the profile.
func (p *Profile) AddRequiredKind(kind string) {
	p.RequiredKinds = append(p.RequiredKinds, kind)
}

// AddRequiredRelation adds a required relation type to the profile.
func (p *Profile) AddRequiredRelation(relType string) {
	p.RequiredRelations = append(p.RequiredRelations, relType)
}

// HasRule checks if the profile includes the given rule ID.
func (p *Profile) HasRule(ruleID string) bool {
	for _, r := range p.Rules {
		if r == ruleID {
			return true
		}
	}
	return false
}
