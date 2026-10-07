package foreman

import (
	"context"
	"testing"
)

// TemplateInputs are left empty here: when non-empty, CreateJobTemplate
// loops and calls CreateTemplateInput for each one against a nested
// endpoint - that path is exercised directly in template_inputs_test.go.
func TestJobTemplate_CRUD(t *testing.T) {
	_, _, client := newDummyServer(t)
	ctx := context.Background()

	created, err := client.CreateJobTemplate(ctx, &ForemanJobTemplate{
		ForemanObject: ForemanObject{Name: "run-command"},
		Template:      "<%= input(\"command\") %>",
	})
	if err != nil {
		t.Fatalf("CreateJobTemplate: %v", err)
	}
	if created.Id == 0 || created.Name != "run-command" {
		t.Fatalf("CreateJobTemplate: unexpected result %+v", created)
	}

	read, err := client.ReadJobTemplate(ctx, created.Id)
	if err != nil {
		t.Fatalf("ReadJobTemplate: %v", err)
	}
	if read.Id != created.Id || read.Name != "run-command" {
		t.Fatalf("ReadJobTemplate: unexpected result %+v", read)
	}

	updated, err := client.UpdateJobTemplate(ctx, &ForemanJobTemplate{
		ForemanObject: ForemanObject{Id: created.Id, Name: "run-command-v2"},
	})
	if err != nil {
		t.Fatalf("UpdateJobTemplate: %v", err)
	}
	if updated.Name != "run-command-v2" {
		t.Fatalf("UpdateJobTemplate: expected Name [run-command-v2], got [%s]", updated.Name)
	}

	qr, err := client.QueryJobTemplate(ctx, &ForemanJobTemplate{ForemanObject: ForemanObject{Name: "run-command-v2"}})
	if err != nil {
		t.Fatalf("QueryJobTemplate: %v", err)
	}
	if len(qr.Results) == 0 {
		t.Fatalf("QueryJobTemplate: expected at least one result, got none")
	}

	if err := client.DeleteJobTemplate(ctx, created); err != nil {
		t.Fatalf("DeleteJobTemplate: %v", err)
	}
}
