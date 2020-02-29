package resources

import (
	"encoding/json"
	"testing"

	"github.com/codahale/belvedere/pkg/belvedere/internal/fixtures"
)

func TestBaseResources(t *testing.T) {
	resources := NewBuilder().Base("cornbread.club")

	actual, err := json.MarshalIndent(map[string]interface{}{
		"resources": resources,
	}, "", "  ")
	if err != nil {
		t.Fatal(err)
	}

	fixtures.Compare(t, "base.json", actual)
}
