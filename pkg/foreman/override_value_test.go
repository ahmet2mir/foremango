package foreman

import (
	"context"
	"testing"
)

func TestOverrideValue_CRUD(t *testing.T) {
	_, _, client := newDummyServer(t)
	ctx := context.Background()

	const scpID = 5

	created, err := client.CreateOverrideValue(ctx, &ForemanOverrideValue{
		SmartClassParameterId: scpID,
		MatchType:             "fqdn",
		MatchValue:            "host1.example.com",
		Value:                 "custom-value",
	})
	if err != nil {
		t.Fatalf("CreateOverrideValue: %v", err)
	}
	if created.Id == 0 || created.MatchType != "fqdn" || created.Value != "custom-value" {
		t.Fatalf("CreateOverrideValue: unexpected result %+v", created)
	}

	read, err := client.ReadOverrideValue(ctx, created.Id, scpID)
	if err != nil {
		t.Fatalf("ReadOverrideValue: %v", err)
	}
	if read.Id != created.Id || read.MatchValue != "host1.example.com" {
		t.Fatalf("ReadOverrideValue: unexpected result %+v", read)
	}

	updated, err := client.UpdateOverrideValue(ctx, &ForemanOverrideValue{
		ForemanObject:         ForemanObject{Id: created.Id},
		SmartClassParameterId: scpID,
		MatchType:             "fqdn",
		MatchValue:            "host1.example.com",
		Value:                 "updated-value",
	})
	if err != nil {
		t.Fatalf("UpdateOverrideValue: %v", err)
	}
	if updated.Value != "updated-value" {
		t.Fatalf("UpdateOverrideValue: expected Value [updated-value], got [%s]", updated.Value)
	}

	if err := client.DeleteOverrideValue(ctx, created.Id, scpID); err != nil {
		t.Fatalf("DeleteOverrideValue: %v", err)
	}
}
