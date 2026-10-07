package foreman

import (
	"context"
	"net/http"
	"strconv"
	"testing"
)

// Drives CreateKatelloContentView's publish step, and DeleteKatelloContentView's
// remove step, through a real 202 + Katello task poll - the two branches in
// SendAndParse's statusCode==202 switch that the synchronous-response tests
// in katello_content_views_test.go don't reach.
func TestKatelloContentView_AsyncPublishAndRemove(t *testing.T) {
	mux, d, client := newDummyServer(t)
	ctx := context.Background()

	const collection = "/katello/api/content_views"

	var publishedID int // set by the /publish handler, read back by the task poll below
	mux.HandleFunc("POST "+collection+"/{id}/publish", func(w http.ResponseWriter, r *http.Request) {
		publishedID, _ = strconv.Atoi(r.PathValue("id"))
		writeJSON(w, http.StatusAccepted, map[string]interface{}{
			"id":      "task-publish",
			"pending": true,
			"label":   "Actions::Katello::ContentView::Publish",
		})
	})
	mux.HandleFunc("GET /foreman_tasks/api/tasks/task-publish", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"id":      "task-publish",
			"pending": false,
			"label":   "Actions::Katello::ContentView::Publish",
			"result":  "success",
			// finishedTask.Output is what SendAndParse actually reads
			// content_view_id from - the original 202 body's own content
			// is irrelevant once the task is polled.
			"output": map[string]interface{}{"content_view_id": float64(publishedID)},
		})
	})

	created, err := client.CreateKatelloContentView(ctx, &ContentView{
		ForemanObject: ForemanObject{Name: "cv-async"},
	})
	if err != nil {
		t.Fatalf("CreateKatelloContentView: %v", err)
	}
	if created.Id == 0 || created.Name != "cv-async" {
		t.Fatalf("CreateKatelloContentView: unexpected result %+v", created)
	}

	mux.HandleFunc("PUT "+collection+"/{id}/remove", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusAccepted, map[string]interface{}{
			"id":      "task-remove",
			"pending": true,
			"label":   "Actions::Katello::ContentView::Remove",
		})
	})
	mux.HandleFunc("GET /foreman_tasks/api/tasks/task-remove", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"id":      "task-remove",
			"pending": false,
			"label":   "Actions::Katello::ContentView::Remove",
			"result":  "success",
		})
	})

	if err := client.DeleteKatelloContentView(ctx, created.Id); err != nil {
		t.Fatalf("DeleteKatelloContentView: %v", err)
	}

	// A failed removal (result != "success") must surface as an error.
	d.storeFor(collection).seed(created.Id, map[string]interface{}{"name": "cv-async"})
	mux.HandleFunc("GET /foreman_tasks/api/tasks/task-remove-failed", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"id":        "task-remove-failed",
			"pending":   false,
			"label":     "Actions::Katello::ContentView::Remove",
			"result":    "error",
			"humanized": map[string]interface{}{"errors": []string{"boom"}},
		})
	})
	mux.HandleFunc("PUT "+collection+"/{id}/remove-failing", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusAccepted, map[string]interface{}{
			"id":      "task-remove-failed",
			"pending": true,
			"label":   "Actions::Katello::ContentView::Remove",
		})
	})
	req, err := client.NewRequestWithContext(ctx, http.MethodPut, collection+"/"+strconv.Itoa(created.Id)+"/remove-failing", nil)
	if err != nil {
		t.Fatalf("NewRequestWithContext: %v", err)
	}
	if err := client.SendAndParse(req, nil); err == nil {
		t.Fatalf("expected an error for a failed Katello remove task, got nil")
	}
}
