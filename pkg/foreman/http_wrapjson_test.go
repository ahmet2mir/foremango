package foreman

import (
	"strings"
	"testing"
)

// WrapJSONWithTaxonomy only adds location_id/organization_id when both are
// >= 0; every other test in this package uses the -1/-1 default from
// newDummyServer, which never exercises that branch.
func TestWrapJSONWithTaxonomy_AddsLocationAndOrg(t *testing.T) {
	client := NewClient(Server{}, ClientCredentials{}, ClientConfig{
		LocationID:     1,
		OrganizationID: 2,
	})

	b, err := client.WrapJSONWithTaxonomy("thing", &ForemanObject{Name: "x"})
	if err != nil {
		t.Fatalf("WrapJSONWithTaxonomy: %v", err)
	}

	s := string(b)
	for _, want := range []string{`"location_id":1`, `"organization_id":2`} {
		if !strings.Contains(s, want) {
			t.Errorf("expected output to contain %q, got %s", want, s)
		}
	}
}
