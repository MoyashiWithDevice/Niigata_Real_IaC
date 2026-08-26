package schema

import (
	"testing"

	"IACForge/src/core"
	"IACForge/src/core/kinds"
	"IACForge/src/core/types"
)

func TestNewSchema(t *testing.T) {
	s := NewSchema("1.0.0", "1.0")
	if s == nil {
		t.Fatal("expected non-nil schema")
	}
	if s.Version.SchemaVersion != "1.0.0" {
		t.Errorf("expected schema version 1.0.0, got %s", s.Version.SchemaVersion)
	}
	if s.Version.SpecVersion != "1.0" {
		t.Errorf("expected spec version 1.0, got %s", s.Version.SpecVersion)
	}
	if len(s.EntityKinds) != 0 {
		t.Errorf("expected empty entity kinds, got %d", len(s.EntityKinds))
	}
	if len(s.RelationTypes) != 0 {
		t.Errorf("expected empty relation types, got %d", len(s.RelationTypes))
	}
}

func TestCoreSchemaEntityKinds(t *testing.T) {
	s := CoreSchema()

	expectedKinds := []core.EntityKind{
		kinds.Region, kinds.Rack, kinds.Server, kinds.Interface, kinds.Cable,
		kinds.PowerDistribution, kinds.Network, kinds.VLAN, kinds.Switch,
		kinds.Router, kinds.Firewall, kinds.ACL, kinds.ACLRule,
		kinds.VM, kinds.Container, kinds.Application, kinds.OpenPort,
		kinds.Storage, kinds.Volume, kinds.Cluster, kinds.AvailabilityZone,
	}

	if len(s.EntityKinds) != len(expectedKinds) {
		t.Fatalf("expected %d entity kinds, got %d", len(expectedKinds), len(s.EntityKinds))
	}

	for _, kind := range expectedKinds {
		if !s.HasEntityKind(kind) {
			t.Errorf("expected entity kind %q to be defined", kind)
		}
	}
}

func TestCoreSchemaRelationTypes(t *testing.T) {
	s := CoreSchema()

	expectedTypes := []core.RelationType{
		types.Connects, types.Hosts, types.DependsOn, types.BelongsTo,
		types.ReplicatesTo, types.BacksUp, types.Monitors, types.ManagedBy,
		types.MountedOn, types.AppliesTo, types.ListensOn,
	}

	if len(s.RelationTypes) != len(expectedTypes) {
		t.Fatalf("expected %d relation types, got %d", len(expectedTypes), len(s.RelationTypes))
	}

	for _, relType := range expectedTypes {
		if !s.HasRelationType(relType) {
			t.Errorf("expected relation type %q to be defined", relType)
		}
	}
}

func TestCoreSchemaRelationDirections(t *testing.T) {
	s := CoreSchema()

	tests := []struct {
		relType  core.RelationType
		expected DirectionType
	}{
		{types.Connects, DirectionSymmetric},
		{types.Hosts, DirectionDirected},
		{types.DependsOn, DirectionDirected},
		{types.BelongsTo, DirectionDirected},
		{types.ReplicatesTo, DirectionDirected},
		{types.BacksUp, DirectionDirected},
		{types.Monitors, DirectionDirected},
		{types.ManagedBy, DirectionDirected},
		{types.MountedOn, DirectionDirected},
		{types.AppliesTo, DirectionDirected},
		{types.ListensOn, DirectionDirected},
	}

	for _, tt := range tests {
		def, ok := s.GetRelationTypeDef(tt.relType)
		if !ok {
			t.Fatalf("relation type %q not found", tt.relType)
		}
		if def.Direction != tt.expected {
			t.Errorf("relation type %q: expected direction %q, got %q", tt.relType, tt.expected, def.Direction)
		}
	}
}

