package aws

import (
	"testing"

	"IACForge/src/core"
	"IACForge/src/core/kinds"
	"IACForge/src/core/types"
	"IACForge/src/extension"
	"IACForge/src/schema"
	"IACForge/src/validation"
)

func newTestSetup(t *testing.T) (*schema.Schema, *validation.Engine, *extension.Manager) {
	t.Helper()
	setup, err := extension.NewSetup("")
	if err != nil {
		t.Fatalf("NewSetup failed: %v", err)
	}
	return setup.Schema, setup.Validation, setup.Manager
}

func kindDef(t *testing.T, kind core.EntityKind) *schema.EntityKindDefinition {
	t.Helper()
	for _, c := range KindDefinitions() {
		if c.Kind == kind {
			return c.Definition
		}
	}
	t.Fatalf("no definition for kind %q", kind)
	return nil
}

func propDef(t *testing.T, def *schema.EntityKindDefinition, name string) *schema.PropertyDefinition {
	t.Helper()
	for i := range def.Properties {
		if def.Properties[i].Name == name {
			return &def.Properties[i]
		}
	}
	t.Fatalf("kind %q has no property %q", def.Description, name)
	return nil
}

func containsKind(list []core.EntityKind, k core.EntityKind) bool {
	for _, item := range list {
		if item == k {
			return true
		}
	}
	return false
}

// --- Kind definition completeness ---

func TestKindDefinitionsCompleteness(t *testing.T) {
	contribs := KindDefinitions()
	if len(contribs) != 47 {
		t.Errorf("expected 47 kind definitions, got %d", len(contribs))
	}

	defined := make(map[core.EntityKind]bool, len(contribs))
	for _, c := range contribs {
		if defined[c.Kind] {
			t.Errorf("duplicate kind definition for %q", c.Kind)
		}
		if c.Definition == nil {
			t.Errorf("nil definition for kind %q", c.Kind)
		}
		defined[c.Kind] = true
	}

	for _, k := range AllKinds() {
		if !defined[k] {
			t.Errorf("kind %q has no definition", k)
		}
	}
	if got := len(AllKinds()); got != 47 {
		t.Errorf("expected 47 kinds in AllKinds, got %d", got)
	}
}

func TestKindNamingConvention(t *testing.T) {
	for _, k := range AllKinds() {
		s := string(k)
		if len(s) < 5 || s[:4] != "aws." {
			t.Errorf("kind %q does not use the aws. namespace prefix", k)
		}
		for _, r := range s {
			if r >= 'A' && r <= 'Z' {
				t.Errorf("kind %q contains an uppercase letter", k)
			}
		}
	}
}

func TestKindDefinitionsSorted(t *testing.T) {
	contribs := KindDefinitions()
	for i := 1; i < len(contribs); i++ {
		if contribs[i-1].Kind > contribs[i].Kind {
			t.Errorf("kind definitions not sorted: %q before %q", contribs[i-1].Kind, contribs[i].Kind)
		}
	}
}

// --- Nesting definitions ---

func TestNestingReferencesDefinedKinds(t *testing.T) {
	valid := make(map[core.EntityKind]bool)
	for _, k := range AllKinds() {
		valid[k] = true
	}
	coreKinds := []core.EntityKind{
		kinds.Region, kinds.Rack, kinds.Server, kinds.Interface, kinds.Cable,
		kinds.PowerDistribution, kinds.Network, kinds.VLAN, kinds.Switch,
		kinds.Router, kinds.Firewall, kinds.ACL, kinds.ACLRule, kinds.VM,
		kinds.Container, kinds.Application, kinds.OpenPort, kinds.Storage,
		kinds.Volume, kinds.Cluster, kinds.AvailabilityZone,
	}
	for _, k := range coreKinds {
		valid[k] = true
	}
	for _, c := range KindDefinitions() {
		for _, nd := range c.Definition.NestingDefs {
			if !valid[nd.ChildKind] {
				t.Errorf("kind %q nesting def %q references undefined child kind %q", c.Kind, nd.NestKey, nd.ChildKind)
			}
		}
	}
}

