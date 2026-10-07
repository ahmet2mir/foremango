package foreman

import (
	"context"
	"net/http"
	"strconv"
	"testing"
)

// CreateKatelloContentView always also POSTs .../publish, and
// ReadKatelloContentView/UpdateKatelloContentView always also touch
// .../filters - both registered here since they're true RPC-style actions,
// not generic CRUD, the dummy server's catch-all doesn't model them.
// Filters are left nil throughout, which keeps the nested
// Create/UpdateKatelloContentViewFilters calls as no-op empty loops.
func TestKatelloContentView_CRUD(t *testing.T) {
	mux, d, client := newDummyServer(t)
	ctx := context.Background()

	const collection = "/katello/api/content_views"

	mux.HandleFunc("POST "+collection+"/{id}/publish", func(w http.ResponseWriter, r *http.Request) {
		id, _ := strconv.Atoi(r.PathValue("id"))
		obj, ok := d.storeFor(collection).get(id)
		if !ok {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		writeJSON(w, http.StatusOK, obj)
	})
	mux.HandleFunc("PUT "+collection+"/{id}/remove", func(w http.ResponseWriter, r *http.Request) {
		id, _ := strconv.Atoi(r.PathValue("id"))
		d.storeFor(collection).delete(id)
		w.WriteHeader(http.StatusOK)
	})

	created, err := client.CreateKatelloContentView(ctx, &ContentView{
		ForemanObject: ForemanObject{Name: "cv1"},
	})
	if err != nil {
		t.Fatalf("CreateKatelloContentView: %v", err)
	}
	if created.Id == 0 || created.Name != "cv1" {
		t.Fatalf("CreateKatelloContentView: unexpected result %+v", created)
	}

	read, err := client.ReadKatelloContentView(ctx, &ContentView{ForemanObject: ForemanObject{Id: created.Id}})
	if err != nil {
		t.Fatalf("ReadKatelloContentView: %v", err)
	}
	if read.Id != created.Id || read.Name != "cv1" {
		t.Fatalf("ReadKatelloContentView: unexpected result %+v", read)
	}

	updated, err := client.UpdateKatelloContentView(ctx, &ContentView{
		ForemanObject: ForemanObject{Id: created.Id, Name: "cv2"},
	})
	if err != nil {
		t.Fatalf("UpdateKatelloContentView: %v", err)
	}
	if updated.Name != "cv2" {
		t.Fatalf("UpdateKatelloContentView: expected Name [cv2], got [%s]", updated.Name)
	}

	qr, err := client.QueryContentView(ctx, &ContentView{ForemanObject: ForemanObject{Name: "cv2"}})
	if err != nil {
		t.Fatalf("QueryContentView: %v", err)
	}
	if len(qr.Results) == 0 {
		t.Fatalf("QueryContentView: expected at least one result, got none")
	}

	if err := client.DeleteKatelloContentView(ctx, created.Id); err != nil {
		t.Fatalf("DeleteKatelloContentView: %v", err)
	}
}
