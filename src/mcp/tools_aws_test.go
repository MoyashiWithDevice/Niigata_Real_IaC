package mcp

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"

	_ "IACForge/src/extension/builtin/aws"
)

// callTool invokes an MCP tool handler and returns the result.
func callTool(t *testing.T, s *mcpserver.MCPServer, name string, args map[string]interface{}) *mcp.CallToolResult {
	t.Helper()
	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      name,
			Arguments: args,
		},
	}
	res, err := s.GetTool(name).Handler(context.Background(), req)
	if err != nil {
		t.Fatalf("handler error for %s: %v", name, err)
	}
	return res
}

// seedAWSGraph builds a full AWS ownership hierarchy through the MCP add_entity
// and add_relation tools, then returns the session data.
func seedAWSGraph(t *testing.T, sm *SessionManager, s *mcpserver.MCPServer) *SessionData {
	t.Helper()

	entities := []struct {
		id, kind, name, owner string
		props                 string
	}{
		{"org-01", "aws.organization", "Example Org", "", ""},
		{"acct-01", "aws.account", "Example Account", "org-01", `{"account_id":"123456789012"}`},
		{"region-01", "aws.region", "us-east-1", "acct-01", `{"region_code":"us-east-1"}`},
		{"az-01", "aws.availability_zone", "us-east-1a", "region-01", `{"zone_name":"us-east-1a"}`},
		{"vpc-01", "aws.vpc", "Main VPC", "region-01", `{"cidr_block":"10.0.0.0/16"}`},
		{"subnet-01", "aws.subnet", "Public Subnet", "vpc-01", `{"cidr_block":"10.0.1.0/24"}`},
		{"sg-01", "aws.security_group", "Web SG", "vpc-01", ""},
		{"ec2-01", "aws.ec2", "Web Server", "subnet-01", `{"instance_type":"t3.micro"}`},
		{"rds-01", "aws.rds", "Database", "subnet-01", `{"instance_class":"db.t3.micro"}`},
		{"lambda-01", "aws.lambda_function", "Processor", "acct-01", `{"runtime":"python3.12"}`},
		{"s3-01", "aws.s3_bucket", "Assets Bucket", "az-01", ""},
		{"sqs-01", "aws.sqs_queue", "Job Queue", "acct-01", ""},
		{"sns-01", "aws.sns_topic", "Events Topic", "acct-01", ""},
		{"cw-01", "aws.cloudwatch_alarm", "CPU Alarm", "acct-01", ""},
	}
	for _, e := range entities {
		args := map[string]interface{}{"id": e.id, "kind": e.kind, "name": e.name}
		if e.owner != "" {
			args["owner"] = e.owner
		}
		if e.props != "" {
			args["properties_json"] = e.props
		}
		res := callTool(t, s, "add_entity", args)
		if res.IsError {
			t.Fatalf("add_entity %s failed: %s", e.id, toolResultText(t, res))
		}
	}

	relations := []struct {
		id, typ, source, target string
	}{
		{"rel-ec2-subnet", "belongs_to", "ec2-01", "subnet-01"},
		{"rel-cw-ec2", "monitors", "cw-01", "ec2-01"},
		{"rel-sns-sqs", "aws.subscribes", "sns-01", "sqs-01"},
	}
	for _, r := range relations {
		res := callTool(t, s, "add_relation", map[string]interface{}{
			"id": r.id, "type": r.typ, "source": r.source, "target": r.target,
		})
		if res.IsError {
			t.Fatalf("add_relation %s failed: %s", r.id, toolResultText(t, res))
		}
	}

	return sm.GetOrCreate("default")
}

