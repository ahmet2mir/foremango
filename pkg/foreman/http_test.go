// HTTP request building, sending, and response-parsing tests for Client.
// See client_test.go for NewClient/construction tests.
package foreman

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"testing"
)

// ----------------------------------------------------------------------------
// Client.NewRequestWithContext
// ----------------------------------------------------------------------------

// Ensures Client.NewRequestWithContext() returns an error for a bad HTTP method.
func TestNewRequest_BadHTTPMethodError(t *testing.T) {
	serv := Server{}
	cred := ClientCredentials{}
	conf := ClientConfig{}
	client := NewClient(serv, cred, conf)

	badHTTPMethods := []string{"",
		"FOO",
		"ZZ",
		"fo0",
		"connect",
		"CONNECT",
		"10",
		" GET",
		"\tGET\n",
		"get\n",
	}
	for _, value := range badHTTPMethods {
		_, badReqErr := client.NewRequestWithContext(context.TODO(), value, "/foo", nil)
		if badReqErr == nil {
			t.Fatalf(
				"Client.NewRequestWithContext did not return error when given invalid HTTP method [%s]. "+
					"Expected [error], got [nil].",
				value,
			)
		}
	}

}

// Ensures Client.NewRequestWithContext() does not raise an error when given a valid
// HTTP method.
func TestNewRequest_GoodHTTPMethodNoError(t *testing.T) {
	serv := Server{}
	cred := ClientCredentials{}
	conf := ClientConfig{}
	client := NewClient(serv, cred, conf)

	goodHTTPMethods := []string{
		"GET",
		"gEt",
		"get",
		"Get",
		http.MethodGet,
		http.MethodHead,
		http.MethodPost,
		http.MethodPut,
		http.MethodDelete,
		http.MethodOptions,
		http.MethodTrace,
		http.MethodPatch,
	}
	for _, value := range goodHTTPMethods {
		_, reqErr := client.NewRequestWithContext(context.TODO(), value, "/foo", nil)
		if reqErr != nil {
			t.Fatalf(
				"Client.NewRequestWithContext returned an error when given valid HTTP method [%s]. "+
					"Expected [nil], got [%s].",
				value,
				reqErr.Error(),
			)
		}
	}

}

// Ensures Client.NewRequestWithContext() sets the HTTP request's method to upper case.
func TestNewRequest_RequestMethodToUpper(t *testing.T) {
	serv := Server{}
	cred := ClientCredentials{}
	conf := ClientConfig{}
	client := NewClient(serv, cred, conf)

	testMethods := []string{
		"get",
		"Get",
		"gEt",
		"geT",
		"GEt",
		"gET",
		"GeT",
		"GET",
		http.MethodGet,
	}
	expectedMethod := "GET"

	for _, value := range testMethods {
		req, _ := client.NewRequestWithContext(context.TODO(), value, "/foo", nil)
		if req.Method != expectedMethod {
			t.Fatalf(
				"http.Request returned by Client.NewRequestWithContext() has incorrect Method. "+
					"Expected [%s], got [%s].\n",
				expectedMethod,
				req.Method,
			)
		}
	}

}

// Ensures Client.NewRequestWithContext() sets the correct meta-data on the HTTP
// request.
func TestNewRequest_Header(t *testing.T) {
	serv := Server{}
	cred := ClientCredentials{
		Username: "Admin",
		Password: "ChangeMe",
	}
	conf := ClientConfig{}
	client := NewClient(serv, cred, conf)

	// perform HTTP basic access authorization for the credentials
	// SEE: RFC 7617
	credentialsEncoded := "Basic " + base64.StdEncoding.EncodeToString(
		[]byte(cred.Username+":"+cred.Password),
	)

	req, _ := client.NewRequestWithContext(context.TODO(), http.MethodGet, "/foo", nil)

	expectedHeader := http.Header{}
	expectedHeader.Add("User-Agent", "foremango")
	expectedHeader.Add("Content-Type", "application/json")
	expectedHeader.Add("ACCEPT", "application/json,version="+FOREMAN_API_VERSION)
	expectedHeader.Add("Authorization", credentialsEncoded)

	for key := range expectedHeader {
		if req.Header.Get(key) != expectedHeader.Get(key) {
			t.Fatalf(
				"http.Request returned by Client.NewRequestWithContext() has incorrect HTTP header. "+
					"Expected [%s], got [%s] for Header key [%s].\n",
				expectedHeader.Get(key),
				req.Header.Get(key),
				key,
			)
		}
	}

}

// Ensures Client.NewRequestWithContext() is properly concatenating the server's URL
// and the endpoint when constructing the request's URL.
func TestNewRequest_URL(t *testing.T) {
	cred := ClientCredentials{}
	conf := ClientConfig{}
	_, server, client := NewForemanAPIAndClient(cred, conf)
	defer server.Close()

	// map with the endpoint as key and the expected constructed URL path as
	// the value
	testEndpoints := map[string]string{
		"/foo":     FOREMAN_API_URL_PREFIX + "/foo",
		"/":        FOREMAN_API_URL_PREFIX + "/",
		"":         FOREMAN_API_URL_PREFIX + "/",
		"/foo/bar": FOREMAN_API_URL_PREFIX + "/foo/bar",
		"foo/bar":  FOREMAN_API_URL_PREFIX + "/foo/bar",
	}

	for key, value := range testEndpoints {
		req, _ := client.NewRequestWithContext(context.TODO(), http.MethodGet, key, nil)
		expectedURL := client.Server.URL
		expectedURL.Path = value
		if *(req.URL) != expectedURL {
			t.Fatalf(
				"http.Request returned by Client.NewRequestWithContext() has incorrect URL. "+
					"Expected [%s], got [%s].\n",
				expectedURL.String(),
				req.URL.String(),
			)
		}
	}

}

