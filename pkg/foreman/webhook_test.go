package foreman

import (
	"context"
	"testing"
)

func TestWebhook_CRUD(t *testing.T) {
	_, _, client := newDummyServer(t)
	ctx := context.Background()

	created, err := client.CreateWebhook(ctx, &ForemanWebhook{
		ForemanObject: ForemanObject{Name: "hook1"},
		TargetURL:     "https://example.com/hook",
		HTTPMethod:    "POST",
		Event:         "actions.katello.content_view.promote_succeeded",
	})
	if err != nil {
		t.Fatalf("CreateWebhook: %v", err)
	}
	if created.Id == 0 || created.Name != "hook1" {
		t.Fatalf("CreateWebhook: unexpected result %+v", created)
	}

	read, err := client.ReadWebhook(ctx, created.Id)
	if err != nil {
		t.Fatalf("ReadWebhook: %v", err)
	}
	if read.Id != created.Id || read.Name != "hook1" {
		t.Fatalf("ReadWebhook: unexpected result %+v", read)
	}

	updated, err := client.UpdateWebhook(ctx, &ForemanWebhook{
		ForemanObject: ForemanObject{Id: created.Id, Name: "hook2"},
		TargetURL:     "https://example.com/hook2",
	})
	if err != nil {
		t.Fatalf("UpdateWebhook: %v", err)
	}
	if updated.Name != "hook2" {
		t.Fatalf("UpdateWebhook: expected Name [hook2], got [%s]", updated.Name)
	}

	qr, err := client.QueryWebhook(ctx, &ForemanWebhook{ForemanObject: ForemanObject{Name: "hook2"}})
	if err != nil {
		t.Fatalf("QueryWebhook: %v", err)
	}
	if len(qr.Results) == 0 {
		t.Fatalf("QueryWebhook: expected at least one result, got none")
	}

	if err := client.DeleteWebhook(ctx, created.Id); err != nil {
		t.Fatalf("DeleteWebhook: %v", err)
	}
}
