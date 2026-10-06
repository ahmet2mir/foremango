# foremango

A Go client library for the [Foreman](https://www.theforeman.org/) and
[Katello](https://theforeman.org/plugins/katello/) APIs.

The API layer in [`pkg/api`](pkg/api) started out as a vendored copy of the
`foreman/api` package from the
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
diverge — add, remove, or change anything in `pkg/api` as `foremango`'s own
needs dictate, without worrying about re-syncing with a plain diff.

Concretely, this already means `pkg/api` has been decoupled from the
Terraform ecosystem entirely:

* No `github.com/hashicorp/terraform-plugin-sdk` dependency. The one function
  that pulled it in, `CheckDeleted(d *schema.ResourceData, err error)`, made
  no sense for a plain library anyway (nothing here manages Terraform state);
  it's gone, replaced by a framework-agnostic `IsNotFound(err error) bool`.
* No `github.com/HanseMerkur/terraform-provider-utils` dependency. Its `log`
  package is replaced by a small internal logger in [`pkg/utils`](pkg/utils),
  built on the Go standard library only. Verbosity defaults to `INFO` and can
  be raised with the `FOREMANGO_LOG_LEVEL` environment variable (`trace`,
  `debug`, `info`, `warning`, `error`, `none`) or programmatically via
  `utils.SetLevel`.
* The result: `go.mod` carries a handful of direct dependencies
  (`go-cleanhttp` for the HTTP client, `go-spnego` for optional Kerberos/SPNEGO
  auth) instead of the ~45 packages the Terraform SDK dragged in transitively.

`go-spnego` is kept — it isn't Terraform-specific, it's what backs
`ClientConfig.NegotiateAuthEnabled` for talking to a Foreman server behind
HTTP Negotiate auth, and dropping it would remove real functionality nobody
asked to lose.

## Usage

```go
package main

import (
	"context"
	"log"
	"net/url"

	"github.com/ahmet2mir/foremango/pkg/api"
)

func main() {
	serverURL, err := url.Parse("https://foreman.example.com")
	if err != nil {
		log.Fatal(err)
	}

	client := api.NewClient(
		api.Server{URL: *serverURL},
		api.ClientCredentials{
			Username: "admin",
			Password: "changeme",
		},
		api.ClientConfig{
			TLSInsecureEnabled: false,
			// Foreman < 1.21 with organizations/locations disabled: use < 0.
			LocationID:     -1,
			OrganizationID: -1,
		},
	)

	arch, err := client.ReadArchitecture(context.Background(), 1)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("architecture %d: %s", arch.Id, arch.Name)
}
```

## Development

```sh
go build ./...
go vet ./...
go test ./...
```

## License

Mozilla Public License 2.0 — see [LICENSE](LICENSE). The vendored code retains
its original MPL-2.0 license from upstream.