func TestCoreSchemaEntityKindProperties(t *testing.T) {
	s := CoreSchema()

	// Server should have cpu property (structured list)
	serverDef, ok := s.GetEntityKindDef(kinds.Server)
	if !ok {
		t.Fatal("server kind not found")
	}
	foundCpu := false
	for _, p := range serverDef.Properties {
		if p.Name == "cpu" {
			foundCpu = true
			if p.Type != PropertyTypeList {
				t.Errorf("expected cpu type list, got %s", p.Type)
			}
			if len(p.Properties) != 2 {
				t.Errorf("expected cpu to have 2 sub-properties, got %d", len(p.Properties))
			}
		}
	}
	if !foundCpu {
		t.Error("server kind missing cpu property")
	}

	// Open port should have port property with range 1-65535
	openPortDef, ok := s.GetEntityKindDef(kinds.OpenPort)
	if !ok {
		t.Fatal("open_port kind not found")
	}
	foundPort := false
	for _, p := range openPortDef.Properties {
		if p.Name == "port" {
			foundPort = true
			if !p.Required {
				t.Error("expected port property to be required")
			}
			if p.Constraints == nil || p.Constraints.Min == nil || p.Constraints.Max == nil {
				t.Error("expected port property to have min and max constraints")
			}
		}
	}
	if !foundPort {
		t.Error("open_port kind missing port property")
	}
}

func TestCoreSchemaRelationParticipantConstraints(t *testing.T) {
	s := CoreSchema()

	// Connects should be symmetric with interface participants
	connectsDef, ok := s.GetRelationTypeDef(types.Connects)
	if !ok {
		t.Fatal("connects relation type not found")
	}
	if connectsDef.Participants == nil {
		t.Fatal("connects missing participant constraints")
	}
	if len(connectsDef.Participants.SourceKinds) != 1 || connectsDef.Participants.SourceKinds[0] != kinds.Interface {
		t.Errorf("connects source kinds should be [interface], got %v", connectsDef.Participants.SourceKinds)
	}

	// Hosts should have server, vm, container as source
	hostsDef, ok := s.GetRelationTypeDef(types.Hosts)
	if !ok {
		t.Fatal("hosts relation type not found")
	}
	if hostsDef.Participants == nil {
		t.Fatal("hosts missing participant constraints")
	}
	if len(hostsDef.Participants.SourceKinds) != 4 {
		t.Errorf("hosts should have 4 source kinds, got %d", len(hostsDef.Participants.SourceKinds))
	}
}

func TestAddEntityKind(t *testing.T) {
	s := NewSchema("1.0.0", "1.0")
	s.AddEntityKind("custom_kind", &EntityKindDefinition{
		Description: "Custom kind",
	})

	if !s.HasEntityKind("custom_kind") {
		t.Error("expected custom_kind to exist")
	}

	def, ok := s.GetEntityKindDef("custom_kind")
	if !ok {
		t.Fatal("custom_kind not found")
	}
	if def.Description != "Custom kind" {
		t.Errorf("expected description 'Custom kind', got %q", def.Description)
	}
}

func TestAddRelationType(t *testing.T) {
	s := NewSchema("1.0.0", "1.0")
	s.AddRelationType("custom_type", &RelationTypeDefinition{
		Direction: DirectionDirected,
	})

	if !s.HasRelationType("custom_type") {
		t.Error("expected custom_type to exist")
	}

	def, ok := s.GetRelationTypeDef("custom_type")
	if !ok {
		t.Fatal("custom_type not found")
	}
	if def.Direction != DirectionDirected {
		t.Errorf("expected direction directed, got %s", def.Direction)
	}
}

func TestCoreSchemaNestingDefinitions(t *testing.T) {
	s := CoreSchema()

	tests := []struct {
		parentKind  core.EntityKind
		nestKey     string
		childKind   core.EntityKind
		autoRelType core.RelationType
	}{
		// Server can nest VMs and Containers
		{kinds.Server, "vms", kinds.VM, types.Hosts},
		{kinds.Server, "containers", kinds.Container, types.Hosts},
		// VM can nest Applications and Containers
		{kinds.VM, "applications", kinds.Application, types.Hosts},
		{kinds.VM, "containers", kinds.Container, types.Hosts},
		// Container can nest Applications
		{kinds.Container, "applications", kinds.Application, types.Hosts},
		// Application can nest Containers and OpenPorts
		{kinds.Application, "containers", kinds.Container, types.Hosts},
		{kinds.Application, "open_ports", kinds.OpenPort, types.BelongsTo},
		// Cluster can nest VMs and Servers as nodes (belongs_to membership)
		{kinds.Cluster, "vms", kinds.VM, types.BelongsTo},
		{kinds.Cluster, "servers", kinds.Server, types.BelongsTo},
		// Region can nest racks, clusters, and availability zones
		{kinds.Region, "racks", kinds.Rack, types.BelongsTo},
		{kinds.Region, "clusters", kinds.Cluster, types.BelongsTo},
		{kinds.Region, "availability_zones", kinds.AvailabilityZone, types.BelongsTo},
		// Switch and Router can nest Ports as interface entities (belongs_to)
		{kinds.Switch, "ports", kinds.Interface, types.BelongsTo},
		{kinds.Router, "ports", kinds.Interface, types.BelongsTo},
	}

	for _, tt := range tests {
		defs := s.GetNestingDefs(tt.parentKind)
		found := false
		for _, d := range defs {
			if d.NestKey == tt.nestKey && d.ChildKind == tt.childKind {
				if d.AutoRelationType != tt.autoRelType {
					t.Errorf("%s/%s: expected auto relation %s, got %s",
						tt.parentKind, tt.nestKey, tt.autoRelType, d.AutoRelationType)
				}
				found = true
				break
			}
		}
		if !found {
			t.Errorf("%s/%s -> %s nesting definition not found", tt.parentKind, tt.nestKey, tt.childKind)
		}
	}
}

