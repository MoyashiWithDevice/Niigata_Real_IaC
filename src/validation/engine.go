package validation

import (
	"fmt"
	"net"
	"sort"
	"strings"

	"IACForge/src/core"
	"IACForge/src/core/kinds"
	"IACForge/src/core/types"
	"IACForge/src/schema"
)

// Engine is the validation engine that evaluates rules against a graph.
type Engine struct {
	schema           *schema.Schema
	rules            map[string]RuleFunc
	ruleDefs         map[string]*Rule
	allowedRootKinds map[core.EntityKind]bool
}

// NewEngine creates a new validation engine with the given schema.
func NewEngine(s *schema.Schema) *Engine {
	e := &Engine{
		schema:           s,
		rules:            make(map[string]RuleFunc),
		ruleDefs:         make(map[string]*Rule),
		allowedRootKinds: make(map[core.EntityKind]bool),
	}
	return e
}

// AddAllowedRootKind grants root authority to the given entity kind. A graph may
// have multiple root entities only when every root's kind has been granted root
// authority (e.g. an extension registering aws.organization).
func (e *Engine) AddAllowedRootKind(kind core.EntityKind) {
	e.allowedRootKinds[kind] = true
}

// IsAllowedRootKind reports whether the given kind has been granted root authority.
func (e *Engine) IsAllowedRootKind(kind core.EntityKind) bool {
	return e.allowedRootKinds[kind]
}

// rootEntities returns the root entity IDs (entities without an owner).
func rootEntities(g *core.Graph) []string {
	var roots []string
	for _, e := range g.Entities() {
		if e.Owner == "" {
			roots = append(roots, e.ID)
		}
	}
	sort.Strings(roots)
	return roots
}

// disallowedRoots returns the root entity IDs whose kind has not been granted
// root authority.
func (e *Engine) disallowedRoots(g *core.Graph) []string {
	var disallowed []string
	for _, ent := range g.Entities() {
		if ent.Owner == "" && !e.IsAllowedRootKind(ent.Kind) {
			disallowed = append(disallowed, ent.ID)
		}
	}
	sort.Strings(disallowed)
	return disallowed
}

// RegisterRule registers a validation rule with the engine.
func (e *Engine) RegisterRule(rule *Rule, fn RuleFunc) {
	e.rules[rule.ID] = fn
	e.ruleDefs[rule.ID] = rule
}

// Validate validates the given graph against all registered rules.
func (e *Engine) Validate(g *core.Graph, profile *schema.Profile) *Result {
	result := &Result{}

	for ruleID, fn := range e.rules {
		ruleDef := e.ruleDefs[ruleID]

		if profile != nil && len(profile.Rules) > 0 {
			if !profile.HasRule(ruleID) {
				continue
			}
		}

		ctx := &Context{
			Graph:   g,
			Schema:  e.schema,
			Profile: profile,
		}

		findings := fn(ctx)
		for i := range findings {
			if findings[i].RuleID == "" {
				findings[i].RuleID = ruleID
			}
			if findings[i].Severity == "" && ruleDef != nil {
				findings[i].Severity = ruleDef.Severity
			}
		}
		result.Findings = append(result.Findings, findings...)
	}

	if profile != nil && len(profile.RequiredKinds) > 0 {
		result.Findings = append(result.Findings, e.checkRequiredKinds(g, profile)...)
	}
	if profile != nil && len(profile.RequiredRelations) > 0 {
		result.Findings = append(result.Findings, e.checkRequiredRelations(g, profile)...)
	}

	result.Summary = e.computeSummary(result.Findings)
	result.Passed = result.Summary.Errors == 0

	return result
}

func (e *Engine) checkRequiredKinds(g *core.Graph, profile *schema.Profile) []Finding {
	var findings []Finding
	for _, kind := range profile.RequiredKinds {
		entities := g.EntitiesByKind(core.EntityKind(kind))
		if len(entities) == 0 {
			findings = append(findings, Finding{
				RuleID:   "profile-required-kind",
				Severity: SeverityError,
				Message:  fmt.Sprintf("profile requires at least one entity of kind %q", kind),
			})
		}
	}
	return findings
}

func (e *Engine) checkRequiredRelations(g *core.Graph, profile *schema.Profile) []Finding {
	var findings []Finding
	for _, relType := range profile.RequiredRelations {
		relations := g.RelationsByType(core.RelationType(relType))
		if len(relations) == 0 {
			findings = append(findings, Finding{
				RuleID:   "profile-required-relation",
				Severity: SeverityError,
				Message:  fmt.Sprintf("profile requires at least one relation of type %q", relType),
			})
		}
	}
	return findings
}

func (e *Engine) computeSummary(findings []Finding) Summary {
	s := Summary{
		TotalFindings: len(findings),
	}
	s.TotalRules = len(e.rules)
	for _, f := range findings {
		switch f.Severity {
		case SeverityError:
			s.Errors++
		case SeverityWarning:
			s.Warnings++
		case SeverityInfo:
			s.Infos++
		}
	}
	return s
}

// RegisterCoreRules registers all core validation rules.
func RegisterCoreRules(engine *Engine) {
	registerGraphIntegrityRules(engine)
	registerEntityRules(engine)
	registerRelationRules(engine)
	registerOwnershipRules(engine)
	registerReferenceRules(engine)
	registerNetworkRules(engine)
}