// ----------------------------------------------------------------------------
// Client.Send
// ----------------------------------------------------------------------------

// Ensure Client.Send() returns an error when attempting to send a nil
// http.Request reference.
func TestSend_NilRequestError(t *testing.T) {
	cred := ClientCredentials{}
	conf := ClientConfig{}
	_, server, client := NewForemanAPIAndClient(cred, conf)
	defer server.Close()

	_, _, sendErr := client.Send(nil)
	if sendErr == nil {
		t.Fatalf(
			"Client.Send() did not return error when given nil. " +
				"Expected [error], got [nil].",
		)
	}

}

// Ensure Client.Send() returns the server's HTTP response status code.
func TestSend_StatusCode(t *testing.T) {
	cred := ClientCredentials{}
	conf := ClientConfig{}
	mux, server, client := NewForemanAPIAndClient(cred, conf)
	defer server.Close()

	// dummy '[GET] /foo' endpoint - just returns 200
	mux.HandleFunc(FOREMAN_API_URL_PREFIX+"/foo", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req, _ := client.NewRequestWithContext(context.TODO(), http.MethodGet, "/foo", nil)
	statusCode, _, _ := client.Send(req)

	if statusCode != http.StatusOK {
		t.Fatalf(
			"Client.Send() did not return correct status code from server. "+
				"Expected [%d], got [%d].",
			http.StatusOK,
			statusCode,
		)
	}

}

// Ensure Client.Send() returns the server's response body.
func TestSend_ResponseBody(t *testing.T) {
	cred := ClientCredentials{}
	conf := ClientConfig{}
	mux, server, client := NewForemanAPIAndClient(cred, conf)
	defer server.Close()

	expectedRespStr := "Hello, World!"
	expectedRespBody := []byte(expectedRespStr)

	// dummy '[GET] /foo' endpoint - returns "Hello, World!"
	mux.HandleFunc(FOREMAN_API_URL_PREFIX+"/foo", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(expectedRespBody)
	})

	req, _ := client.NewRequestWithContext(context.TODO(), http.MethodGet, "/foo", nil)
	_, respBody, _ := client.Send(req)

	if string(respBody) != expectedRespStr {
		t.Fatalf(
			"Client.Send() did not return the correct response body from server. "+
				"Expected [%s], got [%s].",
			expectedRespStr,
			respBody,
		)
	}

}

// ----------------------------------------------------------------------------
// IsNotFound
// ----------------------------------------------------------------------------

// Ensures IsNotFound recognizes a bare HTTPError for a 404 response.
func TestIsNotFound_Direct(t *testing.T) {
	err := HTTPError{Endpoint: "/foo", StatusCode: 404, RespBody: "not found"}
	if !IsNotFound(err) {
		t.Fatalf("IsNotFound did not recognize a direct 404 HTTPError")
	}
}

// Ensures IsNotFound recognizes an HTTPError wrapped with fmt.Errorf's %w,
// as errors.As would be expected to.
func TestIsNotFound_Wrapped(t *testing.T) {
	inner := HTTPError{Endpoint: "/foo", StatusCode: 404, RespBody: "not found"}
	wrapped := fmt.Errorf("while reading foo: %w", inner)
	if !IsNotFound(wrapped) {
		t.Fatalf("IsNotFound did not recognize a wrapped 404 HTTPError")
	}
}

// Ensures IsNotFound rejects a non-404 HTTPError and a plain error.
func TestIsNotFound_False(t *testing.T) {
	other := HTTPError{Endpoint: "/foo", StatusCode: 500, RespBody: "boom"}
	if IsNotFound(other) {
		t.Fatalf("IsNotFound incorrectly matched a 500 HTTPError")
	}
	if IsNotFound(fmt.Errorf("unrelated error")) {
		t.Fatalf("IsNotFound incorrectly matched a plain error")
	}
}

// ----------------------------------------------------------------------------
// Client.SendAndParse
// ----------------------------------------------------------------------------

// Ensure SendAndParse() returns an error when the server responds with a
// status code not in the 2xx range
func TestSendAndParseStatusCodeError(t *testing.T) {
	cred := ClientCredentials{
		Username: "Admin",
		Password: "ChangeMe",
	}
	conf := ClientConfig{}
	mux, server, client := NewForemanAPIAndClient(cred, conf)
	defer server.Close()

	// dummy '[GET] /foo' endpoint - returns 500 Internal server error
	mux.HandleFunc(FOREMAN_API_URL_PREFIX+"/foo", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	req, _ := client.NewRequestWithContext(context.TODO(), http.MethodGet, "/foo", nil)
	sendErr := client.SendAndParse(req, nil)
	if sendErr == nil {
		t.Errorf(
			"Client.ParseAndSend() did not return an error when the server responded " +
				"with a non-2xx status code. Expected [error] got [nil]",
		)
	}
}

// Ensure SendAndParse() returns no errors when the server responds with a
// status code in the 2xx range
func TestSendAndParseStatusCodeNoError(t *testing.T) {
	cred := ClientCredentials{
		Username: "Admin",
		Password: "ChangeMe",
	}
	conf := ClientConfig{}
	mux, server, client := NewForemanAPIAndClient(cred, conf)
	defer server.Close()

	// dummy '[GET] /foo' endpoint - returns 500 Internal server error
	mux.HandleFunc(FOREMAN_API_URL_PREFIX+"/foo", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req, _ := client.NewRequestWithContext(context.TODO(), http.MethodGet, "/foo", nil)
	sendErr := client.SendAndParse(req, nil)
	if sendErr != nil {
		t.Errorf(
			"Client.ParseAndSend() did not return an error when the server responded " +
				"with a 2xx status code. Expected [nil] got [error]",
		)
	}
}
