package foreman

import (
	"context"
	"net/http"
	"testing"
)

// Setting is read-only and its Id is a string (Foreman setting names, e.g.
// "http_proxy"), not the numeric id the dummy server's generic catch-all
// expects at the end of a path - so this registers its own handler instead
// of relying on it.
func TestSetting_ReadQuery(t *testing.T) {
	mux, _, client := newDummyServer(t)
	ctx := context.Background()

	const settingID = "http_proxy"
	mux.HandleFunc("GET /api/settings/"+settingID, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"id":    settingID,
			"value": "https://proxy.example.com",
		})
	})
	mux.HandleFunc("GET /api/settings", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"results": []map[string]interface{}{
				{"id": settingID, "value": "https://proxy.example.com"},
			},
			"total":    1,
			"subtotal": 1,
		})
	})

	read, err := client.ReadSetting(ctx, settingID)
	if err != nil {
		t.Fatalf("ReadSetting: %v", err)
	}
	if read.Id != settingID {
		t.Fatalf("ReadSetting: expected Id [%s], got [%s]", settingID, read.Id)
	}

	qr, err := client.QuerySetting(ctx, &ForemanSetting{})
	if err != nil {
		t.Fatalf("QuerySetting: %v", err)
	}
	if len(qr.Results) == 0 {
		t.Fatalf("QuerySetting: expected at least one result, got none")
	}
}