func TestValidateProperty(t *testing.T) {
	s := CoreSchema()

	serverDef, _ := s.GetEntityKindDef(kinds.Server)

	// Find cpu property (structured list)
	var cpuProp *PropertyDefinition
	for i := range serverDef.Properties {
		if serverDef.Properties[i].Name == "cpu" {
			cpuProp = &serverDef.Properties[i]
			break
		}
	}
	if cpuProp == nil {
		t.Fatal("cpu property not found")
	}

	// Test nil value (not required, should pass)
	if err := s.ValidateProperty(cpuProp, nil); err != nil {
		t.Errorf("expected nil error for nil value, got %v", err)
	}

	// Test valid structured list value
	validCpu := []interface{}{
		map[string]interface{}{
			"cores":        16,
			"architecture": "x86_64",
		},
	}
	if err := s.ValidateProperty(cpuProp, validCpu); err != nil {
		t.Errorf("expected no error for valid structured list value, got %v", err)
	}

	// Test scalar value (should fail for list type)
	if err := s.ValidateProperty(cpuProp, 8); err == nil {
		t.Error("expected error for scalar value on list property")
	}

	// Test invalid sub-property type
	invalidCpu := []interface{}{
		map[string]interface{}{
			"cores": "not_a_number",
		},
	}
	if err := s.ValidateProperty(cpuProp, invalidCpu); err == nil {
		t.Error("expected error for invalid sub-property type")
	}
}

func TestValidatePropertyRequired(t *testing.T) {
	s := NewSchema("1.0.0", "1.0")
	prop := &PropertyDefinition{
		Name:     "test_prop",
		Type:     PropertyTypeString,
		Required: true,
	}

	// Test nil value on required property
	if err := s.ValidateProperty(prop, nil); err == nil {
		t.Error("expected error for nil value on required property")
	}

	// Test non-nil value on required property
	if err := s.ValidateProperty(prop, "hello"); err != nil {
		t.Errorf("expected no error for valid value, got %v", err)
	}
}

func TestValidatePropertyStringConstraints(t *testing.T) {
	s := NewSchema("1.0.0", "1.0")
	minLen := 3
	maxLen := 10
	prop := &PropertyDefinition{
		Name: "test_prop",
		Type: PropertyTypeString,
		Constraints: &Constraint{
			MinLength: &minLen,
			MaxLength: &maxLen,
		},
	}

	// Too short
	if err := s.ValidateProperty(prop, "ab"); err == nil {
		t.Error("expected error for string too short")
	}

	// Too long
	if err := s.ValidateProperty(prop, "abcdefghijk"); err == nil {
		t.Error("expected error for string too long")
	}

	// Valid
	if err := s.ValidateProperty(prop, "hello"); err != nil {
		t.Errorf("expected no error for valid string, got %v", err)
	}
}

func TestValidatePropertyEnumConstraints(t *testing.T) {
	s := NewSchema("1.0.0", "1.0")
	prop := &PropertyDefinition{
		Name: "action",
		Type: PropertyTypeString,
		Constraints: &Constraint{
			Enum: []string{"allow", "deny"},
		},
	}

	// Valid enum value
	if err := s.ValidateProperty(prop, "allow"); err != nil {
		t.Errorf("expected no error for valid enum value, got %v", err)
	}

	// Invalid enum value
	if err := s.ValidateProperty(prop, "reject"); err == nil {
		t.Error("expected error for invalid enum value")
	}
}

