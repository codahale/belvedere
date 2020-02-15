package belvedere

import (
	"encoding/json"
	"testing"

	"github.com/codahale/belvedere/pkg/belvedere/internal/fixtures"
	compute "google.golang.org/api/compute/v0.beta"
	"google.golang.org/api/dns/v1"
)

func TestAppResources(t *testing.T) {
	zone := &dns.ManagedZone{
		Name:    "belvedere",
		DnsName: "horse.club",
	}
	config := &Config{
		CDNPolicy: &compute.BackendServiceCdnPolicy{
			SignedUrlCacheMaxAgeSec: 200,
		},
		IAMRoles: []string{
			"roles/dogWalker.dog",
		},
		IAP: &compute.BackendServiceIAP{
			Enabled:            true,
			Oauth2ClientId:     "hello",
			Oauth2ClientSecret: "world",
		},
	}
	resources := appResources("my-project", "my-app", zone, config)

	actual, err := json.MarshalIndent(map[string]interface{}{
		"resources": resources,
	}, "", "  ")
	if err != nil {
		t.Fatal(err)
	}

	fixtures.Compare(t, "app_fixture.json", actual)
}
