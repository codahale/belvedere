package resources

import (
	"encoding/json"
	"testing"

	"github.com/codahale/gubbins/assert"
)

func TestBaseResources(t *testing.T) {
	t.Parallel()

	resources := NewBuilder().Base("cornbread.club")

	got, err := json.MarshalIndent(map[string]interface{}{
		"resources": resources,
	}, "", "  ")
	if err != nil {
		t.Fatal(err)
	}

	assert.EqualFixture(t, "Base()", "base.json", got)
}