func registerGraphIntegrityRules(e *Engine) {
	e.RegisterRule(&Rule{
		ID:       "unique-id",
		Name:     "Unique Identifier",
		Severity: SeverityError,
	}, ruleUniqueID)

	e.RegisterRule(&Rule{
		ID:       "valid-reference",
		Name:     "Valid Reference",
		Severity: SeverityError,
	}, ruleValidReference)

	e.RegisterRule(&Rule{
		ID:       "valid-owner",
		Name:     "Valid Owner",
		Severity: SeverityError,
	}, ruleValidOwner)

	e.RegisterRule(&Rule{
		ID:       "single-owner",
		Name:     "Single Owner",
		Severity: SeverityError,
	}, e.ruleSingleOwner)

	e.RegisterRule(&Rule{
		ID:          "valid-property",
		Name:        "Valid Property",
		Description: "entity spec and relation properties MUST conform to the schema property definitions (type, enum, required, min/max); undefined properties are reported",
		Severity:    SeverityWarning,
	}, ruleValidProperty)
}

func registerEntityRules(e *Engine) {
	e.RegisterRule(&Rule{
		ID:       "required-kind",
		Name:     "Required Kind",
		Severity: SeverityError,
	}, ruleRequiredKind)

	e.RegisterRule(&Rule{
		ID:       "required-name",
		Name:     "Required Name",
		Severity: SeverityError,
	}, ruleRequiredName)

	e.RegisterRule(&Rule{
		ID:       "valid-kind",
		Name:     "Valid Kind",
		Severity: SeverityError,
	}, ruleValidKind)

	e.RegisterRule(&Rule{
		ID:       "valid-status",
		Name:     "Valid Status",
		Severity: SeverityWarning,
	}, ruleValidStatus)

	e.RegisterRule(&Rule{
		ID:       "valid-port-range",
		Name:     "Valid Port Range",
		Severity: SeverityError,
	}, ruleValidPortRange)

	e.RegisterRule(&Rule{
		ID:       "valid-acl-rule-parent",
		Name:     "Valid ACL Rule Parent",
		Severity: SeverityError,
	}, ruleValidACLRULEParent)

	e.RegisterRule(&Rule{
		ID:       "no-slash-in-id",
		Name:     "No Slash in Entity ID",
		Severity: SeverityError,
	}, ruleNoSlashInID)

	e.RegisterRule(&Rule{
		ID:       "valid-nesting-parent",
		Name:     "Valid Nesting Parent",
		Severity: SeverityError,
	}, ruleValidNestingParent)
}

func registerRelationRules(e *Engine) {
	e.RegisterRule(&Rule{
		ID:       "required-type",
		Name:     "Required Type",
		Severity: SeverityError,
	}, ruleRequiredType)

	e.RegisterRule(&Rule{
		ID:       "required-participants",
		Name:     "Required Participants",
		Severity: SeverityError,
	}, ruleRequiredParticipants)

	e.RegisterRule(&Rule{
		ID:       "valid-type",
		Name:     "Valid Type",
		Severity: SeverityError,
	}, ruleValidType)

	e.RegisterRule(&Rule{
		ID:       "valid-direction",
		Name:     "Valid Direction",
		Severity: SeverityError,
	}, ruleValidDirection)

	e.RegisterRule(&Rule{
		ID:       "valid-cardinality",
		Name:     "Valid Cardinality",
		Severity: SeverityError,
	}, ruleValidCardinality)

	e.RegisterRule(&Rule{
		ID:       "valid-participant-kind",
		Name:     "Valid Participant Kind",
		Severity: SeverityWarning,
	}, ruleValidParticipantKind)
}

func registerOwnershipRules(e *Engine) {
	e.RegisterRule(&Rule{
		ID:       "ownership-tree",
		Name:     "Ownership Tree",
		Severity: SeverityError,
	}, e.ruleOwnershipTree)

	e.RegisterRule(&Rule{
		ID:       "no-ownership-cycle",
		Name:     "No Ownership Cycle",
		Severity: SeverityError,
	}, ruleNoOwnershipCycle)

	e.RegisterRule(&Rule{
		ID:       "root-entity",
		Name:     "Root Entity",
		Severity: SeverityError,
	}, e.ruleRootEntity)
}

func registerReferenceRules(e *Engine) {
	e.RegisterRule(&Rule{
		ID:       "dangling-reference",
		Name:     "Dangling Reference",
		Severity: SeverityError,
	}, ruleDanglingReference)

	e.RegisterRule(&Rule{
		ID:       "invalid-path",
		Name:     "Invalid Path",
		Severity: SeverityError,
	}, ruleInvalidPath)
}

func registerNetworkRules(e *Engine) {
	e.RegisterRule(&Rule{
		ID:          "valid-ip-format",
		Name:        "Valid IP Address Format",
		Description: "interface ip_address values MUST be valid IP addresses or CIDR notation",
		Severity:    SeverityWarning,
	}, ruleValidIPFormat)

	e.RegisterRule(&Rule{
		ID:          "ip-requires-network",
		Name:        "IP Requires Network",
		Description: "interface with IP addresses SHOULD reference a network via the network property or a belongs_to relation",
		Severity:    SeverityWarning,
	}, ruleIPRequiresNetwork)

	e.RegisterRule(&Rule{
		ID:          "network-reference-kind",
		Name:        "Network Reference Kind",
		Description: "interface network reference MUST point to an entity of kind network",
		Severity:    SeverityWarning,
	}, ruleNetworkReferenceKind)

	e.RegisterRule(&Rule{
		ID:          "ip-in-cidr",
		Name:        "IP Within Network CIDR",
		Description: "interface IP addresses SHOULD be within the CIDR of a referenced network",
		Severity:    SeverityWarning,
	}, ruleIPInCIDR)

	e.RegisterRule(&Rule{
		ID:          "network-cidr-required",
		Name:        "Network CIDR Required",
		Description: "network with member interfaces that have IP addresses SHOULD define a valid cidr",
		Severity:    SeverityWarning,
	}, ruleNetworkCIDRRequired)

	e.RegisterRule(&Rule{
		ID:          "gateway-in-cidr",
		Name:        "Gateway Within Network CIDR",
		Description: "network gateway SHOULD be within the network cidr",
		Severity:    SeverityWarning,
	}, ruleGatewayInCIDR)

	e.RegisterRule(&Rule{
		ID:          "ip-unique-in-network",
		Name:        "Unique IP Within Network",
		Description: "IP addresses SHOULD be unique within a network",
		Severity:    SeverityWarning,
	}, ruleIPUniqueInNetwork)
}

