package foreman

import (
	"context"
	"testing"
)

func TestComputeResource_CRUD(t *testing.T) {
	_, _, client := newDummyServer(t)
	ctx := context.Background()

	created, err := client.CreateComputeResource(ctx, &ForemanComputeResource{
		Name:     "libvirt1",
		URL:      "qemu+ssh://root@libvirt.example.com/system",
		Provider: "Libvirt",
	})
	if err != nil {
		t.Fatalf("CreateComputeResource: %v", err)
	}
	if created.Id == 0 || created.Name != "libvirt1" {
		t.Fatalf("CreateComputeResource: unexpected result %+v", created)
	}

	read, err := client.ReadComputeResource(ctx, created.Id)
	if err != nil {
		t.Fatalf("ReadComputeResource: %v", err)
	}
	if read.Id != created.Id {
		t.Fatalf("ReadComputeResource: expected Id [%d], got [%d]", created.Id, read.Id)
	}

	updated, err := client.UpdateComputeResource(ctx, &ForemanComputeResource{
		ForemanObject: ForemanObject{Id: created.Id},
		Name:          "libvirt2",
		Provider:      "Libvirt",
	})
	if err != nil {
		t.Fatalf("UpdateComputeResource: %v", err)
	}
	if updated.Name != "libvirt2" {
		t.Fatalf("UpdateComputeResource: expected Name [libvirt2], got [%s]", updated.Name)
	}

	qr, err := client.QueryComputeResource(ctx, &ForemanComputeResource{Name: "libvirt2"})
	if err != nil {
		t.Fatalf("QueryComputeResource: %v", err)
	}
	if len(qr.Results) == 0 {
		t.Fatalf("QueryComputeResource: expected at least one result, got none")
	}

	if err := client.DeleteComputeResource(ctx, created.Id); err != nil {
		t.Fatalf("DeleteComputeResource: %v", err)
	}
}
