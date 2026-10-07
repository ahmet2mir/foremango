package foreman

import (
	"context"
	"testing"
)

func TestPartitionTable_CRUD(t *testing.T) {
	_, _, client := newDummyServer(t)
	ctx := context.Background()

	created, err := client.CreatePartitionTable(ctx, &ForemanPartitionTable{
		ForemanObject: ForemanObject{Name: "Kickstart default"},
		Layout:        "zerombr\nclearpart --all --initlabel",
	})
	if err != nil {
		t.Fatalf("CreatePartitionTable: %v", err)
	}
	if created.Id == 0 || created.Name != "Kickstart default" {
		t.Fatalf("CreatePartitionTable: unexpected result %+v", created)
	}

	read, err := client.ReadPartitionTable(ctx, created.Id)
	if err != nil {
		t.Fatalf("ReadPartitionTable: %v", err)
	}
	if read.Id != created.Id {
		t.Fatalf("ReadPartitionTable: expected Id [%d], got [%d]", created.Id, read.Id)
	}

	updated, err := client.UpdatePartitionTable(ctx, &ForemanPartitionTable{
		ForemanObject: ForemanObject{Id: created.Id, Name: "Kickstart renamed"},
	})
	if err != nil {
		t.Fatalf("UpdatePartitionTable: %v", err)
	}
	if updated.Name != "Kickstart renamed" {
		t.Fatalf("UpdatePartitionTable: expected Name [Kickstart renamed], got [%s]", updated.Name)
	}

	qr, err := client.QueryPartitionTable(ctx, &ForemanPartitionTable{ForemanObject: ForemanObject{Name: "Kickstart renamed"}})
	if err != nil {
		t.Fatalf("QueryPartitionTable: %v", err)
	}
	if len(qr.Results) == 0 {
		t.Fatalf("QueryPartitionTable: expected at least one result, got none")
	}

	if err := client.DeletePartitionTable(ctx, created.Id); err != nil {
		t.Fatalf("DeletePartitionTable: %v", err)
	}
}
