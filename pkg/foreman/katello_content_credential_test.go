package foreman

import (
	"context"
	"testing"
)

func TestKatelloContentCredential_CRUD(t *testing.T) {
	_, _, client := newDummyServer(t)
	ctx := context.Background()

	created, err := client.CreateKatelloContentCredential(ctx, &ForemanKatelloContentCredential{
		ForemanObject: ForemanObject{Name: "gpg-key-1"},
		Content:       "-----BEGIN PGP PUBLIC KEY BLOCK-----",
	})
	if err != nil {
		t.Fatalf("CreateKatelloContentCredential: %v", err)
	}
	if created.Id == 0 || created.Name != "gpg-key-1" {
		t.Fatalf("CreateKatelloContentCredential: unexpected result %+v", created)
	}

	read, err := client.ReadKatelloContentCredential(ctx, created.Id)
	if err != nil {
		t.Fatalf("ReadKatelloContentCredential: %v", err)
	}
	if read.Id != created.Id || read.Name != "gpg-key-1" {
		t.Fatalf("ReadKatelloContentCredential: unexpected result %+v", read)
	}

	updated, err := client.UpdateKatelloContentCredential(ctx, &ForemanKatelloContentCredential{
		ForemanObject: ForemanObject{Id: created.Id, Name: "gpg-key-2"},
	})
	if err != nil {
		t.Fatalf("UpdateKatelloContentCredential: %v", err)
	}
	if updated.Name != "gpg-key-2" {
		t.Fatalf("UpdateKatelloContentCredential: expected Name [gpg-key-2], got [%s]", updated.Name)
	}

	qr, err := client.QueryKatelloContentCredential(ctx, &ForemanKatelloContentCredential{ForemanObject: ForemanObject{Name: "gpg-key-2"}})
	if err != nil {
		t.Fatalf("QueryKatelloContentCredential: %v", err)
	}
	if len(qr.Results) == 0 {
		t.Fatalf("QueryKatelloContentCredential: expected at least one result, got none")
	}

	if err := client.DeleteKatelloContentCredential(ctx, created.Id); err != nil {
		t.Fatalf("DeleteKatelloContentCredential: %v", err)
	}
}
