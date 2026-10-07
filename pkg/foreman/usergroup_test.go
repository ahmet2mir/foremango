package foreman

import (
	"context"
	"testing"
)

func TestUsergroup_CRUD(t *testing.T) {
	_, _, client := newDummyServer(t)
	ctx := context.Background()

	created, err := client.CreateUsergroup(ctx, &ForemanUsergroup{
		ForemanObject: ForemanObject{Name: "admins"},
		Admin:         true,
	})
	if err != nil {
		t.Fatalf("CreateUsergroup: %v", err)
	}
	if created.Id == 0 || created.Name != "admins" {
		t.Fatalf("CreateUsergroup: unexpected result %+v", created)
	}

	read, err := client.ReadUsergroup(ctx, created.Id)
	if err != nil {
		t.Fatalf("ReadUsergroup: %v", err)
	}
	if read.Id != created.Id || read.Name != "admins" {
		t.Fatalf("ReadUsergroup: unexpected result %+v", read)
	}

	updated, err := client.UpdateUsergroup(ctx, &ForemanUsergroup{
		ForemanObject: ForemanObject{Id: created.Id, Name: "operators"},
	})
	if err != nil {
		t.Fatalf("UpdateUsergroup: %v", err)
	}
	if updated.Name != "operators" {
		t.Fatalf("UpdateUsergroup: expected Name [operators], got [%s]", updated.Name)
	}

	qr, err := client.QueryUsergroup(ctx, &ForemanUsergroup{ForemanObject: ForemanObject{Name: "operators"}})
	if err != nil {
		t.Fatalf("QueryUsergroup: %v", err)
	}
	if len(qr.Results) == 0 {
		t.Fatalf("QueryUsergroup: expected at least one result, got none")
	}

	if err := client.DeleteUsergroup(ctx, created.Id); err != nil {
		t.Fatalf("DeleteUsergroup: %v", err)
	}
}
