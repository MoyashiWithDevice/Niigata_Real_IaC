package parser

import (
	"strings"
	"testing"

	"IACForge/src/core"
	"IACForge/src/core/kinds"
	"IACForge/src/core/types"
)

func TestSerializeBasicEntity(t *testing.T) {
	g := core.NewGraph()
	e := core.NewEntity("region-ap-northeast-1", kinds.Region, "Tokyo Datacenter 1")
	if err := g.AddEntity(e); err != nil {
		t.Fatalf("failed to add entity: %v", err)
	}

	serializer := NewSerializer()
	data, err := serializer.Serialize(g)
	if err != nil {
		t.Fatalf("failed to serialize: %v", err)
	}

	// Parse back and verify
	parser := NewParser()
	g2, err := parser.Parse(data)
	if err != nil {
		t.Fatalf("failed to parse serialized data: %v", err)
	}

	e2, ok := g2.GetEntity("region-ap-northeast-1")
	if !ok {
		t.Fatal("entity region-ap-northeast-1 not found in parsed data")
	}

	if e2.ID != e.ID {
		t.Errorf("expected ID %s, got %s", e.ID, e2.ID)
	}
	if e2.Kind != e.Kind {
		t.Errorf("expected kind %s, got %s", e.Kind, e2.Kind)
	}
	if e2.Name != e.Name {
		t.Errorf("expected name %s, got %s", e.Name, e2.Name)
	}
}

func TestSerializeEntityWithAllProperties(t *testing.T) {
	g := core.NewGraph()
	e := core.NewEntity("srv-proxmox-01", kinds.Server, "Proxmox Node 01")
	e.Description = "Primary Proxmox server"
	e.SetStatus(core.StatusActive)
	e.AddTag("production")
	e.AddTag("compute")
	e.SetLabel("region", "ap-northeast-1")
	e.SetLabel("environment", "production")
	e.Extensions = map[string]interface{}{"vendor": "dell"}
	e.SetProperty("platform", "proxmox")
	e.SetProperty("cpu_cores", 32)
	e.SetProperty("memory", []interface{}{
		map[string]interface{}{"size_gb": 64, "speed": 3200, "type": "ddr4"},
		map[string]interface{}{"size_gb": 64, "speed": 3200, "type": "ddr4"},
	})

	if err := g.AddEntity(e); err != nil {
		t.Fatalf("failed to add entity: %v", err)
	}

	serializer := NewSerializer()
	data, err := serializer.Serialize(g)
	if err != nil {
		t.Fatalf("failed to serialize: %v", err)
	}

	// Parse back and verify
	parser := NewParser()
	g2, err := parser.Parse(data)
	if err != nil {
		t.Fatalf("failed to parse serialized data: %v", err)
	}

	e2, ok := g2.GetEntity("srv-proxmox-01")
	if !ok {
		t.Fatal("entity srv-proxmox-01 not found in parsed data")
	}

	if e2.Description != e.Description {
		t.Errorf("expected description %s, got %s", e.Description, e2.Description)
	}
	if e2.Status != e.Status {
		t.Errorf("expected status %s, got %s", e.Status, e2.Status)
	}
	if len(e2.Tags) != 2 {
		t.Errorf("expected 2 tags, got %d", len(e2.Tags))
	}
	if e2.Labels["region"] != "ap-northeast-1" {
		t.Errorf("expected label region=ap-northeast-1, got %s", e2.Labels["region"])
	}
	if platform, ok := e2.GetProperty("platform"); !ok || platform != "proxmox" {
		t.Errorf("expected property platform=proxmox, got %v", platform)
	}
}

