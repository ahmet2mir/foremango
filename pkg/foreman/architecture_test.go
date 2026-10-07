package foreman

import (
	"context"
	"testing"
)

// Exercises every ForemanArchitecture method against the dummy server.
//
// ForemanArchitecture has a custom UnmarshalJSON that reads operating
// system ids back from an "operatingsystems" array of full objects, not the
// "operatingsystem_ids" array of ints it's written as - so the dummy
// server's generic echo (which only knows the request's wire shape) can't
// faithfully round-trip that one field. Id and Name are asserted; whether
// OperatingSystemIds round-trips is Foreman's own API behavior, not this
// client's, and isn't something a generic dummy server can stand in for.
func TestArchitecture_CRUD(t *testing.T) {
	_, _, client := newDummyServer(t)
	ctx := context.Background()

	created, err := client.CreateArchitecture(ctx, &ForemanArchitecture{
		ForemanObject:      ForemanObject{Name: "x86_64"},
		OperatingSystemIds: []int{1, 2},
	})
	if err != nil {
		t.Fatalf("CreateArchitecture: %v", err)
	}
	if created.Id == 0 {
		t.Fatalf("CreateArchitecture: expected a non-zero Id, got 0")
	}
	if created.Name != "x86_64" {
		t.Fatalf("CreateArchitecture: expected Name [x86_64], got [%s]", created.Name)
	}

	read, err := client.ReadArchitecture(ctx, created.Id)
	if err != nil {
		t.Fatalf("ReadArchitecture: %v", err)
	}
	if read.Id != created.Id || read.Name != "x86_64" {
		t.Fatalf("ReadArchitecture: expected {Id: %d, Name: x86_64}, got %+v", created.Id, read)
	}

	updated, err := client.UpdateArchitecture(ctx, &ForemanArchitecture{
		ForemanObject: ForemanObject{Id: created.Id, Name: "arm64"},
	})
	if err != nil {
		t.Fatalf("UpdateArchitecture: %v", err)
	}
	if updated.Name != "arm64" {
		t.Fatalf("UpdateArchitecture: expected Name [arm64], got [%s]", updated.Name)
	}

	qr, err := client.QueryArchitecture(ctx, &ForemanArchitecture{ForemanObject: ForemanObject{Name: "arm64"}})
	if err != nil {
		t.Fatalf("QueryArchitecture: %v", err)
	}
	if len(qr.Results) == 0 {
		t.Fatalf("QueryArchitecture: expected at least one result, got none")
	}

	if err := client.DeleteArchitecture(ctx, created.Id); err != nil {
		t.Fatalf("DeleteArchitecture: %v", err)
	}
	if _, err := client.ReadArchitecture(ctx, created.Id); !IsNotFound(err) {
		t.Fatalf("ReadArchitecture after delete: expected a 404/IsNotFound error, got %v", err)
	}
}
