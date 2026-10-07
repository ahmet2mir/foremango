package foreman

import "testing"

// WrapParameters' name==nil branch marshals item then unmarshals it back
// into a map - most callers pass a struct, which round-trips fine, but
// that isn't the only way to reach either of its two error returns.
func TestClient_WrapParameters_NilNameErrors(t *testing.T) {
	client := NewClient(Server{}, ClientCredentials{}, ClientConfig{})

	// A func value can't be marshaled at all.
	if _, err := client.WrapParameters(nil, func() {}); err == nil {
		t.Errorf("expected a marshal error for an unmarshalable item, got nil")
	}

	// A JSON array marshals fine, but can't be unmarshaled into a map.
	if _, err := client.WrapParameters(nil, []int{1, 2, 3}); err == nil {
		t.Errorf("expected an unmarshal error for a non-object item, got nil")
	}

	// The happy path: a struct marshals to an object and round-trips.
	got, err := client.WrapParameters(nil, ForemanObject{Name: "x"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got["name"] != "x" {
		t.Errorf("expected name=x in the unwrapped map, got %v", got)
	}
}