func TestSerializeDirectedRelation(t *testing.T) {
	g := core.NewGraph()

	srv := core.NewEntity("srv-01", kinds.Server, "Server 01")
	vm := core.NewEntity("vm-01", kinds.VM, "VM 01")

	if err := g.AddEntity(srv); err != nil {
		t.Fatalf("failed to add entity: %v", err)
	}
	if err := g.AddEntity(vm); err != nil {
		t.Fatalf("failed to add entity: %v", err)
	}

	r := core.NewDirectedRelation("rel-hosts-vm", types.Hosts, "srv-01", "vm-01")
	if err := g.AddRelation(r); err != nil {
		t.Fatalf("failed to add relation: %v", err)
	}

	serializer := NewSerializer()
	data, err := serializer.Serialize(g)
	if err != nil {
		t.Fatalf("failed to serialize: %v", err)
	}

	// Parse back and verify
	parser := NewParser()
	g2, err := parser.Parse(data)
	if err != nil {
		t.Fatalf("failed to parse serialized data: %v", err)
	}

	r2, ok := g2.GetRelation("rel-hosts-vm")
	if !ok {
		t.Fatal("relation rel-hosts-vm not found in parsed data")
	}

	if r2.Type != types.Hosts {
		t.Errorf("expected type hosts, got %s", r2.Type)
	}
	if r2.Direction != core.DirectionDirected {
		t.Errorf("expected direction directed, got %s", r2.Direction)
	}
	if r2.Source() != "srv-01" {
		t.Errorf("expected source srv-01, got %s", r2.Source())
	}
	if r2.Target() != "vm-01" {
		t.Errorf("expected target vm-01, got %s", r2.Target())
	}
}

func TestSerializeSymmetricRelation(t *testing.T) {
	g := core.NewGraph()

	// Create parent entities first
	srv := core.NewEntity("srv-01", kinds.Server, "Server 01")
	sw := core.NewEntity("sw-01", kinds.Switch, "Switch 01")
	if err := g.AddEntity(srv); err != nil {
		t.Fatalf("failed to add entity: %v", err)
	}
	if err := g.AddEntity(sw); err != nil {
		t.Fatalf("failed to add entity: %v", err)
	}

	iface1 := core.NewEntity("eno1", kinds.Interface, "eno1")
	iface1.SetOwner("srv-01")
	iface2 := core.NewEntity("port1", kinds.Interface, "port1")
	iface2.SetOwner("sw-01")

	if err := g.AddEntity(iface1); err != nil {
		t.Fatalf("failed to add entity: %v", err)
	}
	if err := g.AddEntity(iface2); err != nil {
		t.Fatalf("failed to add entity: %v", err)
	}

	r := core.NewSymmetricRelation("rel-connects", types.Connects, []string{"eno1", "port1"})
	if err := g.AddRelation(r); err != nil {
		t.Fatalf("failed to add relation: %v", err)
	}

	serializer := NewSerializer()
	data, err := serializer.Serialize(g)
	if err != nil {
		t.Fatalf("failed to serialize: %v", err)
	}

	// Parse back and verify
	parser := NewParser()
	g2, err := parser.Parse(data)
	if err != nil {
		t.Fatalf("failed to parse serialized data: %v", err)
	}

	r2, ok := g2.GetRelation("rel-connects")
	if !ok {
		t.Fatal("relation rel-connects not found in parsed data")
	}

	if r2.Type != types.Connects {
		t.Errorf("expected type connects, got %s", r2.Type)
	}
	if r2.Direction != core.DirectionSymmetric {
		t.Errorf("expected direction symmetric, got %s", r2.Direction)
	}
	if len(r2.Participants.List) != 2 {
		t.Errorf("expected 2 participants, got %d", len(r2.Participants.List))
	}
}

