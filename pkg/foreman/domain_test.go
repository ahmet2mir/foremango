package foreman

import (
	"context"
	"testing"
)

func TestDomain_CRUD(t *testing.T) {
	_, _, client := newDummyServer(t)
	ctx := context.Background()

	created, err := client.CreateDomain(ctx, &ForemanDomain{
		ForemanObject: ForemanObject{Name: "example.com"},
		Fullname:      "Example domain",
	})
	if err != nil {
		t.Fatalf("CreateDomain: %v", err)
	}
	if created.Id == 0 || created.Name != "example.com" {
		t.Fatalf("CreateDomain: unexpected result %+v", created)
	}

	read, err := client.ReadDomain(ctx, created.Id)
	if err != nil {
		t.Fatalf("ReadDomain: %v", err)
	}
	if read.Id != created.Id || read.Name != "example.com" {
		t.Fatalf("ReadDomain: unexpected result %+v", read)
	}

	updated, err := client.UpdateDomain(ctx, &ForemanDomain{Fullname: "Renamed domain"}, created.Id)
	if err != nil {
		t.Fatalf("UpdateDomain: %v", err)
	}
	if updated.Fullname != "Renamed domain" {
		t.Fatalf("UpdateDomain: expected Fullname [Renamed domain], got [%s]", updated.Fullname)
	}

	qr, err := client.QueryDomain(ctx, &ForemanDomain{ForemanObject: ForemanObject{Name: "example.com"}})
	if err != nil {
		t.Fatalf("QueryDomain: %v", err)
	}
	if len(qr.Results) == 0 {
		t.Fatalf("QueryDomain: expected at least one result, got none")
	}

	if err := client.DeleteDomain(ctx, created.Id); err != nil {
		t.Fatalf("DeleteDomain: %v", err)
	}
}
