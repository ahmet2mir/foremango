package foreman

import (
	"context"
	"net/http"
	"testing"
)

// ForemanHost decodes into the private foremanHostDecode type, which has
// several fields Create/UpdateHost also unconditionally reads back via a
// GET to <id>/vm_compute_attributes - the dummy server's generic catch-all
// answers that with an empty list shape, which Host happily treats as "no
// compute attributes". Only Name is asserted on round-trip.
func TestHost_CRUD(t *testing.T) {
	_, _, client := newDummyServer(t)
	ctx := context.Background()

	created, err := client.CreateHost(ctx, &ForemanHost{
		ForemanObject: ForemanObject{Name: "host1.example.com"},
	}, 1)
	if err != nil {
		t.Fatalf("CreateHost: %v", err)
	}
	if created.Id == 0 || created.Name != "host1.example.com" {
		t.Fatalf("CreateHost: unexpected result %+v", created)
	}

	read, err := client.ReadHost(ctx, created.Id)
	if err != nil {
		t.Fatalf("ReadHost: %v", err)
	}
	if read.Id != created.Id || read.Name != "host1.example.com" {
		t.Fatalf("ReadHost: unexpected result %+v", read)
	}

	updated, err := client.UpdateHost(ctx, &ForemanHost{
		ForemanObject: ForemanObject{Id: created.Id, Name: "host2.example.com"},
	}, 1)
	if err != nil {
		t.Fatalf("UpdateHost: %v", err)
	}
	if updated.Name != "host2.example.com" {
		t.Fatalf("UpdateHost: expected Name [host2.example.com], got [%s]", updated.Name)
	}

	if err := client.DeleteHost(ctx, created.Id); err != nil {
		t.Fatalf("DeleteHost: %v", err)
	}
}

// SendPowerCommand parses its response back into the request value itself
// (as a generic map, per encoding/json's behavior unmarshaling into an
// interface{}), then checks response["power"] != false to decide success -
// so the mock must answer specifically {"power": true}, not the generic
// store-backed echo.
func TestHost_SendPowerCommand(t *testing.T) {
	mux, _, client := newDummyServer(t)
	ctx := context.Background()

	mux.HandleFunc("PUT /api/hosts/{id}/power", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]interface{}{"power": true})
	})

	host := &ForemanHost{ForemanObject: ForemanObject{Id: 1, Name: "host1.example.com"}}
	cmd := Power{PowerAction: "start", Power: true}
	if err := client.SendPowerCommand(ctx, host, cmd, 1); err != nil {
		t.Fatalf("SendPowerCommand: %v", err)
	}
}

// Covers SendPowerCommand's other branches: BMCBoot (a different URL
// suffix than Power), an unsupported command type, and a reported failure.
func TestHost_SendPowerCommand_OtherBranches(t *testing.T) {
	mux, _, client := newDummyServer(t)
	ctx := context.Background()
	host := &ForemanHost{ForemanObject: ForemanObject{Id: 1, Name: "host1.example.com"}}

	mux.HandleFunc("PUT /api/hosts/{id}/boot", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"boot": map[string]interface{}{"result": true},
		})
	})
	if err := client.SendPowerCommand(ctx, host, BMCBoot{Device: "disk"}, 1); err != nil {
		t.Fatalf("SendPowerCommand (BMCBoot): %v", err)
	}

	if err := client.SendPowerCommand(ctx, host, "not-a-valid-command", 1); err == nil {
		t.Fatalf("SendPowerCommand: expected an error for an unsupported command type, got nil")
	}

	mux.HandleFunc("PUT /api/hosts/{id}/power", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]interface{}{"power": false})
	})
	if err := client.SendPowerCommand(ctx, host, Power{PowerAction: "start", Power: true}, 1); err == nil {
		t.Fatalf("SendPowerCommand: expected an error when the server reports power=false, got nil")
	}
}