func TestRoundTrip(t *testing.T) {
	yaml := `
objects:
  - id: region-ap-northeast-1
    kind: region
    name: Tokyo Datacenter 1
    attributes:
      status: active
      labels:
        region: ap-northeast-1

  - id: rack-a01
    kind: rack
    name: Rack A01
    attributes:
      owner: region-ap-northeast-1
      status: active
    spec:
      height_units: 42

  - id: srv-proxmox-01
    kind: server
    name: Proxmox Node 01
    attributes:
      owner: rack-a01
      status: active
    spec:
      platform: proxmox
      cpu_cores: 32
      memory:
        - size_gb: 64
          speed: 3200
          type: ddr4

  - id: vm-web-01
    kind: vm
    name: Web Server 01
    attributes:
      owner: srv-proxmox-01
      status: active
    spec:
      cpu_cores: 4
      memory:
        - size_gb: 8
          speed: 3200
          type: ddr4
      os: ubuntu

  - id: rel-hosts-server-vm
    type: hosts
    participants:
      source: srv-proxmox-01
      target: vm-web-01
`

	// Parse original
	parser1 := NewParser()
	g1, err := parser1.Parse([]byte(yaml))
	if err != nil {
		t.Fatalf("failed to parse original: %v", err)
	}

	// Serialize
	serializer := NewSerializer()
	data, err := serializer.Serialize(g1)
	if err != nil {
		t.Fatalf("failed to serialize: %v", err)
	}

	// Parse serialized data
	parser2 := NewParser()
	g2, err := parser2.Parse(data)
	if err != nil {
		t.Fatalf("failed to parse serialized data: %v", err)
	}

	// Verify entities
	if g1.EntityCount() != g2.EntityCount() {
		t.Errorf("entity count mismatch: %d vs %d", g1.EntityCount(), g2.EntityCount())
	}

	for _, e1 := range g1.Entities() {
		e2, ok := g2.GetEntity(e1.ID)
		if !ok {
			t.Errorf("entity %s not found in round-trip", e1.ID)
			continue
		}
		if e1.ID != e2.ID {
			t.Errorf("entity ID mismatch: %s vs %s", e1.ID, e2.ID)
		}
		if e1.Kind != e2.Kind {
			t.Errorf("entity kind mismatch for %s: %s vs %s", e1.ID, e1.Kind, e2.Kind)
		}
		if e1.Name != e2.Name {
			t.Errorf("entity name mismatch for %s: %s vs %s", e1.ID, e1.Name, e2.Name)
		}
		if e1.Owner != e2.Owner {
			t.Errorf("entity owner mismatch for %s: %s vs %s", e1.ID, e1.Owner, e2.Owner)
		}
	}

	// Verify relations
	// Auto-generated relations from nesting may differ in count between g1 and g2
	// because the serializer nests entities based on ownership.
	// Verify that all explicit (non-auto) relations from g1 exist in g2.
	for _, r1 := range g1.Relations() {
		if val, ok := r1.GetLabel("auto_generated"); ok && val == "true" {
			continue
		}
		r2, ok := g2.GetRelation(r1.ID)
		if !ok {
			t.Errorf("explicit relation %s not found in round-trip", r1.ID)
			continue
		}
		if r1.Type != r2.Type {
			t.Errorf("relation type mismatch for %s: %s vs %s", r1.ID, r1.Type, r2.Type)
		}
		if r1.Direction != r2.Direction {
			t.Errorf("relation direction mismatch for %s: %s vs %s", r1.ID, r1.Direction, r2.Direction)
		}
		if r1.Source() != r2.Source() {
			t.Errorf("relation source mismatch for %s: %s vs %s", r1.ID, r1.Source(), r2.Source())
		}
		if r1.Target() != r2.Target() {
			t.Errorf("relation target mismatch for %s: %s vs %s", r1.ID, r1.Target(), r2.Target())
		}
	}

	// Verify all non-auto relations in g2 exist in g1 (or are auto-generated)
	for _, r2 := range g2.Relations() {
		if val, ok := r2.GetLabel("auto_generated"); ok && val == "true" {
			continue
		}
		if _, ok := g1.GetRelation(r2.ID); !ok {
			t.Errorf("relation %s in g2 not found in g1", r2.ID)
		}
	}
}

func TestRoundTripClusterNestedNodes(t *testing.T) {
	yaml := `
objects:
  - id: region-ap-northeast-1
    kind: region
    name: Tokyo Datacenter 1
  - id: k8s-prod
    kind: cluster
    name: Production K8s Cluster
    attributes:
      owner: region-ap-northeast-1
    spec:
      cluster_type: compute
      ha_enabled: true
      vms:
        - id: vm-k8s-node-01
          name: K8s Node 01
          spec:
            cpu:
              - cores: 4
            memory:
              - size_gb: 16
        - id: vm-k8s-node-02
          name: K8s Node 02
      servers:
        - id: srv-k8s-node-01
          name: K8s Bare-metal Node 01
`
	// Parse original
	parser1 := NewParser()
	g1, err := parser1.Parse([]byte(yaml))
	if err != nil {
		t.Fatalf("failed to parse original: %v", err)
	}

	// Serialize
	serializer := NewSerializer()
	data, err := serializer.Serialize(g1)
	if err != nil {
		t.Fatalf("failed to serialize: %v", err)
	}

	// Parse serialized data
	parser2 := NewParser()
	g2, err := parser2.Parse(data)
	if err != nil {
		t.Fatalf("failed to parse serialized data: %v", err)
	}

	if g1.EntityCount() != g2.EntityCount() {
		t.Errorf("entity count mismatch: %d vs %d", g1.EntityCount(), g2.EntityCount())
	}

	// Verify all entities survive round-trip with same owner
	for _, e1 := range g1.Entities() {
		e2, ok := g2.GetEntity(e1.ID)
		if !ok {
			t.Errorf("entity %s not found in round-trip", e1.ID)
			continue
		}
		if e1.Kind != e2.Kind {
			t.Errorf("kind mismatch for %s: %s vs %s", e1.ID, e1.Kind, e2.Kind)
		}
		if e1.Owner != e2.Owner {
			t.Errorf("owner mismatch for %s: %s vs %s", e1.ID, e1.Owner, e2.Owner)
		}
	}

	// Auto-generated belongs_to relations must be preserved
	for _, relID := range []string{
		"rel-auto-belongs_to-vm-k8s-node-01-k8s-prod",
		"rel-auto-belongs_to-vm-k8s-node-02-k8s-prod",
		"rel-auto-belongs_to-srv-k8s-node-01-k8s-prod",
	} {
		r2, ok := g2.GetRelation(relID)
		if !ok {
			t.Errorf("auto relation %s not found after round-trip", relID)
			continue
		}
		if r2.Type != types.BelongsTo {
			t.Errorf("relation %s: expected belongs_to, got %s", relID, r2.Type)
		}
		if r2.Source() == "" || r2.Target() != "k8s-prod" {
			t.Errorf("relation %s: expected source member -> target k8s-prod, got %s -> %s",
				relID, r2.Source(), r2.Target())
		}
	}
}

