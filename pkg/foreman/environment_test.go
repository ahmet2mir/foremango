package foreman

import (
	"context"
	"testing"
)

func TestEnvironment_CRUD(t *testing.T) {
	_, _, client := newDummyServer(t)
	ctx := context.Background()

	created, err := client.CreateEnvironment(ctx, &ForemanEnvironment{
		ForemanObject: ForemanObject{Name: "production"},
	})
	if err != nil {
		t.Fatalf("CreateEnvironment: %v", err)
	}
	if created.Id == 0 || created.Name != "production" {
		t.Fatalf("CreateEnvironment: unexpected result %+v", created)
	}

	read, err := client.ReadEnvironment(ctx, created.Id)
	if err != nil {
		t.Fatalf("ReadEnvironment: %v", err)
	}
	if read.Id != created.Id || read.Name != "production" {
		t.Fatalf("ReadEnvironment: unexpected result %+v", read)
	}

	updated, err := client.UpdateEnvironment(ctx, &ForemanEnvironment{
		ForemanObject: ForemanObject{Id: created.Id, Name: "staging"},
	})
	if err != nil {
		t.Fatalf("UpdateEnvironment: %v", err)
	}
	if updated.Name != "staging" {
		t.Fatalf("UpdateEnvironment: expected Name [staging], got [%s]", updated.Name)
	}

	qr, err := client.QueryEnvironment(ctx, &ForemanEnvironment{ForemanObject: ForemanObject{Name: "staging"}})
	if err != nil {
		t.Fatalf("QueryEnvironment: %v", err)
	}
	if len(qr.Results) == 0 {
		t.Fatalf("QueryEnvironment: expected at least one result, got none")
	}

	if err := client.DeleteEnvironment(ctx, created.Id); err != nil {
		t.Fatalf("DeleteEnvironment: %v", err)
	}
}
