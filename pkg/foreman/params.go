package foreman

import (
	"encoding/json"
	"fmt"
)

// KVParameters are used in all inline Parameter Maps. i.e. Host, HostGroup
type ForemanKVParameter struct {
	Name  string      `json:"name"`
	Value interface{} `json:"value"`
}

// UnmarshalJSON custom unmarshaling for ForemanKVParameter to handle different types.
//
// This method allows for different types in the "value" field of the
// ForemanKVParameter.
func (p *ForemanKVParameter) UnmarshalJSON(data []byte) error {
	// Define a temporary structure to decode JSON into
	var temp struct {
		Name  string          `json:"name"`
		Value json.RawMessage `json:"value"`
	}

	// Unmarshal into the temporary structure
	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}

	// Set the name field
	p.Name = temp.Name

	// Try to unmarshal the value into different types
	var strValue string
	if err := json.Unmarshal(temp.Value, &strValue); err == nil {
		p.Value = strValue
		return nil
	}

	var boolValue bool
	if err := json.Unmarshal(temp.Value, &boolValue); err == nil {
		p.Value = boolValue
		return nil
	}

	var jsonValue map[string]interface{}
	if err := json.Unmarshal(temp.Value, &jsonValue); err == nil {
		p.Value = jsonValue
		return nil
	}

	return fmt.Errorf("unsupported value type for parameter: %s", string(temp.Value))
}

// JSON object for creating and updating puppetattributes on hosts and hostgroups
type PuppetAttribute struct {
	Puppetclass_ids []int `json:"puppetclass_ids"`
	ConfigGroup_ids []int `json:"config_group_ids"`
}

// FromKV returns a map[string]interface{} from a slice of ForemanKVParameters.
//
// This function handles different types in the "value" field of the
// ForemanKVParameter and converts them to the appropriate type in the
// returned map[string]interface{}.
func FromKV(kv []ForemanKVParameter) map[string]interface{} {
	if kv == nil {
		return nil
	}

	ret := make(map[string]interface{}, len(kv))
	for _, pair := range kv {
		switch v := pair.Value.(type) {
		case string:
			ret[pair.Name] = v
		case bool:
			ret[pair.Name] = v
		case map[string]interface{}:
			ret[pair.Name] = v
		default:
			// Handle unexpected types or log an error
			ret[pair.Name] = fmt.Sprintf("%v", v)
		}
	}
	return ret
}

// ToKV converts a map[string]interface{} to a slice of ForemanKVParameters.
//
// This function handles different types in the map and converts them to the
// appropriate type in the returned slice of ForemanKVParameters.
func ToKV(m map[string]interface{}) []ForemanKVParameter {
	if m == nil {
		return nil
	}

	var ret []ForemanKVParameter
	for key, value := range m {
		switch v := value.(type) {
		case string:
			ret = append(ret, ForemanKVParameter{Name: key, Value: v})
		case bool:
			ret = append(ret, ForemanKVParameter{Name: key, Value: v})
		case map[string]interface{}:
			jsonValue, err := json.Marshal(v)
			if err != nil {
				Errorf("Error marshalling JSON for %s: %v", key, err)
				continue
			}
			ret = append(ret, ForemanKVParameter{Name: key, Value: json.RawMessage(jsonValue)})
		default:
			ret = append(ret, ForemanKVParameter{Name: key, Value: fmt.Sprintf("%v", v)})
		}
	}
	return ret
}
