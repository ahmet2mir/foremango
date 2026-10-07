package foreman

import (
	"context"
	"testing"
)

func TestCommonParameter_CRUD(t *testing.T) {
	_, _, client := newDummyServer(t)
	ctx := context.Background()

	created, err := client.CreateCommonParameter(ctx, &ForemanCommonParameter{
		Name:  "key1",
		Value: "value1",
	})
	if err != nil {
		t.Fatalf("CreateCommonParameter: %v", err)
	}
	if created.Id == 0 || created.Value != "value1" {
		t.Fatalf("CreateCommonParameter: unexpected result %+v", created)
	}

	read, err := client.ReadCommonParameter(ctx, &ForemanCommonParameter{}, created.Id)
	if err != nil {
		t.Fatalf("ReadCommonParameter: %v", err)
	}
	if read.Id != created.Id || read.Value != "value1" {
		t.Fatalf("ReadCommonParameter: unexpected result %+v", read)
	}

	updated, err := client.UpdateCommonParameter(ctx, &ForemanCommonParameter{Value: "value2"}, created.Id)
	if err != nil {
		t.Fatalf("UpdateCommonParameter: %v", err)
	}
	if updated.Value != "value2" {
		t.Fatalf("UpdateCommonParameter: expected Value [value2], got [%s]", updated.Value)
	}

	qr, err := client.QueryCommonParameter(ctx, &ForemanCommonParameter{Name: "key1"})
	if err != nil {
		t.Fatalf("QueryCommonParameter: %v", err)
	}
	if len(qr.Results) == 0 {
		t.Fatalf("QueryCommonParameter: expected at least one result, got none")
	}

	if err := client.DeleteCommonParameter(ctx, &ForemanCommonParameter{}, created.Id); err != nil {
		t.Fatalf("DeleteCommonParameter: %v", err)
	}
}