func TestRoundTripSwitchRouterPorts(t *testing.T) {
	yaml := `
objects:
  - id: sw-core-01
    kind: switch
    name: Core Switch 01
    spec:
      port_count: 24
      ports:
        - id: port1
          name: port1
          spec:
            type: ethernet
            speed_mbps: 10000
            mode: trunk
        - id: port2
          name: port2
          spec:
            type: ethernet
            speed_mbps: 1000
            mode: access

  - id: rt-core-01
    kind: router
    name: Core Router 01
    spec:
      ports:
        - id: ge0/0
          name: GigabitEthernet0/0
          spec:
            type: ethernet
            speed_mbps: 1000
`
	// Parse original
	parser1 := NewParser()
	g1, err := parser1.Parse([]byte(yaml))
	if err != nil {
		t.Fatalf("failed to parse original: %v", err)
	}

	// Serialize
	serializer := NewSerializer()
	data, err := serializer.Serialize(g1)
	if err != nil {
		t.Fatalf("failed to serialize: %v", err)
	}

	// The serialized output must use the `ports` nest key for switch/router,
	// not the global `interfaces` nest key.
	if !strings.Contains(string(data), "ports:") {
		t.Errorf("serialized output missing 'ports:' nest key:\n%s", data)
	}
	if strings.Contains(string(data), "interfaces:") {
		t.Errorf("serialized output should not use 'interfaces:' nest key:\n%s", data)
	}

	// Parse serialized data
	parser2 := NewParser()
	g2, err := parser2.Parse(data)
	if err != nil {
		t.Fatalf("failed to parse serialized data: %v", err)
	}

	if g1.EntityCount() != g2.EntityCount() {
		t.Errorf("entity count mismatch: %d vs %d", g1.EntityCount(), g2.EntityCount())
	}

	// Verify all entities survive round-trip with same owner and kind
	for _, e1 := range g1.Entities() {
		e2, ok := g2.GetEntity(e1.ID)
		if !ok {
			t.Errorf("entity %s not found in round-trip", e1.ID)
			continue
		}
		if e1.Kind != e2.Kind {
			t.Errorf("kind mismatch for %s: %s vs %s", e1.ID, e1.Kind, e2.Kind)
		}
		if e1.Owner != e2.Owner {
			t.Errorf("owner mismatch for %s: %s vs %s", e1.ID, e1.Owner, e2.Owner)
		}
	}

	// port_count property must survive round-trip
	sw2, ok := g2.GetEntity("sw-core-01")
	if !ok {
		t.Fatal("entity sw-core-01 not found after round-trip")
	}
	portCount, ok := sw2.GetProperty("port_count")
	if !ok || portCount != 24 {
		t.Errorf("expected port_count 24 after round-trip, got %v", portCount)
	}

	// Auto-generated belongs_to relations must be preserved
	for _, relID := range []string{
		"rel-auto-belongs_to-port1-sw-core-01",
		"rel-auto-belongs_to-port2-sw-core-01",
		"rel-auto-belongs_to-ge0/0-rt-core-01",
	} {
		r2, ok := g2.GetRelation(relID)
		if !ok {
			t.Errorf("auto relation %s not found after round-trip", relID)
			continue
		}
		if r2.Type != types.BelongsTo {
			t.Errorf("relation %s: expected belongs_to, got %s", relID, r2.Type)
		}
	}
}

