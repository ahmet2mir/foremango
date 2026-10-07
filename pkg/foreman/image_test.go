package foreman

import (
	"context"
	"testing"
)

func TestImage_CRUD(t *testing.T) {
	_, _, client := newDummyServer(t)
	ctx := context.Background()

	const computeResourceID = 3

	created, err := client.CreateImage(ctx, &ForemanImage{
		ForemanObject:     ForemanObject{Name: "image1"},
		ComputeResourceID: computeResourceID,
	}, computeResourceID)
	if err != nil {
		t.Fatalf("CreateImage: %v", err)
	}
	if created.Id == 0 || created.Name != "image1" {
		t.Fatalf("CreateImage: unexpected result %+v", created)
	}

	read, err := client.ReadImage(ctx, &ForemanImage{
		ForemanObject:     ForemanObject{Id: created.Id},
		ComputeResourceID: computeResourceID,
	})
	if err != nil {
		t.Fatalf("ReadImage: %v", err)
	}
	if read.Id != created.Id || read.Name != "image1" {
		t.Fatalf("ReadImage: unexpected result %+v", read)
	}

	updated, err := client.UpdateImage(ctx, &ForemanImage{
		ForemanObject:     ForemanObject{Id: created.Id, Name: "image2"},
		ComputeResourceID: computeResourceID,
	})
	if err != nil {
		t.Fatalf("UpdateImage: %v", err)
	}
	if updated.Name != "image2" {
		t.Fatalf("UpdateImage: expected Name [image2], got [%s]", updated.Name)
	}

	qr, err := client.QueryImage(ctx, &ForemanImage{ComputeResourceID: computeResourceID})
	if err != nil {
		t.Fatalf("QueryImage: %v", err)
	}
	if len(qr.Results) == 0 {
		t.Fatalf("QueryImage: expected at least one result, got none")
	}

	if err := client.DeleteImage(ctx, computeResourceID, created.Id); err != nil {
		t.Fatalf("DeleteImage: %v", err)
	}
}
