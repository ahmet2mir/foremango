package foreman

import (
	"context"
	"encoding/json"
	"testing"
)

func TestKatelloContentViewFilters_CRUD(t *testing.T) {
	_, _, client := newDummyServer(t)
	ctx := context.Background()

	const cvID = 4

	created, err := client.CreateKatelloContentViewFilters(ctx, cvID, &[]ContentViewFilter{
		{ForemanObject: ForemanObject{Name: "filter1"}, Type: "rpm", Inclusion: true},
	})
	if err != nil {
		t.Fatalf("CreateKatelloContentViewFilters: %v", err)
	}
	if len(*created) != 1 || (*created)[0].Id == 0 || (*created)[0].Name != "filter1" {
		t.Fatalf("CreateKatelloContentViewFilters: unexpected result %+v", *created)
	}
	filterID := (*created)[0].Id

	rules, err := client.CreateKatelloContentViewFilterRules(ctx, filterID, &[]ContentViewFilterRule{
		{Architecture: "x86_64"},
	})
	if err != nil {
		t.Fatalf("CreateKatelloContentViewFilterRules: %v", err)
	}
	if len(*rules) != 1 || (*rules)[0].Id == 0 {
		t.Fatalf("CreateKatelloContentViewFilterRules: unexpected result %+v", *rules)
	}

	qr, err := client.QueryContentViewFilters(ctx, cvID)
	if err != nil {
		t.Fatalf("QueryContentViewFilters: %v", err)
	}
	if len(qr.Results) == 0 {
		t.Fatalf("QueryContentViewFilters: expected at least one result, got none")
	}

	read, err := client.ReadKatelloContentViewFilters(ctx, cvID)
	if err != nil {
		t.Fatalf("ReadKatelloContentViewFilters: %v", err)
	}
	if len(*read) == 0 {
		t.Fatalf("ReadKatelloContentViewFilters: expected at least one filter, got none")
	}

	ruleID := (*rules)[0].Id
	updated, err := client.UpdateKatelloContentViewFilters(ctx, cvID, &[]ContentViewFilter{
		{
			ForemanObject: ForemanObject{Id: filterID, Name: "filter1-renamed"},
			Rules: []ContentViewFilterRule{
				{ForemanObject: ForemanObject{Id: ruleID}, Architecture: "aarch64"},
			},
		},
	})
	if err != nil {
		t.Fatalf("UpdateKatelloContentViewFilters: %v", err)
	}
	if len(*updated) != 1 || (*updated)[0].Name != "filter1-renamed" {
		t.Fatalf("UpdateKatelloContentViewFilters: unexpected result %+v", *updated)
	}
}

// ContentViewFilter.MarshalJSON has a pointer receiver, but every real call
// site (CreateKatelloContentViewFilters above included) ranges over a
// []ContentViewFilter and passes the loop variable - a value, not a
// pointer - to WrapJSONWithTaxonomy, so this method never actually runs in
// practice; json.Marshal only promotes pointer-receiver methods for values
// it can take the address of itself, and a method value already boxed in
// an interface can't be. Called directly here, since nothing else does.
func TestContentViewFilter_MarshalJSON(t *testing.T) {
	cvf := &ContentViewFilter{
		ForemanObject: ForemanObject{Id: 1, Name: "filter1"},
		Type:          "rpm",
		Inclusion:     true,
		Description:   "desc",
	}
	b, err := cvf.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON: %v", err)
	}
	var m map[string]interface{}
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatalf("decoding MarshalJSON output: %v", err)
	}
	if m["name"] != "filter1" || m["type"] != "rpm" || m["inclusion"] != true {
		t.Fatalf("unexpected marshaled output: %s", b)
	}
}
