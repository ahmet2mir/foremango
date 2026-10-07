package foreman

import (
	"encoding/json"
	"testing"
)

// Covers ForemanKatelloRepository.MarshalJSON's conditional fields and its
// switch over ContentType, called directly per case - exercising every
// branch through CreateKatelloRepository would mean juggling a different
// HTTP round trip per case for no extra benefit, since this method has no
// custom Unmarshal to pair it with.
func TestForemanKatelloRepository_MarshalJSON(t *testing.T) {
	cases := []struct {
		name string
		repo ForemanKatelloRepository
		want map[string]interface{}
	}{
		{
			"deb",
			ForemanKatelloRepository{ContentType: "deb", DebReleases: "bookworm"},
			map[string]interface{}{"content_type": "deb", "deb_releases": "bookworm"},
		},
		{
			"docker",
			ForemanKatelloRepository{ContentType: "docker", DockerUpstreamName: "library/nginx"},
			map[string]interface{}{"content_type": "docker", "docker_upstream_name": "library/nginx"},
		},
		{
			"ansible_collection",
			ForemanKatelloRepository{ContentType: "ansible_collection", AnsibleCollectionRequirements: "foo.bar"},
			map[string]interface{}{"content_type": "ansible_collection", "ansible_collection_requirements": "foo.bar"},
		},
		{
			"download_concurrency set",
			ForemanKatelloRepository{ContentType: "yum", DownloadConcurrency: 5},
			map[string]interface{}{"download_concurrency": 5.0},
		},
		{
			"gpg_key_id set",
			ForemanKatelloRepository{ContentType: "yum", GpgKeyId: 3},
			map[string]interface{}{"gpg_key_id": 3.0},
		},
		{
			"non-default http_proxy_policy includes http_proxy_id",
			ForemanKatelloRepository{ContentType: "yum", HttpProxyPolicy: "use_selected_http_proxy", HttpProxyId: 9},
			map[string]interface{}{"http_proxy_id": 9.0},
		},
		{
			"default http_proxy_policy omits http_proxy_id",
			ForemanKatelloRepository{ContentType: "yum", HttpProxyPolicy: "global_default_http_proxy", HttpProxyId: 9},
			nil, // asserted separately below
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			b, err := c.repo.MarshalJSON()
			if err != nil {
				t.Fatalf("MarshalJSON: %v", err)
			}
			var m map[string]interface{}
			if err := json.Unmarshal(b, &m); err != nil {
				t.Fatalf("decoding MarshalJSON output: %v", err)
			}
			if c.name == "default http_proxy_policy omits http_proxy_id" {
				if _, present := m["http_proxy_id"]; present {
					t.Fatalf("expected http_proxy_id to be omitted, got %v", m["http_proxy_id"])
				}
				return
			}
			for k, want := range c.want {
				if m[k] != want {
					t.Errorf("field %q = %v, want %v", k, m[k], want)
				}
			}
		})
	}
}