// --- Rule Implementations ---

func ruleUniqueID(ctx *Context) []Finding {
	g := ctx.Graph.(*core.Graph)
	var findings []Finding

	seen := make(map[string]bool)
	for _, e := range g.Entities() {
		if seen[e.ID] {
			findings = append(findings, Finding{
				Severity:   SeverityError,
				Message:    fmt.Sprintf("duplicate entity ID: %q", e.ID),
				ObjectID:   e.ID,
				ObjectType: ObjectTypeEntity,
			})
		}
		seen[e.ID] = true
	}
	for _, r := range g.Relations() {
		if seen[r.ID] {
			findings = append(findings, Finding{
				Severity:   SeverityError,
				Message:    fmt.Sprintf("duplicate relation ID: %q", r.ID),
				ObjectID:   r.ID,
				ObjectType: ObjectTypeRelation,
			})
		}
		seen[r.ID] = true
	}

	return findings
}

func ruleValidReference(ctx *Context) []Finding {
	g := ctx.Graph.(*core.Graph)
	var findings []Finding

	for _, r := range g.Relations() {
		for _, participantID := range r.ParticipantIDs() {
			_, found := g.ResolveReference(participantID)
			if !found {
				findings = append(findings, Finding{
					Severity:   SeverityError,
					Message:    fmt.Sprintf("relation %q references non-existent object %q", r.ID, participantID),
					ObjectID:   r.ID,
					ObjectType: ObjectTypeRelation,
				})
			}
		}
	}

	return findings
}

func ruleValidOwner(ctx *Context) []Finding {
	g := ctx.Graph.(*core.Graph)
	var findings []Finding

	for _, e := range g.Entities() {
		if e.Owner != "" {
			_, found := g.GetEntity(e.Owner)
			if !found {
				findings = append(findings, Finding{
					Severity:   SeverityError,
					Message:    fmt.Sprintf("entity %q has owner %q which does not exist", e.ID, e.Owner),
					ObjectID:   e.ID,
					ObjectType: ObjectTypeEntity,
				})
			}
		}
	}

	return findings
}

func ruleValidProperty(ctx *Context) []Finding {
	g := ctx.Graph.(*core.Graph)
	s := ctx.Schema.(*schema.Schema)
	var findings []Finding

	for _, e := range g.Entities() {
		def, ok := s.GetEntityKindDef(e.Kind)
		if !ok {
			continue // undefined kind is handled by valid-kind
		}
		findings = append(findings, validateObjectProperties(s, e.ID, ObjectTypeEntity, def.Properties, e.Properties)...)
	}

	for _, r := range g.Relations() {
		def, ok := s.GetRelationTypeDef(r.Type)
		if !ok {
			continue // undefined type is handled by valid-type
		}
		findings = append(findings, validateObjectProperties(s, r.ID, ObjectTypeRelation, def.Properties, r.Properties)...)
	}

	return findings
}

// validateObjectProperties validates an object's spec properties against the
// schema property definitions, reporting type/enum/constraint violations,
// missing required properties, and undefined properties as warnings.
func validateObjectProperties(s *schema.Schema, objectID string, objectType ObjectType, defs []schema.PropertyDefinition, values map[string]interface{}) []Finding {
	var findings []Finding
	defined := make(map[string]bool, len(defs))

	for i := range defs {
		def := defs[i]
		defined[def.Name] = true

		value, present := values[def.Name]
		if !present {
			if def.Required {
				findings = append(findings, Finding{
					Severity:   SeverityWarning,
					Message:    fmt.Sprintf("%s %q is missing required property %q", objectType, objectID, def.Name),
					ObjectID:   objectID,
					ObjectType: objectType,
				})
			}
			continue
		}

		if err := s.ValidateProperty(&def, value); err != nil {
			findings = append(findings, Finding{
				Severity:   SeverityWarning,
				Message:    fmt.Sprintf("%s %q property %q: %v", objectType, objectID, def.Name, err),
				ObjectID:   objectID,
				ObjectType: objectType,
			})
		}
	}

	for key := range values {
		if !defined[key] {
			findings = append(findings, Finding{
				Severity:   SeverityWarning,
				Message:    fmt.Sprintf("%s %q has undefined property %q", objectType, objectID, key),
				ObjectID:   objectID,
				ObjectType: objectType,
			})
		}
	}

	return findings
}

func (e *Engine) ruleSingleOwner(ctx *Context) []Finding {
	g := ctx.Graph.(*core.Graph)
	var findings []Finding

	roots := rootEntities(g)

	// For each entity with an owner, verify the owner exists
	for _, ent := range g.Entities() {
		if ent.Owner == "" {
			continue
		}
		_, found := g.GetEntity(ent.Owner)
		if !found {
			findings = append(findings, Finding{
				Severity:   SeverityError,
				Message:    fmt.Sprintf("entity %q has no valid owner (references non-existent entity %q)", ent.ID, ent.Owner),
				ObjectID:   ent.ID,
				ObjectType: ObjectTypeEntity,
			})
		}
	}

	// Check that exactly one root exists (complements root-entity rule),
	// unless every root's kind has been granted root authority.
	if len(roots) == 0 {
		findings = append(findings, Finding{
			Severity: SeverityError,
			Message:  "no root entity found (every entity has an owner, but one root is required)",
		})
	} else if len(roots) > 1 {
		disallowed := e.disallowedRoots(g)
		if len(disallowed) > 0 {
			findings = append(findings, Finding{
				Severity: SeverityError,
				Message:  fmt.Sprintf("multiple root entities found: %v (exactly one root is required unless every root is of a root-authorized kind; disallowed roots: %v)", roots, disallowed),
			})
		}
	}

	return findings
}

