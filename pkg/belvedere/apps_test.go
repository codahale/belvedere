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
	resources := appResources("my-project", "my-app", zone,
		&compute.BackendServiceCdnPolicy{
			SignedUrlCacheMaxAgeSec: 200,
		}, &compute.BackendServiceIAP{
			Enabled:            true,
			Oauth2ClientId:     "hello",
			Oauth2ClientSecret: "world",
		}, []string{
			"roles/dogWalker.dog",
		},
	)

	actual, err := json.MarshalIndent(map[string]interface{}{
		"resources": resources,
	}, "", "  ")
	if err != nil {
		t.Fatal(err)
	}

	fixtures.Compare(t, "app_fixture.json", actual)
}
