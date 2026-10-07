// A generic, in-memory dummy Foreman/Katello server used by every resource's
// tests. Rather than hand-coding the exact URL each resource's endpoint
// constants resolve to (they're inconsistent about leading slashes and some
// nest under other resources), dummyForeman reacts generically to whatever
// path and method actually arrive:
//
//   - POST to a path creates an object in that path's collection store,
//     assigns it an id, and returns it.
//   - GET to a path ending in a numeric id reads that object back.
//   - GET to a path NOT ending in a numeric id lists the collection as
//     Foreman's search endpoints do: {"results": [...], "total": N, ...}.
//   - PUT to a path ending in a numeric id merges the request body into the
//     stored object and returns it.
//   - DELETE removes the object.
//
// Each distinct collection path (e.g. "/api/architectures", or a nested one
// like "/api/katello/api/content_views/5/filters") gets its own independent
// store, keyed by that path string - so unrelated resources never collide,
// and nested per-parent collections naturally stay separate per parent id.
//
// A few endpoints are true RPC-style actions rather than CRUD on a
// sub-collection (host power commands, content view publish/remove) and are
// given explicit handlers registered ahead of the generic catch-all.
package foreman

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"testing"
)

// resourceStore is a tiny in-memory id -> object table.
type resourceStore struct {
	mu     sync.Mutex
	nextID int
	byID   map[int]map[string]interface{}
}

func newResourceStore() *resourceStore {
	return &resourceStore{nextID: 1, byID: map[int]map[string]interface{}{}}
}

func (s *resourceStore) create(obj map[string]interface{}) map[string]interface{} {
	s.mu.Lock()
	defer s.mu.Unlock()
	id := s.nextID
	s.nextID++
	obj["id"] = id
	s.byID[id] = obj
	return obj
}

// seed stores obj under a caller-chosen id, for read-only resources that
// have no Create to populate the store through.
func (s *resourceStore) seed(id int, obj map[string]interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()
	obj["id"] = id
	s.byID[id] = obj
	if id >= s.nextID {
		s.nextID = id + 1
	}
}

func (s *resourceStore) get(id int) (map[string]interface{}, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	obj, ok := s.byID[id]
	return obj, ok
}

func (s *resourceStore) update(id int, patch map[string]interface{}) map[string]interface{} {
	s.mu.Lock()
	defer s.mu.Unlock()
	obj, ok := s.byID[id]
	if !ok {
		obj = map[string]interface{}{}
	}
	for k, v := range patch {
		obj[k] = v
	}
	obj["id"] = id
	s.byID[id] = obj
	return obj
}

func (s *resourceStore) delete(id int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.byID, id)
}

func (s *resourceStore) list() []map[string]interface{} {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]map[string]interface{}, 0, len(s.byID))
	for _, obj := range s.byID {
		out = append(out, obj)
	}
	return out
}

// dummyForeman is the generic server: a collection-path -> resourceStore
// table, plus a mux for the handful of non-CRUD action endpoints.
type dummyForeman struct {
	mu     sync.Mutex
	stores map[string]*resourceStore
}

func newDummyForeman() *dummyForeman {
	return &dummyForeman{stores: map[string]*resourceStore{}}
}

func (d *dummyForeman) storeFor(path string) *resourceStore {
	d.mu.Lock()
	defer d.mu.Unlock()
	s, ok := d.stores[path]
	if !ok {
		s = newResourceStore()
		d.stores[path] = s
	}
	return s
}

// splitCollectionAndID splits a request path into its collection path and,
// if the last segment is numeric, the id that names one item in it.
func splitCollectionAndID(p string) (collection string, id int, hasID bool) {
	p = strings.TrimSuffix(p, "/")
	idx := strings.LastIndex(p, "/")
	if idx < 0 {
		return p, 0, false
	}
	last := p[idx+1:]
	n, err := strconv.Atoi(last)
	if err != nil {
		return p, 0, false
	}
	return p[:idx], n, true
}

// decodeUnwrapped reads the request body as JSON and, if it's a single-key
// object whose value is itself an object (the shape WrapJSON/
// WrapJSONWithTaxonomy produce, e.g. {"architecture": {...}}), returns that
// inner object - otherwise returns the decoded body as-is (some resources,
// e.g. katello_content_credential, send the flat object with no wrapper).
func decodeUnwrapped(r *http.Request) map[string]interface{} {
	var body map[string]interface{}
	_ = json.NewDecoder(r.Body).Decode(&body)
	if len(body) == 1 {
		for _, v := range body {
			if inner, ok := v.(map[string]interface{}); ok {
				return inner
			}
		}
	}
	return body
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func (d *dummyForeman) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	collection, id, hasID := splitCollectionAndID(r.URL.Path)
	store := d.storeFor(collection)

	switch r.Method {
	case http.MethodPost:
		writeJSON(w, http.StatusCreated, store.create(decodeUnwrapped(r)))
	case http.MethodGet:
		if hasID {
			obj, ok := store.get(id)
			if !ok {
				http.Error(w, "not found", http.StatusNotFound)
				return
			}
			writeJSON(w, http.StatusOK, obj)
			return
		}
		results := store.list()
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"results":  results,
			"total":    len(results),
			"subtotal": len(results),
		})
	case http.MethodPut, http.MethodPatch:
		writeJSON(w, http.StatusOK, store.update(id, decodeUnwrapped(r)))
	case http.MethodDelete:
		store.delete(id)
		w.WriteHeader(http.StatusNoContent)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

// newDummyServer starts the generic dummy server and returns a *Client
// pointed at it, an *http.ServeMux for registering any resource-specific
// overrides (special action endpoints) ahead of the generic catch-all, and
// the dummyForeman itself for direct store seeding (read-only resources).
// The server is closed automatically via t.Cleanup.
func newDummyServer(t *testing.T) (*http.ServeMux, *dummyForeman, *Client) {
	t.Helper()

	d := newDummyForeman()
	mux := http.NewServeMux()
	mux.Handle("/", d)

	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	serverURL, err := url.Parse(server.URL)
	if err != nil {
		t.Fatalf("parsing test server URL: %v", err)
	}

	client := NewClient(
		Server{URL: *serverURL},
		ClientCredentials{Username: "admin", Password: "changeme"},
		ClientConfig{LocationID: -1, OrganizationID: -1},
	)

	return mux, d, client
}