func ruleRequiredKind(ctx *Context) []Finding {
	g := ctx.Graph.(*core.Graph)
	var findings []Finding

	for _, e := range g.Entities() {
		if e.Kind == "" {
			findings = append(findings, Finding{
				Severity:   SeverityError,
				Message:    fmt.Sprintf("entity %q is missing required 'kind' property", e.ID),
				ObjectID:   e.ID,
				ObjectType: ObjectTypeEntity,
			})
		}
	}

	return findings
}

func ruleRequiredName(ctx *Context) []Finding {
	g := ctx.Graph.(*core.Graph)
	var findings []Finding

	for _, e := range g.Entities() {
		if e.Name == "" {
			findings = append(findings, Finding{
				Severity:   SeverityError,
				Message:    fmt.Sprintf("entity %q is missing required 'name' property", e.ID),
				ObjectID:   e.ID,
				ObjectType: ObjectTypeEntity,
			})
		}
	}

	return findings
}

func ruleValidKind(ctx *Context) []Finding {
	g := ctx.Graph.(*core.Graph)
	s := ctx.Schema.(*schema.Schema)
	var findings []Finding

	for _, e := range g.Entities() {
		if !s.HasEntityKind(e.Kind) {
			findings = append(findings, Finding{
				Severity:   SeverityError,
				Message:    fmt.Sprintf("entity %q has undefined kind %q", e.ID, e.Kind),
				ObjectID:   e.ID,
				ObjectType: ObjectTypeEntity,
			})
		}
	}

	return findings
}

func ruleValidStatus(ctx *Context) []Finding {
	g := ctx.Graph.(*core.Graph)
	var findings []Finding

	for _, e := range g.Entities() {
		if e.Status != "" && !kinds.IsValidStatus(e.Status) {
			findings = append(findings, Finding{
				Severity:   SeverityWarning,
				Message:    fmt.Sprintf("entity %q has invalid status %q", e.ID, e.Status),
				ObjectID:   e.ID,
				ObjectType: ObjectTypeEntity,
			})
		}
	}

	for _, r := range g.Relations() {
		if r.Status != "" && !kinds.IsValidStatus(r.Status) {
			findings = append(findings, Finding{
				Severity:   SeverityWarning,
				Message:    fmt.Sprintf("relation %q has invalid status %q", r.ID, r.Status),
				ObjectID:   r.ID,
				ObjectType: ObjectTypeRelation,
			})
		}
	}

	return findings
}

func ruleValidPortRange(ctx *Context) []Finding {
	g := ctx.Graph.(*core.Graph)
	var findings []Finding

	for _, e := range g.Entities() {
		if e.Kind == kinds.OpenPort {
			portVal, ok := e.GetProperty("port")
			if !ok {
				findings = append(findings, Finding{
					Severity:   SeverityError,
					Message:    fmt.Sprintf("open_port entity %q is missing required 'port' property", e.ID),
					ObjectID:   e.ID,
					ObjectType: ObjectTypeEntity,
				})
				continue
			}

			var portNum int
			switch v := portVal.(type) {
			case int:
				portNum = v
			case int64:
				portNum = int(v)
			case float64:
				portNum = int(v)
			default:
				findings = append(findings, Finding{
					Severity:   SeverityError,
					Message:    fmt.Sprintf("open_port entity %q has non-numeric port value", e.ID),
					ObjectID:   e.ID,
					ObjectType: ObjectTypeEntity,
				})
				continue
			}

			if portNum < 1 || portNum > 65535 {
				findings = append(findings, Finding{
					Severity:   SeverityError,
					Message:    fmt.Sprintf("open_port entity %q has port %d outside valid range 1-65535", e.ID, portNum),
					ObjectID:   e.ID,
					ObjectType: ObjectTypeEntity,
				})
			}
		}
	}

	return findings
}

func ruleValidACLRULEParent(ctx *Context) []Finding {
	g := ctx.Graph.(*core.Graph)
	var findings []Finding

	for _, e := range g.Entities() {
		if e.Kind == kinds.ACLRule {
			if e.Owner == "" {
				findings = append(findings, Finding{
					Severity:   SeverityError,
					Message:    fmt.Sprintf("acl_rule entity %q has no owner", e.ID),
					ObjectID:   e.ID,
					ObjectType: ObjectTypeEntity,
				})
				continue
			}

			ownerEntity, found := g.GetEntity(e.Owner)
			if !found {
				findings = append(findings, Finding{
					Severity:   SeverityError,
					Message:    fmt.Sprintf("acl_rule entity %q has non-existent owner %q", e.ID, e.Owner),
					ObjectID:   e.ID,
					ObjectType: ObjectTypeEntity,
				})
				continue
			}

			if ownerEntity.Kind != kinds.ACL {
				findings = append(findings, Finding{
					Severity:   SeverityError,
					Message:    fmt.Sprintf("acl_rule entity %q must be owned by an acl, but owner %q is of kind %q", e.ID, e.Owner, ownerEntity.Kind),
					ObjectID:   e.ID,
					ObjectType: ObjectTypeEntity,
				})
			}
		}
	}

	return findings
}

func ruleRequiredType(ctx *Context) []Finding {
	g := ctx.Graph.(*core.Graph)
	var findings []Finding

	for _, r := range g.Relations() {
		if r.Type == "" {
			findings = append(findings, Finding{
				Severity:   SeverityError,
				Message:    fmt.Sprintf("relation %q is missing required 'type' property", r.ID),
				ObjectID:   r.ID,
				ObjectType: ObjectTypeRelation,
			})
		}
	}

	return findings
}

