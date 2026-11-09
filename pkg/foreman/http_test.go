package foreman

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"resty.dev/v3"
)

func TestDo_RejectsInvalidMethod(t *testing.T) {
	c := NewClient(Server{}, ClientCredentials{}, ClientConfig{})
	for _, m := range []string{"", "FETCH", "CONNECT", "connect", " GET", "GET\n", "G E T"} {
		if err := c.Do(context.Background(), m, "hosts", nil, nil, nil); err == nil {
			t.Errorf("Do(%q) returned no error", m)
		}
	}
}

func TestDo_SendsMethodUpperCase(t *testing.T) {
	var got string
	c := testClient(t, ClientCredentials{}, ClientConfig{}, func(w http.ResponseWriter, r *http.Request) {
		got = r.Method
	})
	for _, m := range []string{"get", "Post", "pUT", "patch", "delete", "HEAD", "options", "trace"} {
		if err := c.Do(context.Background(), m, "hosts", nil, nil, nil); err != nil {
			t.Fatalf("Do(%q): %v", m, err)
		}
		if got != strings.ToUpper(m) {
			t.Errorf("server got method %q for %q", got, m)
		}
	}
}

func TestDo_Headers(t *testing.T) {
	var got http.Header
	c := testClient(t, ClientCredentials{Username: "admin", Password: "secret"}, ClientConfig{},
		func(w http.ResponseWriter, r *http.Request) { got = r.Header.Clone() })

	if err := c.Do(context.Background(), http.MethodGet, "hosts", nil, nil, nil); err != nil {
		t.Fatalf("Do: %v", err)
	}

	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	req.SetBasicAuth("admin", "secret")
	want := map[string]string{
		"Accept":        "application/json,version=" + apiVersion,
		"Content-Type":  "application/json",
		"User-Agent":    "foremango",
		"Authorization": req.Header.Get("Authorization"),
	}
	for k, v := range want {
		if got.Get(k) != v {
			t.Errorf("header %s = %q, want %q", k, got.Get(k), v)
		}
	}
}

func TestDo_NoBasicAuthWithoutCredentials(t *testing.T) {
	var auth string
	c := testClient(t, ClientCredentials{}, ClientConfig{}, func(w http.ResponseWriter, r *http.Request) {
		auth = r.Header.Get("Authorization")
	})
	if err := c.Do(context.Background(), http.MethodGet, "hosts", nil, nil, nil); err != nil {
		t.Fatalf("Do: %v", err)
	}
	if auth != "" {
		t.Errorf("Authorization = %q, want none", auth)
	}
}

func TestDo_UsesSuppliedRestyClient(t *testing.T) {
	var got string
	rc := resty.New().SetHeader("X-Trace", "abc")
	c := testClient(t, ClientCredentials{}, ClientConfig{RestyClient: rc}, func(w http.ResponseWriter, r *http.Request) {
		got = r.Header.Get("X-Trace")
	})
	if err := c.Do(context.Background(), http.MethodGet, "hosts", nil, nil, nil); err != nil {
		t.Fatalf("Do: %v", err)
	}
	if got != "abc" {
		t.Errorf("X-Trace = %q: supplied resty client not used", got)
	}
}

