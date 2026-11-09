package foreman

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"resty.dev/v3"
)

// apiVersion is the Foreman core API version requested in the Accept header.
const apiVersion = "2"

// apiRoots maps an endpoint prefix to the API root it is served under. An
// endpoint matching none of them belongs to the Foreman core API, /api.
var apiRoots = []struct{ prefix, root string }{
	{"katello/", "/katello/api/"},
	{"puppet/", "/foreman_puppet/api/"},
	{"foreman_tasks/", "/foreman_tasks/api/"},
}

// HTTPError is the error Client.Do returns when Foreman answers with a
// status outside 2xx.
type HTTPError struct {
	Method     string
	URL        string
	StatusCode int
	// Body is the raw response body, usually Foreman's JSON error message.
	Body string
}

func (e HTTPError) Error() string {
	return fmt.Sprintf("foreman: %s %s: %d %s: %s",
		e.Method, e.URL, e.StatusCode, http.StatusText(e.StatusCode), e.Body)
}

// IsNotFound reports whether err is, or wraps, an HTTPError with status 404.
func IsNotFound(err error) bool {
	var httpErr HTTPError
	return errors.As(err, &httpErr) && httpErr.StatusCode == http.StatusNotFound
}

// RestyClient returns the resty client that sends this client's requests.
func (c *Client) RestyClient() *resty.Client {
	return c.resty
}

// SetRestyClient replaces the resty client that sends this client's requests.
func (c *Client) SetRestyClient(rc *resty.Client) {
	c.resty = rc
}

// Payload builds the JSON body Foreman expects to create or update an
// object: obj nested under its resource name, as in {"host": {...}}. The
// client's location and organization (Config.LocationID and
// Config.OrganizationID) are added beside it when they are set (> 0).
func (c *Client) Payload(resource string, obj any) ([]byte, error) {
	body := map[string]any{resource: obj}
	if c.Config.LocationID > 0 {
		body["location_id"] = c.Config.LocationID
	}
	if c.Config.OrganizationID > 0 {
		body["organization_id"] = c.Config.OrganizationID
	}
	return json.Marshal(body)
}

// Do sends one request to the Foreman API and decodes the JSON answer.
//
// endpoint is a path relative to an API root: "hosts/1" is sent to
// /api/hosts/1, and "katello/...", "puppet/..." and "foreman_tasks/..."
// to those plugins' API roots. An endpoint starting with "/" is used as an
// absolute path on the server.
//
// query is appended to the URL (nil for none). body is sent as the JSON
// request body (nil for none; see Payload). The response body is decoded
// into out unless out is nil or the body is empty. A status outside 2xx is
// returned as an HTTPError.
func (c *Client) Do(ctx context.Context, method, endpoint string, query url.Values, body []byte, out any) error {
	TraceFunctionCall()

	if !isValidRequestMethod(method) {
		c.log.Errorf("invalid HTTP request method: [%s]", method)
		return fmt.Errorf("foreman: invalid HTTP request method %q", method)
	}
	method = strings.ToUpper(method)

	reqURL, accept := c.endpointURL(endpoint)
	if len(query) > 0 {
		reqURL.RawQuery = query.Encode()
	}
	target := reqURL.String()

	req := c.resty.R().
		SetContext(ctx).
		SetHeader("Accept", accept).
		SetHeader("Content-Type", "application/json")
	if c.credentials.Username != "" || c.credentials.Password != "" {
		req.SetBasicAuth(c.credentials.Username, c.credentials.Password)
	}
	if body != nil {
		req.SetBody(body)
	}

	c.log.Debugf("request: %s %s", method, target)
	resp, err := req.Execute(method, target)
	if err != nil {
		c.log.Errorf("request failed: %s %s: %s", method, target, err)
		return fmt.Errorf("foreman: %s %s: %w", method, target, err)
	}

	respBody := resp.Bytes()
	c.log.Debugf("response: %s %s: %d: %s", method, target, resp.StatusCode(), respBody)

	if resp.StatusCode() < 200 || resp.StatusCode() > 299 {
		return HTTPError{
			Method:     method,
			URL:        target,
			StatusCode: resp.StatusCode(),
			Body:       string(respBody),
		}
	}

	if out == nil || len(respBody) == 0 {
		return nil
	}
	if err := json.Unmarshal(respBody, out); err != nil {
		return fmt.Errorf("foreman: %s %s: decoding response: %w", method, target, err)
	}
	return nil
}

// endpointURL resolves endpoint (see Do) against the server URL. It also
// returns the Accept header to send: the core API gets the version
// parameter, plugin APIs do not.
func (c *Client) endpointURL(endpoint string) (url.URL, string) {
	path := endpoint
	if !strings.HasPrefix(endpoint, "/") {
		path = "/api/" + endpoint
		for _, r := range apiRoots {
			if strings.HasPrefix(endpoint, r.prefix) {
				path = r.root + strings.TrimPrefix(endpoint, r.prefix)
				break
			}
		}
	}

	accept := "application/json"
	if strings.HasPrefix(path, "/api/") {
		accept += ",version=" + apiVersion
	}

	u := c.Server.URL
	u.Path = strings.TrimSuffix(u.Path, "/") + path
	u.RawPath = ""
	return u, accept
}

// isValidRequestMethod reports whether Do can send method, in any letter
// case. CONNECT is excluded: Go's HTTP client cannot send it.
func isValidRequestMethod(method string) bool {
	switch strings.ToUpper(method) {
	case http.MethodGet, http.MethodHead, http.MethodPost, http.MethodPut,
		http.MethodPatch, http.MethodDelete, http.MethodOptions, http.MethodTrace:
		return true
	}
	return false
}
