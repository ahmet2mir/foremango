package foreman

import (
	"context"
	"testing"
)

func TestWebhookTemplate_CRUD(t *testing.T) {
	_, _, client := newDummyServer(t)
	ctx := context.Background()

	created, err := client.CreateWebhookTemplate(ctx, &ForemanWebhookTemplate{
		ForemanObject: ForemanObject{Name: "default-payload"},
		Template:      `{"event": "<%= @event %>"}`,
	})
	if err != nil {
		t.Fatalf("CreateWebhookTemplate: %v", err)
	}
	if created.Id == 0 || created.Name != "default-payload" {
		t.Fatalf("CreateWebhookTemplate: unexpected result %+v", created)
	}

	read, err := client.ReadWebhookTemplate(ctx, created.Id)
	if err != nil {
		t.Fatalf("ReadWebhookTemplate: %v", err)
	}
	if read.Id != created.Id || read.Name != "default-payload" {
		t.Fatalf("ReadWebhookTemplate: unexpected result %+v", read)
	}

	updated, err := client.UpdateWebhookTemplate(ctx, &ForemanWebhookTemplate{
		ForemanObject: ForemanObject{Id: created.Id, Name: "default-payload-v2"},
	})
	if err != nil {
		t.Fatalf("UpdateWebhookTemplate: %v", err)
	}
	if updated.Name != "default-payload-v2" {
		t.Fatalf("UpdateWebhookTemplate: expected Name [default-payload-v2], got [%s]", updated.Name)
	}

	qr, err := client.QueryWebhookTemplate(ctx, &ForemanWebhookTemplate{ForemanObject: ForemanObject{Name: "default-payload-v2"}})
	if err != nil {
		t.Fatalf("QueryWebhookTemplate: %v", err)
	}
	if len(qr.Results) == 0 {
		t.Fatalf("QueryWebhookTemplate: expected at least one result, got none")
	}

	if err := client.DeleteWebhookTemplate(ctx, created.Id); err != nil {
		t.Fatalf("DeleteWebhookTemplate: %v", err)
	}
}
