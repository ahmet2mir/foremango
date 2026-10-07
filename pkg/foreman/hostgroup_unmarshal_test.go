package foreman

import (
	"encoding/json"
	"testing"
)

// Covers foremanHostGroupDecode.UnmarshalJSON's three shapes for
// "parameters" (slice, map, and genuinely invalid) directly - the
// HTTP-driven test in hostgroup_test.go only exercises the slice shape.
func TestForemanHostGroupDecode_UnmarshalJSON(t *testing.T) {
	t.Run("parameters as slice", func(t *testing.T) {
		var h foremanHostGroupDecode
		err := json.Unmarshal([]byte(`{"name":"g","parameters":[{"name":"k","value":"v"}]}`), &h)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(h.HostGroupParametersDecode) != 1 || h.HostGroupParametersDecode[0].Name != "k" {
			t.Fatalf("unexpected parameters: %+v", h.HostGroupParametersDecode)
		}
	})

	t.Run("parameters as map", func(t *testing.T) {
		var h foremanHostGroupDecode
		err := json.Unmarshal([]byte(`{"name":"g","parameters":{"k":"v"}}`), &h)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(h.HostGroupParametersDecode) != 1 || h.HostGroupParametersDecode[0].Name != "k" || h.HostGroupParametersDecode[0].Value != "v" {
			t.Fatalf("unexpected parameters: %+v", h.HostGroupParametersDecode)
		}
	})

	t.Run("parameters absent errors", func(t *testing.T) {
		var h foremanHostGroupDecode
		if err := json.Unmarshal([]byte(`{"name":"g"}`), &h); err == nil {
			t.Fatalf("expected an error when \"parameters\" is absent, got nil")
		}
	})
}
