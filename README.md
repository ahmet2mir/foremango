# foremango

A Go client library for the [Foreman](https://www.theforeman.org/) API.

* Every resource type and its `Create`/`Read`/`Update`/`Delete`/`Query`
  methods are generated from Foreman's own API description
  (`/apidoc/v2.json`) by [`tools/apigen`](tools/apigen).
* Requests go through [resty](https://resty.dev) (`resty.dev/v3`). You can
  supply and configure your own resty client.
* Two direct dependencies: `resty.dev/v3`, and `go-spnego` for optional
  Kerberos/SPNEGO auth.

## Package layout

Everything is in one package, `pkg/foreman`:

* `client.go` — `Client`, `ClientCredentials`, `ClientConfig`, `NewClient`.
* `http.go` — `Client.Do` (the one call every request goes through),
  endpoint routing (`/api`, `/katello/api`, `/foreman_puppet/api`,
  `/foreman_tasks/api`), `HTTPError`, `IsNotFound`, `Client.Payload`.
* `types.go` — `ForemanObject` (embedded by every resource) and `List[T]`
  (what every `QueryX` returns).
* `server.go` — `Server`.
* `logger.go` — the `Logger` interface and the built-in default logger.
* `<resource>_generated.go` — one file per Foreman resource
  (`architectures_generated.go`, `hosts_generated.go`, ...). Don't edit
  these by hand: change the generator and regenerate (see
  [`tools/apigen/README.md`](tools/apigen/README.md)).

## Usage

```go
package main

import (
	"context"
	"log"
	"net/url"

	"github.com/ahmet2mir/foremango/pkg/foreman"
)

func main() {
	serverURL, err := url.Parse("https://foreman.example.com")
	if err != nil {
		log.Fatal(err)
	}

	c := foreman.NewClient(
		foreman.Server{URL: *serverURL},
		foreman.ClientCredentials{
			Username: "admin",
			Password: "changeme",
		},
		foreman.ClientConfig{
			// Optional: sent with every create/update when > 0.
			LocationID:     2,
			OrganizationID: 1,
		},
	)
	ctx := context.Background()

	arch, err := c.ReadArchitecture(ctx, 1)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("architecture %d: %s", arch.Id, arch.Name)

	// Query methods take a Foreman search expression and return typed results.
	hosts, err := c.QueryHost(ctx, `name ~ "web"`)
	if err != nil {
		log.Fatal(err)
	}
	for _, h := range hosts.Results {
		log.Printf("host %d: %s", h.Id, h.Name)
	}
}
```

Errors from Foreman (non-2xx answers) are `foreman.HTTPError` values;
`foreman.IsNotFound(err)` tells you whether one was a 404.

More in [`examples`](examples) - e.g. [`host_with_parameters`](examples/host_with_parameters)
for creating a host and associating parameters with it.

## Customizing the HTTP client

`NewClient` builds a default resty client from
`ClientConfig.TLSInsecureEnabled`/`NegotiateAuthEnabled`. To control retries,
timeouts, proxies, mTLS, middleware, tracing and so on, build your own
`*resty.Client` and pass it as `ClientConfig.RestyClient`:

```go
rc := resty.New().
	SetTimeout(30 * time.Second).
	SetRetryCount(3).
	SetRootCertificates("/etc/pki/foreman-ca.pem")

c := foreman.NewClient(
	foreman.Server{URL: *serverURL},
	foreman.ClientCredentials{Username: "admin", Password: "changeme"},
	foreman.ClientConfig{
		RestyClient: rc, // TLSInsecureEnabled/NegotiateAuthEnabled are ignored when this is set
	},
)
```

For each request foremango only sets the full URL, the `Accept` and
`Content-Type` headers, and basic auth from `ClientCredentials`. Leave both
credential fields empty if your resty client handles authentication itself.
Everything else comes from your client.

`c.RestyClient()` returns the resty client in use (yours, or the one
`NewClient` built), and `c.SetRestyClient(rc)` replaces it after
construction, e.g. in tests.

`c.Do(ctx, method, endpoint, query, body, &out)` is the one call every
generated method goes through. It is public, so you can reach endpoints the
generator doesn't cover yet:

```go
var status map[string]any
err := c.Do(ctx, http.MethodGet, "status", nil, nil, &status) // GET /api/status
```

## Custom logger

Every `*Client` holds its own logger, `c.log`, that every method on it -
every resource's `CreateX`/`ReadX`/..., `Do`, everything in `http.go` - logs
through. The default resty client also routes resty's own warnings and errors
through it. Set it per client via `ClientConfig.Logger`:

```go
type Logger interface {
	Tracef(format string, a ...interface{})
	Debugf(format string, a ...interface{})
	Infof(format string, a ...interface{})
	Warningf(format string, a ...interface{})
	Errorf(format string, a ...interface{})
}

c := foreman.NewClient(server, creds, foreman.ClientConfig{
	Logger: myZapOrLogrusOrSlogAdapter, // implements foreman.Logger
})
```

Leave `ClientConfig.Logger` unset and the client falls back to the built-in
default logger - a small stdlib-based leveled logger. Configure *that* one
(level `INFO` by default, raise it with the `FOREMANGO_LOG_LEVEL` env var -
`trace`, `debug`, `info`, `warning`, `error`, `none` - or programmatically)
with the package-level functions in `logger.go`:

```go
foreman.SetLevel(foreman.LevelDebug)
foreman.SetOutput(myWriter)
```

These same package-level functions (`Tracef`, `Debugf`, `Infof`, ...) are
also what the rare piece of code with no `*Client` around uses directly - a
resource type's `UnmarshalJSON`, say. That logging always goes through the
default logger; `ClientConfig.Logger` only affects `c.log` on `Client`
methods.

## Development

```sh
go build ./...
go vet ./...
go test ./...
```

## License

Mozilla Public License 2.0 — see [LICENSE](LICENSE).

## Thanks

foremango started as a fork of the `foreman/api` package of
[terraform-coop/terraform-provider-foreman](https://github.com/terraform-coop/terraform-provider-foreman),
itself a fork of the provider originally built by the SRE - Orchestration
pod at Wayfair. Thanks to everyone who built and maintains it, in
particular Lennart Weller, Dominik Pataky, Arthur Outhenin-Chalandre,
Michael Gusek, Kirill Shirinkin and
[all other contributors](https://github.com/terraform-coop/terraform-provider-foreman/graphs/contributors).

foremango has since been rewritten (resty-based transport, resources
generated from Foreman's apidoc) and is not kept in sync with it.
