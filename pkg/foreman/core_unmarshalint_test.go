package foreman

import "testing"

// unmarshalInteger's float64 branch is practically unreachable through a
// real JSON round-trip for its one caller (webhook.go's
// WebhookTemplateID): the write side marshals that field through
// intIdToJSONString, which produces a JSON string or null, never a bare
// number - so decoding it back always lands in the "not a float64" branch.
// Called directly here to cover the other branch too.
func TestUnmarshalInteger(t *testing.T) {
	if got := unmarshalInteger(5.0); got != 5 {
		t.Errorf("unmarshalInteger(5.0) = %d, want 5", got)
	}
	if got := unmarshalInteger("5"); got != 0 {
		t.Errorf(`unmarshalInteger("5") = %d, want 0`, got)
	}
	if got := unmarshalInteger(nil); got != 0 {
		t.Errorf("unmarshalInteger(nil) = %d, want 0", got)
	}
}