func TestNestingAutoRelation(t *testing.T) {
	for _, c := range KindDefinitions() {
		for _, nd := range c.Definition.NestingDefs {
			if nd.AutoRelationType != types.BelongsTo {
				t.Errorf("kind %q nesting %q has auto relation type %q, expected belongs_to", c.Kind, nd.NestKey, nd.AutoRelationType)
			}
			if nd.AutoRelationSource != "child" {
				t.Errorf("kind %q nesting %q has auto relation source %q, expected child", c.Kind, nd.NestKey, nd.AutoRelationSource)
			}
		}
	}
}

func TestAccountNestingCoverage(t *testing.T) {
	def := kindDef(t, Account)
	nestKeys := make(map[string]bool)
	for _, nd := range def.NestingDefs {
		if nestKeys[nd.NestKey] {
			t.Errorf("account has duplicate nest key %q", nd.NestKey)
		}
		nestKeys[nd.NestKey] = true
	}
	// Regional account-level resources must be nestable under an account.
	if !nestKeys["regions"] {
		t.Error("account is missing the regions nest key")
	}
	for _, required := range []core.EntityKind{Region, LambdaFunction, SQSQueue, SNSTopic, DynamoDBTable, APIGateway, Route53Zone, CloudWatchAlarm, EventBridgeRule} {
		found := false
		for _, nd := range def.NestingDefs {
			if nd.ChildKind == required {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("account nesting is missing child kind %q", required)
		}
	}
}

func TestRegionNestingIncludesRegionalResources(t *testing.T) {
	def := kindDef(t, Region)
	nestKeys := make(map[string]bool)
	for _, nd := range def.NestingDefs {
		if nestKeys[nd.NestKey] {
			t.Errorf("region has duplicate nest key %q", nd.NestKey)
		}
		nestKeys[nd.NestKey] = true
	}
	for _, required := range []struct {
		key  string
		kind core.EntityKind
	}{
		{"s3_buckets", S3Bucket},
		{"elasticache_clusters", ElastiCache},
		{"efs_filesystems", EFS},
	} {
		if !nestKeys[required.key] {
			t.Errorf("region nesting is missing nest key %q", required.key)
		}
		found := false
		for _, nd := range def.NestingDefs {
			if nd.ChildKind == required.kind {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("region nesting is missing child kind %q", required.kind)
		}
	}
}

func TestRouteGatewayPropertiesAreReferences(t *testing.T) {
	def := kindDef(t, Route)
	for _, name := range []string{"gateway_id", "nat_gateway_id", "transit_gateway_id", "vpc_peering_connection_id"} {
		p := propDef(t, def, name)
		if p.Type != schema.PropertyTypeReference {
			t.Errorf("route property %q must be a reference type, got %q", name, p.Type)
		}
	}
}

func TestEC2NestsApplications(t *testing.T) {
	def := kindDef(t, EC2)
	found := false
	for _, nd := range def.NestingDefs {
		if nd.ChildKind == kinds.Application {
			found = true
			break
		}
	}
	if !found {
		t.Error("ec2 nesting is missing applications child kind")
	}
}

// --- Property constraints spot checks ---

func TestSpotCheckConstraints(t *testing.T) {
	cases := []struct {
		kind      core.EntityKind
		prop      string
		check     func(p *schema.PropertyDefinition)
		checkEnum bool
		wantEnum  int
	}{
		{VPC, "cidr_block", func(p *schema.PropertyDefinition) {
			if !p.Required {
				t.Error("vpc cidr_block should be required")
			}
		}, false, 0},
		{VPC, "tenancy", nil, true, 3},
		{SecurityGroupRule, "from_port", func(p *schema.PropertyDefinition) {
			if p.Constraints == nil || p.Constraints.Min == nil || *p.Constraints.Min != 0 || p.Constraints.Max == nil || *p.Constraints.Max != 65535 {
				t.Error("security_group_rule from_port must be constrained to 0-65535")
			}
		}, false, 0},
		{SecurityGroupRule, "to_port", func(p *schema.PropertyDefinition) {
			if p.Constraints == nil || p.Constraints.Min == nil || *p.Constraints.Min != 0 || p.Constraints.Max == nil || *p.Constraints.Max != 65535 {
				t.Error("security_group_rule to_port must be constrained to 0-65535")
			}
		}, false, 0},
		{LambdaFunction, "memory_size", func(p *schema.PropertyDefinition) {
			if p.Constraints == nil || p.Constraints.Min == nil || *p.Constraints.Min != 128 || p.Constraints.Max == nil || *p.Constraints.Max != 10240 {
				t.Error("lambda memory_size must be constrained to 128-10240")
			}
		}, false, 0},
		{CloudWatchLogGroup, "retention_days", func(p *schema.PropertyDefinition) {
			if p.Constraints == nil || p.Constraints.Min == nil || *p.Constraints.Min != 1 || p.Constraints.Max == nil || *p.Constraints.Max != 3653 {
				t.Error("cloudwatch_log_group retention_days must be constrained to 1-3653")
			}
		}, false, 0},
		{TargetGroup, "port", func(p *schema.PropertyDefinition) {
			if p.Constraints == nil || p.Constraints.Min == nil || *p.Constraints.Min != 1 || p.Constraints.Max == nil || *p.Constraints.Max != 65535 {
				t.Error("target_group port must be constrained to 1-65535")
			}
		}, false, 0},
		{AutoScalingGroup, "min_size", func(p *schema.PropertyDefinition) {
			if p.Constraints == nil || p.Constraints.Min == nil || *p.Constraints.Min != 0 || p.Constraints.Max == nil || *p.Constraints.Max != 5000 {
				t.Error("auto_scaling_group min_size must be constrained to 0-5000")
			}
		}, false, 0},
		{AutoScalingGroup, "max_size", func(p *schema.PropertyDefinition) {
			if p.Constraints == nil || p.Constraints.Min == nil || *p.Constraints.Min != 0 || p.Constraints.Max == nil || *p.Constraints.Max != 5000 {
				t.Error("auto_scaling_group max_size must be constrained to 0-5000")
			}
		}, false, 0},
		{AutoScalingGroup, "desired_capacity", func(p *schema.PropertyDefinition) {
			if p.Constraints == nil || p.Constraints.Min == nil || *p.Constraints.Min != 0 || p.Constraints.Max == nil || *p.Constraints.Max != 5000 {
				t.Error("auto_scaling_group desired_capacity must be constrained to 0-5000")
			}
		}, false, 0},
		{RDS, "engine", nil, true, 8},
		{EBSVolume, "volume_type", nil, true, 7},
		{Subnet, "cidr_block", func(p *schema.PropertyDefinition) {
			if !p.Required {
				t.Error("subnet cidr_block should be required")
			}
		}, false, 0},
	}

	for _, tc := range cases {
		def := kindDef(t, tc.kind)
		p := propDef(t, def, tc.prop)
		if tc.check != nil {
			tc.check(p)
		}
		if tc.checkEnum {
			if p.Constraints == nil || len(p.Constraints.Enum) != tc.wantEnum {
				t.Errorf("kind %q property %q: expected %d enum values, got %v", tc.kind, tc.prop, tc.wantEnum, p.Constraints)
			}
		}
	}
}

// --- Relation definitions ---

func TestNewRelationTypesRegistered(t *testing.T) {
	s, _, _ := newTestSetup(t)
	for _, c := range RelationTypeDefinitions() {
		def, ok := s.GetRelationTypeDef(c.Type)
		if !ok {
			t.Errorf("relation type %q not registered in schema", c.Type)
			continue
		}
		if def.Direction != schema.DirectionDirected {
			t.Errorf("relation type %q expected directed direction, got %q", c.Type, def.Direction)
		}
		if def.Participants == nil || def.Participants.MinParticipants != 2 || def.Participants.MaxParticipants != 2 {
			t.Errorf("relation type %q must be a binary relation with min/max 2 participants", c.Type)
		}
	}
}

func TestRelationParticipantKindsDefined(t *testing.T) {
	valid := make(map[core.EntityKind]bool)
	for _, k := range AllKinds() {
		valid[k] = true
	}
	for _, c := range RelationTypeDefinitions() {
		p := c.Definition.Participants
		for _, k := range p.SourceKinds {
			if !valid[k] {
				t.Errorf("relation %q references undefined source kind %q", c.Type, k)
			}
		}
		for _, k := range p.TargetKinds {
			if !valid[k] {
				t.Errorf("relation %q references undefined target kind %q", c.Type, k)
			}
		}
	}
}

func TestRegistersRelationDefinition(t *testing.T) {
	s, _, _ := newTestSetup(t)
	def, ok := s.GetRelationTypeDef(Registers)
	if !ok {
		t.Fatal("aws.registers not registered in schema")
	}
	if def.Direction != schema.DirectionDirected {
		t.Errorf("aws.registers expected directed, got %q", def.Direction)
	}
	if def.Participants == nil || len(def.Participants.SourceKinds) != 1 || len(def.Participants.TargetKinds) != 2 {
		t.Fatalf("aws.registers participant kinds unexpected: %+v", def.Participants)
	}
	if def.Participants.SourceKinds[0] != TargetGroup {
		t.Errorf("aws.registers source kind must be aws.target_group, got %q", def.Participants.SourceKinds[0])
	}
	if !containsKind(def.Participants.TargetKinds, EC2) || !containsKind(def.Participants.TargetKinds, LambdaFunction) {
		t.Errorf("aws.registers target kinds must include aws.ec2 and aws.lambda_function, got %v", def.Participants.TargetKinds)
	}
}

func TestAttachesTargetsIncludeNATGateway(t *testing.T) {
	s, _, _ := newTestSetup(t)
	def, ok := s.GetRelationTypeDef(Attaches)
	if !ok {
		t.Fatal("aws.attaches not registered in schema")
	}
	if !containsKind(def.Participants.TargetKinds, NATGateway) {
		t.Error("aws.attaches target kinds must include aws.nat_gateway")
	}
}

func TestAugmentDefinitionsTargetCore(t *testing.T) {
	coreTypes := []core.RelationType{types.BelongsTo, types.DependsOn, types.Hosts, types.Monitors, types.BacksUp, types.MountedOn}
	augmented := make(map[core.RelationType]bool)
	for _, c := range AugmentDefinitions() {
		if !c.Augment {
			t.Errorf("augment contribution for %q must set Augment=true", c.Type)
		}
		augmented[c.Type] = true
		if c.Definition.Direction != "" {
			t.Errorf("augment contribution for %q must not change direction", c.Type)
		}
		if len(c.Definition.Properties) > 0 {
			t.Errorf("augment contribution for %q must not add properties", c.Type)
		}
		if c.Definition.Participants == nil {
			t.Errorf("augment contribution for %q requires participants", c.Type)
		}
		if c.Definition.Participants.MinParticipants != 0 || c.Definition.Participants.MaxParticipants != 0 {
			t.Errorf("augment contribution for %q must not change participant limits", c.Type)
		}
	}
	for _, ct := range coreTypes {
		if !augmented[ct] {
			t.Errorf("core relation type %q is missing an augment contribution", ct)
		}
	}
}

func TestAugmentMergesParticipantKinds(t *testing.T) {
	s, _, _ := newTestSetup(t)

	belongsTo, ok := s.GetRelationTypeDef(types.BelongsTo)
	if !ok {
		t.Fatal("belongs_to not in schema")
	}
	if !containsKind(belongsTo.Participants.SourceKinds, VPC) || !containsKind(belongsTo.Participants.TargetKinds, Subnet) {
		t.Error("belongs_to was not augmented with aws participant kinds")
	}

	monitors, ok := s.GetRelationTypeDef(types.Monitors)
	if !ok {
		t.Fatal("monitors not in schema")
	}
	if !containsKind(monitors.Participants.SourceKinds, CloudWatchAlarm) {
		t.Error("monitors source kinds were not augmented with aws.cloudwatch_alarm")
	}
	if !containsKind(monitors.Participants.TargetKinds, EC2) {
		t.Error("monitors target kinds were not augmented with aws.ec2")
	}

	backsUp, ok := s.GetRelationTypeDef(types.BacksUp)
	if !ok {
		t.Fatal("backs_up not in schema")
	}
	if !containsKind(backsUp.Participants.TargetKinds, EBSSnapshot) {
		t.Error("backs_up target kinds were not augmented with aws.ebs_snapshot")
	}
}

// --- Integration: registration + validation of a sample AWS graph ---

func TestValidateSampleGraph(t *testing.T) {
	_, v, _ := newTestSetup(t)

	g := core.NewGraph()
	addEntity := func(id string, kind core.EntityKind, name, owner string, props map[string]interface{}) {
		t.Helper()
		e := core.NewEntity(id, kind, name)
		e.Owner = owner
		for k, val := range props {
			e.SetProperty(k, val)
		}
		if err := g.AddEntity(e); err != nil {
			t.Fatalf("AddEntity %s: %v", id, err)
		}
	}
	addRelation := func(id string, relType core.RelationType, source, target string) {
		t.Helper()
		r := core.NewRelation(id, relType, core.DirectionDirected)
		r.Participants.Source = source
		r.Participants.Target = target
		if err := g.AddRelation(r); err != nil {
			t.Fatalf("AddRelation %s: %v", id, err)
		}
	}

	addEntity("org-01", Organization, "Example Org", "", nil)
	addEntity("acct-01", Account, "Example Account", "org-01", map[string]interface{}{"account_id": "123456789012"})
	addEntity("region-01", Region, "us-east-1", "acct-01", map[string]interface{}{"region_code": "us-east-1"})
	addEntity("az-01", AvailabilityZone, "us-east-1a", "region-01", map[string]interface{}{"zone_name": "us-east-1a"})
	addEntity("vpc-01", VPC, "Main VPC", "region-01", map[string]interface{}{"cidr_block": "10.0.0.0/16"})
	addEntity("subnet-01", Subnet, "Public Subnet", "vpc-01", map[string]interface{}{"cidr_block": "10.0.1.0/24"})
	addEntity("sg-01", SecurityGroup, "Web SG", "vpc-01", nil)
	addEntity("ec2-01", EC2, "Web Server", "subnet-01", map[string]interface{}{
		"instance_type":   "t3.micro",
		"subnet":          core.NewReferenceValue("subnet-01"),
		"security_groups": []interface{}{core.NewReferenceValue("sg-01")},
	})
	addEntity("s3-01", S3Bucket, "Assets Bucket", "az-01", nil)
	addEntity("rds-01", RDS, "Database", "subnet-01", map[string]interface{}{"instance_class": "db.t3.micro"})
	addEntity("lambda-01", LambdaFunction, "Processor", "acct-01", map[string]interface{}{"runtime": "python3.12"})
	addEntity("sqs-01", SQSQueue, "Job Queue", "acct-01", nil)
	addEntity("sns-01", SNSTopic, "Events Topic", "acct-01", nil)
	addEntity("cw-01", CloudWatchAlarm, "CPU Alarm", "acct-01", nil)

	addRelation("rel-ec2-subnet", types.BelongsTo, "ec2-01", "subnet-01")
	addRelation("rel-cw-ec2", types.Monitors, "cw-01", "ec2-01")
	addRelation("rel-sns-sqs", Subscribes, "sns-01", "sqs-01")

	result := v.Validate(g, nil)
	if !result.Passed {
		t.Errorf("expected sample AWS graph to pass validation, got:\n%s", resultSummary(result))
	}
}

func resultSummary(r *validation.Result) string {
	out := ""
	for _, f := range r.Findings {
		out += string(f.Severity) + ": " + f.Message + "\n"
	}
	return out
}
