package foreman

import "testing"

func TestForemanParameter_apiEndpoint(t *testing.T) {
	cases := []struct {
		name     string
		param    ForemanParameter
		wantType string
		wantID   int
	}{
		{"host", ForemanParameter{HostID: 1}, "hosts", 1},
		{"hostgroup", ForemanParameter{HostGroupID: 2}, "hostgroups", 2},
		{"domain", ForemanParameter{DomainID: 3}, "domains", 3},
		{"operatingsystem", ForemanParameter{OperatingSystemID: 4}, "operatingsystems", 4},
		{"subnet", ForemanParameter{SubnetID: 5}, "subnets", 5},
		{"none set", ForemanParameter{}, "", -1},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			gotType, gotID := c.param.apiEndpoint()
			if gotType != c.wantType || gotID != c.wantID {
				t.Errorf("apiEndpoint() = (%q, %d), want (%q, %d)", gotType, gotID, c.wantType, c.wantID)
			}
		})
	}
}
