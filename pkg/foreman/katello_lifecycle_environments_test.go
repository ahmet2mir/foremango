package foreman

import (
	"context"
	"testing"
)

func TestKatelloLifecycleEnvironment_CRUD(t *testing.T) {
	_, _, client := newDummyServer(t)
	ctx := context.Background()

	created, err := client.CreateKatelloLifecycleEnvironment(ctx, &LifecycleEnvironment{
		ForemanObject: ForemanObject{Name: "Dev"},
		Label:         "dev",
	})
	if err != nil {
		t.Fatalf("CreateKatelloLifecycleEnvironment: %v", err)
	}
	if created.Id == 0 || created.Name != "Dev" {
		t.Fatalf("CreateKatelloLifecycleEnvironment: unexpected result %+v", created)
	}

	read, err := client.ReadKatelloLifecycleEnvironment(ctx, &LifecycleEnvironment{ForemanObject: ForemanObject{Id: created.Id}})
	if err != nil {
		t.Fatalf("ReadKatelloLifecycleEnvironment: %v", err)
	}
	if read.Id != created.Id || read.Name != "Dev" {
		t.Fatalf("ReadKatelloLifecycleEnvironment: unexpected result %+v", read)
	}

	updated, err := client.UpdateKatelloLifecycleEnvironment(ctx, &LifecycleEnvironment{
		ForemanObject: ForemanObject{Id: created.Id, Name: "Staging"},
	})
	if err != nil {
		t.Fatalf("UpdateKatelloLifecycleEnvironment: %v", err)
	}
	if updated.Name != "Staging" {
		t.Fatalf("UpdateKatelloLifecycleEnvironment: expected Name [Staging], got [%s]", updated.Name)
	}

	qr, err := client.QueryLifecycleEnvironment(ctx, &LifecycleEnvironment{ForemanObject: ForemanObject{Name: "Staging"}})
	if err != nil {
		t.Fatalf("QueryLifecycleEnvironment: %v", err)
	}
	if len(qr.Results) == 0 {
		t.Fatalf("QueryLifecycleEnvironment: expected at least one result, got none")
	}

	if err := client.DeleteKatelloLifecycleEnvironment(ctx, created.Id); err != nil {
		t.Fatalf("DeleteKatelloLifecycleEnvironment: %v", err)
	}
}
