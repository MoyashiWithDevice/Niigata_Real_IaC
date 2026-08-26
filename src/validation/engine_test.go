package validation

import (
	"strings"
	"testing"

	"IACForge/src/core"
	"IACForge/src/core/kinds"
	"IACForge/src/core/types"
	"IACForge/src/schema"
)

func newTestGraph() *core.Graph {
	g := core.NewGraph()
	region := core.NewEntity("region-01", kinds.Region, "Region 01")
	_ = g.AddEntity(region)
	rack := core.NewEntity("rack-01", kinds.Rack, "Rack 01")
	rack.SetOwner("region-01")
	_ = g.AddEntity(rack)
	server := core.NewEntity("srv-01", kinds.Server, "Server 01")
	server.SetOwner("rack-01")
	_ = g.AddEntity(server)
	return g
}

func newTestEngine() *Engine {
	s := schema.CoreSchema()
	e := NewEngine(s)
	RegisterCoreRules(e)
	return e
}

func TestEngineCoreRulesRegistered(t *testing.T) {
	e := newTestEngine()
	expectedRules := []string{
		"unique-id", "valid-reference", "valid-owner", "single-owner",
		"valid-property",
		"required-kind", "required-name", "valid-kind", "valid-status",
		"valid-port-range", "valid-acl-rule-parent",
		"required-type", "required-participants", "valid-type", "valid-direction",
		"valid-cardinality", "valid-participant-kind",
		"ownership-tree", "no-ownership-cycle", "root-entity",
		"dangling-reference", "invalid-path",
		"valid-ip-format", "ip-requires-network", "network-reference-kind",
		"ip-in-cidr", "network-cidr-required", "gateway-in-cidr",
		"ip-unique-in-network",
	}

	for _, ruleID := range expectedRules {
		if _, ok := e.ruleDefs[ruleID]; !ok {
			t.Errorf("expected rule %q to be registered", ruleID)
		}
	}
}

func TestValidateValidGraph(t *testing.T) {
	e := newTestEngine()
	graph := newTestGraph()

	result := e.Validate(graph, nil)
	if !result.Passed {
		t.Errorf("expected validation to pass, but found errors:")
		for _, f := range result.Findings {
			if f.Severity == SeverityError {
				t.Errorf("  %s: %s", f.RuleID, f.Message)
			}
		}
	}
}

func TestValidateDuplicateEntityID(t *testing.T) {
	e := newTestEngine()
	graph := core.NewGraph()

	region := core.NewEntity("dup-01", kinds.Region, "Region 1")
	if err := graph.AddEntity(region); err != nil {
		t.Fatalf("failed to addentity: %v", err)
	}

	result := e.Validate(graph, nil)
	for _, f := range result.Findings {
		if f.RuleID == "unique-id" && f.Severity == SeverityError {
			t.Errorf("unexpected unique-id error: %s", f.Message)
		}
	}
}

func TestValidateDuplicateRelationID(t *testing.T) {
	e := newTestEngine()
	graph := core.NewGraph()

	region := core.NewEntity("region-01", kinds.Region, "Region 1")
	if err := graph.AddEntity(region); err != nil {
		t.Fatalf("failed to addentity: %v", err)
	}
	server := core.NewEntity("srv-01", kinds.Server, "Server 1")
	server.SetOwner("region-01")
	if err := graph.AddEntity(server); err != nil {
		t.Fatalf("failed to addentity: %v", err)
	}

	r1 := core.NewDirectedRelation("rel-01", types.Hosts, "srv-01", "region-01")
	if err := graph.AddRelation(r1); err != nil {
		t.Fatalf("failed to addrelation: %v", err)
	}

	result := e.Validate(graph, nil)
	for _, f := range result.Findings {
		if f.RuleID == "unique-id" && f.ObjectType == ObjectTypeRelation {
			t.Errorf("unexpected unique-id error: %s", f.Message)
		}
	}
}

func TestValidateMissingKind(t *testing.T) {
	e := newTestEngine()
	graph := core.NewGraph()

	region := core.NewEntity("region-01", "", "Region 1")
	graph.ForceAddEntity(region)

	result := e.Validate(graph, nil)
	found := false
	for _, f := range result.Findings {
		if f.RuleID == "required-kind" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected required-kind error")
	}
}

func TestValidateMissingName(t *testing.T) {
	e := newTestEngine()
	graph := core.NewGraph()

	region := core.NewEntity("region-01", kinds.Region, "")
	graph.ForceAddEntity(region)

	result := e.Validate(graph, nil)
	found := false
	for _, f := range result.Findings {
		if f.RuleID == "required-name" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected required-name error")
	}
}

func TestValidateInvalidKind(t *testing.T) {
	e := newTestEngine()
	graph := core.NewGraph()

	region := core.NewEntity("region-01", "nonexistent_kind", "Region 1")
	graph.ForceAddEntity(region)

	result := e.Validate(graph, nil)
	found := false
	for _, f := range result.Findings {
		if f.RuleID == "valid-kind" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected valid-kind error")
	}
}

