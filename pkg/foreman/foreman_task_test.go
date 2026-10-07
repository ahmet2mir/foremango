package foreman

import (
	"bytes"
	"context"
	"net/http"
	"testing"
)

// SendAndParse's handling of a 202 response is the one piece of non-trivial
// protocol logic in this whole client: it polls /foreman_tasks/api/tasks/:id
// (via waitForKatelloAsyncTask) until the task stops being pending, then -
// for any task whose Label isn't one of the two Katello content-view
// special cases - parses the ORIGINAL 202 response body into obj (not a
// re-fetch of the finished task). This exercises that path end to end.
func TestForemanTask_AsyncSendAndParse(t *testing.T) {
	mux, _, client := newDummyServer(t)
	ctx := context.Background()

	mux.HandleFunc("POST /api/myresource", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusAccepted, map[string]interface{}{
			"id":      "task-1",
			"pending": true,
			"label":   "Actions::Something::Else",
			"name":    "accepted-resource",
		})
	})
	mux.HandleFunc("GET /foreman_tasks/api/tasks/task-1", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"id":      "task-1",
			"pending": false,
			"label":   "Actions::Something::Else",
			"result":  "success",
		})
	})

	req, err := client.NewRequestWithContext(ctx, http.MethodPost, "/myresource", bytes.NewBufferString("{}"))
	if err != nil {
		t.Fatalf("NewRequestWithContext: %v", err)
	}

	var result struct {
		Id   string `json:"id"`
		Name string `json:"name"`
	}
	if err := client.SendAndParse(req, &result); err != nil {
		t.Fatalf("SendAndParse: %v", err)
	}
	if result.Name != "accepted-resource" {
		t.Fatalf("SendAndParse: expected the original 202 body's Name [accepted-resource], got [%s]", result.Name)
	}
}
