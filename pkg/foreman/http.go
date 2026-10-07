// HTTP request building, sending, and response handling for Client. See
// client.go for Client itself and how to build one.
package foreman

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

const (
	// Every Foreman API call has the following prefix to the path component
	// of the URL.  The client hepler functions utilize this to automatically
	// create endpoint URLs.
	FOREMAN_API_URL_PREFIX = "/api"

	// FOREMAN_KATELLO_API_URL_PREFIX is the Foreman Katello API endpoint
	FOREMAN_KATELLO_API_URL_PREFIX = "/katello/api"

	// FOREMAN_TASKS_API_URL_PREFIX is the prefix for async tasks
	FOREMAN_TASKS_API_URL_PREFIX = "/foreman_tasks/api"

	// API Prefix for Puppet plugin
	FOREMAN_PUPPET_API_URL_PREFIX = "/foreman_puppet/api"

	// The Foreman API allows you to request a specific API version in the
	// Accept header of the HTTP request.  The two supported versions (at
	// the time of writing) are 1 and 2, which version 1 planning on being
	// deprecated after version 1.17.
	FOREMAN_API_VERSION = "2"
)

// HTTPDoer is the minimal interface Client needs to send a request.
// *http.Client satisfies it as-is, so it is the default; set
// ClientConfig.HTTPClient to swap in any other implementation (custom
// transport, retries, instrumentation, a test double, ...).
type HTTPDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

// HTTPError represents a Foreman API response with a non-2xx status code.
type HTTPError struct {
	Endpoint   string
	StatusCode int
	RespBody   string
}

func (e HTTPError) Error() string {
	return fmt.Sprintf(
		"HTTP Error:{\n"+
			"  endpoint:   [%s]\n"+
			"  statusCode: [%d]\n"+
			"  respBody:   [%s]\n"+
			"}",
		e.Endpoint,
		e.StatusCode,
		e.RespBody,
	)
}

// IsNotFound reports whether err is (or wraps) an HTTPError for a 404 (Not
// Found) response. Useful for callers that want to treat a missing remote
// object as "already gone" rather than as a hard error.
func IsNotFound(err error) bool {
	var httpErr HTTPError
	return errors.As(err, &httpErr) && httpErr.StatusCode == 404
}

// HTTPClient returns the HTTPDoer currently used to send requests - either
// the one supplied via ClientConfig.HTTPClient, or the default client
// NewClient built from TLSInsecureEnabled/NegotiateAuthEnabled.
func (c *Client) HTTPClient() HTTPDoer {
	return c.httpClient
}

// SetHTTPClient replaces the HTTPDoer used to send requests. Useful to swap
// transports after construction, e.g. in tests.
func (c *Client) SetHTTPClient(d HTTPDoer) {
	c.httpClient = d
}

// NewRequestWithContext constructs an HTTP request using the client configuration.
// Common request functionality is abstracted and wrapped into this function
// (ie: headers, cookies, MIME-info, etc).  The client should never interact
// with the underlying HTTP client or request object directly.
//
// If the user provides an invalid HTTP method, the function returns 'nil'
// for the request and will return an Error.
//
// The following headers are added and set automatically:
//
//	User-Agent
//	ACCEPT
//	Content-Type
//	Authorization
//
// method
//
//	The HTTP Verb to use.  This should correspond to a 'Method*' constant
//	from 'net/http'.
//
// endpoint
//
//	The server's endpoint to send the request.  The endpoint value is
//	appended to the client's server URL to construct the full URL for the
//	request.  NewRequestWithContext() will automatically prepend the Foreman API URL
//	prefix to the endpoint.
//
// body
//
//	Functions exactly like net/http/NewRequestWithContext()
func (c *Client) NewRequestWithContext(ctx context.Context, method string, endpoint string, body io.Reader) (*http.Request, error) {
	TraceFunctionCall()
	c.log.Debugf(
		"method: [%s], endpoint: [%s]",
		method,
		endpoint,
	)

	if !isValidRequestMethod(method) {
		c.log.Errorf("invalid HTTP request method: [%s]", method)
		return nil, fmt.Errorf("invalid HTTP request method: [%s]", method)
	}

	var version_append string

	// Build the URL for the request
	reqURL := c.Server.URL

	// Check for katello endpoint
	if strings.HasPrefix(endpoint, "katello") {
		reqURL.Path = FOREMAN_KATELLO_API_URL_PREFIX + strings.TrimPrefix(endpoint, "katello")
	} else if strings.HasPrefix(endpoint, "/katello/api") {
		reqURL.Path = endpoint
	} else if strings.HasPrefix(endpoint, "puppet") {
		reqURL.Path = FOREMAN_PUPPET_API_URL_PREFIX + strings.TrimPrefix(endpoint, "puppet")
	} else if strings.HasPrefix(endpoint, "foreman_tasks") || strings.HasPrefix(endpoint, "/foreman_tasks") {
		reqURL.Path = endpoint
	} else {
		if strings.HasPrefix(endpoint, "/") {
			reqURL.Path = FOREMAN_API_URL_PREFIX + endpoint
		} else {
			reqURL.Path = FOREMAN_API_URL_PREFIX + "/" + endpoint
		}
		version_append = "version=" + FOREMAN_API_VERSION
	}

	c.log.Debugf(
		"reqURL: [%s]",
		reqURL.String(),
	)

	// Create the request object, bubble up errors if any were encountered
	req, reqErr := http.NewRequestWithContext(
		ctx,
		strings.ToUpper(method),
		reqURL.String(),
		body,
	)
	if reqErr != nil {
		c.log.Errorf(
			"failed to construct a new HTTP request\n"+
				"  Error: %s",
			reqErr.Error(),
		)
		return req, reqErr
	}
	// Add common meta-data and header information for the request
	req.Header.Add("User-Agent", "foremango")
	req.Header.Add("Accept", "application/json,"+version_append)
	req.Header.Add("Content-Type", "application/json")
	req.SetBasicAuth(c.credentials.Username, c.credentials.Password)
	return req, nil
}