func TestValidateInvalidStatus(t *testing.T) {
	e := newTestEngine()
	graph := core.NewGraph()

	region := core.NewEntity("region-01", kinds.Region, "Region 1")
	region.SetStatus("invalid_status")
	if err := graph.AddEntity(region); err != nil {
		t.Fatalf("failed to addentity: %v", err)
	}

	result := e.Validate(graph, nil)
	found := false
	for _, f := range result.Findings {
		if f.RuleID == "valid-status" && f.Severity == SeverityWarning {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected valid-status warning")
	}
}

func TestValidateInvalidPortRange(t *testing.T) {
	e := newTestEngine()
	graph := core.NewGraph()

	region := core.NewEntity("region-01", kinds.Region, "Region 1")
	if err := graph.AddEntity(region); err != nil {
		t.Fatalf("failed to addentity: %v", err)
	}
	server := core.NewEntity("srv-01", kinds.Server, "Server 1")
	server.SetOwner("region-01")
	if err := graph.AddEntity(server); err != nil {
		t.Fatalf("failed to addentity: %v", err)
	}

	port := core.NewEntity("port-01", kinds.OpenPort, "Port 1")
	port.SetOwner("srv-01")
	port.SetProperty("port", 70000)
	if err := graph.AddEntity(port); err != nil {
		t.Fatalf("failed to addentity: %v", err)
	}

	result := e.Validate(graph, nil)
	found := false
	for _, f := range result.Findings {
		if f.RuleID == "valid-port-range" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected valid-port-range error")
	}
}

func TestValidateValidPortRange(t *testing.T) {
	e := newTestEngine()
	graph := core.NewGraph()

	region := core.NewEntity("region-01", kinds.Region, "Region 1")
	if err := graph.AddEntity(region); err != nil {
		t.Fatalf("failed to addentity: %v", err)
	}
	server := core.NewEntity("srv-01", kinds.Server, "Server 1")
	server.SetOwner("region-01")
	if err := graph.AddEntity(server); err != nil {
		t.Fatalf("failed to addentity: %v", err)
	}

	port := core.NewEntity("port-01", kinds.OpenPort, "Port 1")
	port.SetOwner("srv-01")
	port.SetProperty("port", 443)
	if err := graph.AddEntity(port); err != nil {
		t.Fatalf("failed to addentity: %v", err)
	}

	result := e.Validate(graph, nil)
	for _, f := range result.Findings {
		if f.RuleID == "valid-port-range" && f.Severity == SeverityError {
			t.Errorf("unexpected valid-port-range error: %s", f.Message)
		}
	}
}

func TestValidateACLRULEParent(t *testing.T) {
	e := newTestEngine()
	graph := core.NewGraph()

	region := core.NewEntity("region-01", kinds.Region, "Region 1")
	if err := graph.AddEntity(region); err != nil {
		t.Fatalf("failed to addentity: %v", err)
	}
	acl := core.NewEntity("acl-01", kinds.ACL, "ACL 1")
	acl.SetOwner("region-01")
	if err := graph.AddEntity(acl); err != nil {
		t.Fatalf("failed to addentity: %v", err)
	}

	rule := core.NewEntity("rule-01", kinds.ACLRule, "Rule 1")
	rule.SetOwner("acl-01")
	rule.SetProperty("action", "allow")
	if err := graph.AddEntity(rule); err != nil {
		t.Fatalf("failed to addentity: %v", err)
	}

	result := e.Validate(graph, nil)
	for _, f := range result.Findings {
		if f.RuleID == "valid-acl-rule-parent" && f.Severity == SeverityError {
			t.Errorf("unexpected valid-acl-rule-parent error: %s", f.Message)
		}
	}
}

func TestValidateACLRULEWrongParent(t *testing.T) {
	e := newTestEngine()
	graph := core.NewGraph()

	region := core.NewEntity("region-01", kinds.Region, "Region 1")
	if err := graph.AddEntity(region); err != nil {
		t.Fatalf("failed to addentity: %v", err)
	}
	server := core.NewEntity("srv-01", kinds.Server, "Server 1")
	server.SetOwner("region-01")
	if err := graph.AddEntity(server); err != nil {
		t.Fatalf("failed to addentity: %v", err)
	}

	rule := core.NewEntity("rule-01", kinds.ACLRule, "Rule 1")
	rule.SetOwner("srv-01")
	rule.SetProperty("action", "allow")
	if err := graph.AddEntity(rule); err != nil {
		t.Fatalf("failed to addentity: %v", err)
	}

	result := e.Validate(graph, nil)
	found := false
	for _, f := range result.Findings {
		if f.RuleID == "valid-acl-rule-parent" && f.Severity == SeverityError {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected valid-acl-rule-parent error for wrong parent kind")
	}
}

func TestValidateMissingRelationType(t *testing.T) {
	e := newTestEngine()
	graph := core.NewGraph()

	region := core.NewEntity("region-01", kinds.Region, "Region 1")
	if err := graph.AddEntity(region); err != nil {
		t.Fatalf("failed to addentity: %v", err)
	}
	server := core.NewEntity("srv-01", kinds.Server, "Server 1")
	server.SetOwner("region-01")
	if err := graph.AddEntity(server); err != nil {
		t.Fatalf("failed to addentity: %v", err)
	}

	r := core.NewRelation("rel-01", "", core.DirectionDirected)
	r.Participants.Source = "srv-01"
	r.Participants.Target = "region-01"
	graph.ForceAddRelation(r)

	result := e.Validate(graph, nil)
	found := false
	for _, f := range result.Findings {
		if f.RuleID == "required-type" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected required-type error")
	}
}

func TestValidateInvalidRelationType(t *testing.T) {
	e := newTestEngine()
	graph := core.NewGraph()

	region := core.NewEntity("region-01", kinds.Region, "Region 1")
	if err := graph.AddEntity(region); err != nil {
		t.Fatalf("failed to addentity: %v", err)
	}
	server := core.NewEntity("srv-01", kinds.Server, "Server 1")
	server.SetOwner("region-01")
	if err := graph.AddEntity(server); err != nil {
		t.Fatalf("failed to addentity: %v", err)
	}

	r := core.NewDirectedRelation("rel-01", "nonexistent_type", "srv-01", "region-01")
	if err := graph.AddRelation(r); err != nil {
		t.Fatalf("failed to addrelation: %v", err)
	}

	result := e.Validate(graph, nil)
	found := false
	for _, f := range result.Findings {
		if f.RuleID == "valid-type" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected valid-type error")
	}
}

func TestValidateDirectedRelationMissingTarget(t *testing.T) {
	e := newTestEngine()
	graph := core.NewGraph()

	region := core.NewEntity("region-01", kinds.Region, "Region 1")
	if err := graph.AddEntity(region); err != nil {
		t.Fatalf("failed to addentity: %v", err)
	}
	server := core.NewEntity("srv-01", kinds.Server, "Server 1")
	server.SetOwner("region-01")
	if err := graph.AddEntity(server); err != nil {
		t.Fatalf("failed to addentity: %v", err)
	}

	r := core.NewRelation("rel-01", types.Hosts, core.DirectionDirected)
	r.Participants.Source = "srv-01"
	graph.ForceAddRelation(r)

	result := e.Validate(graph, nil)
	found := false
	for _, f := range result.Findings {
		if f.RuleID == "valid-direction" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected valid-direction error for directed relation without target")
	}
}

func TestValidateParticipantKindWarning(t *testing.T) {
	e := newTestEngine()
	graph := core.NewGraph()

	region := core.NewEntity("region-01", kinds.Region, "Region 1")
	if err := graph.AddEntity(region); err != nil {
		t.Fatalf("failed to addentity: %v", err)
	}
	server := core.NewEntity("srv-01", kinds.Server, "Server 1")
	server.SetOwner("region-01")
	if err := graph.AddEntity(server); err != nil {
		t.Fatalf("failed to addentity: %v", err)
	}

	// connects should have interface participants, not server
	r := core.NewDirectedRelation("rel-01", types.Connects, "srv-01", "region-01")
	if err := graph.AddRelation(r); err != nil {
		t.Fatalf("failed to addrelation: %v", err)
	}

	result := e.Validate(graph, nil)
	found := false
	for _, f := range result.Findings {
		if f.RuleID == "valid-participant-kind" && f.Severity == SeverityWarning {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected valid-participant-kind warning for server in connects relation")
	}
}

func TestValidateParticipantKindDirection(t *testing.T) {
	e := newTestEngine()
	graph := core.NewGraph()

	server := core.NewEntity("srv-01", kinds.Server, "Server 1")
	if err := graph.AddEntity(server); err != nil {
		t.Fatalf("failed to addentity: %v", err)
	}
	vm := core.NewEntity("vm-01", kinds.VM, "VM 1")
	if err := graph.AddEntity(vm); err != nil {
		t.Fatalf("failed to addentity: %v", err)
	}

	// hosts source kinds are {server,vm,container,application} and target
	// kinds are {vm,container,application}, so a target of kind server is only
	// caught when the direction is respected.
	r := core.NewDirectedRelation("rel-rev", types.Hosts, "vm-01", "srv-01")
	if err := graph.AddRelation(r); err != nil {
		t.Fatalf("failed to addrelation: %v", err)
	}

	result := e.Validate(graph, nil)
	var found []Finding
	for _, f := range result.Findings {
		if f.RuleID == "valid-participant-kind" && f.Severity == SeverityWarning {
			found = append(found, f)
		}
	}
	if len(found) != 1 {
		t.Fatalf("expected exactly 1 valid-participant-kind warning for reversed hosts relation, got %d", len(found))
	}
	if !strings.Contains(found[0].Message, "target") {
		t.Errorf("expected target role in message, got: %s", found[0].Message)
	}
}

func TestValidateParticipantKindDirectionValid(t *testing.T) {
	e := newTestEngine()
	graph := core.NewGraph()

	server := core.NewEntity("srv-01", kinds.Server, "Server 1")
	if err := graph.AddEntity(server); err != nil {
		t.Fatalf("failed to addentity: %v", err)
	}
	vm := core.NewEntity("vm-01", kinds.VM, "VM 1")
	if err := graph.AddEntity(vm); err != nil {
		t.Fatalf("failed to addentity: %v", err)
	}

	r := core.NewDirectedRelation("rel-ok", types.Hosts, "srv-01", "vm-01")
	if err := graph.AddRelation(r); err != nil {
		t.Fatalf("failed to addrelation: %v", err)
	}

	result := e.Validate(graph, nil)
	for _, f := range result.Findings {
		if f.RuleID == "valid-participant-kind" && f.Severity == SeverityWarning {
			t.Errorf("unexpected valid-participant-kind warning: %s", f.Message)
		}
	}
}

func TestValidateOwnershipCycle(t *testing.T) {
	e := newTestEngine()
	graph := core.NewGraph()

	a := core.NewEntity("a", kinds.Server, "A")
	a.SetOwner("b")
	graph.ForceAddEntity(a)
	b := core.NewEntity("b", kinds.Server, "B")
	b.SetOwner("a")
	graph.ForceAddEntity(b)

	result := e.Validate(graph, nil)
	found := false
	for _, f := range result.Findings {
		if f.RuleID == "no-ownership-cycle" && f.Severity == SeverityError {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected no-ownership-cycle error")
	}
}

func TestValidateMultipleRoots(t *testing.T) {
	e := newTestEngine()
	graph := core.NewGraph()

	root1 := core.NewEntity("root-1", kinds.Region, "Root 1")
	if err := graph.AddEntity(root1); err != nil {
		t.Fatalf("failed to addentity: %v", err)
	}
	root2 := core.NewEntity("root-2", kinds.Region, "Root 2")
	if err := graph.AddEntity(root2); err != nil {
		t.Fatalf("failed to addentity: %v", err)
	}

	result := e.Validate(graph, nil)
	found := false
	for _, f := range result.Findings {
		if f.RuleID == "root-entity" && f.Severity == SeverityError {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected root-entity error for multiple roots")
	}
}

func TestValidateNoRoot(t *testing.T) {
	e := newTestEngine()
	graph := core.NewGraph()

	a := core.NewEntity("a", kinds.Server, "A")
	a.SetOwner("b")
	graph.ForceAddEntity(a)
	b := core.NewEntity("b", kinds.Server, "B")
	b.SetOwner("a")
	graph.ForceAddEntity(b)

	result := e.Validate(graph, nil)
	found := false
	for _, f := range result.Findings {
		if f.RuleID == "root-entity" && f.Severity == SeverityError {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected root-entity error")
	}
}

func TestValidatePathReferenceParticipant(t *testing.T) {
	e := newTestEngine()
	graph := core.NewGraph()

	region := core.NewEntity("region-01", kinds.Region, "Region 1")
	if err := graph.AddEntity(region); err != nil {
		t.Fatalf("failed to addentity: %v", err)
	}
	server := core.NewEntity("srv-01", kinds.Server, "Server 1")
	server.SetOwner("region-01")
	if err := graph.AddEntity(server); err != nil {
		t.Fatalf("failed to addentity: %v", err)
	}
	network := core.NewEntity("net-01", kinds.Network, "Network 1")
	network.SetOwner("region-01")
	if err := graph.AddEntity(network); err != nil {
		t.Fatalf("failed to addentity: %v", err)
	}
	intf := core.NewEntity("eth0", kinds.Interface, "eth0")
	intf.SetOwner("srv-01")
	if err := graph.AddEntity(intf); err != nil {
		t.Fatalf("failed to addentity: %v", err)
	}

	r := core.NewDirectedRelation("rel-01", types.BelongsTo, "srv-01/eth0", "net-01")
	if err := graph.AddRelation(r); err != nil {
		t.Fatalf("failed to addrelation: %v", err)
	}

	result := e.Validate(graph, nil)
	for _, f := range result.Findings {
		if (f.RuleID == "valid-reference" || f.RuleID == "dangling-reference") && f.Severity == SeverityError {
			t.Errorf("unexpected %s error: %s", f.RuleID, f.Message)
		}
	}
}

func TestValidateDanglingReference(t *testing.T) {
	e := newTestEngine()
	graph := core.NewGraph()

	server := core.NewEntity("srv-01", kinds.Server, "Server 1")
	if err := graph.AddEntity(server); err != nil {
		t.Fatalf("failed to addentity: %v", err)
	}

	r := core.NewDirectedRelation("rel-01", types.Hosts, "srv-01", "nonexistent")
	graph.ForceAddRelation(r)

	result := e.Validate(graph, nil)
	found := false
	for _, f := range result.Findings {
		if f.RuleID == "dangling-reference" && f.Severity == SeverityError {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected dangling-reference error")
	}
}

func TestValidateValidRelationParticipants(t *testing.T) {
	e := newTestEngine()
	graph := core.NewGraph()

	region := core.NewEntity("region-01", kinds.Region, "Region 1")
	if err := graph.AddEntity(region); err != nil {
		t.Fatalf("failed to addentity: %v", err)
	}
	server := core.NewEntity("srv-01", kinds.Server, "Server 1")
	server.SetOwner("region-01")
	if err := graph.AddEntity(server); err != nil {
		t.Fatalf("failed to addentity: %v", err)
	}

	r := core.NewDirectedRelation("rel-01", types.Hosts, "srv-01", "region-01")
	if err := graph.AddRelation(r); err != nil {
		t.Fatalf("failed to addrelation: %v", err)
	}

	result := e.Validate(graph, nil)
	for _, f := range result.Findings {
		if f.RuleID == "required-participants" && f.Severity == SeverityError {
			t.Errorf("unexpected required-participants error: %s", f.Message)
		}
	}
}

func TestValidateSummary(t *testing.T) {
	e := newTestEngine()
	graph := core.NewGraph()

	result := e.Validate(graph, nil)

	if result.Summary.TotalRules == 0 {
		t.Error("expected non-zero total rules")
	}
	if result.Summary.TotalFindings != len(result.Findings) {
		t.Errorf("expected summary total findings %d to match findings count %d",
			result.Summary.TotalFindings, len(result.Findings))
	}
}

func TestValidateWithProfile(t *testing.T) {
	e := newTestEngine()
	graph := newTestGraph()

	profile := schema.NewProfile("minimal")
	profile.Rules = []string{"required-kind", "required-name"}

	result := e.Validate(graph, profile)

	if !result.Passed {
		t.Errorf("expected validation to pass with minimal profile, but found errors:")
		for _, f := range result.Findings {
			if f.Severity == SeverityError {
				t.Errorf("  %s: %s", f.RuleID, f.Message)
			}
		}
	}

	for _, f := range result.Findings {
		if f.RuleID != "required-kind" && f.RuleID != "required-name" {
			t.Errorf("unexpected rule %q found in profile-filtered validation", f.RuleID)
		}
	}
}

func TestValidateProfileRequiredKinds(t *testing.T) {
	e := newTestEngine()
	graph := core.NewGraph()

	region := core.NewEntity("region-01", kinds.Region, "Region 1")
	if err := graph.AddEntity(region); err != nil {
		t.Fatalf("failed to addentity: %v", err)
	}

	profile := schema.NewProfile("with-kinds")
	profile.AddRequiredKind("server")

	result := e.Validate(graph, profile)

	found := false
	for _, f := range result.Findings {
		if f.RuleID == "profile-required-kind" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected profile-required-kind error")
	}
}

func TestValidateProfileRequiredRelations(t *testing.T) {
	e := newTestEngine()
	graph := newTestGraph()

	profile := schema.NewProfile("with-relations")
	profile.AddRequiredRelation("connects")

	result := e.Validate(graph, profile)

	found := false
	for _, f := range result.Findings {
		if f.RuleID == "profile-required-relation" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected profile-required-relation error")
	}
}

func TestValidateResultPassed(t *testing.T) {
	e := newTestEngine()
	graph := core.NewGraph()
	if err := graph.AddEntity(core.NewEntity("region-01", kinds.Region, "Region 1")); err != nil {
		t.Fatalf("failed to add entity: %v", err)
	}

	result := e.Validate(graph, nil)

	errorCount := 0
	for _, f := range result.Findings {
		if f.Severity == SeverityError {
			errorCount++
		}
	}

	if errorCount > 0 && result.Passed {
		t.Error("expected Passed=false when there are errors")
	}
	if errorCount == 0 && !result.Passed {
		t.Error("expected Passed=true when there are no errors")
	}
}

func TestDanglingPropertyReference(t *testing.T) {
	e := newTestEngine()
	graph := core.NewGraph()

	region := core.NewEntity("region-01", kinds.Region, "Region 01")
	if err := graph.AddEntity(region); err != nil {
		t.Fatalf("failed to addentity: %v", err)
	}

	net := core.NewEntity("net-mgmt", kinds.Network, "Management Network")
	net.SetOwner("region-01")
	if err := graph.AddEntity(net); err != nil {
		t.Fatalf("failed to addentity: %v", err)
	}

	vlan := core.NewEntity("vlan-100", kinds.VLAN, "VLAN 100")
	vlan.SetOwner("region-01")
	vlan.SetProperty("associated_network", core.NewReferenceValue("@net-mgmt"))
	if err := graph.AddEntity(vlan); err != nil {
		t.Fatalf("failed to addentity: %v", err)
	}

	vlan2 := core.NewEntity("vlan-200", kinds.VLAN, "VLAN 200")
	vlan2.SetOwner("region-01")
	vlan2.SetProperty("associated_network", core.NewReferenceValue("@nonexistent"))
	if err := graph.AddEntity(vlan2); err != nil {
		t.Fatalf("failed to addentity: %v", err)
	}

	result := e.Validate(graph, nil)

	foundDangling := false
	for _, f := range result.Findings {
		if f.RuleID == "dangling-reference" && f.ObjectID == "vlan-200" {
			foundDangling = true
			break
		}
	}
	if !foundDangling {
		t.Error("expected dangling-reference error for vlan-200 property reference")
	}

	// vlan-100 should NOT have a dangling reference error
	for _, f := range result.Findings {
		if f.RuleID == "dangling-reference" && f.ObjectID == "vlan-100" {
			t.Error("vlan-100 should not have dangling-reference error")
		}
	}
}

func TestDanglingReferenceInListProperty(t *testing.T) {
	e := newTestEngine()
	graph := core.NewGraph()

	region := core.NewEntity("region-01", kinds.Region, "Region 01")
	if err := graph.AddEntity(region); err != nil {
		t.Fatalf("failed to addentity: %v", err)
	}

	net := core.NewEntity("net-mgmt", kinds.Network, "Management Network")
	net.SetOwner("region-01")
	if err := graph.AddEntity(net); err != nil {
		t.Fatalf("failed to addentity: %v", err)
	}

	sg := core.NewEntity("sg-01", kinds.ACL, "Web SG")
	sg.SetOwner("region-01")
	if err := graph.AddEntity(sg); err != nil {
		t.Fatalf("failed to addentity: %v", err)
	}

	// ec2-01 has one valid list reference (sg-01) and one dangling (ghost).
	ec2 := core.NewEntity("ec2-01", kinds.Server, "Web Server")
	ec2.SetOwner("region-01")
	ec2.SetProperty("security_groups", []interface{}{
		core.NewReferenceValue("@sg-01"),
		core.NewReferenceValue("@ghost"),
	})
	if err := graph.AddEntity(ec2); err != nil {
		t.Fatalf("failed to addentity: %v", err)
	}

	// rds-01 has a nested-map reference that is dangling (nonexistent).
	rds := core.NewEntity("rds-01", kinds.Server, "Database")
	rds.SetOwner("region-01")
	rds.SetProperty("vpc_config", map[string]interface{}{
		"subnets": []interface{}{
			core.NewReferenceValue("@net-mgmt"),
			core.NewReferenceValue("@missing-subnet"),
		},
	})
	if err := graph.AddEntity(rds); err != nil {
		t.Fatalf("failed to addentity: %v", err)
	}

	// lb-01 has only valid list references.
	lb := core.NewEntity("lb-01", kinds.Server, "Load Balancer")
	lb.SetOwner("region-01")
	lb.SetProperty("subnets", []interface{}{
		core.NewReferenceValue("@net-mgmt"),
	})
	if err := graph.AddEntity(lb); err != nil {
		t.Fatalf("failed to addentity: %v", err)
	}

	result := e.Validate(graph, nil)

	findings := map[string]bool{}
	for _, f := range result.Findings {
		if f.RuleID == "dangling-reference" {
			findings[f.Message] = true
		}
	}

	if !findings["entity \"ec2-01\" property \"security_groups\" references non-existent object \"ghost\""] {
		t.Error("expected dangling-reference error for list element @ghost on ec2-01")
	}
	if !findings["entity \"rds-01\" property \"vpc_config\" references non-existent object \"missing-subnet\""] {
		t.Error("expected dangling-reference error for nested-map element @missing-subnet on rds-01")
	}
	for _, f := range result.Findings {
		if f.RuleID == "dangling-reference" && f.ObjectID == "lb-01" {
			t.Errorf("lb-01 should not have dangling-reference error, got: %v", f.Message)
		}
	}
}

func TestValidPropertyReference(t *testing.T) {
	e := newTestEngine()
	graph := core.NewGraph()

	region := core.NewEntity("region-01", kinds.Region, "Region 01")
	if err := graph.AddEntity(region); err != nil {
		t.Fatalf("failed to addentity: %v", err)
	}

	net := core.NewEntity("net-mgmt", kinds.Network, "Management Network")
	net.SetOwner("region-01")
	if err := graph.AddEntity(net); err != nil {
		t.Fatalf("failed to addentity: %v", err)
	}

	vlan := core.NewEntity("vlan-100", kinds.VLAN, "VLAN 100")
	vlan.SetOwner("region-01")
	vlan.SetProperty("associated_network", core.NewReferenceValue("@net-mgmt"))
	if err := graph.AddEntity(vlan); err != nil {
		t.Fatalf("failed to addentity: %v", err)
	}

	result := e.Validate(graph, nil)

	for _, f := range result.Findings {
		if f.RuleID == "dangling-reference" && f.ObjectID == "vlan-100" {
			t.Error("valid property reference should not cause dangling-reference error")
		}
	}
}

func TestInvalidPathNonExistentEntity(t *testing.T) {
	e := newTestEngine()
	graph := core.NewGraph()

	region := core.NewEntity("region-01", kinds.Region, "Region 01")
	region.SetPath("/region-01")
	if err := graph.AddEntity(region); err != nil {
		t.Fatalf("failed to addentity: %v", err)
	}

	server := core.NewEntity("srv-01", kinds.Server, "Server 01")
	server.SetOwner("region-01")
	server.SetPath("/region-01/nonexistent/srv-01")
	if err := graph.AddEntity(server); err != nil {
		t.Fatalf("failed to addentity: %v", err)
	}

	result := e.Validate(graph, nil)
	found := false
	for _, f := range result.Findings {
		if f.RuleID == "invalid-path" && f.Severity == SeverityError {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected invalid-path error for path referencing non-existent entity")
	}
}

func TestInvalidPathWrongOwnership(t *testing.T) {
	e := newTestEngine()
	graph := core.NewGraph()

	region := core.NewEntity("region-01", kinds.Region, "Region 01")
	region.SetPath("/region-01")
	if err := graph.AddEntity(region); err != nil {
		t.Fatalf("failed to addentity: %v", err)
	}

	rack := core.NewEntity("rack-01", kinds.Rack, "Rack 01")
	rack.SetOwner("region-01")
	rack.SetPath("/region-01/rack-01")
	if err := graph.AddEntity(rack); err != nil {
		t.Fatalf("failed to addentity: %v", err)
	}

	server := core.NewEntity("srv-01", kinds.Server, "Server 01")
	server.SetOwner("rack-01")
	server.SetPath("/region-01/rack-01/srv-01")
	if err := graph.AddEntity(server); err != nil {
		t.Fatalf("failed to addentity: %v", err)
	}

	// Manually set wrong path (not matching ownership)
	server.SetPath("/region-01/srv-01")
	result := e.Validate(graph, nil)
	found := false
	for _, f := range result.Findings {
		if f.RuleID == "invalid-path" && f.Severity == SeverityError {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected invalid-path error for path with wrong ownership")
	}
}

func TestValidPath(t *testing.T) {
	e := newTestEngine()
	graph := core.NewGraph()

	region := core.NewEntity("region-01", kinds.Region, "Region 01")
	region.SetPath("/region-01")
	if err := graph.AddEntity(region); err != nil {
		t.Fatalf("failed to addentity: %v", err)
	}

	rack := core.NewEntity("rack-01", kinds.Rack, "Rack 01")
	rack.SetOwner("region-01")
	rack.SetPath("/region-01/rack-01")
	if err := graph.AddEntity(rack); err != nil {
		t.Fatalf("failed to addentity: %v", err)
	}

	server := core.NewEntity("srv-01", kinds.Server, "Server 01")
	server.SetOwner("rack-01")
	server.SetPath("/region-01/rack-01/srv-01")
	if err := graph.AddEntity(server); err != nil {
		t.Fatalf("failed to addentity: %v", err)
	}

	result := e.Validate(graph, nil)
	for _, f := range result.Findings {
		if f.RuleID == "invalid-path" && f.Severity == SeverityError {
			t.Errorf("unexpected invalid-path error: %s", f.Message)
		}
	}
}

// --- Network Rules ---

func newNetworkGraph() *core.Graph {
	g := core.NewGraph()
	region := core.NewEntity("region-01", kinds.Region, "Region 01")
	_ = g.AddEntity(region)
	server := core.NewEntity("srv-01", kinds.Server, "Server 01")
	server.SetOwner("region-01")
	_ = g.AddEntity(server)
	net := core.NewEntity("net-01", kinds.Network, "Network 01")
	net.SetOwner("region-01")
	net.SetProperty("cidr", "10.0.1.0/24")
	_ = g.AddEntity(net)
	return g
}

func newServerInterface(graph *core.Graph, id string) *core.Entity {
	intf := core.NewEntity(id, kinds.Interface, id)
	intf.SetOwner("srv-01")
	_ = graph.AddEntity(intf)
	return intf
}

func hasFindingByRule(result *Result, ruleID string) bool {
	for _, f := range result.Findings {
		if f.RuleID == ruleID {
			return true
		}
	}
	return false
}

func TestRuleValidIPFormat(t *testing.T) {
	e := newTestEngine()
	graph := core.NewGraph()
	if err := graph.AddEntity(core.NewEntity("region-01", kinds.Region, "Region 01")); err != nil {
		t.Fatalf("failed to add entity: %v", err)
	}

	valid := core.NewEntity("eth0", kinds.Interface, "eth0")
	valid.SetOwner("region-01")
	valid.SetProperty("ip_address", "10.0.1.10")
	if err := graph.AddEntity(valid); err != nil {
		t.Fatalf("failed to addentity: %v", err)
	}

	invalid := core.NewEntity("eth1", kinds.Interface, "eth1")
	invalid.SetOwner("region-01")
	invalid.SetProperty("ip_address", []interface{}{"10.0.1.10", "not-an-ip"})
	if err := graph.AddEntity(invalid); err != nil {
		t.Fatalf("failed to addentity: %v", err)
	}

	result := e.Validate(graph, nil)
	if hasFindingByRule(result, "valid-ip-format") == false {
		t.Error("expected valid-ip-format warning for invalid IP address")
	}
	for _, f := range result.Findings {
		if f.RuleID == "valid-ip-format" && f.ObjectID == "eth0" {
			t.Error("eth0 should not have valid-ip-format warning")
		}
		if f.RuleID == "valid-ip-format" && f.Severity != SeverityWarning {
			t.Errorf("valid-ip-format should be warning severity, got %s", f.Severity)
		}
	}
}

func TestRuleIPRequiresNetwork(t *testing.T) {
	e := newTestEngine()
	graph := newNetworkGraph()

	withNetwork := newServerInterface(graph, "eth0")
	withNetwork.SetProperty("ip_address", []interface{}{"10.0.1.10"})
	withNetwork.SetProperty("network", core.NewReferenceValue("@net-01"))

	withoutNetwork := newServerInterface(graph, "eth1")
	withoutNetwork.SetProperty("ip_address", []interface{}{"10.0.1.11"})

	newServerInterface(graph, "eth2")

	result := e.Validate(graph, nil)
	if !hasFindingByRule(result, "ip-requires-network") {
		t.Error("expected ip-requires-network warning for interface without network")
	}
	for _, f := range result.Findings {
		if f.RuleID == "ip-requires-network" && f.ObjectID == "eth0" {
			t.Error("eth0 references a network and should not have ip-requires-network warning")
		}
		if f.RuleID == "ip-requires-network" && f.ObjectID == "eth2" {
			t.Error("eth2 has no IP address and should not have ip-requires-network warning")
		}
	}
}

func TestRuleIPRequiresNetworkViaBelongsTo(t *testing.T) {
	e := newTestEngine()
	graph := newNetworkGraph()

	intf := newServerInterface(graph, "eth0")
	intf.SetProperty("ip_address", []interface{}{"10.0.1.10"})

	rel := core.NewDirectedRelation("rel-intf-net", types.BelongsTo, "eth0", "net-01")
	if err := graph.AddRelation(rel); err != nil {
		t.Fatalf("failed to addrelation: %v", err)
	}

	result := e.Validate(graph, nil)
	if hasFindingByRule(result, "ip-requires-network") {
		t.Error("interface with belongs_to relation to a network should not have ip-requires-network warning")
	}
}

func TestRuleNetworkReferenceKind(t *testing.T) {
	e := newTestEngine()
	graph := newNetworkGraph()

	// net-01 is referenced correctly (kind network); create a server reference too
	intf := newServerInterface(graph, "eth0")
	intf.SetProperty("ip_address", []interface{}{"10.0.1.10"})
	intf.SetProperty("network", core.NewReferenceValue("@net-01"))

	bad := newServerInterface(graph, "eth1")
	bad.SetProperty("ip_address", []interface{}{"10.0.1.11"})
	bad.SetProperty("network", core.NewReferenceValue("@srv-01"))

	result := e.Validate(graph, nil)
	if !hasFindingByRule(result, "network-reference-kind") {
		t.Error("expected network-reference-kind warning for network property referencing non-network entity")
	}
	for _, f := range result.Findings {
		if f.RuleID == "network-reference-kind" && f.ObjectID == "eth0" {
			t.Error("eth0 references a network and should not have network-reference-kind warning")
		}
	}
}

func TestRuleIPInCIDR(t *testing.T) {
	e := newTestEngine()
	graph := newNetworkGraph()

	inRange := newServerInterface(graph, "eth0")
	inRange.SetProperty("ip_address", []interface{}{"10.0.1.10/24"})
	inRange.SetProperty("network", core.NewReferenceValue("@net-01"))

	outOfRange := newServerInterface(graph, "eth1")
	outOfRange.SetProperty("ip_address", []interface{}{"192.168.0.10"})
	outOfRange.SetProperty("network", core.NewReferenceValue("@net-01"))

	result := e.Validate(graph, nil)
	if !hasFindingByRule(result, "ip-in-cidr") {
		t.Error("expected ip-in-cidr warning for IP outside network CIDR")
	}
	for _, f := range result.Findings {
		if f.RuleID == "ip-in-cidr" && f.ObjectID == "eth0" {
			t.Error("eth0 IP is within network CIDR and should not have ip-in-cidr warning")
		}
	}
}

func TestRuleNetworkCIDRRequired(t *testing.T) {
	e := newTestEngine()
	graph := newNetworkGraph()

	net2 := core.NewEntity("net-02", kinds.Network, "Network 02")
	net2.SetOwner("region-01")
	if err := graph.AddEntity(net2); err != nil {
		t.Fatalf("failed to addentity: %v", err)
	}

	withNetwork := newServerInterface(graph, "eth0")
	withNetwork.SetProperty("ip_address", []interface{}{"10.0.1.10"})
	withNetwork.SetProperty("network", core.NewReferenceValue("@net-01"))

	noCidr := newServerInterface(graph, "eth1")
	noCidr.SetProperty("ip_address", []interface{}{"10.0.2.10"})
	noCidr.SetProperty("network", core.NewReferenceValue("@net-02"))

	result := e.Validate(graph, nil)
	if !hasFindingByRule(result, "network-cidr-required") {
		t.Error("expected network-cidr-required warning for network without cidr that has IP members")
	}
	for _, f := range result.Findings {
		if f.RuleID == "network-cidr-required" && f.ObjectID == "net-01" {
			t.Error("net-01 defines a cidr and should not have network-cidr-required warning")
		}
	}
}

func TestRuleGatewayInCIDR(t *testing.T) {
	e := newTestEngine()
	graph := newNetworkGraph()

	good := core.NewEntity("net-good", kinds.Network, "Good Network")
	good.SetOwner("region-01")
	good.SetProperty("cidr", "10.0.1.0/24")
	good.SetProperty("gateway", "10.0.1.1")
	if err := graph.AddEntity(good); err != nil {
		t.Fatalf("failed to addentity: %v", err)
	}

	bad := core.NewEntity("net-bad", kinds.Network, "Bad Network")
	bad.SetOwner("region-01")
	bad.SetProperty("cidr", "10.0.1.0/24")
	bad.SetProperty("gateway", "192.168.0.1")
	if err := graph.AddEntity(bad); err != nil {
		t.Fatalf("failed to addentity: %v", err)
	}

	invalid := core.NewEntity("net-invalid", kinds.Network, "Invalid Gateway Network")
	invalid.SetOwner("region-01")
	invalid.SetProperty("cidr", "10.0.1.0/24")
	invalid.SetProperty("gateway", "not-an-ip")
	if err := graph.AddEntity(invalid); err != nil {
		t.Fatalf("failed to addentity: %v", err)
	}

	result := e.Validate(graph, nil)
	if !hasFindingByRule(result, "gateway-in-cidr") {
		t.Error("expected gateway-in-cidr warning for gateway outside CIDR or invalid gateway")
	}
	for _, f := range result.Findings {
		if f.RuleID == "gateway-in-cidr" && f.ObjectID == "net-good" {
			t.Error("net-good gateway is within CIDR and should not have gateway-in-cidr warning")
		}
	}
}

func TestRuleIPUniqueInNetwork(t *testing.T) {
	e := newTestEngine()
	graph := newNetworkGraph()

	intf1 := newServerInterface(graph, "eth0")
	intf1.SetProperty("ip_address", []interface{}{"10.0.1.10"})
	intf1.SetProperty("network", core.NewReferenceValue("@net-01"))

	intf2 := newServerInterface(graph, "eth1")
	intf2.SetProperty("ip_address", []interface{}{"10.0.1.10"})
	intf2.SetProperty("network", core.NewReferenceValue("@net-01"))

	intf3 := newServerInterface(graph, "eth2")
	intf3.SetProperty("ip_address", []interface{}{"10.0.1.11"})
	intf3.SetProperty("network", core.NewReferenceValue("@net-01"))

	result := e.Validate(graph, nil)
	if !hasFindingByRule(result, "ip-unique-in-network") {
		t.Error("expected ip-unique-in-network warning for duplicate IP within a network")
	}
}

func TestNetworkRulesNoWarningsForCompliantGraph(t *testing.T) {
	e := newTestEngine()
	graph := newNetworkGraph()

	intf := newServerInterface(graph, "eth0")
	intf.SetProperty("ip_address", []interface{}{"10.0.1.10"})
	intf.SetProperty("network", core.NewReferenceValue("@net-01"))

	result := e.Validate(graph, nil)
	networkRules := map[string]bool{
		"valid-ip-format": false, "ip-requires-network": false,
		"network-reference-kind": false, "ip-in-cidr": false,
		"network-cidr-required": false, "gateway-in-cidr": false,
		"ip-unique-in-network": false,
	}
	for _, f := range result.Findings {
		if _, ok := networkRules[f.RuleID]; ok {
			t.Errorf("unexpected %s warning: %s", f.RuleID, f.Message)
		}
	}
}

// --- Valid Property Rule ---

func hasFindingFor(result *Result, ruleID, objectID string) bool {
	for _, f := range result.Findings {
		if f.RuleID == ruleID && f.ObjectID == objectID {
			return true
		}
	}
	return false
}

func TestValidPropertyCompliant(t *testing.T) {
	e := newTestEngine()
	graph := core.NewGraph()
	region := core.NewEntity("region-01", kinds.Region, "Region 1")
	if err := graph.AddEntity(region); err != nil {
		t.Fatalf("failed to addentity: %v", err)
	}
	server := core.NewEntity("srv-01", kinds.Server, "Server 1")
	server.SetOwner("region-01")
	server.SetProperty("platform", "proxmox")
	if err := graph.AddEntity(server); err != nil {
		t.Fatalf("failed to addentity: %v", err)
	}

	result := e.Validate(graph, nil)
	for _, f := range result.Findings {
		if f.RuleID == "valid-property" {
			t.Errorf("unexpected valid-property warning: %s", f.Message)
		}
	}
}

func TestValidPropertyTypeMismatch(t *testing.T) {
	e := newTestEngine()
	graph := core.NewGraph()
	region := core.NewEntity("region-01", kinds.Region, "Region 1")
	if err := graph.AddEntity(region); err != nil {
		t.Fatalf("failed to addentity: %v", err)
	}
	rack := core.NewEntity("rack-01", kinds.Rack, "Rack 1")
	rack.SetOwner("region-01")
	rack.SetProperty("height_units", "42") // defined as integer
	if err := graph.AddEntity(rack); err != nil {
		t.Fatalf("failed to addentity: %v", err)
	}

	result := e.Validate(graph, nil)
	if !hasFindingFor(result, "valid-property", "rack-01") {
		t.Error("expected valid-property warning for rack-01 with non-integer height_units")
	}
}

func TestValidPropertyEnumViolation(t *testing.T) {
	e := newTestEngine()
	graph := core.NewGraph()
	region := core.NewEntity("region-01", kinds.Region, "Region 1")
	if err := graph.AddEntity(region); err != nil {
		t.Fatalf("failed to addentity: %v", err)
	}
	intf := core.NewEntity("eth0", kinds.Interface, "eth0")
	intf.SetOwner("region-01")
	intf.SetProperty("type", "ethernet")
	if err := graph.AddEntity(intf); err != nil {
		t.Fatalf("failed to addentity: %v", err)
	}
	bad := core.NewEntity("eth1", kinds.Interface, "eth1")
	bad.SetOwner("region-01")
	bad.SetProperty("type", "warp") // not in enum
	if err := graph.AddEntity(bad); err != nil {
		t.Fatalf("failed to addentity: %v", err)
	}

	result := e.Validate(graph, nil)
	if !hasFindingFor(result, "valid-property", "eth1") {
		t.Error("expected valid-property warning for eth1 with invalid interface type")
	}
	if hasFindingFor(result, "valid-property", "eth0") {
		t.Error("eth0 with valid interface type should not have valid-property warning")
	}
}

func TestValidPropertyMissingRequired(t *testing.T) {
	e := newTestEngine()
	graph := core.NewGraph()
	region := core.NewEntity("region-01", kinds.Region, "Region 1")
	if err := graph.AddEntity(region); err != nil {
		t.Fatalf("failed to addentity: %v", err)
	}
	vlan := core.NewEntity("vlan-100", kinds.VLAN, "VLAN 100")
	vlan.SetOwner("region-01")
	// vlan_id is required but not set
	if err := graph.AddEntity(vlan); err != nil {
		t.Fatalf("failed to addentity: %v", err)
	}

	result := e.Validate(graph, nil)
	if !hasFindingFor(result, "valid-property", "vlan-100") {
		t.Error("expected valid-property warning for vlan-100 missing required vlan_id")
	}
}

func TestValidPropertyUndefinedProperty(t *testing.T) {
	e := newTestEngine()
	graph := core.NewGraph()
	region := core.NewEntity("region-01", kinds.Region, "Region 1")
	if err := graph.AddEntity(region); err != nil {
		t.Fatalf("failed to addentity: %v", err)
	}
	server := core.NewEntity("srv-01", kinds.Server, "Server 1")
	server.SetOwner("region-01")
	server.SetProperty("not_a_defined_property", "value")
	if err := graph.AddEntity(server); err != nil {
		t.Fatalf("failed to addentity: %v", err)
	}

	result := e.Validate(graph, nil)
	if !hasFindingFor(result, "valid-property", "srv-01") {
		t.Error("expected valid-property warning for undefined property on srv-01")
	}
}

func TestValidPropertyRelationProperty(t *testing.T) {
	e := newTestEngine()
	graph := core.NewGraph()
	region := core.NewEntity("region-01", kinds.Region, "Region 1")
	if err := graph.AddEntity(region); err != nil {
		t.Fatalf("failed to addentity: %v", err)
	}
	server := core.NewEntity("srv-01", kinds.Server, "Server 1")
	server.SetOwner("region-01")
	if err := graph.AddEntity(server); err != nil {
		t.Fatalf("failed to addentity: %v", err)
	}
	app := core.NewEntity("app-01", kinds.Application, "App 1")
	app.SetOwner("srv-01")
	if err := graph.AddEntity(app); err != nil {
		t.Fatalf("failed to addentity: %v", err)
	}

	// depends_on defines dependency_type (string) and critical (boolean).
	bad := core.NewDirectedRelation("rel-bad", types.DependsOn, "app-01", "srv-01")
	bad.SetProperty("critical", "not-a-bool")
	if err := graph.AddRelation(bad); err != nil {
		t.Fatalf("failed to addrelation: %v", err)
	}
	good := core.NewDirectedRelation("rel-good", types.DependsOn, "app-01", "srv-01")
	good.SetProperty("critical", true)
	if err := graph.AddRelation(good); err != nil {
		t.Fatalf("failed to addrelation: %v", err)
	}

	result := e.Validate(graph, nil)
	if !hasFindingFor(result, "valid-property", "rel-bad") {
		t.Error("expected valid-property warning for relation with non-boolean critical")
	}
	if hasFindingFor(result, "valid-property", "rel-good") {
		t.Error("relation with valid critical boolean should not have valid-property warning")
	}
}

func TestValidPropertyCoreKindEnumViolation(t *testing.T) {
	e := newTestEngine()
	graph := core.NewGraph()
	region := core.NewEntity("region-01", kinds.Region, "Region 1")
	if err := graph.AddEntity(region); err != nil {
		t.Fatalf("failed to addentity: %v", err)
	}
	net := core.NewEntity("net-01", kinds.Network, "Network 1")
	net.SetOwner("region-01")
	net.SetProperty("network_type", "carrier") // not in enum
	if err := graph.AddEntity(net); err != nil {
		t.Fatalf("failed to addentity: %v", err)
	}
	good := core.NewEntity("net-02", kinds.Network, "Network 2")
	good.SetOwner("region-01")
	good.SetProperty("network_type", "storage")
	if err := graph.AddEntity(good); err != nil {
		t.Fatalf("failed to addentity: %v", err)
	}

	result := e.Validate(graph, nil)
	if !hasFindingFor(result, "valid-property", "net-01") {
		t.Error("expected valid-property warning for net-01 with invalid network_type")
	}
	if hasFindingFor(result, "valid-property", "net-02") {
		t.Error("net-02 with valid network_type should not have valid-property warning")
	}
}

func TestValidPropertyListUnknownKeyWarning(t *testing.T) {
	e := newTestEngine()
	graph := core.NewGraph()
	region := core.NewEntity("region-01", kinds.Region, "Region 1")
	if err := graph.AddEntity(region); err != nil {
		t.Fatalf("failed to addentity: %v", err)
	}
	server := core.NewEntity("srv-01", kinds.Server, "Server 1")
	server.SetOwner("region-01")
	server.SetProperty("cpu", []interface{}{
		map[string]interface{}{"cores": 4, "architecture": "x86_64"},
		map[string]interface{}{"cores": 8, "sockets": 2}, // unknown key
	})
	if err := graph.AddEntity(server); err != nil {
		t.Fatalf("failed to addentity: %v", err)
	}

	result := e.Validate(graph, nil)
	if !hasFindingFor(result, "valid-property", "srv-01") {
		t.Error("expected valid-property warning for srv-01 cpu item with unknown key")
	}
}

// --- Root-Authorized Kind Hook ---

func TestAllowedRootKindsMethods(t *testing.T) {
	e := NewEngine(schema.CoreSchema())
	if e.IsAllowedRootKind("region") {
		t.Error("region should not be an allowed root kind by default")
	}
	e.AddAllowedRootKind("aws.organization")
	e.AddAllowedRootKind("aws.organization") // idempotent
	e.AddAllowedRootKind("aws.account")

	if !e.IsAllowedRootKind("aws.organization") {
		t.Error("aws.organization should be an allowed root kind")
	}
	if e.IsAllowedRootKind("region") {
		t.Error("region should not be an allowed root kind by default")
	}
}

func TestMultipleRootsAllAllowedKinds(t *testing.T) {
	e := newTestEngine()
	e.AddAllowedRootKind("aws.organization")

	graph := core.NewGraph()
	org1 := core.NewEntity("org-1", core.EntityKind("aws.organization"), "Org 1")
	if err := graph.AddEntity(org1); err != nil {
		t.Fatalf("failed to addentity: %v", err)
	}
	org2 := core.NewEntity("org-2", core.EntityKind("aws.organization"), "Org 2")
	if err := graph.AddEntity(org2); err != nil {
		t.Fatalf("failed to addentity: %v", err)
	}

	result := e.Validate(graph, nil)
	for _, f := range result.Findings {
		if f.RuleID == "root-entity" || f.RuleID == "single-owner" || f.RuleID == "ownership-tree" {
			t.Errorf("unexpected %s error for all-authorized roots: %s", f.RuleID, f.Message)
		}
	}
}

func TestMultipleRootsMixedKinds(t *testing.T) {
	e := newTestEngine()
	e.AddAllowedRootKind("aws.organization")

	graph := core.NewGraph()
	org := core.NewEntity("org-1", core.EntityKind("aws.organization"), "Org 1")
	if err := graph.AddEntity(org); err != nil {
		t.Fatalf("failed to addentity: %v", err)
	}
	region := core.NewEntity("region-01", kinds.Region, "Region 1")
	if err := graph.AddEntity(region); err != nil {
		t.Fatalf("failed to addentity: %v", err)
	}

	result := e.Validate(graph, nil)
	for _, f := range result.Findings {
		if f.RuleID == "root-entity" && f.Severity == SeverityError {
			return // expected
		}
	}
	t.Error("expected root-entity error when a non-authorized root coexists with an authorized root")
}

func TestOwnershipTreeForestWithAuthorizedRoots(t *testing.T) {
	e := newTestEngine()
	e.AddAllowedRootKind("aws.organization")

	graph := core.NewGraph()
	org1 := core.NewEntity("org-1", core.EntityKind("aws.organization"), "Org 1")
	if err := graph.AddEntity(org1); err != nil {
		t.Fatalf("failed to addentity: %v", err)
	}
	account1 := core.NewEntity("acct-1", core.EntityKind("aws.account"), "Account 1")
	account1.SetOwner("org-1")
	if err := graph.AddEntity(account1); err != nil {
		t.Fatalf("failed to addentity: %v", err)
	}
	org2 := core.NewEntity("org-2", core.EntityKind("aws.organization"), "Org 2")
	if err := graph.AddEntity(org2); err != nil {
		t.Fatalf("failed to addentity: %v", err)
	}
	account2 := core.NewEntity("acct-2", core.EntityKind("aws.account"), "Account 2")
	account2.SetOwner("org-2")
	if err := graph.AddEntity(account2); err != nil {
		t.Fatalf("failed to addentity: %v", err)
	}

	result := e.Validate(graph, nil)
	for _, f := range result.Findings {
		if f.RuleID == "ownership-tree" {
			t.Errorf("unexpected ownership-tree error for connected forest: %s", f.Message)
		}
		if f.RuleID == "single-owner" {
			t.Errorf("unexpected single-owner error for authorized forest: %s", f.Message)
		}
		if f.RuleID == "root-entity" {
			t.Errorf("unexpected root-entity error for authorized forest: %s", f.Message)
		}
	}
}

func TestOwnershipTreeDisconnectedForest(t *testing.T) {
	e := newTestEngine()
	e.AddAllowedRootKind("aws.organization")

	graph := core.NewGraph()
	org1 := core.NewEntity("org-1", core.EntityKind("aws.organization"), "Org 1")
	if err := graph.AddEntity(org1); err != nil {
		t.Fatalf("failed to addentity: %v", err)
	}
	// acct-1 has an owner that does not exist (dangling), so it is unreachable
	// from any root and the forest is disconnected.
	orphan := core.NewEntity("orphan-1", core.EntityKind("aws.account"), "Orphan")
	orphan.SetOwner("nonexistent")
	graph.ForceAddEntity(orphan)

	result := e.Validate(graph, nil)
	found := false
	for _, f := range result.Findings {
		if f.RuleID == "ownership-tree" && f.Severity == SeverityError {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected ownership-tree error for disconnected forest")
	}
}

func TestValidPropertyIntegerJSONFloat(t *testing.T) {
	e := newTestEngine()
	graph := core.NewGraph()
	region := core.NewEntity("region-01", kinds.Region, "Region 1")
	if err := graph.AddEntity(region); err != nil {
		t.Fatalf("failed to addentity: %v", err)
	}
	rack := core.NewEntity("rack-01", kinds.Rack, "Rack 1")
	rack.SetOwner("region-01")
	rack.SetProperty("height_units", float64(42)) // JSON numbers decode to float64
	if err := graph.AddEntity(rack); err != nil {
		t.Fatalf("failed to addentity: %v", err)
	}

	result := e.Validate(graph, nil)
	if hasFindingFor(result, "valid-property", "rack-01") {
		t.Error("integral float64 should not trigger valid-property warning")
	}
}

func TestValidPropertyIntegerNonIntegralFloat(t *testing.T) {
	e := newTestEngine()
	graph := core.NewGraph()
	region := core.NewEntity("region-01", kinds.Region, "Region 1")
	if err := graph.AddEntity(region); err != nil {
		t.Fatalf("failed to addentity: %v", err)
	}
	rack := core.NewEntity("rack-01", kinds.Rack, "Rack 1")
	rack.SetOwner("region-01")
	rack.SetProperty("height_units", 42.5)
	if err := graph.AddEntity(rack); err != nil {
		t.Fatalf("failed to addentity: %v", err)
	}

	result := e.Validate(graph, nil)
	if !hasFindingFor(result, "valid-property", "rack-01") {
		t.Error("expected valid-property warning for non-integral float height_units")
	}
}

func TestValidPropertyStringWithReference(t *testing.T) {
	e := newTestEngine()
	graph := core.NewGraph()
	region := core.NewEntity("region-01", kinds.Region, "Region 1")
	if err := graph.AddEntity(region); err != nil {
		t.Fatalf("failed to addentity: %v", err)
	}
	net := core.NewEntity("net-01", kinds.Network, "Network 1")
	net.SetOwner("region-01")
	// gateway is a string property; the parser converts @-prefixed strings to
	// ReferenceValue, which must not trigger a type warning.
	net.SetProperty("gateway", core.NewReferenceValue("@gw-01"))
	if err := graph.AddEntity(net); err != nil {
		t.Fatalf("failed to addentity: %v", err)
	}

	result := e.Validate(graph, nil)
	if hasFindingFor(result, "valid-property", "net-01") {
		t.Error("string property holding a ReferenceValue should not warn")
	}
}
