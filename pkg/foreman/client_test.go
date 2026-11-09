package foreman

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"

	"github.com/dpotapov/go-spnego"
	"resty.dev/v3"
)

func TestMain(m *testing.M) {
	SetOutput(io.Discard)
	os.Exit(m.Run())
}

// testClient starts an HTTP server answering with h and returns a Client
// pointed at it. The server is closed when the test ends.
func testClient(t *testing.T, cred ClientCredentials, cfg ClientConfig, h http.HandlerFunc) *Client {
	t.Helper()
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)
	u, err := url.Parse(srv.URL)
	if err != nil {
		t.Fatalf("parsing test server URL: %v", err)
	}
	return NewClient(Server{URL: *u}, cred, cfg)
}

func TestNewClient_KeepsServerAndConfig(t *testing.T) {
	u, _ := url.Parse("https://foreman.example.com")
	cfg := ClientConfig{LocationID: 3, OrganizationID: 4}
	c := NewClient(Server{URL: *u}, ClientCredentials{Username: "u", Password: "p"}, cfg)

	if c.Server.URL != *u {
		t.Errorf("Server.URL = %s, want %s", c.Server.URL.String(), u)
	}
	if c.Config.LocationID != 3 || c.Config.OrganizationID != 4 {
		t.Errorf("Config = %+v, want location 3 organization 4", c.Config)
	}
	if c.credentials.Username != "u" || c.credentials.Password != "p" {
		t.Errorf("credentials not kept")
	}
}

func TestNewClient_TLSInsecureEnabled(t *testing.T) {
	for _, insecure := range []bool{false, true} {
		c := NewClient(Server{}, ClientCredentials{}, ClientConfig{TLSInsecureEnabled: insecure})
		tlsCfg := c.RestyClient().TLSClientConfig()
		if tlsCfg == nil {
			t.Fatalf("default resty client has no TLS config")
		}
		if tlsCfg.InsecureSkipVerify != insecure {
			t.Errorf("InsecureSkipVerify = %t, want %t", tlsCfg.InsecureSkipVerify, insecure)
		}
	}
}

func TestNewClient_NegotiateAuthEnabled(t *testing.T) {
	c := NewClient(Server{}, ClientCredentials{}, ClientConfig{NegotiateAuthEnabled: true})
	if _, ok := c.RestyClient().Transport().(*spnego.Transport); !ok {
		t.Errorf("transport = %T, want *spnego.Transport", c.RestyClient().Transport())
	}
}

func TestNewClient_RestyClientOverride(t *testing.T) {
	rc := resty.New()
	c := NewClient(Server{}, ClientCredentials{}, ClientConfig{RestyClient: rc})
	if c.RestyClient() != rc {
		t.Errorf("ClientConfig.RestyClient was not used")
	}

	other := resty.New()
	c.SetRestyClient(other)
	if c.RestyClient() != other {
		t.Errorf("SetRestyClient did not replace the resty client")
	}
}