func TestAWSKindsInSessionSchema(t *testing.T) {
	sm := NewSessionManager()
	s := NewMCPServer(sm)

	// The AWS extension must be visible through list_entity_kinds.
	res := callTool(t, s, "list_entity_kinds", map[string]interface{}{})
	if res.IsError {
		t.Fatalf("list_entity_kinds failed: %+v", res)
	}
	text := toolResultText(t, res)
	for _, want := range []string{`"kind": "aws.organization"`, `"kind": "aws.vpc"`, `"kind": "aws.ec2"`, `"kind": "aws.s3_bucket"`} {
		if !strings.Contains(text, want) {
			t.Errorf("expected %q in list_entity_kinds output:\n%s", want, text)
		}
	}

	// The AWS extension must be listed by the extension tools.
	extRes := callTool(t, s, "list_extensions", map[string]interface{}{})
	extText := toolResultText(t, extRes)
	if !strings.Contains(extText, `"id": "iacforge.aws"`) {
		t.Errorf("expected iacforge.aws in list_extensions output:\n%s", extText)
	}

	kindRes := callTool(t, s, "list_extension_kinds", map[string]interface{}{})
	kindText := toolResultText(t, kindRes)
	if !strings.Contains(kindText, "aws.ec2") {
		t.Errorf("expected aws.ec2 in list_extension_kinds output:\n%s", kindText)
	}
}

func TestAWSGetEntityKindSchema(t *testing.T) {
	sm := NewSessionManager()
	s := NewMCPServer(sm)

	res := callTool(t, s, "get_entity_kind", map[string]interface{}{"kind": "aws.vpc"})
	if res.IsError {
		t.Fatalf("get_entity_kind failed: %+v", res)
	}
	text := toolResultText(t, res)
	for _, want := range []string{`"kind": "aws.vpc"`, `"properties"`, `"cidr_block"`} {
		if !strings.Contains(text, want) {
			t.Errorf("expected %q in get_entity_kind output:\n%s", want, text)
		}
	}

	relRes := callTool(t, s, "list_relation_types", map[string]interface{}{})
	relText := toolResultText(t, relRes)
	for _, want := range []string{`"type": "aws.subscribes"`, `"type": "aws.registers"`} {
		if !strings.Contains(relText, want) {
			t.Errorf("expected %s in list_relation_types output:\n%s", want, relText)
		}
	}
}

func TestAWSAddEntityAndGet(t *testing.T) {
	sm := NewSessionManager()
	s := NewMCPServer(sm)
	sd := seedAWSGraph(t, sm, s)

	g := sd.Graph
	if got := len(g.Entities()); got != 14 {
		t.Errorf("expected 14 entities, got %d", got)
	}

	// get_entity returns the aws entity with its kind preserved.
	res := callTool(t, s, "get_entity", map[string]interface{}{"id": "ec2-01"})
	if res.IsError {
		t.Fatalf("get_entity failed: %+v", res)
	}
	text := toolResultText(t, res)
	for _, want := range []string{`"id": "ec2-01"`, `"kind": "aws.ec2"`, `"instance_type"`} {
		if !strings.Contains(text, want) {
			t.Errorf("expected %q in get_entity output:\n%s", want, text)
		}
	}

	// list_entities with a kind filter only returns matching aws entities.
	listRes := callTool(t, s, "list_entities", map[string]interface{}{"kind": "aws.ec2"})
	listText := toolResultText(t, listRes)
	if !strings.Contains(listText, `"id": "ec2-01"`) {
		t.Errorf("expected ec2-01 in list_entities kind filter output:\n%s", listText)
	}
	if strings.Contains(listText, `"id": "s3-01"`) {
		t.Errorf("list_entities kind filter leaked a non-ec2 entity:\n%s", listText)
	}
}

