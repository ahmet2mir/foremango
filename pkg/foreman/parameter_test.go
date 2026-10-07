package foreman

import (
	"context"
	"testing"
)

// ForemanParameter's custom UnmarshalJSON reads "name"/"value" from the
// top level of the response, but MarshalJSON (the default, struct-tag
// based one - there is no custom Marshal here) nests them under
// "parameter": {...} per the Parameter field's json tag. That asymmetry
// means Parameter.Name/Value don't round-trip through the dummy server's
// generic echo; only Id is asserted after Read.
func TestParameter_CRUD(t *testing.T) {
	_, _, client := newDummyServer(t)
	ctx := context.Background()

	created, err := client.CreateParameter(ctx, &ForemanParameter{
		HostID:    7,
		Parameter: ForemanKVParameter{Name: "env", Value: "prod"},
	})
	if err != nil {
		t.Fatalf("CreateParameter: %v", err)
	}
	if created.Id == 0 {
		t.Fatalf("CreateParameter: expected a non-zero Id, got 0")
	}

	read, err := client.ReadParameter(ctx, &ForemanParameter{HostID: 7}, created.Id)
	if err != nil {
		t.Fatalf("ReadParameter: %v", err)
	}
	if read.Id != created.Id {
		t.Fatalf("ReadParameter: expected Id [%d], got [%d]", created.Id, read.Id)
	}

	updated, err := client.UpdateParameter(ctx, &ForemanParameter{
		HostID:    7,
		Parameter: ForemanKVParameter{Name: "env", Value: "staging"},
	}, created.Id)
	if err != nil {
		t.Fatalf("UpdateParameter: %v", err)
	}
	if updated.Id != created.Id {
		t.Fatalf("UpdateParameter: expected Id [%d], got [%d]", created.Id, updated.Id)
	}

	if _, err := client.QueryParameter(ctx, &ForemanParameter{HostID: 7}); err != nil {
		t.Fatalf("QueryParameter: %v", err)
	}

	if err := client.DeleteParameter(ctx, &ForemanParameter{HostID: 7}, created.Id); err != nil {
		t.Fatalf("DeleteParameter: %v", err)
	}
}
