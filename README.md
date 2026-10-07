# foremango

A Go client library for the [Foreman](https://www.theforeman.org/) and
[Katello](https://theforeman.org/plugins/katello/) APIs.

`pkg/foreman` started out as a vendored copy of the `foreman/api` package from
the
[terraform-coop/terraform-provider-foreman](https://github.com/terraform-coop/terraform-provider-foreman)
Terraform provider. Credit for that original code belongs to its authors —
see below. `foremango` has since diverged from it (see "Goal of this
repository") and is not kept in sync with upstream.

## Credits and provenance

| | |
|---|---|
| Upstream project | [terraform-coop/terraform-provider-foreman](https://github.com/terraform-coop/terraform-provider-foreman) |
| Vendored from | commit [`457a810`](https://github.com/terraform-coop/terraform-provider-foreman/commit/457a810cddb50394d08420859b9dc006cca7e22a) (`v0.7.0-9-g457a810`, 2026-02-22) |
| License | Mozilla Public License 2.0 — same as upstream (see [LICENSE](LICENSE)) |

The upstream provider is itself a fork of a provider originally developed, owned
and maintained by the **SRE - Orchestration pod at Wayfair**. Thanks to everyone
who built and maintains it, in particular its most prolific contributors:
Lennart Weller, Dominik Pataky, Arthur Outhenin-Chalandre, Michael Gusek,
Kirill Shirinkin, and [all other contributors](https://github.com/terraform-coop/terraform-provider-foreman/graphs/contributors).

## Goal of this repository

`foremango` is an independent Go client library, not a Terraform-provider
component. It started life as vendored code from
`terraform-provider-foreman`, but that lineage is history now, not a
constraint: this project is **not kept in sync with upstream** and is free to
diverge — add, remove, or change anything in `pkg/foreman` as `foremango`'s
own needs dictate, without worrying about re-syncing with a plain diff.

Concretely, this already means `pkg/foreman` has been decoupled from the
Terraform ecosystem entirely:

* No `github.com/hashicorp/terraform-plugin-sdk` dependency. The one function
  that pulled it in, `CheckDeleted(d *schema.ResourceData, err error)`, made
  no sense for a plain library anyway (nothing here manages Terraform state);
  it's gone, replaced by a framework-agnostic `IsNotFound(err error) bool`
  (built on `errors.As`, so it also recognizes an `HTTPError` wrapped with
  `%w` further up the call chain, not just a bare one).
* No `github.com/HanseMerkur/terraform-provider-utils` dependency. Its `log`
  package is replaced by a small `Logger` interface (see "Custom logger"
  below), built on the Go standard library only. Verbosity defaults to
  `INFO` and can be raised with the `FOREMANGO_LOG_LEVEL` environment
  variable (`trace`, `debug`, `info`, `warning`, `error`, `none`) or
  programmatically via `foreman.SetLevel`.
* The result: `go.mod` carries a handful of direct dependencies
  (`go-cleanhttp` for the HTTP client, `go-spnego` for optional Kerberos/SPNEGO
  auth) instead of the ~45 packages the Terraform SDK dragged in transitively.

`go-spnego` is kept — it isn't Terraform-specific, it's what backs
`ClientConfig.NegotiateAuthEnabled` for talking to a Foreman server behind
HTTP Negotiate auth, and dropping it would remove real functionality nobody
asked to lose.

## Package layout

Everything lives in one package, `pkg/foreman` - there's no separate
transport package to embed or wrap. Every Foreman/Katello resource method
(`CreateArchitecture`, `ReadHost`, `QueryDomain`, ...) is defined directly on
`*foreman.Client`, right alongside the HTTP plumbing that sends its requests.

Files are split by concern so each one is easy to find:

* `client.go` — just `Client`, `ClientCredentials`, `ClientConfig`, and how
  to build one (`NewClient`). Nothing else.
* `http.go` — request building and sending: `NewRequestWithContext`, `Send`,
  `SendAndParse` (Foreman/Katello-protocol aware: it knows how to wait out
  Katello's async-task responses), `WrapJSON`/`WrapJSONWithTaxonomy`,
  `HTTPDoer`, `HTTPError`, `IsNotFound`, `HTTPClient()`/`SetHTTPClient()`.
* `logger.go` — the `Logger` interface, the built-in default logger, and the
  handful of package-level logging functions used by code that has no
  `*Client` at hand (see "Custom logger" below).
* `server.go` — the `Server` type.
* One file per resource (`architecture.go`, `host.go`, `katello_content_views.go`,
  ...): its `Foreman<X>` type and its `CreateX`/`ReadX`/`UpdateX`/`DeleteX`/`QueryX`
  methods on `*Client`.

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
			TLSInsecureEnabled: false,
			// Foreman < 1.21 with organizations/locations disabled: use < 0.
			LocationID:     -1,
			OrganizationID: -1,
		},
	)

	arch, err := c.ReadArchitecture(context.Background(), 1)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("architecture %d: %s", arch.Id, arch.Name)
}
```

More in [`examples`](examples) - e.g. [`host_with_parameters`](examples/host_with_parameters)
for creating a host and associating parameters with it.

## Using your own HTTP client

`NewClient` builds a default `*http.Client` from
`ClientConfig.TLSInsecureEnabled`/`NegotiateAuthEnabled`, but you can supply
your own via `ClientConfig.HTTPClient` instead - any instrumentation,
retries, custom `*http.Transport` (proxies, mTLS, connection pooling, ...),
or test double is fine, as long as it satisfies `HTTPDoer`:

```go
type HTTPDoer interface {
	Do(req *http.Request) (*http.Response, error)
}
```

`*http.Client` already satisfies it, so this is a drop-in:

```go
myHTTPClient := &http.Client{
	Timeout:   30 * time.Second,
	Transport: myInstrumentedTransport,
}

c := foreman.NewClient(
	foreman.Server{URL: *serverURL},
	foreman.ClientCredentials{Username: "admin", Password: "changeme"},
	foreman.ClientConfig{
		HTTPClient: myHTTPClient, // TLSInsecureEnabled/NegotiateAuthEnabled are ignored when this is set
	},
)
```

`c.HTTPClient()` returns whichever `HTTPDoer` is currently in use (the one
you supplied, or the one `NewClient` built); `c.SetHTTPClient(d)` swaps it
after construction, e.g. in tests.

## Custom logger

Every `*Client` holds its own logger, `c.log`, that every method on it -
every resource's `CreateX`/`ReadX`/..., `Send`, `SendAndParse`, everything in
`http.go` - logs through. Set it per client via `ClientConfig.Logger`:

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

Mozilla Public License 2.0 — see [LICENSE](LICENSE). The vendored code retains
its original MPL-2.0 license from upstream.