func TestAWSRelations(t *testing.T) {
	sm := NewSessionManager()
	s := NewMCPServer(sm)
	sd := seedAWSGraph(t, sm, s)

	if got := len(sd.Graph.Relations()); got != 3 {
		t.Errorf("expected 3 relations, got %d", got)
	}

	// list_relations with a type filter.
	res := callTool(t, s, "list_relations", map[string]interface{}{"type": "aws.subscribes"})
	text := toolResultText(t, res)
	if !strings.Contains(text, `"id": "rel-sns-sqs"`) {
		t.Errorf("expected rel-sns-sqs in list_relations output:\n%s", text)
	}
	if strings.Contains(text, `"id": "rel-cw-ec2"`) {
		t.Errorf("list_relations type filter leaked a non-subscribes relation:\n%s", text)
	}

	// get_relation returns the relation.
	getRes := callTool(t, s, "get_relation", map[string]interface{}{"id": "rel-sns-sqs"})
	if getRes.IsError {
		t.Fatalf("get_relation failed: %+v", getRes)
	}
	getText := toolResultText(t, getRes)
	if !strings.Contains(getText, `"aws.subscribes"`) {
		t.Errorf("expected aws.subscribes in get_relation output:\n%s", getText)
	}
}

func TestAWSQueryEntities(t *testing.T) {
	sm := NewSessionManager()
	s := NewMCPServer(sm)
	_ = seedAWSGraph(t, sm, s)

	res := callTool(t, s, "query_entities", map[string]interface{}{"kind": "aws.ec2"})
	if res.IsError {
		t.Fatalf("query_entities failed: %+v", res)
	}
	text := toolResultText(t, res)
	if !strings.Contains(text, `"id": "ec2-01"`) {
		t.Errorf("expected ec2-01 in query_entities output:\n%s", text)
	}
	if !strings.Contains(text, `"kind": "aws.ec2"`) {
		t.Errorf("expected aws.ec2 kind in query_entities output:\n%s", text)
	}
	if !strings.Contains(text, `"count": 1`) {
		t.Errorf("expected count 1 in query_entities output:\n%s", text)
	}
}

func TestAWSQueryRelatedTraversal(t *testing.T) {
	sm := NewSessionManager()
	s := NewMCPServer(sm)
	_ = seedAWSGraph(t, sm, s)

	res := callTool(t, s, "query_related", map[string]interface{}{
		"from": "org-01", "operation": "descendants",
	})
	if res.IsError {
		t.Fatalf("query_related failed: %+v", res)
	}
	text := toolResultText(t, res)
	for _, want := range []string{`"id": "acct-01"`, `"id": "vpc-01"`, `"id": "ec2-01"`} {
		if !strings.Contains(text, want) {
			t.Errorf("expected %q in descendants output:\n%s", want, text)
		}
	}

	// query_related with a relation type filter.
	relRes := callTool(t, s, "query_related", map[string]interface{}{
		"from": "sns-01", "operation": "targets", "relation_type": "aws.subscribes",
	})
	if relRes.IsError {
		t.Fatalf("query_related targets failed: %+v", relRes)
	}
	relText := toolResultText(t, relRes)
	if !strings.Contains(relText, `"id": "sqs-01"`) {
		t.Errorf("expected sqs-01 as subscribes target:\n%s", relText)
	}
}

func TestAWSResolvePathAndWhoReferences(t *testing.T) {
	sm := NewSessionManager()
	s := NewMCPServer(sm)
	_ = seedAWSGraph(t, sm, s)

	// Ownership path resolution.
	res := callTool(t, s, "resolve_path", map[string]interface{}{"ref": "org-01/acct-01/region-01/vpc-01/subnet-01/ec2-01"})
	if res.IsError {
		t.Fatalf("resolve_path failed: %+v", res)
	}
	text := toolResultText(t, res)
	if !strings.Contains(text, `"id": "ec2-01"`) {
		t.Errorf("expected ec2-01 from path resolution:\n%s", text)
	}

	// who_references on an entity referenced by a relation and a property.
	whoRes := callTool(t, s, "who_references", map[string]interface{}{"id": "ec2-01"})
	if whoRes.IsError {
		t.Fatalf("who_references failed: %+v", whoRes)
	}
	whoText := toolResultText(t, whoRes)
	if !strings.Contains(whoText, `"id": "rel-cw-ec2"`) {
		t.Errorf("expected monitors relation in who_references output:\n%s", whoText)
	}
}

