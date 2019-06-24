package belvedere

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"testing"

	"github.com/sergi/go-diff/diffmatchpatch"
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
	}
	resources := appResources("my-project", "my-app", zone, config)

	have, err := json.MarshalIndent(map[string]interface{}{
		"resources": resources,
	}, "", "  ")
	if err != nil {
		t.Fatal(err)
	}

	//_ = ioutil.WriteFile("app_fixture.json", have, 0644)

	want, err := ioutil.ReadFile("app_fixture.json")
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(have, want) {
		dmp := diffmatchpatch.New()
		diffs := dmp.DiffMain(string(want), string(have), false)
		if len(diffs) > 0 {
			t.Fatal(dmp.DiffPrettyText(diffs))
		}
	}
}
