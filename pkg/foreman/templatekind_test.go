package foreman

import (
	"context"
	"testing"
)

// TemplateKind is read-only (no Create); the store is seeded directly.
func TestTemplateKind_ReadQuery(t *testing.T) {
	_, d, client := newDummyServer(t)
	ctx := context.Background()

	d.storeFor("/api/template_kinds").seed(1, map[string]interface{}{"name": "iPXE"})

	read, err := client.ReadTemplateKind(ctx, 1)
	if err != nil {
		t.Fatalf("ReadTemplateKind: %v", err)
	}
	if read.Id != 1 || read.Name != "iPXE" {
		t.Fatalf("ReadTemplateKind: unexpected result %+v", read)
	}

	qr, err := client.QueryTemplateKind(ctx, &ForemanTemplateKind{})
	if err != nil {
		t.Fatalf("QueryTemplateKind: %v", err)
	}
	if len(qr.Results) == 0 {
		t.Fatalf("QueryTemplateKind: expected at least one result, got none")
	}
}
