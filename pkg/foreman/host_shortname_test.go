package foreman

import "testing"

func TestConstructShortname(t *testing.T) {
	cases := []struct {
		name          string
		host          foremanHostDecode
		wantShortname string
	}{
		{
			name:          "shortname already set is left alone",
			host:          foremanHostDecode{ForemanHost: ForemanHost{Shortname: "preset"}},
			wantShortname: "preset",
		},
		{
			name: "no dot in name",
			host: foremanHostDecode{
				ForemanHost: ForemanHost{ForemanObject: ForemanObject{Name: "host1"}},
			},
			wantShortname: "host1",
		},
		{
			name: "dotted name, no DomainName set",
			host: foremanHostDecode{
				ForemanHost: ForemanHost{ForemanObject: ForemanObject{Name: "host1.example.com"}},
			},
			wantShortname: "host1",
		},
		{
			name: "dotted name, matching DomainName",
			host: foremanHostDecode{
				ForemanHost: ForemanHost{
					ForemanObject: ForemanObject{Name: "host1.example.com"},
					DomainName:    "example.com",
				},
			},
			wantShortname: "host1",
		},
		{
			name: "dotted name, mismatched DomainName logs but still succeeds",
			host: foremanHostDecode{
				ForemanHost: ForemanHost{
					ForemanObject: ForemanObject{Name: "host1.example.com"},
					DomainName:    "other.example.org",
				},
			},
			wantShortname: "host1",
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			host := c.host
			if err := constructShortname(&host); err != nil {
				t.Fatalf("constructShortname: %v", err)
			}
			if host.Shortname != c.wantShortname {
				t.Errorf("Shortname = %q, want %q", host.Shortname, c.wantShortname)
			}
		})
	}
}
