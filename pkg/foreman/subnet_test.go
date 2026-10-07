package foreman

import (
	"context"
	"testing"
)

func TestSubnet_CRUD(t *testing.T) {
	_, _, client := newDummyServer(t)
	ctx := context.Background()

	created, err := client.CreateSubnet(ctx, &ForemanSubnet{
		ForemanObject: ForemanObject{Name: "subnet1"},
		Network:       "192.168.100.0",
		Mask:          "255.255.255.0",
	})
	if err != nil {
		t.Fatalf("CreateSubnet: %v", err)
	}
	if created.Id == 0 || created.Name != "subnet1" {
		t.Fatalf("CreateSubnet: unexpected result %+v", created)
	}

	read, err := client.ReadSubnet(ctx, created.Id)
	if err != nil {
		t.Fatalf("ReadSubnet: %v", err)
	}
	if read.Id != created.Id || read.Name != "subnet1" {
		t.Fatalf("ReadSubnet: unexpected result %+v", read)
	}

	updated, err := client.UpdateSubnet(ctx, &ForemanSubnet{
		ForemanObject: ForemanObject{Id: created.Id, Name: "subnet2"},
	})
	if err != nil {
		t.Fatalf("UpdateSubnet: %v", err)
	}
	if updated.Name != "subnet2" {
		t.Fatalf("UpdateSubnet: expected Name [subnet2], got [%s]", updated.Name)
	}

	qr, err := client.QuerySubnet(ctx, &ForemanSubnet{ForemanObject: ForemanObject{Name: "subnet2"}})
	if err != nil {
		t.Fatalf("QuerySubnet: %v", err)
	}
	if len(qr.Results) == 0 {
		t.Fatalf("QuerySubnet: expected at least one result, got none")
	}

	if err := client.DeleteSubnet(ctx, created.Id); err != nil {
		t.Fatalf("DeleteSubnet: %v", err)
	}
}
