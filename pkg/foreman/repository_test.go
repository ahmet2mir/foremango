package foreman

import (
	"context"
	"testing"
)

func TestKatelloRepository_CRUD(t *testing.T) {
	_, _, client := newDummyServer(t)
	ctx := context.Background()

	created, err := client.CreateKatelloRepository(ctx, &ForemanKatelloRepository{
		ForemanObject: ForemanObject{Name: "repo1"},
		ContentType:   "yum",
		Url:           "https://repo.example.com/repo1",
	})
	if err != nil {
		t.Fatalf("CreateKatelloRepository: %v", err)
	}
	if created.Id == 0 || created.Name != "repo1" {
		t.Fatalf("CreateKatelloRepository: unexpected result %+v", created)
	}

	read, err := client.ReadKatelloRepository(ctx, created.Id)
	if err != nil {
		t.Fatalf("ReadKatelloRepository: %v", err)
	}
	if read.Id != created.Id || read.Name != "repo1" {
		t.Fatalf("ReadKatelloRepository: unexpected result %+v", read)
	}

	updated, err := client.UpdateKatelloRepository(ctx, &ForemanKatelloRepository{
		ForemanObject: ForemanObject{Id: created.Id, Name: "repo2"},
	})
	if err != nil {
		t.Fatalf("UpdateKatelloRepository: %v", err)
	}
	if updated.Name != "repo2" {
		t.Fatalf("UpdateKatelloRepository: expected Name [repo2], got [%s]", updated.Name)
	}

	qr, err := client.QueryKatelloRepository(ctx, &ForemanKatelloRepository{ForemanObject: ForemanObject{Name: "repo2"}})
	if err != nil {
		t.Fatalf("QueryKatelloRepository: %v", err)
	}
	if len(qr.Results) == 0 {
		t.Fatalf("QueryKatelloRepository: expected at least one result, got none")
	}

	if err := client.DeleteKatelloRepository(ctx, created.Id); err != nil {
		t.Fatalf("DeleteKatelloRepository: %v", err)
	}
}
