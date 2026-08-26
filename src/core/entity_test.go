package core

import (
	"fmt"
	"testing"
)

func TestNewEntity(t *testing.T) {
	e := NewEntity("srv-01", "server", "Server 01")
	if e.ID != "srv-01" {
		t.Errorf("expected ID srv-01, got %s", e.ID)
	}
	if e.Kind != "server" {
		t.Errorf("expected Kind server, got %s", e.Kind)
	}
	if e.Name != "Server 01" {
		t.Errorf("expected Name Server 01, got %s", e.Name)
	}
	if e.Owner != "" {
		t.Errorf("expected empty Owner, got %s", e.Owner)
	}
	if e.Properties == nil {
		t.Error("expected Properties to be initialized")
	}
}

func TestEntitySetOwner(t *testing.T) {
	e := NewEntity("rack-01", "rack", "Rack 01")
	if !e.IsRoot() {
		t.Error("expected entity to be root initially")
	}
	e.SetOwner("region-01")
	if e.IsRoot() {
		t.Error("expected entity not to be root after setting owner")
	}
	if e.Owner != "region-01" {
		t.Errorf("expected owner region-01, got %s", e.Owner)
	}
}

func TestEntitySetStatus(t *testing.T) {
	e := NewEntity("srv-01", "server", "Server 01")
	e.SetStatus(StatusActive)
	if e.Status != StatusActive {
		t.Errorf("expected status active, got %s", e.Status)
	}
}

func TestEntityTags(t *testing.T) {
	e := NewEntity("srv-01", "server", "Server 01")
	e.AddTag("production")
	e.AddTag("ap-northeast-1")

	if len(e.Tags) != 2 {
		t.Errorf("expected 2 tags, got %d", len(e.Tags))
	}
	if !e.HasTag("production") {
		t.Error("expected to have tag production")
	}
	if !e.HasTag("ap-northeast-1") {
		t.Error("expected to have tag ap-northeast-1")
	}
	if e.HasTag("nonexistent") {
		t.Error("expected not to have tag nonexistent")
	}
}

func TestEntityLabels(t *testing.T) {
	e := NewEntity("srv-01", "server", "Server 01")
	e.SetLabel("region", "asia-pacific")
	e.SetLabel("tier", "primary")

	if v, ok := e.GetLabel("region"); !ok || v != "asia-pacific" {
		t.Errorf("expected label region=asia-pacific, got %s", v)
	}
	if v, ok := e.GetLabel("tier"); !ok || v != "primary" {
		t.Errorf("expected label tier=primary, got %s", v)
	}
	if _, ok := e.GetLabel("nonexistent"); ok {
		t.Error("expected no label nonexistent")
	}
}

func TestEntityProperties(t *testing.T) {
	e := NewEntity("srv-01", "server", "Server 01")
	e.SetProperty("cpu_cores", 32)
	e.SetProperty("memory", []interface{}{
		map[string]interface{}{"size_gb": 64, "speed": 3200, "type": "ddr4"},
		map[string]interface{}{"size_gb": 64, "speed": 3200, "type": "ddr4"},
	})

	if v, ok := e.GetProperty("cpu_cores"); !ok || v != 32 {
		t.Errorf("expected property cpu_cores=32, got %v", v)
	}
	mem, ok := e.GetProperty("memory")
	if !ok {
		t.Error("expected property memory to exist")
	}
	memList, ok := mem.([]interface{})
	if !ok || len(memList) != 2 {
		t.Errorf("expected memory to be a list of 2, got %v", mem)
	}
	if _, ok := e.GetProperty("nonexistent"); ok {
		t.Error("expected no property nonexistent")
	}
}

