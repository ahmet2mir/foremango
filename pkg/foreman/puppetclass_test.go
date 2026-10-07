package foreman

import (
	"context"
	"net/http"
	"testing"
)

// ReadPuppetClass/QueryPuppetClass both GET the bare puppetclasses
// collection and ignore the id argument entirely (that's the library's own
// doing, not this test's) - so there's exactly one canned response to
// register, not a per-id store.
func TestPuppetClass_ReadQuery(t *testing.T) {
	mux, _, client := newDummyServer(t)
	ctx := context.Background()

	mux.HandleFunc("GET /foreman_puppet/api/puppetclasses", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]interface{}{"id": 1, "name": "ntp"})
	})

	read, err := client.ReadPuppetClass(ctx, 1)
	if err != nil {
		t.Fatalf("ReadPuppetClass: %v", err)
	}
	if read.Id != 1 || read.Name != "ntp" {
		t.Fatalf("ReadPuppetClass: unexpected result %+v", read)
	}

	if _, err := client.QueryPuppetClass(ctx, &ForemanPuppetClass{}); err != nil {
		t.Fatalf("QueryPuppetClass: %v", err)
	}
}
