package foreman

import (
	"context"
	"net/http"
	"testing"
)

// Like ReadPuppetClass, ReadSmartClassParameter GETs a bare collection
// endpoint and ignores its id argument. QuerySmartClassParameter uses a
// different, puppet-class-scoped endpoint.
func TestSmartClassParameter_ReadQuery(t *testing.T) {
	mux, _, client := newDummyServer(t)
	ctx := context.Background()

	mux.HandleFunc("GET /foreman_puppet/api/smarts_class_paramaters", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]interface{}{"id": 1, "parameter": "ntp_servers"})
	})
	mux.HandleFunc("GET /foreman_puppet/api/puppetclasses/{puppetClassID}/smart_class_parameters", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"results":  []map[string]interface{}{{"id": 1, "parameter": "ntp_servers"}},
			"total":    1,
			"subtotal": 1,
		})
	})

	read, err := client.ReadSmartClassParameter(ctx, 1)
	if err != nil {
		t.Fatalf("ReadSmartClassParameter: %v", err)
	}
	if read.Id != 1 || read.Parameter != "ntp_servers" {
		t.Fatalf("ReadSmartClassParameter: unexpected result %+v", read)
	}

	qr, err := client.QuerySmartClassParameter(ctx, &ForemanSmartClassParameter{PuppetClassId: 3})
	if err != nil {
		t.Fatalf("QuerySmartClassParameter: %v", err)
	}
	if len(qr.Results) == 0 {
		t.Fatalf("QuerySmartClassParameter: expected at least one result, got none")
	}
}
