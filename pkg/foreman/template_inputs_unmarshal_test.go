package foreman

import (
	"encoding/json"
	"testing"
)

// Exercises every branch of ForemanTemplateInput.UnmarshalJSON's "id" and
// "template_id" type-switches directly via crafted JSON, rather than
// relying on what a real Foreman response happens to send ids as
// (encoding/json only ever decodes a JSON number into float64 when the
// destination is interface{}, so the int/float32 cases can't be reached
// through JSON at all - only float64, string, and "anything else" are
// reachable, and this test sticks to those).
func TestForemanTemplateInput_UnmarshalJSON(t *testing.T) {
	cases := []struct {
		name      string
		json      string
		wantErr   bool
		wantID    int
		wantTplID int
	}{
		{"numeric ids", `{"id":5,"template_id":7,"name":"x"}`, false, 5, 7},
		{"string ids", `{"id":"5","template_id":"7","name":"x"}`, false, 5, 7},
		{"empty string ids", `{"id":"","template_id":"","name":"x"}`, false, 0, 0},
		{"invalid string id", `{"id":"not-a-number","name":"x"}`, true, 0, 0},
		{"invalid string template_id", `{"id":5,"template_id":"not-a-number","name":"x"}`, true, 5, 0},
		{"unsupported id type", `{"id":true,"name":"x"}`, false, 0, 0},
		{"unsupported template_id type", `{"id":5,"template_id":true,"name":"x"}`, false, 5, 0},
		{"missing id key", `{"name":"x"}`, false, 0, 0},
		{"missing template_id key", `{"id":5,"name":"x"}`, false, 5, 0},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var fti ForemanTemplateInput
			err := json.Unmarshal([]byte(c.json), &fti)
			if c.wantErr {
				if err == nil {
					t.Fatalf("expected an error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if fti.Id != c.wantID {
				t.Errorf("Id = %d, want %d", fti.Id, c.wantID)
			}
			if fti.TemplateId != c.wantTplID {
				t.Errorf("TemplateId = %d, want %d", fti.TemplateId, c.wantTplID)
			}
		})
	}
}
