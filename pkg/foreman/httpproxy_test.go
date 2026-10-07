package foreman

import (
	"context"
	"testing"
)

func TestHTTPProxy_CRUD(t *testing.T) {
	_, _, client := newDummyServer(t)
	ctx := context.Background()

	created, err := client.CreateHTTPProxy(ctx, &ForemanHTTPProxy{
		ForemanObject: ForemanObject{Name: "proxy1"},
		URL:           "https://proxy.example.com:8080",
	})
	if err != nil {
		t.Fatalf("CreateHTTPProxy: %v", err)
	}
	if created.Id == 0 || created.Name != "proxy1" {
		t.Fatalf("CreateHTTPProxy: unexpected result %+v", created)
	}

	read, err := client.ReadHTTPProxy(ctx, created.Id)
	if err != nil {
		t.Fatalf("ReadHTTPProxy: %v", err)
	}
	if read.Id != created.Id || read.Name != "proxy1" {
		t.Fatalf("ReadHTTPProxy: unexpected result %+v", read)
	}

	updated, err := client.UpdateHTTPProxy(ctx, &ForemanHTTPProxy{
		ForemanObject: ForemanObject{Id: created.Id, Name: "proxy2"},
	})
	if err != nil {
		t.Fatalf("UpdateHTTPProxy: %v", err)
	}
	if updated.Name != "proxy2" {
		t.Fatalf("UpdateHTTPProxy: expected Name [proxy2], got [%s]", updated.Name)
	}

	qr, err := client.QueryHTTPProxy(ctx, &ForemanHTTPProxy{ForemanObject: ForemanObject{Name: "proxy2"}})
	if err != nil {
		t.Fatalf("QueryHTTPProxy: %v", err)
	}
	if len(qr.Results) == 0 {
		t.Fatalf("QueryHTTPProxy: expected at least one result, got none")
	}

	if err := client.DeleteHTTPProxy(ctx, created.Id); err != nil {
		t.Fatalf("DeleteHTTPProxy: %v", err)
	}
}