func TestEndpointURL(t *testing.T) {
	core := "application/json,version=" + apiVersion
	plugin := "application/json"
	cases := []struct {
		base, endpoint, wantURL, wantAccept string
	}{
		{"https://f.example", "hosts", "https://f.example/api/hosts", core},
		{"https://f.example", "hosts/3/parameters", "https://f.example/api/hosts/3/parameters", core},
		{"https://f.example", "", "https://f.example/api/", core},
		{"https://f.example", "katello/products/2", "https://f.example/katello/api/products/2", plugin},
		{"https://f.example", "puppet/puppetclasses", "https://f.example/foreman_puppet/api/puppetclasses", plugin},
		{"https://f.example", "foreman_tasks/tasks/9", "https://f.example/foreman_tasks/api/tasks/9", plugin},
		{"https://f.example", "/api/status", "https://f.example/api/status", core},
		{"https://f.example", "/katello/api/ping", "https://f.example/katello/api/ping", plugin},
		{"https://f.example/foreman/", "hosts", "https://f.example/foreman/api/hosts", core},
	}
	for _, tc := range cases {
		u, _ := url.Parse(tc.base)
		c := NewClient(Server{URL: *u}, ClientCredentials{}, ClientConfig{})
		gotURL, gotAccept := c.endpointURL(tc.endpoint)
		if gotURL.String() != tc.wantURL || gotAccept != tc.wantAccept {
			t.Errorf("endpointURL(%q) on %s = %s, %q; want %s, %q",
				tc.endpoint, tc.base, gotURL.String(), gotAccept, tc.wantURL, tc.wantAccept)
		}
		if c.Server.URL.String() != u.String() {
			t.Errorf("endpointURL changed the server URL to %s", c.Server.URL.String())
		}
	}
}

func TestDo_QueryBodyAndDecode(t *testing.T) {
	var search, body string
	c := testClient(t, ClientCredentials{}, ClientConfig{}, func(w http.ResponseWriter, r *http.Request) {
		search = r.URL.Query().Get("search")
		b, _ := io.ReadAll(r.Body)
		body = string(b)
		_, _ = w.Write([]byte(`{"id":7,"name":"web01"}`))
	})

	var out ForemanObject
	q := url.Values{"search": {`name = "web01"`}}
	if err := c.Do(context.Background(), http.MethodPost, "hosts", q, []byte(`{"host":{}}`), &out); err != nil {
		t.Fatalf("Do: %v", err)
	}
	if search != `name = "web01"` {
		t.Errorf("search = %q", search)
	}
	if body != `{"host":{}}` {
		t.Errorf("body = %q", body)
	}
	if out.Id != 7 || out.Name != "web01" {
		t.Errorf("decoded %+v", out)
	}
}

func TestDo_EmptyResponseBody(t *testing.T) {
	c := testClient(t, ClientCredentials{}, ClientConfig{}, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	var out ForemanObject
	if err := c.Do(context.Background(), http.MethodDelete, "hosts/1", nil, nil, &out); err != nil {
		t.Fatalf("Do: %v", err)
	}
}

func TestDo_InvalidJSON(t *testing.T) {
	c := testClient(t, ClientCredentials{}, ClientConfig{}, func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`not json`))
	})
	var out ForemanObject
	if err := c.Do(context.Background(), http.MethodGet, "hosts/1", nil, nil, &out); err == nil {
		t.Fatalf("Do decoded invalid JSON without error")
	}
}

func TestDo_HTTPError(t *testing.T) {
	c := testClient(t, ClientCredentials{}, ClientConfig{}, func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/missing") {
			http.Error(w, `{"error":"not found"}`, http.StatusNotFound)
			return
		}
		http.Error(w, "boom", http.StatusInternalServerError)
	})

	err := c.Do(context.Background(), http.MethodGet, "hosts/broken", nil, nil, nil)
	var httpErr HTTPError
	if !errors.As(err, &httpErr) {
		t.Fatalf("err = %v (%T), want HTTPError", err, err)
	}
	if httpErr.StatusCode != http.StatusInternalServerError || httpErr.Method != http.MethodGet ||
		!strings.HasSuffix(httpErr.URL, "/api/hosts/broken") || strings.TrimSpace(httpErr.Body) != "boom" {
		t.Errorf("HTTPError = %+v", httpErr)
	}
	if IsNotFound(err) {
		t.Errorf("IsNotFound(500) = true")
	}

	err = c.Do(context.Background(), http.MethodGet, "hosts/missing", nil, nil, nil)
	if !IsNotFound(err) {
		t.Errorf("IsNotFound(%v) = false", err)
	}
	if !IsNotFound(fmt.Errorf("reading host: %w", err)) {
		t.Errorf("IsNotFound does not see through wrapping")
	}
	if IsNotFound(errors.New("other")) {
		t.Errorf("IsNotFound(plain error) = true")
	}
}

