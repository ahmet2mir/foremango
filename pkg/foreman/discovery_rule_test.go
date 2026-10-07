package foreman

import (
	"context"
	"testing"
)

// ReadDiscoveryRule unmarshals into ForemanDiscoveryRuleResponse, a
// different type from the ForemanDiscoveryRule Create/Update/Query use -
// fields they don't share by JSON tag (e.g. HostsLimitMaxCount: "max_count"
// on the way in, "hosts_limit" on the way out) won't round-trip through the
// dummy server's generic echo, so only Name (shared) is asserted on Read.
func TestDiscoveryRule_CRUD(t *testing.T) {
	_, _, client := newDummyServer(t)
	ctx := context.Background()

	created, err := client.CreateDiscoveryRule(ctx, &ForemanDiscoveryRule{
		Name:     "rule1",
		Search:   "cpu_count = 2",
		Priority: 10,
		Enabled:  true,
	})
	if err != nil {
		t.Fatalf("CreateDiscoveryRule: %v", err)
	}
	if created.Id == 0 || created.Name != "rule1" {
		t.Fatalf("CreateDiscoveryRule: unexpected result %+v", created)
	}

	read, err := client.ReadDiscoveryRule(ctx, created.Id)
	if err != nil {
		t.Fatalf("ReadDiscoveryRule: %v", err)
	}
	if read.Id != created.Id || read.Name != "rule1" {
		t.Fatalf("ReadDiscoveryRule: unexpected result %+v", read)
	}

	updated, err := client.UpdateDiscoveryRule(ctx, &ForemanDiscoveryRule{
		ForemanObject: ForemanObject{Id: created.Id},
		Name:          "rule1-renamed",
		Priority:      10,
	})
	if err != nil {
		t.Fatalf("UpdateDiscoveryRule: %v", err)
	}
	if updated.Name != "rule1-renamed" {
		t.Fatalf("UpdateDiscoveryRule: expected Name [rule1-renamed], got [%s]", updated.Name)
	}

	qr, err := client.QueryDiscoveryRule(ctx, &ForemanDiscoveryRule{Name: "rule1-renamed"})
	if err != nil {
		t.Fatalf("QueryDiscoveryRule: %v", err)
	}
	if len(qr.Results) == 0 {
		t.Fatalf("QueryDiscoveryRule: expected at least one result, got none")
	}

	if err := client.DeleteDiscoveryRule(ctx, created.Id); err != nil {
		t.Fatalf("DeleteDiscoveryRule: %v", err)
	}
}
