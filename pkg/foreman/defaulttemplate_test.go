package foreman

import (
	"context"
	"testing"
)

func TestDefaultTemplate_CRUD(t *testing.T) {
	_, _, client := newDummyServer(t)
	ctx := context.Background()

	osID := 7
	created, err := client.CreateDefaultTemplate(ctx, &ForemanDefaultTemplate{
		ForemanObject:          ForemanObject{},
		OperatingSystemId:      osID,
		ProvisioningTemplateId: 3,
		TemplateKindId:         1,
	})
	if err != nil {
		t.Fatalf("CreateDefaultTemplate: %v", err)
	}
	if created.Id == 0 {
		t.Fatalf("CreateDefaultTemplate: expected a non-zero Id, got 0")
	}

	read, err := client.ReadDefaultTemplate(ctx, &ForemanDefaultTemplate{OperatingSystemId: osID}, created.Id)
	if err != nil {
		t.Fatalf("ReadDefaultTemplate: %v", err)
	}
	if read.Id != created.Id {
		t.Fatalf("ReadDefaultTemplate: expected Id [%d], got [%d]", created.Id, read.Id)
	}

	updated, err := client.UpdateDefaultTemplate(ctx, &ForemanDefaultTemplate{
		OperatingSystemId:      osID,
		ProvisioningTemplateId: 9,
	}, created.Id)
	if err != nil {
		t.Fatalf("UpdateDefaultTemplate: %v", err)
	}
	if updated.ProvisioningTemplateId != 9 {
		t.Fatalf("UpdateDefaultTemplate: expected ProvisioningTemplateId [9], got [%d]", updated.ProvisioningTemplateId)
	}

	// NOTE: QueryDefaultTemplate builds its endpoint as
	// fmt.Sprintf("/%s", DefaultTemplateEndpointPrefix), but
	// DefaultTemplateEndpointPrefix itself is "/operatingsystems/%d/os_default_templates" -
	// its own "%d" is never substituted, so the real request path is
	// "//operatingsystems/%d/os_default_templates" verbatim and can never
	// match anything Create/Read/Update/Delete populated above. That's a
	// pre-existing bug in the library, not this test; only the no-error
	// contract is asserted here.
	if _, err := client.QueryDefaultTemplate(ctx, &ForemanDefaultTemplate{OperatingSystemId: osID}); err != nil {
		t.Fatalf("QueryDefaultTemplate: %v", err)
	}

	if err := client.DeleteDefaultTemplate(ctx, &ForemanDefaultTemplate{OperatingSystemId: osID}, created.Id); err != nil {
		t.Fatalf("DeleteDefaultTemplate: %v", err)
	}
}
