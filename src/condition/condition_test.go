package condition

import (
	"testing"
)

func TestToFloat64(t *testing.T) {
	tests := []struct {
		input    interface{}
		expected float64
		ok       bool
	}{
		{int(4), 4, true},
		{int64(8), 8, true},
		{float64(1.5), 1.5, true},
		{float32(2.5), 2.5, true},
		{"string", 0, false},
		{nil, 0, false},
	}

	for _, tt := range tests {
		got, ok := ToFloat64(tt.input)
		if ok != tt.ok || (ok && got != tt.expected) {
			t.Errorf("ToFloat64(%v) = %v, %v; want %v, %v", tt.input, got, ok, tt.expected, tt.ok)
		}
	}
}

func TestEvaluate(t *testing.T) {
	tests := []struct {
		name     string
		value    interface{}
		expected interface{}
		operator string
		want     bool
		wantErr  bool
	}{
		{"eq string", "active", "active", "eq", true, false},
		{"eq mismatch", "active", "maintenance", "eq", false, false},
		{"eq nil", nil, nil, "eq", true, false},
		{"ne", "active", "maintenance", "ne", true, false},
		{"in", "vm", []interface{}{"vm", "container"}, "in", true, false},
		{"in not found", "server", []interface{}{"vm", "container"}, "in", false, false},
		{"nin", "server", []interface{}{"vm", "container"}, "nin", true, false},
		{"gt", 8, 4, "gt", true, false},
		{"ge", 4, 4, "ge", true, false},
		{"lt", 4, 8, "lt", true, false},
		{"le", 4, 4, "le", true, false},
		{"gt non-numeric", "abc", 4, "gt", false, true},
		{"contains string", "hello world", "world", "contains", true, false},
		{"contains slice", []interface{}{"a", "b"}, "a", "contains", true, false},
		{"starts_with", "vm-web", "vm-", "starts_with", true, false},
		{"ends_with", "vm-web", "web", "ends_with", true, false},
		{"matches", "vm-01", "^vm-.*", "matches", true, false},
		{"defined nil", nil, true, "defined", false, false},
		{"defined empty string", "", true, "defined", false, false},
		{"defined empty false", "", false, "defined", true, false},
		{"undefined nil", nil, true, "undefined", true, false},
		{"undefined empty false", "", false, "undefined", false, false},
		{"unsupported", "x", "y", "unknown-op", false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Evaluate(tt.value, tt.expected, tt.operator)
			if tt.wantErr {
				if err == nil {
					t.Errorf("Evaluate(%v, %v, %s) expected error, got nil", tt.value, tt.expected, tt.operator)
				}
				return
			}
			if err != nil {
				t.Errorf("Evaluate(%v, %v, %s) unexpected error: %v", tt.value, tt.expected, tt.operator, err)
			}
			if got != tt.want {
				t.Errorf("Evaluate(%v, %v, %s) = %v, want %v", tt.value, tt.expected, tt.operator, got, tt.want)
			}
		})
	}
}
