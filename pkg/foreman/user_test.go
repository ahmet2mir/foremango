package foreman

import (
	"context"
	"testing"
)

func TestUser_CRUD(t *testing.T) {
	_, _, client := newDummyServer(t)
	ctx := context.Background()

	created, err := client.CreateUser(ctx, &ForemanUser{
		ForemanObject: ForemanObject{Name: "alice"},
		Login:         "alice",
		Firstname:     "Alice",
	})
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}
	if created.Id == 0 || created.Login != "alice" {
		t.Fatalf("CreateUser: unexpected result %+v", created)
	}

	read, err := client.ReadUser(ctx, created.Id)
	if err != nil {
		t.Fatalf("ReadUser: %v", err)
	}
	if read.Id != created.Id || read.Login != "alice" {
		t.Fatalf("ReadUser: unexpected result %+v", read)
	}

	updated, err := client.UpdateUser(ctx, &ForemanUser{
		ForemanObject: ForemanObject{Id: created.Id},
		Login:         "alice2",
	})
	if err != nil {
		t.Fatalf("UpdateUser: %v", err)
	}
	if updated.Login != "alice2" {
		t.Fatalf("UpdateUser: expected Login [alice2], got [%s]", updated.Login)
	}

	qr, err := client.QueryUser(ctx, &ForemanUser{Login: "alice2"})
	if err != nil {
		t.Fatalf("QueryUser: %v", err)
	}
	if len(qr.Results) == 0 {
		t.Fatalf("QueryUser: expected at least one result, got none")
	}

	if err := client.DeleteUser(ctx, created.Id); err != nil {
		t.Fatalf("DeleteUser: %v", err)
	}
}

// QueryUser builds its search query from the first of Description,
// Firstname, Lastname, Mail, or Login that's set; each is its own branch.
func TestQueryUser_SearchFields(t *testing.T) {
	_, _, client := newDummyServer(t)
	ctx := context.Background()

	for _, u := range []*ForemanUser{
		{Description: "desc"},
		{Firstname: "First"},
		{Lastname: "Last"},
		{Mail: "a@b.com"},
		{Login: "login"},
		{}, // none set
	} {
		if _, err := client.QueryUser(ctx, u); err != nil {
			t.Fatalf("QueryUser(%+v): %v", u, err)
		}
	}
}
