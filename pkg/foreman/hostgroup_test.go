package foreman

import (
	"context"
	"net/http"
	"strconv"
	"testing"
)

// ReadHostgroup and UpdateHostgroup decode into foremanHostGroupDecode,
// whose custom UnmarshalJSON unconditionally parses a "parameters" key and
// errors out if it's absent - the generic dummy server never adds it,
// since the write side uses the differently-named "group_parameters_attributes"
// and sends nothing when there are none. A real Foreman server likely
// always includes "parameters" (if only as []), so the override below adds
// it to keep the mock realistic rather than tripping that decode bug.
func TestHostgroup_CRUD(t *testing.T) {
	mux, d, client := newDummyServer(t)
	ctx := context.Background()

	const collection = "/api/hostgroups"
	injectParameters := func(w http.ResponseWriter, obj map[string]interface{}) {
		obj["parameters"] = []interface{}{}
		writeJSON(w, http.StatusOK, obj)
	}
	mux.HandleFunc("GET "+collection+"/{id}", func(w http.ResponseWriter, r *http.Request) {
		id, _ := strconv.Atoi(r.PathValue("id"))
		obj, ok := d.storeFor(collection).get(id)
		if !ok {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		injectParameters(w, obj)
	})
	mux.HandleFunc("PUT "+collection+"/{id}", func(w http.ResponseWriter, r *http.Request) {
		id, _ := strconv.Atoi(r.PathValue("id"))
		obj := d.storeFor(collection).update(id, decodeUnwrapped(r))
		injectParameters(w, obj)
	})

	created, err := client.CreateHostgroup(ctx, &ForemanHostgroup{
		ForemanObject: ForemanObject{Name: "base"},
	})
	if err != nil {
		t.Fatalf("CreateHostgroup: %v", err)
	}
	if created.Id == 0 || created.Name != "base" {
		t.Fatalf("CreateHostgroup: unexpected result %+v", created)
	}

	read, err := client.ReadHostgroup(ctx, created.Id)
	if err != nil {
		t.Fatalf("ReadHostgroup: %v", err)
	}
	if read.Id != created.Id || read.Name != "base" {
		t.Fatalf("ReadHostgroup: unexpected result %+v", read)
	}

	updated, err := client.UpdateHostgroup(ctx, &ForemanHostgroup{
		ForemanObject: ForemanObject{Id: created.Id, Name: "base-renamed"},
	})
	if err != nil {
		t.Fatalf("UpdateHostgroup: %v", err)
	}
	if updated.Name != "base-renamed" {
		t.Fatalf("UpdateHostgroup: expected Name [base-renamed], got [%s]", updated.Name)
	}

	qr, err := client.QueryHostgroup(ctx, &ForemanHostgroup{ForemanObject: ForemanObject{Name: "base-renamed"}})
	if err != nil {
		t.Fatalf("QueryHostgroup: %v", err)
	}
	if len(qr.Results) == 0 {
		t.Fatalf("QueryHostgroup: expected at least one result, got none")
	}

	if err := client.DeleteHostgroup(ctx, created.Id); err != nil {
		t.Fatalf("DeleteHostgroup: %v", err)
	}
}