// isValidRequestMethod is a helper function used to determine if an HTTP
// request method is valid.
//
// NOTE(ALL): Go's HTTP client does not support sending a request with
//
//	the 'CONNECT' method and therefore is not counted as a valid request
//	method. See http.Transport, http.Client for more information.
func isValidRequestMethod(method string) bool {
	// Slice of valid HTTP methods for sending and creating requests
	validHTTPMethods := []string{
		http.MethodGet,
		http.MethodHead,
		http.MethodPost,
		http.MethodPut,
		http.MethodPatch,
		http.MethodDelete,
		http.MethodOptions,
		http.MethodTrace,
	}
	// list isn't large - use linear search to validate the method.  Use
	// strings.EqualFold to perform case-insensitive comparisons
	for _, value := range validHTTPMethods {
		if strings.EqualFold(value, method) {
			return true
		}
	}
	return false
}

// Send sends an HTTP request generated by Client.NewRequestWithContext() and returns
// the StatusCode, response. Serves as a facade to the Client's underlying
// HTTP client.
//
// If an error is encountered when reading the server's response, the returned
// StatusCode will be -1.  If an error is encountered during any step of the
// the send and response parsing, an empty slice will be returned as the
// request body.
//
// request
//
//	An HTTP request generated by Client.NewRequestWithContext()
func (c *Client) Send(request *http.Request) (int, []byte, error) {
	TraceFunctionCall()

	emptySlice := []byte{}

	if request == nil {
		c.log.Errorf("client trying to send a nil request")
		return -1, emptySlice, errors.New("client trying to send a nil request")
	}

	// Send the request to the server
	resp, respErr := c.httpClient.Do(request)
	if respErr != nil {
		c.log.Errorf(
			"error encountered when sending HTTP request to server\n"+
				"  Error: %s",
			respErr.Error(),
		)
		return -1, emptySlice, respErr
	}
	// NOTE(ALL): Golang stdlib dictates that it is the caller's resposibility
	//   to close the response body.  See net/http Response type for more
	//   information.
	defer func() { _ = resp.Body.Close() }()

	// Read the server's response
	respBody, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		c.log.Errorf(
			"error encountered when reading HTTP response from server\n"+
				"  Error: %s",
			readErr.Error(),
		)
		return resp.StatusCode, emptySlice, readErr
	}

	return resp.StatusCode, respBody, nil
}

