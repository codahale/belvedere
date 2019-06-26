package belvedere

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"testing"

	"github.com/google/go-cmp/cmp"
	"google.golang.org/api/dns/v1"
)

func TestAppResources(t *testing.T) {
	zone := &dns.ManagedZone{
		Name:    "belvedere",
		DnsName: "horse.club",
	}
	config := &Config{
		IAMRoles: []string{
			"roles/dogWalker.dog",
		},
		IAP: IAPConfig{
			Enabled:            true,
			OAuth2ClientID:     "hello",
			OAuth2ClientSecret: "world",
		},
	}
	resources := appResources("my-project", "my-app", zone, config)

	actual, err := json.MarshalIndent(map[string]interface{}{
		"resources": resources,
	}, "", "  ")
	if err != nil {
		t.Fatal(err)
	}

	//_ = ioutil.WriteFile("app_fixture.json", actual, 0644)

	expected, err := ioutil.ReadFile("app_fixture.json")
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(expected, actual) {
		t.Error(cmp.Diff(string(expected), string(actual)))
	}
}
