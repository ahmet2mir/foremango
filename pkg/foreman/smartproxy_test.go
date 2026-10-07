package foreman

import (
	"context"
	"testing"
)

func TestSmartProxy_CRUD(t *testing.T) {
	_, _, client := newDummyServer(t)
	ctx := context.Background()

	created, err := client.CreateSmartProxy(ctx, &ForemanSmartProxy{
		ForemanObject: ForemanObject{Name: "proxy1"},
		URL:           "https://smartproxy.example.com:8443",
	})
	if err != nil {
		t.Fatalf("CreateSmartProxy: %v", err)
	}
	if created.Id == 0 || created.Name != "proxy1" {
		t.Fatalf("CreateSmartProxy: unexpected result %+v", created)
	}

	read, err := client.ReadSmartProxy(ctx, created.Id)
	if err != nil {
		t.Fatalf("ReadSmartProxy: %v", err)
	}
	if read.Id != created.Id || read.Name != "proxy1" {
		t.Fatalf("ReadSmartProxy: unexpected result %+v", read)
	}

	updated, err := client.UpdateSmartProxy(ctx, &ForemanSmartProxy{
		ForemanObject: ForemanObject{Id: created.Id, Name: "proxy2"},
	})
	if err != nil {
		t.Fatalf("UpdateSmartProxy: %v", err)
	}
	if updated.Name != "proxy2" {
		t.Fatalf("UpdateSmartProxy: expected Name [proxy2], got [%s]", updated.Name)
	}

	qr, err := client.QuerySmartProxy(ctx, &ForemanSmartProxy{ForemanObject: ForemanObject{Name: "proxy2"}})
	if err != nil {
		t.Fatalf("QuerySmartProxy: %v", err)
	}
	if len(qr.Results) == 0 {
		t.Fatalf("QuerySmartProxy: expected at least one result, got none")
	}

	if err := client.DeleteSmartProxy(ctx, created.Id); err != nil {
		t.Fatalf("DeleteSmartProxy: %v", err)
	}
}