func TestValidatePropertyPatternConstraints(t *testing.T) {
	s := NewSchema("1.0.0", "1.0")
	pattern := `^srv-[a-z0-9-]+$`
	prop := &PropertyDefinition{
		Name: "hostname",
		Type: PropertyTypeString,
		Constraints: &Constraint{
			Pattern: &pattern,
		},
	}

	// Valid match
	if err := s.ValidateProperty(prop, "srv-web-01"); err != nil {
		t.Errorf("expected no error for matching string, got %v", err)
	}

	// Invalid match
	if err := s.ValidateProperty(prop, "web_server!"); err == nil {
		t.Error("expected error for string not matching pattern")
	}

	// Invalid pattern (schema definition error)
	bad := `[`
	badProp := &PropertyDefinition{
		Name: "hostname",
		Type: PropertyTypeString,
		Constraints: &Constraint{
			Pattern: &bad,
		},
	}
	if err := s.ValidateProperty(badProp, "anything"); err == nil {
		t.Error("expected error for invalid pattern")
	}
}

func TestValidatePropertyListUnknownKey(t *testing.T) {
	s := NewSchema("1.0.0", "1.0")
	prop := &PropertyDefinition{
		Name: "cpu",
		Type: PropertyTypeList,
		Properties: []PropertyDefinition{
			{Name: "cores", Type: PropertyTypeInteger},
			{Name: "architecture", Type: PropertyTypeString},
		},
	}

	// Valid item
	valid := []interface{}{
		map[string]interface{}{"cores": 4, "architecture": "x86_64"},
	}
	if err := s.ValidateProperty(prop, valid); err != nil {
		t.Errorf("expected no error for valid list item, got %v", err)
	}

	// Item with unknown key
	invalid := []interface{}{
		map[string]interface{}{"cores": 4, "sockets": 2},
	}
	if err := s.ValidateProperty(prop, invalid); err == nil {
		t.Error("expected error for list item with unknown key")
	}
}

func TestValidatePropertyMapValueProperty(t *testing.T) {
	s := NewSchema("1.0.0", "1.0")
	valueProp := &PropertyDefinition{Name: "entry", Type: PropertyTypeInteger}
	prop := &PropertyDefinition{
		Name:          "ports",
		Type:          PropertyTypeMap,
		ValueProperty: valueProp,
	}

	// Valid map values
	valid := map[string]interface{}{"web": 8080, "db": 5432}
	if err := s.ValidateProperty(prop, valid); err != nil {
		t.Errorf("expected no error for valid map values, got %v", err)
	}

	// map[string]string values must fail against integer value schema
	invalid := map[string]string{"web": "8080"}
	if err := s.ValidateProperty(prop, invalid); err == nil {
		t.Error("expected error for map value not matching value property type")
	}

	// Non-numeric value in a numeric map
	badVal := map[string]interface{}{"web": "8080"}
	if err := s.ValidateProperty(prop, badVal); err == nil {
		t.Error("expected error for non-integer map value")
	}

	// Map without value property still validates as a map
	plain := &PropertyDefinition{Name: "tags", Type: PropertyTypeMap}
	if err := s.ValidateProperty(plain, map[string]interface{}{"a": 1, "b": "two"}); err != nil {
		t.Errorf("expected no error for map without value property, got %v", err)
	}
}

func TestValidatePropertyNumericConstraints(t *testing.T) {
	s := NewSchema("1.0.0", "1.0")
	min := float64(1)
	max := float64(65535)
	prop := &PropertyDefinition{
		Name: "port",
		Type: PropertyTypeInteger,
		Constraints: &Constraint{
			Min: &min,
			Max: &max,
		},
	}

	// Below min
	if err := s.ValidateProperty(prop, 0); err == nil {
		t.Error("expected error for value below min")
	}

	// Above max
	if err := s.ValidateProperty(prop, 70000); err == nil {
		t.Error("expected error for value above max")
	}

	// Valid
	if err := s.ValidateProperty(prop, 8080); err != nil {
		t.Errorf("expected no error for valid value, got %v", err)
	}
}

