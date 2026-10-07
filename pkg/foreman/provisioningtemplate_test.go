package foreman

import (
	"context"
	"testing"
)

// Name, Template, Snippet, AuditComment, Locked and Description all
// round-trip symmetrically through ForemanProvisioningTemplate's custom
// Marshal/Unmarshal. OperatingSystemIds does not (same "ids out, full
// objects in" asymmetry as ForemanArchitecture) and isn't asserted.
func TestProvisioningTemplate_CRUD(t *testing.T) {
	_, _, client := newDummyServer(t)
	ctx := context.Background()

	created, err := client.CreateProvisioningTemplate(ctx, &ForemanProvisioningTemplate{
		ForemanObject: ForemanObject{Name: "kickstart-default"},
		Template:      "<%# kickstart %>",
	})
	if err != nil {
		t.Fatalf("CreateProvisioningTemplate: %v", err)
	}
	if created.Id == 0 || created.Name != "kickstart-default" || created.Template != "<%# kickstart %>" {
		t.Fatalf("CreateProvisioningTemplate: unexpected result %+v", created)
	}

	read, err := client.ReadProvisioningTemplate(ctx, created.Id)
	if err != nil {
		t.Fatalf("ReadProvisioningTemplate: %v", err)
	}
	if read.Id != created.Id || read.Name != "kickstart-default" {
		t.Fatalf("ReadProvisioningTemplate: unexpected result %+v", read)
	}

	updated, err := client.UpdateProvisioningTemplate(ctx, &ForemanProvisioningTemplate{
		ForemanObject: ForemanObject{Id: created.Id, Name: "kickstart-renamed"},
	})
	if err != nil {
		t.Fatalf("UpdateProvisioningTemplate: %v", err)
	}
	if updated.Name != "kickstart-renamed" {
		t.Fatalf("UpdateProvisioningTemplate: expected Name [kickstart-renamed], got [%s]", updated.Name)
	}

	qr, err := client.QueryProvisioningTemplate(ctx, &ForemanProvisioningTemplate{ForemanObject: ForemanObject{Name: "kickstart-renamed"}})
	if err != nil {
		t.Fatalf("QueryProvisioningTemplate: %v", err)
	}
	if len(qr.Results) == 0 {
		t.Fatalf("QueryProvisioningTemplate: expected at least one result, got none")
	}

	if err := client.DeleteProvisioningTemplate(ctx, created.Id); err != nil {
		t.Fatalf("DeleteProvisioningTemplate: %v", err)
	}
}
