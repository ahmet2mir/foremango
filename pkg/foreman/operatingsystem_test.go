package foreman

import (
	"context"
	"testing"
)

func TestOperatingSystem_CRUD(t *testing.T) {
	_, _, client := newDummyServer(t)
	ctx := context.Background()

	created, err := client.CreateOperatingSystem(ctx, &ForemanOperatingSystem{
		ForemanObject: ForemanObject{Name: "RedHat"},
		Major:         "9",
	})
	if err != nil {
		t.Fatalf("CreateOperatingSystem: %v", err)
	}
	if created.Id == 0 || created.Name != "RedHat" {
		t.Fatalf("CreateOperatingSystem: unexpected result %+v", created)
	}

	read, err := client.ReadOperatingSystem(ctx, created.Id)
	if err != nil {
		t.Fatalf("ReadOperatingSystem: %v", err)
	}
	if read.Id != created.Id {
		t.Fatalf("ReadOperatingSystem: expected Id [%d], got [%d]", created.Id, read.Id)
	}

	updated, err := client.UpdateOperatingSystem(ctx, &ForemanOperatingSystem{
		ForemanObject: ForemanObject{Id: created.Id, Name: "RedHat-renamed"},
	})
	if err != nil {
		t.Fatalf("UpdateOperatingSystem: %v", err)
	}
	if updated.Name != "RedHat-renamed" {
		t.Fatalf("UpdateOperatingSystem: expected Name [RedHat-renamed], got [%s]", updated.Name)
	}

	qr, err := client.QueryOperatingSystem(ctx, &ForemanOperatingSystem{ForemanObject: ForemanObject{Name: "RedHat-renamed"}})
	if err != nil {
		t.Fatalf("QueryOperatingSystem: %v", err)
	}
	if len(qr.Results) == 0 {
		t.Fatalf("QueryOperatingSystem: expected at least one result, got none")
	}

	if err := client.DeleteOperatingSystem(ctx, created.Id); err != nil {
		t.Fatalf("DeleteOperatingSystem: %v", err)
	}
}
