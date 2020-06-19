package resources

import (
	"encoding/json"
	"testing"

	"github.com/codahale/belvedere/pkg/belvedere/cfg"
	"github.com/codahale/belvedere/pkg/belvedere/internal/fixtures"
	compute "google.golang.org/api/compute/v1"
	"google.golang.org/api/dns/v1"
)

func TestAppResources(t *testing.T) {
	zone := &dns.ManagedZone{
		Name:    "belvedere",
		DnsName: "horse.club",
	}
	resources := NewBuilder().App("my-project", "my-app", zone,
		&cfg.Config{
			CDNPolicy: &compute.BackendServiceCdnPolicy{
				SignedUrlCacheMaxAgeSec: 200,
			},
			IAP: &compute.BackendServiceIAP{
				Enabled:            true,
				Oauth2ClientId:     "hello",
				Oauth2ClientSecret: "world",
			},
			IAMRoles: []string{
				"roles/dogWalker.dog",
			},
		},
	)

	actual, err := json.MarshalIndent(map[string]interface{}{
		"resources": resources,
	}, "", "  ")
	if err != nil {
		t.Fatal(err)
	}

	fixtures.Compare(t, "app.json", actual)
}
