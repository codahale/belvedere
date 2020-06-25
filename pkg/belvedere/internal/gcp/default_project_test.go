package gcp

import (
	"path/filepath"
	"testing"

	"github.com/codahale/belvedere/internal/assert"
)

func TestDefaultProject(t *testing.T) {
	got, err := DefaultProject(func() (string, error) {
		return filepath.Abs("./fixtures")
	})
	if err != nil {
		t.Fatal(err)
	}

	want := "boop"

	assert.Equal(t, "DefaultProject()", want, got)
}