func TestAWSValidateGraphPasses(t *testing.T) {
	sm := NewSessionManager()
	s := NewMCPServer(sm)
	_ = seedAWSGraph(t, sm, s)

	res := callTool(t, s, "validate_graph", map[string]interface{}{})
	if res.IsError {
		t.Fatalf("validate_graph failed: %+v", res)
	}
	text := toolResultText(t, res)
	if !strings.Contains(text, `"passed": true`) {
		t.Errorf("expected AWS graph to validate, got:\n%s", text)
	}
	if strings.Contains(text, `"errors": 0`) == false {
		t.Errorf("expected zero validation errors, got:\n%s", text)
	}
}

func TestAWSGraphSummary(t *testing.T) {
	sm := NewSessionManager()
	s := NewMCPServer(sm)
	_ = seedAWSGraph(t, sm, s)

	res := callTool(t, s, "graph_summary", map[string]interface{}{})
	if res.IsError {
		t.Fatalf("graph_summary failed: %+v", res)
	}
	text := toolResultText(t, res)
	for _, want := range []string{`"total_entities": 14`, `"total_relations": 3`, `"aws.ec2": 1`} {
		if !strings.Contains(text, want) {
			t.Errorf("expected %q in graph_summary output:\n%s", want, text)
		}
	}
}

func TestAWSUpdateAndRemove(t *testing.T) {
	sm := NewSessionManager()
	s := NewMCPServer(sm)
	sd := seedAWSGraph(t, sm, s)

	// update_entity on an aws entity.
	upd := callTool(t, s, "update_entity", map[string]interface{}{
		"id": "ec2-01", "name": "Renamed Web Server",
		"properties_json": `{"instance_type":"m5.large"}`,
	})
	if upd.IsError {
		t.Fatalf("update_entity failed: %+v", upd)
	}
	e, ok := sd.Graph.GetEntity("ec2-01")
	if !ok {
		t.Fatal("ec2-01 missing after update")
	}
	if e.Name != "Renamed Web Server" {
		t.Errorf("expected updated name, got %q", e.Name)
	}
	if e.Properties["instance_type"] != "m5.large" {
		t.Errorf("expected updated instance_type, got %v", e.Properties["instance_type"])
	}

	// update_relation on an aws relation.
	updRel := callTool(t, s, "update_relation", map[string]interface{}{
		"id": "rel-sns-sqs", "description": "Primary queue subscription",
	})
	if updRel.IsError {
		t.Fatalf("update_relation failed: %+v", updRel)
	}
	rel, ok := sd.Graph.GetRelation("rel-sns-sqs")
	if !ok {
		t.Fatal("rel-sns-sqs missing after update")
	}
	if rel.Description != "Primary queue subscription" {
		t.Errorf("expected updated relation description, got %q", rel.Description)
	}

	// remove_relation then remove_entity.
	rmRel := callTool(t, s, "remove_relation", map[string]interface{}{"id": "rel-cw-ec2"})
	if rmRel.IsError {
		t.Fatalf("remove_relation failed: %+v", rmRel)
	}
	if _, ok := sd.Graph.GetRelation("rel-cw-ec2"); ok {
		t.Error("rel-cw-ec2 still present after removal")
	}

	rm := callTool(t, s, "remove_entity", map[string]interface{}{"id": "sqs-01"})
	if rm.IsError {
		t.Fatalf("remove_entity failed: %+v", rm)
	}
	if _, ok := sd.Graph.GetEntity("sqs-01"); ok {
		t.Error("sqs-01 still present after removal")
	}
}

