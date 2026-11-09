// Command host_with_parameters creates a Foreman host and associates
// parameters with it, both ways the API allows:
//
//   - inline, at creation time, via ForemanHost.HostParametersAttributes
//     ("host_parameters_attributes" on the wire): "environment" and "role"
//     below.
//   - afterwards, through the host's parameters endpoint
//     (CreateParameter): "owner" below, added once the host exists.
//
// It then lists the host's parameters with QueryParameter.
//
// The server and credentials come from the environment:
//
//	FOREMAN_URL=https://foreman.example.com \
//	FOREMAN_USERNAME=admin FOREMAN_PASSWORD=... \
//	go run ./examples/host_with_parameters
//
// The host's IDs (domain, operating system, host group) are placeholders:
// use IDs that exist on your Foreman to run it for real.
package main

import (
	"context"
	"log"
	"net/url"
	"os"

	"github.com/ahmet2mir/foremango/pkg/foreman"
)

func main() {
	serverURL, err := url.Parse(getenv("FOREMAN_URL", "https://foreman.example.com"))
	if err != nil {
		log.Fatal(err)
	}

	client := foreman.NewClient(
		foreman.Server{URL: *serverURL},
		foreman.ClientCredentials{
			Username: getenv("FOREMAN_USERNAME", "admin"),
			Password: os.Getenv("FOREMAN_PASSWORD"),
		},
		foreman.ClientConfig{
			// Set these (> 0) to create the host in a location and
			// organization; left at 0, none is sent.
			LocationID:     0,
			OrganizationID: 0,
		},
	)

	ctx := context.Background()

	// 1. Create the host, with two parameters attached right away.
	host, err := client.CreateHost(ctx, &foreman.ForemanHost{
		ForemanObject:     foreman.ForemanObject{Name: "web01.example.com"},
		DomainId:          1,
		OperatingsystemId: 1,
		HostgroupId:       1,
		Managed:           true,
		HostParametersAttributes: []foreman.HostHostParametersAttributes{
			{Name: "environment", Value: "staging"},
			{Name: "role", Value: "webserver"},
		},
	})
	if err != nil {
		log.Fatalf("CreateHost: %v", err)
	}
	log.Printf("created host %d: %s", host.Id, host.Name)

	// 2. Add one more parameter afterwards, without resending the host.
	param, err := client.CreateParameter(ctx, host.Id, &foreman.ForemanParameter{
		ForemanObject: foreman.ForemanObject{Name: "owner"},
		Value:         "platform-team",
	})
	if err != nil {
		log.Fatalf("CreateParameter: %v", err)
	}
	log.Printf("added parameter %d %s=%s", param.Id, param.Name, param.Value)

	// 3. List the host's parameters (an empty search lists all of them).
	params, err := client.QueryParameter(ctx, host.Id, "")
	if err != nil {
		log.Fatalf("QueryParameter: %v", err)
	}
	for _, p := range params.Results {
		log.Printf("host %d parameter %s=%s", host.Id, p.Name, p.Value)
	}
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
