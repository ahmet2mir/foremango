package foreman

import (
	"context"
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

	updated, err := client.UpdateKatelloContentViewFilters(ctx, cvID, &[]ContentViewFilter{
		{ForemanObject: ForemanObject{Id: filterID, Name: "filter1-renamed"}},
	})
	if err != nil {
		t.Fatalf("UpdateKatelloContentViewFilters: %v", err)
	}
	if len(*updated) != 1 || (*updated)[0].Name != "filter1-renamed" {
		t.Fatalf("UpdateKatelloContentViewFilters: unexpected result %+v", *updated)
	}
}
