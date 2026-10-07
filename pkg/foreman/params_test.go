package foreman

import (
	"encoding/json"
	"testing"
)

func TestForemanKVParameter_UnmarshalJSON(t *testing.T) {
	cases := []struct {
		name    string
		json    string
		want    interface{}
		wantErr bool
	}{
		{"string value", `{"name":"k","value":"v"}`, "v", false},
		{"bool value", `{"name":"k","value":true}`, true, false},
		{"map value", `{"name":"k","value":{"a":1.0}}`, map[string]interface{}{"a": 1.0}, false},
		{"unsupported array value", `{"name":"k","value":[1,2,3]}`, nil, true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var p ForemanKVParameter
			err := json.Unmarshal([]byte(c.json), &p)
			if c.wantErr {
				if err == nil {
					t.Fatalf("expected an error, got nil (value: %+v)", p.Value)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if p.Name != "k" {
				t.Fatalf("expected Name [k], got [%s]", p.Name)
			}
			switch want := c.want.(type) {
			case map[string]interface{}:
				got, ok := p.Value.(map[string]interface{})
				if !ok || got["a"] != want["a"] {
					t.Fatalf("expected Value %+v, got %+v", want, p.Value)
				}
			default:
				if p.Value != c.want {
					t.Fatalf("expected Value %v, got %v", c.want, p.Value)
				}
			}
		})
	}
}

func TestFromKV(t *testing.T) {
	if FromKV(nil) != nil {
		t.Fatalf("FromKV(nil) should return nil")
	}

	got := FromKV([]ForemanKVParameter{
		{Name: "str", Value: "hello"},
		{Name: "flag", Value: true},
		{Name: "obj", Value: map[string]interface{}{"x": 1}},
		{Name: "num", Value: 42}, // falls into the default branch
	})

	if got["str"] != "hello" {
		t.Errorf("expected str=hello, got %v", got["str"])
	}
	if got["flag"] != true {
		t.Errorf("expected flag=true, got %v", got["flag"])
	}
	if m, ok := got["obj"].(map[string]interface{}); !ok || m["x"] != 1 {
		t.Errorf("expected obj={x:1}, got %v", got["obj"])
	}
	if got["num"] != "42" {
		t.Errorf("expected num to fall through to its string form \"42\", got %v", got["num"])
	}
}

func TestToKV(t *testing.T) {
	if ToKV(nil) != nil {
		t.Fatalf("ToKV(nil) should return nil")
	}

	got := ToKV(map[string]interface{}{
		"str":  "hello",
		"flag": true,
		"obj":  map[string]interface{}{"x": 1.0},
		"num":  42, // falls into the default branch
	})

	byName := map[string]ForemanKVParameter{}
	for _, kv := range got {
		byName[kv.Name] = kv
	}

	if byName["str"].Value != "hello" {
		t.Errorf("expected str=hello, got %v", byName["str"].Value)
	}
	if byName["flag"].Value != true {
		t.Errorf("expected flag=true, got %v", byName["flag"].Value)
	}
	if _, ok := byName["obj"].Value.(json.RawMessage); !ok {
		t.Errorf("expected obj to be encoded as json.RawMessage, got %T", byName["obj"].Value)
	}
	if byName["num"].Value != "42" {
		t.Errorf("expected num to fall through to its string form \"42\", got %v", byName["num"].Value)
	}
}
