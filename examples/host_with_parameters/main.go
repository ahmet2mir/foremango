// Command host_with_parameters shows how to create a Foreman host and
// associate parameters with it, both ways the library supports:
//
//   - inline, at creation time, via ForemanHost.HostParameters
//     ("host_parameters_attributes" on the wire) - the two parameters set
//     on the host below, "environment" and "role".
//   - afterwards, as a separate call, via the dedicated Parameter resource
//     (CreateParameter) - the "owner" parameter added once the host
//     already exists.
//
// Every value here - the server URL, credentials, and the host's own
// fields - is a placeholder. Point ServerURL/Username/Password at a real
// Foreman instance (and use IDs that actually exist there for
// DomainId/OperatingSystemId/HostgroupId) to run this for real; as
// written, it's meant to be read as a reference, not executed verbatim.
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

	client := foreman.NewClient(
		foreman.Server{URL: *serverURL},
		foreman.ClientCredentials{
			Username: "admin",
			Password: "changeme",
		},
		foreman.ClientConfig{
			// Foreman < 1.21 with organizations/locations disabled: use < 0.
			LocationID:     -1,
			OrganizationID: -1,
		},
	)

	ctx := context.Background()

	// 1. Create the host, with a couple of parameters attached right away.
	host, err := client.CreateHost(ctx, &foreman.ForemanHost{
		ForemanObject:     foreman.ForemanObject{Name: "web01.example.com"},
		DomainId:          intPtr(1),
		OperatingSystemId: intPtr(1),
		HostgroupId:       intPtr(1),
		Managed:           true,
		HostParameters: []foreman.ForemanKVParameter{
			{Name: "environment", Value: "staging"},
			{Name: "role", Value: "webserver"},
		},
	}, 1 /* retryCount */)
	if err != nil {
		log.Fatalf("CreateHost: %v", err)
	}
	log.Printf("created host %d: %s", host.Id, host.Name)

	// 2. Associate one more parameter after the fact, via the dedicated
	// Parameter resource instead of resending the whole host.
	param, err := client.CreateParameter(ctx, &foreman.ForemanParameter{
		HostID: host.Id,
		Parameter: foreman.ForemanKVParameter{
			Name:  "owner",
			Value: "platform-team",
		},
	})
	if err != nil {
		log.Fatalf("CreateParameter: %v", err)
	}
	log.Printf("associated parameter %q=%q with host %d", param.Parameter.Name, param.Parameter.Value, host.Id)
}

// intPtr is a small helper since ForemanHost's relational fields
// (DomainId, OperatingSystemId, HostgroupId, ...) are *int - nil meaning
// "don't set this" - so a literal can't be used directly.
func intPtr(i int) *int { return &i }
