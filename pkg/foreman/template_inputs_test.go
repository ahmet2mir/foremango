package foreman

import (
	"context"
	"testing"
)

func TestTemplateInput_CRUD(t *testing.T) {
	_, _, client := newDummyServer(t)
	ctx := context.Background()

	const templateID = 9

	created, err := client.CreateTemplateInput(ctx, &ForemanTemplateInput{
		ForemanObject: ForemanObject{Name: "hostname"},
		TemplateId:    templateID,
		InputType:     "user",
	})
	if err != nil {
		t.Fatalf("CreateTemplateInput: %v", err)
	}
	if created.Id == 0 || created.Name != "hostname" {
		t.Fatalf("CreateTemplateInput: unexpected result %+v", created)
	}

	read, err := client.ReadTemplateInput(ctx, &ForemanTemplateInput{
		ForemanObject: ForemanObject{Id: created.Id},
		TemplateId:    templateID,
	})
	if err != nil {
		t.Fatalf("ReadTemplateInput: %v", err)
	}
	if read.Id != created.Id || read.Name != "hostname" {
		t.Fatalf("ReadTemplateInput: unexpected result %+v", read)
	}

	updated, err := client.UpdateTemplateInput(ctx, &ForemanTemplateInput{
		ForemanObject: ForemanObject{Id: created.Id, Name: "hostname-renamed"},
		TemplateId:    templateID,
	})
	if err != nil {
		t.Fatalf("UpdateTemplateInput: %v", err)
	}
	if updated.Name != "hostname-renamed" {
		t.Fatalf("UpdateTemplateInput: expected Name [hostname-renamed], got [%s]", updated.Name)
	}

	qr, err := client.QueryTemplateInput(ctx, &ForemanTemplateInput{
		ForemanObject: ForemanObject{Name: "hostname-renamed"},
		TemplateId:    templateID,
	})
	if err != nil {
		t.Fatalf("QueryTemplateInput: %v", err)
	}
	if len(qr.Results) == 0 {
		t.Fatalf("QueryTemplateInput: expected at least one result, got none")
	}

	if err := client.DeleteTemplateInput(ctx, &ForemanTemplateInput{
		ForemanObject: ForemanObject{Id: created.Id},
		TemplateId:    templateID,
	}); err != nil {
		t.Fatalf("DeleteTemplateInput: %v", err)
	}
}
