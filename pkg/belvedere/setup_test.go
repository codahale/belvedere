package belvedere

import (
	"encoding/json"
	"testing"

	"github.com/codahale/belvedere/pkg/belvedere/internal/fixtures"
)

func TestSetupResources(t *testing.T) {
	resources := setupResources("cornbread.club")

	actual, err := json.MarshalIndent(map[string]interface{}{
		"resources": resources,
	}, "", "  ")
	if err != nil {
		t.Fatal(err)
	}

	fixtures.Compare(t, "setup_fixture.json", actual)
}