func TestDo_TransportError(t *testing.T) {
	c := testClient(t, ClientCredentials{}, ClientConfig{}, func(w http.ResponseWriter, r *http.Request) {})
	c.Server.URL.Host = "127.0.0.1:1" // nothing listens there

	err := c.Do(context.Background(), http.MethodGet, "hosts", nil, nil, nil)
	if err == nil {
		t.Fatalf("Do returned no error")
	}
	var httpErr HTTPError
	if errors.As(err, &httpErr) {
		t.Errorf("transport failure returned as HTTPError: %v", err)
	}
}

func TestPayload(t *testing.T) {
	obj := map[string]string{"name": "web01"}
	cases := []struct {
		cfg  ClientConfig
		want string
	}{
		{ClientConfig{}, `{"host":{"name":"web01"}}`},
		{ClientConfig{LocationID: -1, OrganizationID: -1}, `{"host":{"name":"web01"}}`},
		{ClientConfig{LocationID: 2}, `{"host":{"name":"web01"},"location_id":2}`},
		{ClientConfig{LocationID: 2, OrganizationID: 5}, `{"host":{"name":"web01"},"location_id":2,"organization_id":5}`},
	}
	for _, tc := range cases {
		c := NewClient(Server{}, ClientCredentials{}, tc.cfg)
		got, err := c.Payload("host", obj)
		if err != nil {
			t.Fatalf("Payload: %v", err)
		}
		if string(got) != tc.want {
			t.Errorf("Payload with %+v = %s, want %s", tc.cfg, got, tc.want)
		}
	}
}

// Exercises a generated resource end to end: create, read, query, delete.
func TestGeneratedResource_RoundTrip(t *testing.T) {
	var lastPath, lastSearch string
	var lastBody map[string]any
	c := testClient(t, ClientCredentials{}, ClientConfig{LocationID: 1}, func(w http.ResponseWriter, r *http.Request) {
		lastPath = r.Method + " " + r.URL.Path
		lastSearch = r.URL.Query().Get("search")
		lastBody = nil
		_ = json.NewDecoder(r.Body).Decode(&lastBody)
		switch r.Method {
		case http.MethodGet:
			if r.URL.Path == "/api/architectures" {
				_, _ = w.Write([]byte(`{"total":2,"subtotal":1,"page":1,"per_page":20,"search":"name = x86_64","results":[{"id":1,"name":"x86_64"}]}`))
				return
			}
			_, _ = w.Write([]byte(`{"id":1,"name":"x86_64"}`))
		case http.MethodPost:
			_, _ = w.Write([]byte(`{"id":1,"name":"x86_64"}`))
		}
	})
	ctx := context.Background()

	created, err := c.CreateArchitecture(ctx, &ForemanArchitecture{ForemanObject: ForemanObject{Name: "x86_64"}})
	if err != nil || created.Id != 1 {
		t.Fatalf("CreateArchitecture = %+v, %v", created, err)
	}
	if lastPath != "POST /api/architectures" || lastBody["location_id"] != float64(1) {
		t.Errorf("create sent %s %v", lastPath, lastBody)
	}

	read, err := c.ReadArchitecture(ctx, 1)
	if err != nil || read.Name != "x86_64" || lastPath != "GET /api/architectures/1" {
		t.Errorf("ReadArchitecture = %+v, %v via %s", read, err, lastPath)
	}

	list, err := c.QueryArchitecture(ctx, "name = x86_64")
	if err != nil || lastSearch != "name = x86_64" || list.Subtotal != 1 || len(list.Results) != 1 || list.Results[0].Name != "x86_64" {
		t.Errorf("QueryArchitecture = %+v, %v (search %q)", list, err, lastSearch)
	}

	if err := c.DeleteArchitecture(ctx, 1); err != nil || lastPath != "DELETE /api/architectures/1" {
		t.Errorf("DeleteArchitecture: %v via %s", err, lastPath)
	}
}