func ruleRequiredParticipants(ctx *Context) []Finding {
	g := ctx.Graph.(*core.Graph)
	var findings []Finding

	for _, r := range g.Relations() {
		if r.Participants.Count() < 2 {
			findings = append(findings, Finding{
				Severity:   SeverityError,
				Message:    fmt.Sprintf("relation %q must have at least 2 participants", r.ID),
				ObjectID:   r.ID,
				ObjectType: ObjectTypeRelation,
			})
		}
	}

	return findings
}

func ruleValidType(ctx *Context) []Finding {
	g := ctx.Graph.(*core.Graph)
	s := ctx.Schema.(*schema.Schema)
	var findings []Finding

	for _, r := range g.Relations() {
		if r.Type != "" && !s.HasRelationType(r.Type) {
			findings = append(findings, Finding{
				Severity:   SeverityError,
				Message:    fmt.Sprintf("relation %q has undefined type %q", r.ID, r.Type),
				ObjectID:   r.ID,
				ObjectType: ObjectTypeRelation,
			})
		}
	}

	return findings
}

func ruleValidDirection(ctx *Context) []Finding {
	g := ctx.Graph.(*core.Graph)
	s := ctx.Schema.(*schema.Schema)
	var findings []Finding

	for _, r := range g.Relations() {
		typeDef, ok := s.GetRelationTypeDef(r.Type)
		if !ok {
			continue
		}

		if typeDef.Direction == schema.DirectionDirected {
			if r.Participants.Source == "" || r.Participants.Target == "" {
				findings = append(findings, Finding{
					Severity:   SeverityError,
					Message:    fmt.Sprintf("directed relation %q must have both source and target", r.ID),
					ObjectID:   r.ID,
					ObjectType: ObjectTypeRelation,
				})
			}
		}
	}

	return findings
}

func ruleValidCardinality(ctx *Context) []Finding {
	g := ctx.Graph.(*core.Graph)
	s := ctx.Schema.(*schema.Schema)
	var findings []Finding

	for _, r := range g.Relations() {
		typeDef, ok := s.GetRelationTypeDef(r.Type)
		if !ok || typeDef.Participants == nil {
			continue
		}

		count := r.Participants.Count()
		if typeDef.Participants.MinParticipants > 0 && count < typeDef.Participants.MinParticipants {
			findings = append(findings, Finding{
				Severity:   SeverityError,
				Message:    fmt.Sprintf("relation %q has %d participants, minimum required is %d", r.ID, count, typeDef.Participants.MinParticipants),
				ObjectID:   r.ID,
				ObjectType: ObjectTypeRelation,
			})
		}
		if typeDef.Participants.MaxParticipants > 0 && count > typeDef.Participants.MaxParticipants {
			findings = append(findings, Finding{
				Severity:   SeverityError,
				Message:    fmt.Sprintf("relation %q has %d participants, maximum allowed is %d", r.ID, count, typeDef.Participants.MaxParticipants),
				ObjectID:   r.ID,
				ObjectType: ObjectTypeRelation,
			})
		}
	}

	return findings
}

func ruleValidParticipantKind(ctx *Context) []Finding {
	g := ctx.Graph.(*core.Graph)
	s := ctx.Schema.(*schema.Schema)
	var findings []Finding

	for _, r := range g.Relations() {
		typeDef, ok := s.GetRelationTypeDef(r.Type)
		if !ok || typeDef.Participants == nil {
			continue
		}

		// Directed relations with explicit source and target participants are
		// checked positionally: the source must be a source kind and the target
		// a target kind. Symmetric relations, undirected relations, and
		// list-form participants fall back to the union of source and target
		// kinds (direction mismatches of that shape are handled by
		// valid-direction).
		directional := typeDef.Direction == schema.DirectionDirected && r.Source() != "" && r.Target() != ""
		if directional {
			positional := func(participantID string, allowedKinds []core.EntityKind) {
				entity, found := g.GetEntity(participantID)
				if !found {
					return
				}
				for _, allowedKind := range allowedKinds {
					if entity.Kind == allowedKind {
						return
					}
				}
				findings = append(findings, Finding{
					Severity:   SeverityWarning,
					Message:    fmt.Sprintf("relation %q has participant %q of kind %q which is not a valid %s for type %q", r.ID, participantID, entity.Kind, participantRole(r, participantID), r.Type),
					ObjectID:   r.ID,
					ObjectType: ObjectTypeRelation,
				})
			}
			positional(r.Source(), typeDef.Participants.SourceKinds)
			positional(r.Target(), typeDef.Participants.TargetKinds)
			continue
		}

		for _, participantID := range r.ParticipantIDs() {
			entity, found := g.GetEntity(participantID)
			if !found {
				continue
			}

			kindValid := false
			for _, allowedKind := range typeDef.Participants.SourceKinds {
				if entity.Kind == allowedKind {
					kindValid = true
					break
				}
			}
			if !kindValid {
				for _, allowedKind := range typeDef.Participants.TargetKinds {
					if entity.Kind == allowedKind {
						kindValid = true
						break
					}
				}
			}

			if !kindValid {
				findings = append(findings, Finding{
					Severity:   SeverityWarning,
					Message:    fmt.Sprintf("relation %q has participant %q of kind %q which is not typically allowed for type %q", r.ID, participantID, entity.Kind, r.Type),
					ObjectID:   r.ID,
					ObjectType: ObjectTypeRelation,
				})
			}
		}
	}

	return findings
}

// participantRole returns the role ("source" or "target") a participant plays
// in a directed relation.
func participantRole(r *core.Relation, participantID string) string {
	if r.Source() == participantID {
		return "source"
	}
	return "target"
}

