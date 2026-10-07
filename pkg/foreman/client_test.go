package foreman

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"reflect"
	"testing"
)

// ----------------------------------------------------------------------------
// Test Helper Functions
// ----------------------------------------------------------------------------

// Creates a mock Foreman API server for testing with the client.
//
// NOTE(ALL): It is the caller's responsibility to call Close() on the server
// when finished to prevent a resource leak
func NewForemanAPI() (*http.ServeMux, *httptest.Server) {
	urlMux := http.NewServeMux()
	server := httptest.NewServer(urlMux)
	return urlMux, server
}

// In addition to creating and setting up the mock Foreman API server,
// initialize and set up a Client to communicate with the server.
//
// NOTE(ALL): It is the caller's responsibility to call Close() on the server
// when finished to prevent a resource leak
//
// cred
//
//	A set of credentials used to authenticate the client
func NewForemanAPIAndClient(cred ClientCredentials, conf ClientConfig) (*http.ServeMux, *httptest.Server, *Client) {
	urlMux, server := NewForemanAPI()
	// Server's URL is stored as a string, parse into a url.URL and point the
	// client at it.  url.Parse() will return *url.URL so dereference before
	// passing to the client.  Safely ignore the error, the server's URL should
	// be valid or there is a problem in Golang stdlib.
	serverURL, _ := url.Parse(server.URL)
	s := Server{
		URL: *serverURL,
	}
	// use unsafe TLS when talking to the mock server
	client := NewClient(s, cred, conf)
	return urlMux, server, client
}

func TestMain(m *testing.M) {
	SetOutput(io.Discard)
	os.Exit(m.Run())
}

// ----------------------------------------------------------------------------
// NewClient
// ----------------------------------------------------------------------------

// Ensures the client is not modifying the server's URL when creating new
// client struct.
func TestNewClient_ServerURL(t *testing.T) {
	cred := ClientCredentials{}
	conf := ClientConfig{}
	_, server := NewForemanAPI()
	defer server.Close()
	// create an instance of the client and point it to the server
	serverURL, _ := url.Parse(server.URL)
	client := NewClient(
		Server{
			URL: *serverURL,
		},
		cred,
		conf,
	)

	// Client should have its Server URL set to the mock Foreman API
	if client.Server.URL != *serverURL {
		t.Fatalf(
			"Server URL does not match the Client's server URL. "+
				"Expected [%s], got [%s].\n",
			serverURL.String(),
			client.Server.URL.String(),
		)
	}

}

// Ensures the client is not modifying the passed credentials when creating
// new client struct
func TestNewClient_Credentials(t *testing.T) {
	serv := Server{}
	cred := ClientCredentials{
		Username: "Admin",
		Password: "ChangeMe",
	}
	conf := ClientConfig{}
	client := NewClient(serv, cred, conf)

	// Client should have its Server URL set to the mock Foreman API
	if !reflect.DeepEqual(cred, client.credentials) {
		t.Fatalf(
			"Client credentials do not match the expected values. "+
				"Expected [%+v], got [%+v].\n",
			cred,
			client.credentials,
		)
	}
}

// Ensures if the client has enabled TLS insecure, then the client's
// underlying HTTP transport has disabled TLS verification. Otherwise,
// TLS verification should be enabled.
func TestNewClient_ConfigTLSInsecureEnabled(t *testing.T) {
	serv := Server{}
	cred := ClientCredentials{}

	testCases := []struct {
		Insecure      bool
		ExpectedValue bool
	}{
		{
			Insecure:      true,
			ExpectedValue: true,
		},
		{
			Insecure:      false,
			ExpectedValue: false,
		},
	}

	for _, testCase := range testCases {

		conf := ClientConfig{
			TLSInsecureEnabled: testCase.Insecure,
		}

		client := NewClient(serv, cred, conf)

		// http.Client.Transport is *http.RoundTripper (interface). Type assert
		// the underlying *http.Transport (struct) to read the transport
		// configuration
		httpClient, _ := client.HTTPClient().(*http.Client)
		transCfg, _ := httpClient.Transport.(*http.Transport)
		tlsCfg := transCfg.TLSClientConfig

		if tlsCfg.InsecureSkipVerify != testCase.ExpectedValue {
			t.Fatalf(
				"Client did not properly set TLS config from configuration. "+
					"Expected TLSClientConfig.InsecureSkipVerify to be [%t], got "+
					"[%t] for insecure [%t]",
				testCase.ExpectedValue,
				tlsCfg.InsecureSkipVerify,
				testCase.Insecure,
			)
		}
	}
}

// Ensures that ClientConfig.HTTPClient, when set, is used as-is instead of
// the default client NewClient would otherwise build.
func TestNewClient_ConfigHTTPClientOverride(t *testing.T) {
	serv := Server{}
	cred := ClientCredentials{}
	custom := &http.Client{}

	conf := ClientConfig{
		HTTPClient: custom,
	}
	client := NewClient(serv, cred, conf)

	if client.HTTPClient() != HTTPDoer(custom) {
		t.Fatalf(
			"Client did not use the HTTPDoer supplied via ClientConfig.HTTPClient.",
		)
	}
}

func TestClient_SetHTTPClient(t *testing.T) {
	client := NewClient(Server{}, ClientCredentials{}, ClientConfig{})

	replacement := &http.Client{}
	client.SetHTTPClient(replacement)

	if client.HTTPClient() != HTTPDoer(replacement) {
		t.Fatalf("SetHTTPClient did not replace the Client's HTTPDoer")
	}
}
