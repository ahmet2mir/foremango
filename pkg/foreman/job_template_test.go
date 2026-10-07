package foreman

import (
	"context"
	"fmt"
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

// Exercises CreateJobTemplate's nested-create loop: when TemplateInputs is
// non-empty, it POSTs each one individually (see template_inputs_test.go
// for that endpoint directly) before returning.
func TestJobTemplate_CreateWithTemplateInputs(t *testing.T) {
	_, _, client := newDummyServer(t)
	ctx := context.Background()

	created, err := client.CreateJobTemplate(ctx, &ForemanJobTemplate{
		ForemanObject: ForemanObject{Name: "with-inputs"},
		TemplateInputs: []ForemanTemplateInput{
			{ForemanObject: ForemanObject{Name: "command"}, InputType: "user"},
		},
	})
	if err != nil {
		t.Fatalf("CreateJobTemplate: %v", err)
	}
	if len(created.TemplateInputs) != 1 || created.TemplateInputs[0].Id == 0 {
		t.Fatalf("CreateJobTemplate: expected one created template input, got %+v", created.TemplateInputs)
	}
}

// Exercises ReadJobTemplate's mirror-image loop: when the stored job
// template's own "template_inputs" is non-empty, it re-reads each one
// individually (and sorts them by Id first). Both stores are seeded
// directly since the data has to be consistent between them from the
// start, not pieced together through Create.
func TestJobTemplate_ReadWithTemplateInputs(t *testing.T) {
	_, d, client := newDummyServer(t)
	ctx := context.Background()

	const jtID = 3
	d.storeFor("/api/job_templates").seed(jtID, map[string]interface{}{
		"name": "with-inputs",
		"template_inputs": []map[string]interface{}{
			{"id": 9, "name": "command"},
			{"id": 8, "name": "verbosity"},
		},
	})
	inputsPath := fmt.Sprintf("/api/templates/%d/template_inputs", jtID)
	d.storeFor(inputsPath).seed(9, map[string]interface{}{"name": "command"})
	d.storeFor(inputsPath).seed(8, map[string]interface{}{"name": "verbosity"})

	read, err := client.ReadJobTemplate(ctx, jtID)
	if err != nil {
		t.Fatalf("ReadJobTemplate: %v", err)
	}
	if len(read.TemplateInputs) != 2 {
		t.Fatalf("ReadJobTemplate: expected 2 template inputs, got %+v", read.TemplateInputs)
	}
	// sort.SliceStable orders them by Id ascending.
	if read.TemplateInputs[0].Id != 8 || read.TemplateInputs[1].Id != 9 {
		t.Fatalf("ReadJobTemplate: expected template inputs sorted by Id [8, 9], got [%d, %d]",
			read.TemplateInputs[0].Id, read.TemplateInputs[1].Id)
	}
}

// Exercises UpdateJobTemplate's own nested loop: when TemplateInputs is
// non-empty, it PUTs each one individually via UpdateTemplateInput.
func TestJobTemplate_UpdateWithTemplateInputs(t *testing.T) {
	_, _, client := newDummyServer(t)
	ctx := context.Background()

	const jtID = 4
	updated, err := client.UpdateJobTemplate(ctx, &ForemanJobTemplate{
		ForemanObject: ForemanObject{Id: jtID, Name: "renamed"},
		TemplateInputs: []ForemanTemplateInput{
			{ForemanObject: ForemanObject{Id: 1, Name: "command"}, TemplateId: jtID},
		},
	})
	if err != nil {
		t.Fatalf("UpdateJobTemplate: %v", err)
	}
	if len(updated.TemplateInputs) != 1 {
		t.Fatalf("UpdateJobTemplate: expected one updated template input, got %+v", updated.TemplateInputs)
	}
}