func TestProfile(t *testing.T) {
	p := NewProfile("test-profile")
	if p.Name != "test-profile" {
		t.Errorf("expected name 'test-profile', got %q", p.Name)
	}

	p.Rules = append(p.Rules, "unique-id", "valid-reference")

	if !p.HasRule("unique-id") {
		t.Error("expected profile to have rule 'unique-id'")
	}
	if !p.HasRule("valid-reference") {
		t.Error("expected profile to have rule 'valid-reference'")
	}
	if p.HasRule("nonexistent") {
		t.Error("expected profile not to have rule 'nonexistent'")
	}

	p.AddRequiredKind("server")
	if len(p.RequiredKinds) != 1 || p.RequiredKinds[0] != "server" {
		t.Error("expected profile to require kind 'server'")
	}

	p.AddRequiredRelation("connects")
	if len(p.RequiredRelations) != 1 || p.RequiredRelations[0] != "connects" {
		t.Error("expected profile to require relation 'connects'")
	}
}

func TestValidatePropertyReferenceType(t *testing.T) {
	s := NewSchema("1.0.0", "1.0")
	propDef := &PropertyDefinition{
		Name:     "associated_network",
		Type:     PropertyTypeReference,
		Required: false,
	}

	// Valid ReferenceValue
	ref := core.NewReferenceValue("@net-mgmt")
	if err := s.ValidateProperty(propDef, ref); err != nil {
		t.Errorf("expected no error for valid ReferenceValue, got %v", err)
	}

	// Valid @ prefix string (not yet converted)
	if err := s.ValidateProperty(propDef, "@net-mgmt"); err != nil {
		t.Errorf("expected no error for @ prefix string, got %v", err)
	}

	// Invalid: plain string without @ prefix
	if err := s.ValidateProperty(propDef, "net-mgmt"); err == nil {
		t.Error("expected error for plain string without @ prefix")
	}

	// Invalid: non-string value
	if err := s.ValidateProperty(propDef, 42); err == nil {
		t.Error("expected error for non-string value")
	}

	// nil is OK (not required)
	if err := s.ValidateProperty(propDef, nil); err != nil {
		t.Errorf("expected no error for nil value, got %v", err)
	}
}

func TestCoreSchemaInterfaceModeProperty(t *testing.T) {
	s := CoreSchema()
	def, ok := s.GetEntityKindDef(kinds.Interface)
	if !ok {
		t.Fatal("interface kind not found")
	}

	var modeProp *PropertyDefinition
	for _, p := range def.Properties {
		if p.Name == "mode" {
			modeProp = &p
			break
		}
	}
	if modeProp == nil {
		t.Fatal("mode property not found on interface")
	}
	if modeProp.Type != PropertyTypeString {
		t.Errorf("expected type string, got %s", modeProp.Type)
	}
	if modeProp.Default != "none" {
		t.Errorf("expected default 'none', got %v", modeProp.Default)
	}
	if modeProp.Constraints == nil || len(modeProp.Constraints.Enum) != 4 {
		t.Fatalf("expected 4 enum values, got %v", modeProp.Constraints)
	}

	validModes := map[string]bool{"access": true, "trunk": true, "hybrid": true, "none": true}
	for _, v := range modeProp.Constraints.Enum {
		if !validModes[v] {
			t.Errorf("unexpected enum value %q", v)
		}
	}
}

func TestCoreSchemaVlanTaggedProperty(t *testing.T) {
	s := CoreSchema()
	def, ok := s.GetEntityKindDef(kinds.VLAN)
	if !ok {
		t.Fatal("vlan kind not found")
	}

	var taggedProp *PropertyDefinition
	for _, p := range def.Properties {
		if p.Name == "tagged" {
			taggedProp = &p
			break
		}
	}
	if taggedProp == nil {
		t.Fatal("tagged property not found on vlan")
	}
	if taggedProp.Type != PropertyTypeBoolean {
		t.Errorf("expected type boolean, got %s", taggedProp.Type)
	}
	if taggedProp.Default != false {
		t.Errorf("expected default false, got %v", taggedProp.Default)
	}
}

