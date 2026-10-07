package foreman

import (
	"context"
	"testing"
)

func TestKatelloProduct_CRUD(t *testing.T) {
	_, _, client := newDummyServer(t)
	ctx := context.Background()

	created, err := client.CreateKatelloProduct(ctx, &ForemanKatelloProduct{
		ForemanObject: ForemanObject{Name: "product1"},
	})
	if err != nil {
		t.Fatalf("CreateKatelloProduct: %v", err)
	}
	if created.Id == 0 || created.Name != "product1" {
		t.Fatalf("CreateKatelloProduct: unexpected result %+v", created)
	}

	read, err := client.ReadKatelloProduct(ctx, created.Id)
	if err != nil {
		t.Fatalf("ReadKatelloProduct: %v", err)
	}
	if read.Id != created.Id || read.Name != "product1" {
		t.Fatalf("ReadKatelloProduct: unexpected result %+v", read)
	}

	updated, err := client.UpdateKatelloProduct(ctx, &ForemanKatelloProduct{
		ForemanObject: ForemanObject{Id: created.Id, Name: "product2"},
	})
	if err != nil {
		t.Fatalf("UpdateKatelloProduct: %v", err)
	}
	if updated.Name != "product2" {
		t.Fatalf("UpdateKatelloProduct: expected Name [product2], got [%s]", updated.Name)
	}

	qr, err := client.QueryKatelloProduct(ctx, &ForemanKatelloProduct{ForemanObject: ForemanObject{Name: "product2"}})
	if err != nil {
		t.Fatalf("QueryKatelloProduct: %v", err)
	}
	if len(qr.Results) == 0 {
		t.Fatalf("QueryKatelloProduct: expected at least one result, got none")
	}

	if err := client.DeleteKatelloProduct(ctx, created.Id); err != nil {
		t.Fatalf("DeleteKatelloProduct: %v", err)
	}
}
