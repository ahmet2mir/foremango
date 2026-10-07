package foreman

import (
	"context"
	"testing"
)

func TestModel_CRUD(t *testing.T) {
	_, _, client := newDummyServer(t)
	ctx := context.Background()

	created, err := client.CreateModel(ctx, &ForemanModel{
		ForemanObject: ForemanObject{Name: "PowerEdge R630"},
		VendorClass:   "Dell",
	})
	if err != nil {
		t.Fatalf("CreateModel: %v", err)
	}
	if created.Id == 0 || created.Name != "PowerEdge R630" {
		t.Fatalf("CreateModel: unexpected result %+v", created)
	}

	read, err := client.ReadModel(ctx, created.Id)
	if err != nil {
		t.Fatalf("ReadModel: %v", err)
	}
	if read.Id != created.Id || read.Name != "PowerEdge R630" {
		t.Fatalf("ReadModel: unexpected result %+v", read)
	}

	updated, err := client.UpdateModel(ctx, &ForemanModel{
		ForemanObject: ForemanObject{Id: created.Id, Name: "PowerEdge R640"},
	})
	if err != nil {
		t.Fatalf("UpdateModel: %v", err)
	}
	if updated.Name != "PowerEdge R640" {
		t.Fatalf("UpdateModel: expected Name [PowerEdge R640], got [%s]", updated.Name)
	}

	qr, err := client.QueryModel(ctx, &ForemanModel{ForemanObject: ForemanObject{Name: "PowerEdge R640"}})
	if err != nil {
		t.Fatalf("QueryModel: %v", err)
	}
	if len(qr.Results) == 0 {
		t.Fatalf("QueryModel: expected at least one result, got none")
	}

	if err := client.DeleteModel(ctx, created.Id); err != nil {
		t.Fatalf("DeleteModel: %v", err)
	}
}