func TestAWSYAMLRoundTrip(t *testing.T) {
	sm := NewSessionManager()
	s := NewMCPServer(sm)

	// Build the graph via YAML string.
	doc := `
objects:
  - id: org-01
    kind: aws.organization
    name: Example Org
  - id: acct-01
    kind: aws.account
    name: Example Account
    attributes:
      owner: org-01
    spec:
      account_id: "123456789012"
  - id: region-01
    kind: aws.region
    name: us-east-1
    attributes:
      owner: acct-01
    spec:
      region_code: us-east-1
  - id: vpc-01
    kind: aws.vpc
    name: Main VPC
    attributes:
      owner: region-01
    spec:
      cidr_block: 10.0.0.0/16
  - id: subnet-01
    kind: aws.subnet
    name: Public Subnet
    attributes:
      owner: vpc-01
    spec:
      cidr_block: 10.0.1.0/24
  - id: ec2-01
    kind: aws.ec2
    name: Web Server
    attributes:
      owner: subnet-01
    spec:
      instance_type: t3.micro
  - id: s3-01
    kind: aws.s3_bucket
    name: Assets Bucket
    attributes:
      owner: region-01
`
	parse := callTool(t, s, "parse_yaml_string", map[string]interface{}{"yaml_content": doc})
	if parse.IsError {
		t.Fatalf("parse_yaml_string failed: %+v", parse)
	}
	if !strings.Contains(toolResultText(t, parse), "Parsed 7 entities") {
		t.Errorf("unexpected parse summary: %s", toolResultText(t, parse))
	}

	// Serialize to a YAML string.
	ser := callTool(t, s, "serialize_to_string", map[string]interface{}{})
	if ser.IsError {
		t.Fatalf("serialize_to_string failed: %+v", ser)
	}
	serText := toolResultText(t, ser)
	if !strings.Contains(serText, "s3_buckets") || !strings.Contains(serText, "ec2-01") {
		t.Errorf("serialized output missing aws kinds/entities:\n%s", serText)
	}

	// Validate the graph built from YAML.
	val := callTool(t, s, "validate_graph", map[string]interface{}{})
	if val.IsError {
		t.Fatalf("validate_graph failed: %+v", val)
	}
	if !strings.Contains(toolResultText(t, val), `"passed": true`) {
		t.Errorf("expected YAML-built AWS graph to validate:\n%s", toolResultText(t, val))
	}
}

func TestAWSLoadAndSaveYAML(t *testing.T) {
	sm := NewSessionManager()
	s := NewMCPServer(sm)

	dir := t.TempDir()
	src := filepath.Join(dir, "model.yaml")
	doc := `
objects:
  - id: org-01
    kind: aws.organization
    name: Example Org
  - id: acct-01
    kind: aws.account
    name: Example Account
    attributes:
      owner: org-01
  - id: vpc-01
    kind: aws.vpc
    name: Main VPC
    attributes:
      owner: acct-01
    spec:
      cidr_block: 10.0.0.0/16
`
	if err := os.WriteFile(src, []byte(doc), 0644); err != nil {
		t.Fatal(err)
	}

	load := callTool(t, s, "load_yaml", map[string]interface{}{"path": src})
	if load.IsError {
		t.Fatalf("load_yaml failed: %+v", load)
	}
	if !strings.Contains(toolResultText(t, load), "Loaded 3 entities") {
		t.Errorf("unexpected load summary: %s", toolResultText(t, load))
	}

	// Save the graph and re-load it to confirm round-trip.
	out := filepath.Join(dir, "out.yaml")
	save := callTool(t, s, "save_yaml", map[string]interface{}{"path": out})
	if save.IsError {
		t.Fatalf("save_yaml failed: %+v", save)
	}

	reload := callTool(t, s, "load_yaml", map[string]interface{}{"path": out})
	if reload.IsError {
		t.Fatalf("reload after save failed: %+v", reload)
	}
	if !strings.Contains(toolResultText(t, reload), "Loaded 3 entities") {
		t.Errorf("unexpected reload summary: %s", toolResultText(t, reload))
	}
}

func TestAWSClearGraph(t *testing.T) {
	sm := NewSessionManager()
	s := NewMCPServer(sm)
	sd := seedAWSGraph(t, sm, s)

	res := callTool(t, s, "clear_graph", map[string]interface{}{})
	if res.IsError {
		t.Fatalf("clear_graph failed: %+v", res)
	}
	if got := len(sd.Graph.Entities()); got != 0 {
		t.Errorf("expected empty graph after clear, got %d entities", got)
	}
}