func TestSerializeFile(t *testing.T) {
	g := core.NewGraph()
	e := core.NewEntity("test-entity", kinds.Server, "Test Entity")
	if err := g.AddEntity(e); err != nil {
		t.Fatalf("failed to add entity: %v", err)
	}

	serializer := NewSerializer()
	err := serializer.SerializeFile(g, "/tmp/test_serialize.yaml")
	if err != nil {
		t.Fatalf("failed to serialize to file: %v", err)
	}

	// Parse back and verify
	parser := NewParser()
	g2, err := parser.ParseFile("/tmp/test_serialize.yaml")
	if err != nil {
		t.Fatalf("failed to parse file: %v", err)
	}

	e2, ok := g2.GetEntity("test-entity")
	if !ok {
		t.Fatal("entity test-entity not found in file")
	}

	if e2.ID != e.ID {
		t.Errorf("expected ID %s, got %s", e.ID, e2.ID)
	}
}

func TestRoundTripPropertyReference(t *testing.T) {
	input := `
objects:
  - id: net-mgmt
    kind: network
    name: Management Network
    spec:
      cidr: 10.0.0.0/24

  - id: vlan-100
    kind: vlan
    name: VLAN 100
    spec:
      vlan_id: 100
      associated_network: "@net-mgmt"
`
	parser := NewParser()
	g, err := parser.Parse([]byte(input))
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	serializer := NewSerializer()
	data, err := serializer.Serialize(g)
	if err != nil {
		t.Fatalf("failed to serialize: %v", err)
	}

	// Verify @ prefix is preserved in output (YAML may use single or double quotes)
	output := string(data)
	if !strings.Contains(output, "@net-mgmt") {
		t.Errorf("serialized output should contain @net-mgmt reference, got:\n%s", output)
	}

	// Parse back and verify round-trip
	g2, err := parser.Parse(data)
	if err != nil {
		t.Fatalf("failed to re-parse: %v", err)
	}

	vlan, ok := g2.GetEntity("vlan-100")
	if !ok {
		t.Fatal("expected entity vlan-100")
	}

	v, ok := vlan.GetProperty("associated_network")
	if !ok {
		t.Fatal("expected property associated_network")
	}
	ref, ok := v.(core.ReferenceValue)
	if !ok {
		t.Fatalf("expected ReferenceValue after round-trip, got %T", v)
	}
	if ref.RefTargetID() != "net-mgmt" {
		t.Errorf("expected reference target net-mgmt, got %s", ref.RefTargetID())
	}
}

func TestRoundTripInterfaceNetworkReference(t *testing.T) {
	input := `
objects:
  - id: net-mgmt
    kind: network
    name: Management Network
    spec:
      cidr: 10.0.0.0/24

  - id: eno1
    kind: interface
    name: eno1
    spec:
      network: "@net-mgmt"
      ip_address:
        - 10.0.0.10
`
	parser := NewParser()
	g, err := parser.Parse([]byte(input))
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	serializer := NewSerializer()
	data, err := serializer.Serialize(g)
	if err != nil {
		t.Fatalf("failed to serialize: %v", err)
	}

	output := string(data)
	if !strings.Contains(output, "@net-mgmt") {
		t.Errorf("serialized output should contain @net-mgmt reference, got:\n%s", output)
	}

	g2, err := parser.Parse(data)
	if err != nil {
		t.Fatalf("failed to re-parse: %v", err)
	}

	intf, ok := g2.GetEntity("eno1")
	if !ok {
		t.Fatal("expected entity eno1")
	}

	v, ok := intf.GetProperty("network")
	if !ok {
		t.Fatal("expected property network")
	}
	ref, ok := v.(core.ReferenceValue)
	if !ok {
		t.Fatalf("expected ReferenceValue after round-trip, got %T", v)
	}
	if ref.RefTargetID() != "net-mgmt" {
		t.Errorf("expected reference target net-mgmt, got %s", ref.RefTargetID())
	}
}
