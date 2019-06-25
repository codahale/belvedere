package belvedere

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestSetupResources(t *testing.T) {
	resources := setupResources("cornbread.club")

	actual, err := json.MarshalIndent(map[string]interface{}{
		"resources": resources,
	}, "", "  ")
	if err != nil {
		t.Fatal(err)
	}

	//_ = ioutil.WriteFile("setup_fixture.json", actual, 0644)

	expected, err := ioutil.ReadFile("setup_fixture.json")
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(actual, expected) {
		t.Error(cmp.Diff(string(expected), string(actual)))
	}
}
