package gcp

import (
	"path/filepath"
	"testing"

	"github.com/codahale/belvedere/internal/assert"
)

func TestDefaultProject(t *testing.T) {
	defer func(f func() (string, error)) { sdkPath = f }(sdkPath)

	sdkPath = func() (string, error) {
		return filepath.Abs("./fixtures")
	}

	got, err := DefaultProject()
	if err != nil {
		t.Fatal(err)
	}

	want := "boop"

	assert.Equal(t, "DefaultProject()", want, got)
}