func TestCoreSchemaInterfaceTypeEnum(t *testing.T) {
	s := CoreSchema()
	def, ok := s.GetEntityKindDef(kinds.Interface)
	if !ok {
		t.Fatal("interface kind not found")
	}

	var typeProp *PropertyDefinition
	for _, p := range def.Properties {
		if p.Name == "type" {
			typeProp = &p
			break
		}
	}
	if typeProp == nil {
		t.Fatal("type property not found on interface")
	}
	if typeProp.Constraints == nil || len(typeProp.Constraints.Enum) != 8 {
		t.Fatalf("expected 8 enum values, got %v", typeProp.Constraints)
	}

	validTypes := map[string]bool{
		"ethernet": true, "fiber": true, "wireless": true, "virtual": true,
		"bond": true, "vlan": true, "bridge": true, "loopback": true,
	}
	for _, v := range typeProp.Constraints.Enum {
		if !validTypes[v] {
			t.Errorf("unexpected enum value %q", v)
		}
	}
}

func TestCoreSchemaInterfaceVlanIDProperty(t *testing.T) {
	s := CoreSchema()
	def, ok := s.GetEntityKindDef(kinds.Interface)
	if !ok {
		t.Fatal("interface kind not found")
	}

	var vlanIDProp *PropertyDefinition
	for _, p := range def.Properties {
		if p.Name == "vlan_id" {
			vlanIDProp = &p
			break
		}
	}
	if vlanIDProp == nil {
		t.Fatal("vlan_id property not found on interface")
	}
	if vlanIDProp.Type != PropertyTypeInteger {
		t.Errorf("expected type integer, got %s", vlanIDProp.Type)
	}
	if vlanIDProp.Required {
		t.Errorf("vlan_id should be optional, but is required")
	}
	if vlanIDProp.Constraints == nil || vlanIDProp.Constraints.Min == nil || vlanIDProp.Constraints.Max == nil {
		t.Fatal("vlan_id should have min/max constraints")
	}
	if *vlanIDProp.Constraints.Min != 1 {
		t.Errorf("expected min 1, got %v", *vlanIDProp.Constraints.Min)
	}
	if *vlanIDProp.Constraints.Max != 4094 {
		t.Errorf("expected max 4094, got %v", *vlanIDProp.Constraints.Max)
	}
}

func TestValidatePropertyIntegerFloat(t *testing.T) {
	s := CoreSchema()

	// Integral float64 (as produced by encoding/json) is acceptable for integer.
	integral := &PropertyDefinition{Name: "height_units", Type: PropertyTypeInteger}
	if err := s.ValidateProperty(integral, float64(42)); err != nil {
		t.Errorf("integer property with integral float64 should be accepted, got %v", err)
	}

	// Non-integral float64 must be rejected.
	if err := s.ValidateProperty(integral, 42.5); err == nil {
		t.Error("integer property with non-integral float64 should be rejected")
	}

	// Plain int remains accepted.
	if err := s.ValidateProperty(integral, 42); err != nil {
		t.Errorf("integer property with int should be accepted, got %v", err)
	}
}

func TestValidatePropertyNumberUnsigned(t *testing.T) {
	s := CoreSchema()
	def := &PropertyDefinition{Name: "voltage", Type: PropertyTypeNumber}
	if err := s.ValidateProperty(def, uint(240)); err != nil {
		t.Errorf("number property with uint should be accepted, got %v", err)
	}
	if err := s.ValidateProperty(def, uint64(240)); err != nil {
		t.Errorf("number property with uint64 should be accepted, got %v", err)
	}
}

func TestValidateNumericConstraintsUnsigned(t *testing.T) {
	s := CoreSchema()
	def := &PropertyDefinition{
		Name:        "vlan_id",
		Type:        PropertyTypeInteger,
		Constraints: &Constraint{Min: intPtr(1), Max: intPtr(4094)},
	}
	if err := s.ValidateProperty(def, uint(100)); err != nil {
		t.Errorf("unsigned integer with min/max constraints should pass, got %v", err)
	}
	if err := s.ValidateProperty(def, uint(5000)); err == nil {
		t.Error("unsigned integer exceeding max should be rejected")
	}
}

func TestValidatePropertyStringReferenceValue(t *testing.T) {
	minLength := 1
	s := CoreSchema()
	def := &PropertyDefinition{
		Name:        "gateway",
		Type:        PropertyTypeString,
		Constraints: &Constraint{MinLength: &minLength},
	}
	if err := s.ValidateProperty(def, core.NewReferenceValue("@gw-01")); err != nil {
		t.Errorf("string property with ReferenceValue should be accepted, got %v", err)
	}
	if err := s.ValidateProperty(def, "10.0.0.1"); err != nil {
		t.Errorf("string property with plain string should be accepted, got %v", err)
	}
}
