package resources

import (
	"encoding/json"
	"testing"

	"github.com/codahale/belvedere/internal/assert"
)

func TestBaseResources(t *testing.T) {
	resources := NewBuilder().Base("cornbread.club")

	got, err := json.MarshalIndent(map[string]interface{}{
		"resources": resources,
	}, "", "  ")
	if err != nil {
		t.Fatal(err)
	}

	assert.EqualFixture(t, "Base()", "base.json", got)
}
