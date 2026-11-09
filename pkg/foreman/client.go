// Package foreman is a Go client for the Foreman API.
//
// A Client sends every request through Client.Do, on top of a resty client
// (resty.dev/v3) that callers can supply and configure themselves. The
// resource types and their CreateX/ReadX/UpdateX/DeleteX/QueryX methods
// live in the *_generated.go files, produced by tools/apigen from
// Foreman's apidoc.
//
//	client.go   - Client, its configuration, NewClient.
//	http.go     - Client.Do, endpoint routing, HTTPError, Payload.
//	types.go    - ForemanObject and List, shared by every resource.
//	server.go   - Server.
//	logger.go   - the Logger interface and the built-in default logger.
package foreman

import (
	"crypto/tls"
	"net/http"

	"github.com/dpotapov/go-spnego"
	"resty.dev/v3"
)

// ClientCredentials are the username and password sent as HTTP basic auth
// with every request. Leave both empty to send no basic auth, when the
// resty client authenticates by other means (SPNEGO, a token header, ...).
type ClientCredentials struct {
	Username string
	Password string
}

// ClientConfig configures a Client.
type ClientConfig struct {
	// TLSInsecureEnabled turns off verification of the server's TLS
	// certificate. Only for test servers with a self-signed certificate;
	// prefer trusting the server's CA on your own RestyClient. Ignored when
	// RestyClient is set.
	TLSInsecureEnabled bool

	// NegotiateAuthEnabled authenticates with HTTP Negotiate (SPNEGO,
	// Kerberos) instead of basic auth. Ignored when RestyClient is set.
	NegotiateAuthEnabled bool

	// LocationID and OrganizationID, when > 0, are sent with every create
	// and update (see Client.Payload) so new objects land in that location
	// and organization.
	LocationID     int
	OrganizationID int

	// RestyClient, when set, sends every request as-is; TLSInsecureEnabled
	// and NegotiateAuthEnabled are then ignored. Configure retries,
	// timeouts, transport, CA certificates, middleware, tracing and so on
	// on it directly. Per request, foremango only sets the URL, the Accept
	// and Content-Type headers, and basic auth from ClientCredentials.
	RestyClient *resty.Client

	// Logger receives this client's diagnostics. Defaults to the built-in
	// logger (see logger.go).
	Logger Logger
}

// Client sends requests to one Foreman server. Every resource method is
// defined on it.
type Client struct {
	// Server is the Foreman server requests are sent to.
	Server Server
	// Config is the configuration the client was built with.
	Config ClientConfig

	// credentials is unexported so the password can't be read back from
	// a Client.
	credentials ClientCredentials
	resty       *resty.Client
	log         Logger
}

// NewClient returns a Client for server s, authenticating with cred.
func NewClient(s Server, cred ClientCredentials, cfg ClientConfig) *Client {
	TraceFunctionCall()

	log := cfg.Logger
	if log == nil {
		log = defaultLogger
	}

	rc := cfg.RestyClient
	if rc == nil {
		rc = newDefaultRestyClient(cfg, log)
	}

	return &Client{
		Server:      s,
		Config:      cfg,
		credentials: cred,
		resty:       rc,
		log:         log,
	}
}

// newDefaultRestyClient builds the resty client used when
// ClientConfig.RestyClient is not set.
func newDefaultRestyClient(cfg ClientConfig, log Logger) *resty.Client {
	tlsConfig := &tls.Config{
		MinVersion: tls.VersionTLS12,
		// #nosec G402 -- opt-in through ClientConfig.TLSInsecureEnabled,
		// false by default.
		InsecureSkipVerify: cfg.TLSInsecureEnabled,
	}

	rc := resty.New().
		SetLogger(restyLogger{log}).
		SetHeader("User-Agent", "foremango")

	if cfg.NegotiateAuthEnabled {
		transport := &spnego.Transport{}
		transport.Proxy = http.ProxyFromEnvironment
		transport.TLSClientConfig = tlsConfig
		rc.SetTransport(transport)
	} else {
		rc.SetTLSClientConfig(tlsConfig)
	}

	return rc
}

// restyLogger sends resty's own messages to a Logger.
type restyLogger struct{ l Logger }

func (r restyLogger) Errorf(format string, v ...any) { r.l.Errorf(format, v...) }
func (r restyLogger) Warnf(format string, v ...any)  { r.l.Warningf(format, v...) }
func (r restyLogger) Debugf(format string, v ...any) { r.l.Debugf(format, v...) }