func TestEntityValidate(t *testing.T) {
	tests := []struct {
		name    string
		entity  *Entity
		wantErr error
	}{
		{
			name:    "valid entity",
			entity:  NewEntity("srv-01", "server", "Server 01"),
			wantErr: nil,
		},
		{
			name:    "missing ID",
			entity:  NewEntity("", "server", "Server 01"),
			wantErr: ErrEntityMissingID,
		},
		{
			name:    "missing kind",
			entity:  NewEntity("srv-01", "", "Server 01"),
			wantErr: ErrEntityMissingKind,
		},
		{
			name:    "missing name",
			entity:  NewEntity("srv-01", "server", ""),
			wantErr: ErrEntityMissingName,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.entity.Validate()
			if err != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestEntityPath(t *testing.T) {
	e := NewEntity("eno1", "interface", "eno1")
	if e.Path() != "" {
		t.Errorf("expected empty path, got %s", e.Path())
	}
	e.SetPath("/region01/rack01/pve01/eno1")
	if e.Path() != "/region01/rack01/pve01/eno1" {
		t.Errorf("expected path /region01/rack01/pve01/eno1, got %s", e.Path())
	}
}

func TestEntityFullPath(t *testing.T) {
	e := NewEntity("eno1", "interface", "eno1")
	if e.FullPath() != "/eno1" {
		t.Errorf("expected full path /eno1, got %s", e.FullPath())
	}
	e.SetPath("/region01/rack01/pve01/eno1")
	if e.FullPath() != "/region01/rack01/pve01/eno1" {
		t.Errorf("expected full path /region01/rack01/pve01/eno1, got %s", e.FullPath())
	}
}

func TestEntityString(t *testing.T) {
	e := NewEntity("srv-01", "server", "Server 01")
	e.SetStatus(StatusActive)
	s := e.String()
	if s == "" {
		t.Error("expected non-empty string representation")
	}
}

func TestStatusConstants(t *testing.T) {
	statuses := []Status{
		StatusPlanned, StatusActive, StatusMaintenance,
		StatusDeprecated, StatusOffline, StatusStandby,
	}
	expected := []string{"planned", "active", "maintenance", "deprecated", "offline", "standby"}
	for i, s := range statuses {
		if string(s) != expected[i] {
			t.Errorf("expected status %s, got %s", expected[i], s)
		}
	}
}

func TestReferenceValue(t *testing.T) {
	tests := []struct {
		name       string
		raw        string
		wantTarget string
		wantString string
	}{
		{
			name:       "simple reference",
			raw:        "@net-mgmt",
			wantTarget: "net-mgmt",
			wantString: "@net-mgmt",
		},
		{
			name:       "reference without prefix",
			raw:        "net-mgmt",
			wantTarget: "net-mgmt",
			wantString: "@net-mgmt",
		},
		{
			name:       "path reference",
			raw:        "@/region01/rack01/server01",
			wantTarget: "/region01/rack01/server01",
			wantString: "@/region01/rack01/server01",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ref := NewReferenceValue(tt.raw)
			if ref.RefTargetID() != tt.wantTarget {
				t.Errorf("RefTargetID() = %q, want %q", ref.RefTargetID(), tt.wantTarget)
			}
			if ref.String() != tt.wantString {
				t.Errorf("String() = %q, want %q", ref.String(), tt.wantString)
			}
		})
	}
}

func TestIsReferenceValue(t *testing.T) {
	ref := NewReferenceValue("@net-mgmt")
	if !IsReferenceValue(ref) {
		t.Error("expected IsReferenceValue to return true for ReferenceValue")
	}
	if IsReferenceValue("net-mgmt") {
		t.Error("expected IsReferenceValue to return false for string")
	}
	if IsReferenceValue(42) {
		t.Error("expected IsReferenceValue to return false for int")
	}
}

func TestExtractReferenceValue(t *testing.T) {
	ref := NewReferenceValue("@net-mgmt")
	targetID, ok := ExtractReferenceValue(ref)
	if !ok || targetID != "net-mgmt" {
		t.Errorf("ExtractReferenceValue(ReferenceValue) = (%q, %v), want (\"net-mgmt\", true)", targetID, ok)
	}

	targetID, ok = ExtractReferenceValue("net-mgmt")
	if ok || targetID != "" {
		t.Errorf("ExtractReferenceValue(string) = (%q, %v), want (\"\", false)", targetID, ok)
	}
}

func TestExtractReferenceTargets(t *testing.T) {
	tests := []struct {
		name  string
		value interface{}
		want  []string
	}{
		{
			name:  "nil",
			value: nil,
			want:  nil,
		},
		{
			name:  "plain string",
			value: "net-mgmt",
			want:  nil,
		},
		{
			name:  "scalar reference",
			value: NewReferenceValue("@net-mgmt"),
			want:  []string{"net-mgmt"},
		},
		{
			name: "list of references",
			value: []interface{}{
				NewReferenceValue("@sg-01"),
				"plain",
				NewReferenceValue("@sg-02"),
			},
			want: []string{"sg-01", "sg-02"},
		},
		{
			name: "nested map",
			value: map[string]interface{}{
				"network": NewReferenceValue("@net-mgmt"),
				"name":    "mgmt",
			},
			want: []string{"net-mgmt"},
		},
		{
			name: "list of maps",
			value: []interface{}{
				map[string]interface{}{
					"network": NewReferenceValue("@net-storage"),
					"vlan":    float64(100),
				},
				map[string]interface{}{
					"target": NewReferenceValue("@net-mgmt"),
				},
			},
			want: []string{"net-storage", "net-mgmt"},
		},
		{
			name: "empty containers",
			value: []interface{}{
				map[string]interface{}{"a": float64(1)},
				"plain",
			},
			want: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExtractReferenceTargets(tt.value)
			if len(got) != len(tt.want) {
				t.Fatalf("ExtractReferenceTargets() = %v, want %v", got, tt.want)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("ExtractReferenceTargets() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}

func TestEntityPropertyReference(t *testing.T) {
	e := NewEntity("vlan-100", "vlan", "VLAN 100")
	e.SetProperty("associated_network", NewReferenceValue("@net-mgmt"))

	v, ok := e.GetProperty("associated_network")
	if !ok {
		t.Fatal("expected property associated_network to exist")
	}
	ref, ok := v.(ReferenceValue)
	if !ok {
		t.Fatalf("expected ReferenceValue, got %T", v)
	}
	if ref.RefTargetID() != "net-mgmt" {
		t.Errorf("expected reference target net-mgmt, got %s", ref.RefTargetID())
	}
}

func TestResolvePropertyPath(t *testing.T) {
	e := NewEntity("srv-01", "server", "Server 01")
	e.SetProperty("memory", []interface{}{
		map[string]interface{}{"size_gb": float64(64), "speed": float64(3200), "type": "ddr4"},
		map[string]interface{}{"size_gb": float64(64), "speed": float64(3200), "type": "ddr4"},
	})
	e.SetProperty("cpu", []interface{}{
		map[string]interface{}{"cores": float64(32), "architecture": "x86_64"},
	})
	e.SetProperty("platform", "proxmox")

	tests := []struct {
		name     string
		path     string
		expected interface{}
		isNil    bool
	}{
		{
			name:     "simple property",
			path:     "platform",
			expected: "proxmox",
		},
		{
			name:  "nonexistent property",
			path:  "nonexistent",
			isNil: true,
		},
		{
			name: "list property returns full list",
			path: "memory",
			expected: []interface{}{
				map[string]interface{}{"size_gb": float64(64), "speed": float64(3200), "type": "ddr4"},
				map[string]interface{}{"size_gb": float64(64), "speed": float64(3200), "type": "ddr4"},
			},
		},
		{
			name:     "dot-notation extracts sub-property from list items",
			path:     "memory.size_gb",
			expected: []interface{}{float64(64), float64(64)},
		},
		{
			name:     "dot-notation extracts speed from list items",
			path:     "memory.speed",
			expected: []interface{}{float64(3200), float64(3200)},
		},
		{
			name:     "dot-notation extracts cpu cores",
			path:     "cpu.cores",
			expected: []interface{}{float64(32)},
		},
		{
			name:  "dot-notation on nonexistent top-level",
			path:  "nonexistent.size_gb",
			isNil: true,
		},
		{
			name:  "dot-notation on nonexistent sub-property",
			path:  "memory.nonexistent",
			isNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := e.ResolvePropertyPath(tt.path)
			if tt.isNil {
				if result != nil {
					t.Errorf("expected nil, got %v", result)
				}
				return
			}
			if result == nil {
				t.Fatalf("expected non-nil result, got nil")
			}
			// Compare as string representations to handle type differences
			if fmt.Sprintf("%v", result) != fmt.Sprintf("%v", tt.expected) {
				t.Errorf("ResolvePropertyPath(%q) = %v, want %v", tt.path, result, tt.expected)
			}
		})
	}
}