func (e *Engine) ruleOwnershipTree(ctx *Context) []Finding {
	g := ctx.Graph.(*core.Graph)
	var findings []Finding

	roots := rootEntities(g)

	if len(roots) == 0 {
		findings = append(findings, Finding{
			Severity: SeverityError,
			Message:  "no root entity found (ownership tree must have exactly one root)",
		})
		return findings
	}

	if len(roots) > 1 {
		disallowed := e.disallowedRoots(g)
		if len(disallowed) > 0 {
			findings = append(findings, Finding{
				Severity: SeverityError,
				Message:  fmt.Sprintf("multiple root entities found: %v (ownership tree must have exactly one root unless every root is of a root-authorized kind; disallowed roots: %v)", roots, disallowed),
			})
		}
	}

	// Verify the forest reaches every entity from the set of roots.
	visited := make(map[string]bool)
	for _, rootID := range roots {
		visited[rootID] = true
		queue := []string{rootID}
		for len(queue) > 0 {
			current := queue[0]
			queue = queue[1:]
			for _, child := range g.Children(current) {
				if visited[child.ID] {
					continue
				}
				visited[child.ID] = true
				queue = append(queue, child.ID)
			}
		}
	}
	if len(visited) != g.EntityCount() {
		findings = append(findings, Finding{
			Severity: SeverityError,
			Message:  fmt.Sprintf("ownership tree is disconnected: roots %v reach %d of %d entities", roots, len(visited), g.EntityCount()),
		})
	}

	return findings
}

func ruleNoOwnershipCycle(ctx *Context) []Finding {
	g := ctx.Graph.(*core.Graph)
	var findings []Finding

	visited := make(map[string]bool)
	for _, e := range g.Entities() {
		if e.Owner == "" {
			continue
		}
		if err := detectCycle(g, e.ID, visited); err != nil {
			findings = append(findings, Finding{
				Severity:   SeverityError,
				Message:    err.Error(),
				ObjectID:   e.ID,
				ObjectType: ObjectTypeEntity,
			})
		}
	}

	return findings
}

func detectCycle(g *core.Graph, entityID string, visited map[string]bool) error {
	path := make(map[string]bool)
	current := entityID

	for current != "" {
		if path[current] {
			return fmt.Errorf("ownership cycle detected involving entity %q", current)
		}
		path[current] = true

		e, found := g.GetEntity(current)
		if !found {
			break
		}
		current = e.Owner
	}

	for id := range path {
		visited[id] = true
	}

	return nil
}

func (e *Engine) ruleRootEntity(ctx *Context) []Finding {
	g := ctx.Graph.(*core.Graph)
	var findings []Finding

	roots := rootEntities(g)

	if len(roots) == 0 {
		findings = append(findings, Finding{
			Severity: SeverityError,
			Message:  "no root entity found (graph must have exactly one root)",
		})
	} else if len(roots) > 1 {
		sort.Strings(roots)
		disallowed := e.disallowedRoots(g)
		if len(disallowed) > 0 {
			findings = append(findings, Finding{
				Severity: SeverityError,
				Message:  fmt.Sprintf("multiple root entities found: %v (disallowed roots: %v)", roots, disallowed),
			})
		}
	}

	return findings
}

func ruleDanglingReference(ctx *Context) []Finding {
	g := ctx.Graph.(*core.Graph)
	var findings []Finding

	for _, r := range g.Relations() {
		for _, participantID := range r.ParticipantIDs() {
			_, found := g.ResolveReference(participantID)
			if !found {
				findings = append(findings, Finding{
					Severity:   SeverityError,
					Message:    fmt.Sprintf("relation %q references non-existent object %q", r.ID, participantID),
					ObjectID:   r.ID,
					ObjectType: ObjectTypeRelation,
				})
			}
		}
	}

	for _, e := range g.Entities() {
		if e.Owner != "" {
			_, found := g.GetEntity(e.Owner)
			if !found {
				findings = append(findings, Finding{
					Severity:   SeverityError,
					Message:    fmt.Sprintf("entity %q references non-existent owner %q", e.ID, e.Owner),
					ObjectID:   e.ID,
					ObjectType: ObjectTypeEntity,
				})
			}
		}

		// Check entity property references (@ prefix)
		for key, value := range e.Properties {
			for _, targetID := range core.ExtractReferenceTargets(value) {
				if _, found := g.ResolveReference(targetID); !found {
					findings = append(findings, Finding{
						Severity:   SeverityError,
						Message:    fmt.Sprintf("entity %q property %q references non-existent object %q", e.ID, key, targetID),
						ObjectID:   e.ID,
						ObjectType: ObjectTypeEntity,
					})
				}
			}
		}
	}

	// Check relation property references (@ prefix)
	for _, r := range g.Relations() {
		for key, value := range r.Properties {
			for _, targetID := range core.ExtractReferenceTargets(value) {
				if _, found := g.ResolveReference(targetID); !found {
					findings = append(findings, Finding{
						Severity:   SeverityError,
						Message:    fmt.Sprintf("relation %q property %q references non-existent object %q", r.ID, key, targetID),
						ObjectID:   r.ID,
						ObjectType: ObjectTypeRelation,
					})
				}
			}
		}
	}

	return findings
}

