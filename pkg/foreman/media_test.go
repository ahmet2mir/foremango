package foreman

import (
	"context"
	"testing"
)

func TestMedia_CRUD(t *testing.T) {
	_, _, client := newDummyServer(t)
	ctx := context.Background()

	created, err := client.CreateMedia(ctx, &ForemanMedia{
		ForemanObject: ForemanObject{Name: "CentOS Mirror"},
	})
	if err != nil {
		t.Fatalf("CreateMedia: %v", err)
	}
	if created.Id == 0 || created.Name != "CentOS Mirror" {
		t.Fatalf("CreateMedia: unexpected result %+v", created)
	}

	read, err := client.ReadMedia(ctx, created.Id)
	if err != nil {
		t.Fatalf("ReadMedia: %v", err)
	}
	if read.Id != created.Id {
		t.Fatalf("ReadMedia: expected Id [%d], got [%d]", created.Id, read.Id)
	}

	updated, err := client.UpdateMedia(ctx, &ForemanMedia{
		ForemanObject: ForemanObject{Id: created.Id, Name: "Fedora Mirror"},
	})
	if err != nil {
		t.Fatalf("UpdateMedia: %v", err)
	}
	if updated.Name != "Fedora Mirror" {
		t.Fatalf("UpdateMedia: expected Name [Fedora Mirror], got [%s]", updated.Name)
	}

	qr, err := client.QueryMedia(ctx, &ForemanMedia{ForemanObject: ForemanObject{Name: "Fedora Mirror"}})
	if err != nil {
		t.Fatalf("QueryMedia: %v", err)
	}
	if len(qr.Results) == 0 {
		t.Fatalf("QueryMedia: expected at least one result, got none")
	}

	if err := client.DeleteMedia(ctx, created.Id); err != nil {
		t.Fatalf("DeleteMedia: %v", err)
	}
}
