package foreman

import (
	"context"
	"testing"
)

func TestKatelloSyncPlan_CRUD(t *testing.T) {
	_, _, client := newDummyServer(t)
	ctx := context.Background()

	created, err := client.CreateKatelloSyncPlan(ctx, &ForemanKatelloSyncPlan{
		ForemanObject: ForemanObject{Name: "nightly"},
		Interval:      "daily",
	})
	if err != nil {
		t.Fatalf("CreateKatelloSyncPlan: %v", err)
	}
	if created.Id == 0 || created.Name != "nightly" {
		t.Fatalf("CreateKatelloSyncPlan: unexpected result %+v", created)
	}

	read, err := client.ReadKatelloSyncPlan(ctx, created.Id)
	if err != nil {
		t.Fatalf("ReadKatelloSyncPlan: %v", err)
	}
	if read.Id != created.Id || read.Name != "nightly" {
		t.Fatalf("ReadKatelloSyncPlan: unexpected result %+v", read)
	}

	updated, err := client.UpdateKatelloSyncPlan(ctx, &ForemanKatelloSyncPlan{
		ForemanObject: ForemanObject{Id: created.Id, Name: "weekly"},
	})
	if err != nil {
		t.Fatalf("UpdateKatelloSyncPlan: %v", err)
	}
	if updated.Name != "weekly" {
		t.Fatalf("UpdateKatelloSyncPlan: expected Name [weekly], got [%s]", updated.Name)
	}

	qr, err := client.QueryKatelloSyncPlan(ctx, &ForemanKatelloSyncPlan{ForemanObject: ForemanObject{Name: "weekly"}})
	if err != nil {
		t.Fatalf("QueryKatelloSyncPlan: %v", err)
	}
	if len(qr.Results) == 0 {
		t.Fatalf("QueryKatelloSyncPlan: expected at least one result, got none")
	}

	if err := client.DeleteKatelloSyncPlan(ctx, created.Id); err != nil {
		t.Fatalf("DeleteKatelloSyncPlan: %v", err)
	}
}
