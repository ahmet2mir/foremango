# foremango

A Go client library for the [Foreman](https://www.theforeman.org/) and
[Katello](https://theforeman.org/plugins/katello/) APIs.

The API layer in [`pkg/api`](pkg/api) is **not original work**. It is a vendored
copy of the `foreman/api` package from the
[terraform-coop/terraform-provider-foreman](https://github.com/terraform-coop/terraform-provider-foreman)
Terraform provider. All credit for that code belongs to its original authors.

## Credits and provenance

| | |
|---|---|
| Upstream project | [terraform-coop/terraform-provider-foreman](https://github.com/terraform-coop/terraform-provider-foreman) |
| Vendored at | commit [`457a810`](https://github.com/terraform-coop/terraform-provider-foreman/commit/457a810cddb50394d08420859b9dc006cca7e22a) (`v0.7.0-9-g457a810`, 2026-02-22) |
| License | Mozilla Public License 2.0 — same as upstream (see [LICENSE](LICENSE)) |

The upstream provider is itself a fork of a provider originally developed, owned
and maintained by the **SRE - Orchestration pod at Wayfair**. Thanks to everyone
who built and maintains it, in particular its most prolific contributors:
Lennart Weller, Dominik Pataky, Arthur Outhenin-Chalandre, Michael Gusek,
Kirill Shirinkin, and [all other contributors](https://github.com/terraform-coop/terraform-provider-foreman/graphs/contributors).

## Goal of this repository

**Keep the vendored API code as close to upstream as possible**, changing only
what is strictly required to make it build and work as a standalone package
inside `foremango`.

This is deliberate: staying byte-for-byte close to upstream means improvements
and fixes from `terraform-provider-foreman` can be re-synced with a plain diff
instead of a merge conflict. New functionality for `foremango` belongs in new
packages, not in `pkg/api`.

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