// SendAndParse sends an HTTP request generated by Client.NewRequestWithContext(),
// checks the response's status code, and unmarshals its body into obj (if
// obj is not nil). A non-2xx status code is returned as an HTTPError.
//
// It also handles Foreman/Katello's asynchronous task protocol: a status
// code of 202 means the server accepted the request but is processing it as
// a background Katello task, which this function polls until it finishes.
//
// Two task outcomes get extra handling because the REST response they're
// attached to needs it:
//   - "Actions::Katello::ContentView::Publish" has no useful response body
//     of its own, so the freshly published ContentView is read back and
//     returned as if the original request had produced it directly.
//   - "Actions::Katello::ContentView::Remove" only needs its success
//     checked; a failed removal is turned into an error.
//
// Any other finished task is otherwise accepted and its last known response
// body (from the original 202) is parsed as usual.
func (c *Client) SendAndParse(req *http.Request, obj interface{}) error {
	TraceFunctionCall()

	statusCode, respBody, sendErr := c.Send(req)
	if sendErr != nil {
		return sendErr
	}

	c.log.Debugf(
		"server response:{\n"+
			"  endpoint:   [%s]\n"+
			"  method:     [%s]\n"+
			"  statusCode: [%d]\n"+
			"  respBody:   [%s]\n"+
			"}",
		req.URL,
		req.Method,
		statusCode,
		respBody,
	)

	// Handle Katello async responses.
	// Be aware, that waitForKatelloAsyncTask also lands here. We just need to trust that the
	// foreman_tasks API endpoint does not omit 202 as well.
	// Officially, 202 is the code for "accepted, but not processed yet".
	if statusCode == 202 {
		var asyncTask ForemanTask
		if err := json.Unmarshal(respBody, &asyncTask); err != nil {
			return err
		}
		c.log.Debugf("foremanAsyncTask asyncTask: %+v", asyncTask)

		if asyncTask.Pending {
			c.log.Debugf("KatelloResponse is pending")
			finishedTask, err := c.waitForKatelloAsyncTask(asyncTask.Id)
			if err != nil {
				return err
			}

			switch finishedTask.Label {

			case "Actions::Katello::ContentView::Publish":
				// Used by endpoint POST /katello/api/content_views/:id/publish
				output := finishedTask.Output.(map[string]interface{})
				cvToRead := ContentView{
					ForemanObject: ForemanObject{Id: int(output["content_view_id"].(float64))},
				}

				ctx := context.TODO()
				updatedCv, err := c.ReadKatelloContentView(ctx, &cvToRead)
				if err != nil {
					return err
				}

				respBody, err = json.Marshal(updatedCv)
				if err != nil {
					return err
				}

			case "Actions::Katello::ContentView::Remove":
				// Used by endpoint PUT /katello/api/content_views/:id/remove
				success := finishedTask.Result == "success"
				if !success {
					return fmt.Errorf("error removing content_view: %v", finishedTask.Humanized.Errors)
				}
			}
		}
	}

	if statusCode < 200 || statusCode > 299 {
		return HTTPError{
			Endpoint:   req.URL.String(),
			StatusCode: statusCode,
			RespBody:   string(respBody),
		}
	}

	if obj != nil {
		return json.Unmarshal(respBody, &obj)
	}
	return nil
}

// WrapParameters wraps the given parameters as an object of its own name.
func (c *Client) WrapParameters(name interface{}, item interface{}) (map[string]interface{}, error) {
	var wrapped map[string]interface{}

	if name != nil {
		wrapped = map[string]interface{}{
			fmt.Sprintf("%v", name): item,
		}
	} else {
		data, err := json.Marshal(item)
		if err != nil {
			return nil, err
		}

		if err := json.Unmarshal(data, &wrapped); err != nil {
			return nil, err
		}
	}

	return wrapped, nil
}

// WrapJSON wraps the given parameters as an object of its own name and marshals it to JSON
func (c *Client) WrapJSON(name interface{}, item interface{}) ([]byte, error) {
	wrapped, _ := c.WrapParameters(name, item)
	return json.Marshal(wrapped)
}

// WrapJSONWithTaxonomy wraps the given parameters as an object of its own name,
// includes additional information for the api call and marshals it to JSON
func (c *Client) WrapJSONWithTaxonomy(name interface{}, item interface{}) ([]byte, error) {
	TraceFunctionCall()

	wrapped, err := c.WrapParameters(name, item)
	if err != nil {
		return nil, err
	}

	// Workaround for Foreman versions < 1.21 in case no default location/organization was defined for resources
	if c.Config.LocationID >= 0 && c.Config.OrganizationID >= 0 {
		wrapped["location_id"] = c.Config.LocationID
		wrapped["organization_id"] = c.Config.OrganizationID
		c.log.Debugf("http.go#WrapJSONWithTaxonomy: item %+v", wrapped)
	}

	return json.Marshal(wrapped)
}
