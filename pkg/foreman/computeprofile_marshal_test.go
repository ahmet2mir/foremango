package foreman

import (
	"encoding/json"
	"testing"
)

// Covers every branch of ForemanComputeAttribute.MarshalJSON's VMAttrs
// type-switch directly; the CRUD test only exercises a subset.
func TestForemanComputeAttribute_MarshalJSON_VMAttrsTypes(t *testing.T) {
	type unhandled struct{ X int }

	ca := &ForemanComputeAttribute{
		ForemanObject:     ForemanObject{Id: 1, Name: "attr1"},
		ComputeResourceId: 2,
		VMAttrs: map[string]interface{}{
			"int_val":            4,
			"float32_val":        float32(1.5),
			"float64_val":        2.5,
			"bool_val":           true,
			"nil_val":            nil,
			"string_json_val":    "42",       // valid JSON -> unmarshals to a number
			"string_plain_val":   "not-json", // invalid JSON -> kept as the raw string
			"map_val":            map[string]interface{}{"a": 1},
			"slice_val":          []interface{}{1, 2},
			"unhandled_type_val": unhandled{X: 1},
		},
	}

	b, err := ca.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON: %v", err)
	}

	var decoded struct {
		VMAttrs map[string]interface{} `json:"vm_attrs"`
	}
	if err := json.Unmarshal(b, &decoded); err != nil {
		t.Fatalf("decoding MarshalJSON output: %v", err)
	}

	attrs := decoded.VMAttrs
	if attrs["int_val"] != "4" {
		t.Errorf("int_val = %v, want \"4\"", attrs["int_val"])
	}
	if attrs["float32_val"] != "1.5" {
		t.Errorf("float32_val = %v, want \"1.5\"", attrs["float32_val"])
	}
	if attrs["float64_val"] != "2.5" {
		t.Errorf("float64_val = %v, want \"2.5\"", attrs["float64_val"])
	}
	if attrs["bool_val"] != "true" {
		t.Errorf("bool_val = %v, want \"true\"", attrs["bool_val"])
	}
	if attrs["nil_val"] != nil {
		t.Errorf("nil_val = %v, want nil", attrs["nil_val"])
	}
	if attrs["string_json_val"] != 42.0 {
		t.Errorf("string_json_val = %v, want 42.0 (unmarshaled from its JSON form)", attrs["string_json_val"])
	}
	if attrs["string_plain_val"] != "not-json" {
		t.Errorf("string_plain_val = %v, want \"not-json\"", attrs["string_plain_val"])
	}
	if _, ok := attrs["map_val"].(string); !ok {
		t.Errorf("map_val = %v (%T), want a JSON-encoded string", attrs["map_val"], attrs["map_val"])
	}
	if _, ok := attrs["slice_val"].(string); !ok {
		t.Errorf("slice_val = %v (%T), want a JSON-encoded string", attrs["slice_val"], attrs["slice_val"])
	}
	// unhandled_type_val hits the default case, which only logs - no value
	// is ever set for that key.
	if _, present := attrs["unhandled_type_val"]; present {
		t.Errorf("unhandled_type_val = %v, want absent (default case sets nothing)", attrs["unhandled_type_val"])
	}
}