func ruleInvalidPath(ctx *Context) []Finding {
	g := ctx.Graph.(*core.Graph)
	var findings []Finding

	for _, e := range g.Entities() {
		path := e.Path()
		if path == "" || path == "/"+e.ID {
			continue
		}

		segments := strings.Split(strings.TrimPrefix(path, "/"), "/")
		if len(segments) == 0 {
			continue
		}

		if segments[len(segments)-1] != e.ID {
			findings = append(findings, Finding{
				Severity:   SeverityError,
				Message:    fmt.Sprintf("entity %q path %q does not end with entity ID", e.ID, path),
				ObjectID:   e.ID,
				ObjectType: ObjectTypeEntity,
				Path:       path,
			})
			continue
		}

		for i := 1; i < len(segments); i++ {
			childID := segments[i]
			parentID := segments[i-1]

			child, ok := g.GetEntity(childID)
			if !ok {
				findings = append(findings, Finding{
					Severity:   SeverityError,
					Message:    fmt.Sprintf("entity %q path %q references non-existent entity %q", e.ID, path, childID),
					ObjectID:   e.ID,
					ObjectType: ObjectTypeEntity,
					Path:       path,
				})
				break
			}
			if child.Owner != parentID {
				findings = append(findings, Finding{
					Severity:   SeverityError,
					Message:    fmt.Sprintf("entity %q path %q: entity %q is not owned by %q", e.ID, path, childID, parentID),
					ObjectID:   e.ID,
					ObjectType: ObjectTypeEntity,
					Path:       path,
				})
				break
			}
		}
	}

	return findings
}

func ruleNoSlashInID(ctx *Context) []Finding {
	g := ctx.Graph.(*core.Graph)
	var findings []Finding

	for _, e := range g.Entities() {
		if strings.Contains(e.ID, "/") {
			findings = append(findings, Finding{
				Severity:   SeverityError,
				Message:    fmt.Sprintf("entity %q has invalid ID containing slash", e.ID),
				ObjectID:   e.ID,
				ObjectType: ObjectTypeEntity,
			})
		}
	}

	return findings
}

func ruleValidNestingParent(ctx *Context) []Finding {
	g := ctx.Graph.(*core.Graph)
	s := ctx.Schema.(*schema.Schema)
	var findings []Finding

	for _, e := range g.Entities() {
		if e.IsRoot() {
			continue
		}
		parent, ok := g.GetEntity(e.Owner)
		if !ok {
			continue
		}

		_, ok = s.FindNestingByChildKind(parent.Kind, e.Kind)
		if !ok {
			findings = append(findings, Finding{
				Severity:   SeverityWarning,
				Message:    fmt.Sprintf("entity %q (kind=%q) is owned by %q (kind=%q) but this parent-child nesting is not defined in the schema", e.ID, e.Kind, parent.ID, parent.Kind),
				ObjectID:   e.ID,
				ObjectType: ObjectTypeEntity,
			})
		}
	}

	return findings
}

// --- Network Rules ---

// entityIPs returns the list of ip_address values configured on an interface.
func entityIPs(e *core.Entity) []string {
	v, ok := e.GetProperty("ip_address")
	if !ok {
		return nil
	}
	var ips []string
	switch val := v.(type) {
	case string:
		if val != "" {
			ips = append(ips, val)
		}
	case []interface{}:
		for _, item := range val {
			if s, ok := item.(string); ok && s != "" {
				ips = append(ips, s)
			}
		}
	case []string:
		ips = append(ips, val...)
	}
	return ips
}

// parseIPEntry parses an ip_address entry which may be a bare IP or CIDR notation.
// It returns the IP address and whether parsing succeeded.
func parseIPEntry(s string) (net.IP, bool) {
	if ip := net.ParseIP(s); ip != nil {
		return ip, true
	}
	if _, ipnet, err := net.ParseCIDR(s); err == nil {
		return ipnet.IP, true
	}
	return nil, false
}

// entityNetworkRef returns the target ID of the interface's network property reference.
func entityNetworkRef(e *core.Entity) (string, bool) {
	v, ok := e.GetProperty("network")
	if !ok {
		return "", false
	}
	if target, ok := core.ExtractReferenceValue(v); ok {
		return target, true
	}
	if s, ok := v.(string); ok {
		return strings.TrimPrefix(s, "@"), true
	}
	return "", false
}

// networkCIDR returns the parsed CIDR of a network entity.
func networkCIDR(n *core.Entity) (*net.IPNet, error) {
	v, ok := n.GetProperty("cidr")
	if !ok {
		return nil, fmt.Errorf("cidr not defined")
	}
	cidrStr, ok := v.(string)
	if !ok {
		return nil, fmt.Errorf("cidr is not a string")
	}
	_, ipnet, err := net.ParseCIDR(cidrStr)
	if err != nil {
		return nil, err
	}
	return ipnet, nil
}

// interfaceNetworks returns the network entities an interface belongs to.
// Membership is determined by the network property and belongs_to relations.
func interfaceNetworks(g *core.Graph, e *core.Entity) []*core.Entity {
	var nets []*core.Entity
	seen := make(map[string]bool)

	if ref, ok := entityNetworkRef(e); ok {
		if n, found := g.GetEntity(ref); found && n.Kind == kinds.Network {
			seen[n.ID] = true
			nets = append(nets, n)
		}
	}

	for _, r := range g.RelationsByType(types.BelongsTo) {
		if r.Participants.Source != e.ID {
			continue
		}
		t, found := g.GetEntity(r.Participants.Target)
		if found && t.Kind == kinds.Network && !seen[t.ID] {
			seen[t.ID] = true
			nets = append(nets, t)
		}
	}

	return nets
}

func ruleValidIPFormat(ctx *Context) []Finding {
	g := ctx.Graph.(*core.Graph)
	var findings []Finding

	for _, e := range g.Entities() {
		for _, ip := range entityIPs(e) {
			if _, ok := parseIPEntry(ip); !ok {
				findings = append(findings, Finding{
					Severity:   SeverityWarning,
					Message:    fmt.Sprintf("entity %q has invalid IP address %q", e.ID, ip),
					ObjectID:   e.ID,
					ObjectType: ObjectTypeEntity,
				})
			}
		}
	}

	return findings
}

