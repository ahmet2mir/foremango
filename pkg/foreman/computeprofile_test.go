package foreman

import (
	"context"
	"testing"
)

// ComputeAttributes are exercised by their own CreateComputeprofile loop
// against a nested endpoint; this test sticks to a profile with none, which
// keeps CreateComputeprofile to its single plain POST.
func TestComputeProfile_CRUD(t *testing.T) {
	_, _, client := newDummyServer(t)
	ctx := context.Background()

	created, err := client.CreateComputeprofile(ctx, &ForemanComputeProfile{
		ForemanObject: ForemanObject{Name: "1-small"},
	})
	if err != nil {
		t.Fatalf("CreateComputeprofile: %v", err)
	}
	if created.Id == 0 || created.Name != "1-small" {
		t.Fatalf("CreateComputeprofile: unexpected result %+v", created)
	}

	read, err := client.ReadComputeProfile(ctx, created.Id)
	if err != nil {
		t.Fatalf("ReadComputeProfile: %v", err)
	}
	if read.Id != created.Id {
		t.Fatalf("ReadComputeProfile: expected Id [%d], got [%d]", created.Id, read.Id)
	}

	updated, err := client.UpdateComputeProfile(ctx, &ForemanComputeProfile{
		ForemanObject: ForemanObject{Id: created.Id, Name: "2-medium"},
	})
	if err != nil {
		t.Fatalf("UpdateComputeProfile: %v", err)
	}
	if updated.Name != "2-medium" {
		t.Fatalf("UpdateComputeProfile: expected Name [2-medium], got [%s]", updated.Name)
	}

	qr, err := client.QueryComputeProfile(ctx, &ForemanComputeProfile{ForemanObject: ForemanObject{Name: "2-medium"}})
	if err != nil {
		t.Fatalf("QueryComputeProfile: %v", err)
	}
	if len(qr.Results) == 0 {
		t.Fatalf("QueryComputeProfile: expected at least one result, got none")
	}

	if err := client.DeleteComputeProfile(ctx, created.Id); err != nil {
		t.Fatalf("DeleteComputeProfile: %v", err)
	}
}
