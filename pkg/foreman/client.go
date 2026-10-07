// Package foreman is foremango's Foreman/Katello API client: connection
// config (Server, ClientCredentials, ClientConfig), the Client that builds
// and sends requests and exposes every Foreman/Katello resource method
// (CreateX/ReadX/UpdateX/DeleteX/QueryX), and the pluggable Logger it uses.
// Everything lives in this one package - there is no separate transport
// package to embed or wrap.
//
// The package is split by concern across files so each piece is easy to
// find:
//
//	client.go   - this file: just Client and how to build one (NewClient).
//	http.go     - request building/sending, HTTPError, HTTPDoer.
//	logger.go   - the Logger interface and the built-in default logger.
//	server.go   - the Server type.
//	<resource>.go - one file per Foreman/Katello resource (architecture.go,
//	  host.go, ...): its type plus its CreateX/ReadX/UpdateX/DeleteX/QueryX
//	  methods on *Client.
package foreman

import (
	"crypto/tls"
	"net/http"

	"github.com/dpotapov/go-spnego"

	cleanhttp "github.com/hashicorp/go-cleanhttp"
)

// ClientCredentials used to authenticate the client against the remote
// server - in this case, the Foreman API.
type ClientCredentials struct {
	Username string
	Password string
}

// ClientConfig holds configurable features of the REST client.
type ClientConfig struct {
	// Whether or not to verify the server's certificate/hostname.  This flag
	// is passed to the TLS config when initializing the REST client for API
	// communication. Ignored if HTTPClient is set.
	//
	// See 'pkg/crypto/tls/#Config.InsecureSkipVerify' for more information
	TLSInsecureEnabled bool

	// Whether or not the client should try to authenticate to foreman
	// through the HTTP negotiate mechanism. Ignored if HTTPClient is set.
	NegotiateAuthEnabled bool

	// Information as required by all API calls
	LocationID     int
	OrganizationID int

	// HTTPClient, when set, is used as-is to send every request instead of
	// the client NewClient would otherwise build from TLSInsecureEnabled and
	// NegotiateAuthEnabled (which are then ignored). Use this to fully
	// control transport behavior: a custom *http.Client with its own
	// *http.Transport (proxies, connection pooling, mTLS, ...), a
	// retrying/instrumented wrapper, or a test double satisfying HTTPDoer.
	//
	// *http.Client satisfies HTTPDoer without any change. See http.go.
	HTTPClient HTTPDoer

	// Logger, when set, is used as this client's logger instead of the
	// built-in default (see logger.go). Implement the Logger interface to
	// route this client's diagnostics through zap, logrus, slog, or
	// anything else.
	Logger Logger
}

// Client is the REST client used to communicate with the Foreman API, and
// the receiver of every resource method in this package.
type Client struct {
	// Server used to communicate and interact with the API.
	Server Server
	// Configuration this client was built with - the same ClientConfig
	// passed to NewClient (including the resolved HTTPClient would be
	// confusing here, so HTTPClient is accessible via the HTTPClient()
	// method instead; see http.go).
	Config ClientConfig

	// credentials to authenticate the client. Deliberately unexported: there
	// is no legitimate need for a caller holding a *Client to read the
	// password back out of it.
	credentials ClientCredentials

	// httpClient is the HTTPDoer actually used to send requests - either
	// Config.HTTPClient verbatim, or the one NewClient built from
	// Config.TLSInsecureEnabled / Config.NegotiateAuthEnabled. See http.go.
	httpClient HTTPDoer

	// log is this client's Logger - Config.Logger if set, else
	// defaultLogger (see logger.go). Every method on Client logs through
	// c.log; call sites with no Client at hand (a resource type's
	// UnmarshalJSON, say) use the package-level functions in logger.go
	// instead, which always go through defaultLogger.
	log Logger
}

// NewClient creates a new instance of the REST client for communication with
// the API gateway.
func NewClient(s Server, cred ClientCredentials, cfg ClientConfig) *Client {
	TraceFunctionCall()
	Debugf(
		"Server: [%+v], "+
			"ClientConfig: [%+v]",
		s,
		cfg,
	)

	doer := cfg.HTTPClient
	if doer == nil {
		// Initialize the HTTP client for use by the provider.  The insecure
		// flag from the provider config is used when configuring the TLS
		// settings of the HTTP client.
		cleanClient := cleanhttp.DefaultClient()
		tlsClientConfig := &tls.Config{
			InsecureSkipVerify: cfg.TLSInsecureEnabled,
		}
		if cfg.NegotiateAuthEnabled {
			transCfg := &spnego.Transport{}
			transCfg.TLSClientConfig = tlsClientConfig
			cleanClient.Transport = transCfg
		} else {
			transCfg := &http.Transport{}
			transCfg.TLSClientConfig = tlsClientConfig
			cleanClient.Transport = transCfg
		}
		doer = cleanClient
	}

	log := cfg.Logger
	if log == nil {
		log = defaultLogger
	}

	return &Client{
		Server:      s,
		Config:      cfg,
		credentials: cred,
		httpClient:  doer,
		log:         log,
	}
}