func ruleIPRequiresNetwork(ctx *Context) []Finding {
	g := ctx.Graph.(*core.Graph)
	var findings []Finding

	for _, e := range g.Entities() {
		if e.Kind != kinds.Interface {
			continue
		}
		if len(entityIPs(e)) == 0 {
			continue
		}
		if len(interfaceNetworks(g, e)) > 0 {
			continue
		}
		findings = append(findings, Finding{
			Severity:   SeverityWarning,
			Message:    fmt.Sprintf("interface %q has IP addresses but does not reference a network (use the 'network' property or a belongs_to relation to a network)", e.ID),
			ObjectID:   e.ID,
			ObjectType: ObjectTypeEntity,
		})
	}

	return findings
}

func ruleNetworkReferenceKind(ctx *Context) []Finding {
	g := ctx.Graph.(*core.Graph)
	var findings []Finding

	for _, e := range g.Entities() {
		if e.Kind != kinds.Interface {
			continue
		}
		ref, ok := entityNetworkRef(e)
		if !ok {
			continue
		}
		n, found := g.GetEntity(ref)
		if !found {
			continue // non-existent references are handled by dangling-reference
		}
		if n.Kind != kinds.Network {
			findings = append(findings, Finding{
				Severity:   SeverityWarning,
				Message:    fmt.Sprintf("interface %q references %q as its network, but that entity is of kind %q (expected network)", e.ID, ref, n.Kind),
				ObjectID:   e.ID,
				ObjectType: ObjectTypeEntity,
			})
		}
	}

	return findings
}

func ruleIPInCIDR(ctx *Context) []Finding {
	g := ctx.Graph.(*core.Graph)
	var findings []Finding

	for _, e := range g.Entities() {
		if e.Kind != kinds.Interface {
			continue
		}
		ips := entityIPs(e)
		if len(ips) == 0 {
			continue
		}
		nets := interfaceNetworks(g, e)

		for _, ipEntry := range ips {
			ip, ok := parseIPEntry(ipEntry)
			if !ok {
				continue // handled by valid-ip-format
			}

			anyCidr := false
			contained := false
			for _, n := range nets {
				ipnet, err := networkCIDR(n)
				if err != nil {
					continue // handled by network-cidr-required
				}
				anyCidr = true
				if ipnet.Contains(ip) {
					contained = true
					break
				}
			}

			if anyCidr && !contained {
				findings = append(findings, Finding{
					Severity:   SeverityWarning,
					Message:    fmt.Sprintf("interface %q IP address %q is not within the CIDR of any referenced network", e.ID, ipEntry),
					ObjectID:   e.ID,
					ObjectType: ObjectTypeEntity,
				})
			}
		}
	}

	return findings
}

func ruleNetworkCIDRRequired(ctx *Context) []Finding {
	g := ctx.Graph.(*core.Graph)
	var findings []Finding
	flagged := make(map[string]bool)

	for _, e := range g.Entities() {
		if e.Kind != kinds.Interface || len(entityIPs(e)) == 0 {
			continue
		}
		for _, n := range interfaceNetworks(g, e) {
			if flagged[n.ID] {
				continue
			}
			if _, err := networkCIDR(n); err != nil {
				flagged[n.ID] = true
				findings = append(findings, Finding{
					Severity:   SeverityWarning,
					Message:    fmt.Sprintf("network %q has interfaces with IP addresses but does not define a valid 'cidr'", n.ID),
					ObjectID:   n.ID,
					ObjectType: ObjectTypeEntity,
				})
			}
		}
	}

	return findings
}

func ruleGatewayInCIDR(ctx *Context) []Finding {
	g := ctx.Graph.(*core.Graph)
	var findings []Finding

	for _, n := range g.EntitiesByKind(kinds.Network) {
		gwVal, ok := n.GetProperty("gateway")
		if !ok {
			continue
		}
		gwStr, ok := gwVal.(string)
		if !ok {
			continue
		}

		gw := net.ParseIP(gwStr)
		if gw == nil {
			findings = append(findings, Finding{
				Severity:   SeverityWarning,
				Message:    fmt.Sprintf("network %q has invalid gateway address %q", n.ID, gwStr),
				ObjectID:   n.ID,
				ObjectType: ObjectTypeEntity,
			})
			continue
		}

		ipnet, err := networkCIDR(n)
		if err != nil {
			continue // handled by network-cidr-required
		}
		if !ipnet.Contains(gw) {
			findings = append(findings, Finding{
				Severity:   SeverityWarning,
				Message:    fmt.Sprintf("network %q gateway %q is not within its CIDR %q", n.ID, gwStr, ipnet.String()),
				ObjectID:   n.ID,
				ObjectType: ObjectTypeEntity,
			})
		}
	}

	return findings
}

func ruleIPUniqueInNetwork(ctx *Context) []Finding {
	g := ctx.Graph.(*core.Graph)
	var findings []Finding
	seen := make(map[string]map[string]string) // network ID -> IP -> interface ID

	for _, e := range g.Entities() {
		if e.Kind != kinds.Interface {
			continue
		}
		for _, n := range interfaceNetworks(g, e) {
			if seen[n.ID] == nil {
				seen[n.ID] = make(map[string]string)
			}
			for _, ipEntry := range entityIPs(e) {
				ip, ok := parseIPEntry(ipEntry)
				if !ok {
					continue
				}
				key := ip.String()
				if prev, dup := seen[n.ID][key]; dup && prev != e.ID {
					findings = append(findings, Finding{
						Severity:   SeverityWarning,
						Message:    fmt.Sprintf("IP address %q is used by both interface %q and interface %q within network %q", key, prev, e.ID, n.ID),
						ObjectID:   e.ID,
						ObjectType: ObjectTypeEntity,
					})
				} else {
					seen[n.ID][key] = e.ID
				}
			}
		}
	}

	return findings
}
