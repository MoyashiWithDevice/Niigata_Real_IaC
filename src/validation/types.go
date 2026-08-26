package validation

// Severity represents the severity level of a validation finding.
type Severity string

const (
	SeverityInfo    Severity = "info"
	SeverityWarning Severity = "warning"
	SeverityError   Severity = "error"
)

// ObjectType represents the type of object a finding relates to.
type ObjectType string

const (
	ObjectTypeEntity   ObjectType = "entity"
	ObjectTypeRelation ObjectType = "relation"
)

// Finding represents a single validation finding.
type Finding struct {
	RuleID     string     `yaml:"rule_id"`
	Severity   Severity   `yaml:"severity"`
	Message    string     `yaml:"message"`
	ObjectID   string     `yaml:"object_id,omitempty"`
	ObjectType ObjectType `yaml:"object_type,omitempty"`
	Path       string     `yaml:"path,omitempty"`
}

// Rule defines a validation rule.
type Rule struct {
	ID          string   `yaml:"id"`
	Name        string   `yaml:"name"`
	Description string   `yaml:"description,omitempty"`
	Severity    Severity `yaml:"severity"`
}

// Result holds the complete validation result.
type Result struct {
	Findings []Finding `yaml:"findings"`
	Passed   bool      `yaml:"passed"`
	Summary  Summary   `yaml:"summary"`
}

// Summary provides statistics about the validation result.
type Summary struct {
	TotalRules    int `yaml:"total_rules"`
	TotalFindings int `yaml:"total_findings"`
	Errors        int `yaml:"errors"`
	Warnings      int `yaml:"warnings"`
	Infos         int `yaml:"infos"`
}

// RuleFunc is the function signature for a validation rule implementation.
// It receives the context and returns findings for that rule.
type RuleFunc func(ctx *Context) []Finding

// Context provides access to the graph and schema during validation.
type Context struct {
	// Graph is the graph being validated.
	Graph interface{} // *core.Graph - using interface to avoid circular import
	// Schema is the schema to validate against.
	Schema interface{} // *schema.Schema
	// Profile is the optional validation profile.
	Profile interface{} // *schema.Profile
}
