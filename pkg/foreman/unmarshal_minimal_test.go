package foreman

import (
	"encoding/json"
	"testing"
)

// Every field in these custom UnmarshalJSON implementations falls back to
// its zero value when absent or the wrong type - a path a normal
// Marshal/Unmarshal round trip (as in each resource's own _test.go) never
// takes, since whatever was written is always exactly what's read back.
// Minimal JSON missing everything exercises every one of those fallbacks.

func TestForemanWebhook_UnmarshalJSON_Minimal(t *testing.T) {
	var fw ForemanWebhook
	if err := json.Unmarshal([]byte(`{"id":1,"name":"x"}`), &fw); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fw.TargetURL != "" || fw.HTTPMethod != "" || fw.Enabled != false || fw.WebhookTemplateID != 0 {
		t.Errorf("expected zero values for absent fields, got %+v", fw)
	}
}

func TestForemanPartitionTable_UnmarshalJSON_Minimal(t *testing.T) {
	var ft ForemanPartitionTable
	if err := json.Unmarshal([]byte(`{"id":1,"name":"x"}`), &ft); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ft.Layout != "" || ft.Snippet != false || ft.OSFamily != "" {
		t.Errorf("expected zero values for absent fields, got %+v", ft)
	}
}

func TestForemanProvisioningTemplate_UnmarshalJSON_Minimal(t *testing.T) {
	var ft ForemanProvisioningTemplate
	if err := json.Unmarshal([]byte(`{"id":1,"name":"x"}`), &ft); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ft.Template != "" || ft.Locked != false || ft.TemplateKindId != 0 {
		t.Errorf("expected zero values for absent fields, got %+v", ft)
	}

	// The one field with its own explicit branch, taken when present.
	var withKind ForemanProvisioningTemplate
	if err := json.Unmarshal([]byte(`{"id":1,"name":"x","template_kind_id":5}`), &withKind); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if withKind.TemplateKindId != 5 {
		t.Errorf("expected TemplateKindId 5, got %d", withKind.TemplateKindId)
	}
}

func TestForemanOperatingSystem_UnmarshalJSON_Minimal(t *testing.T) {
	var o ForemanOperatingSystem
	if err := json.Unmarshal([]byte(`{"id":1,"name":"x"}`), &o); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if o.Title != "" || o.Major != "" || o.Family != "" || len(o.OperatingSystemParameters) != 0 {
		t.Errorf("expected zero values for absent fields, got %+v", o)
	}
}

// Covers the branches ForemanOverrideValue's round-trip CRUD test doesn't:
// a Value that parses as a number/bool (MarshalJSON then writes it as that
// type, so UnmarshalJSON's tmpMap["value"].(string) fails and it falls back
// to re-encoding whatever's there), and match prefixes other than "fqdn".
func TestForemanOverrideValue_UnmarshalJSON_NonStringValue(t *testing.T) {
	cases := []struct {
		name      string
		json      string
		wantType  string
		wantValue string
	}{
		{"numeric value falls back to its JSON encoding", `{"match":"fqdn=host1","omit":false,"value":42}`, "fqdn", "42"},
		{"bool value falls back to its JSON encoding", `{"match":"fqdn=host1","omit":false,"value":true}`, "fqdn", "true"},
		{"hostgroup match", `{"match":"hostgroup=webservers","omit":false,"value":"x"}`, "hostgroup", "x"},
		{"domain match", `{"match":"domain=example.com","omit":false,"value":"x"}`, "domain", "x"},
		{"os match", `{"match":"os=RedHat","omit":false,"value":"x"}`, "os", "x"},
		{"unrecognized match prefix", `{"match":"weird=1","omit":false,"value":"x"}`, "", "x"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var ov ForemanOverrideValue
			if err := json.Unmarshal([]byte(c.json), &ov); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if ov.MatchType != c.wantType {
				t.Errorf("MatchType = %q, want %q", ov.MatchType, c.wantType)
			}
			if ov.Value != c.wantValue {
				t.Errorf("Value = %q, want %q", ov.Value, c.wantValue)
			}
		})
	}
}
