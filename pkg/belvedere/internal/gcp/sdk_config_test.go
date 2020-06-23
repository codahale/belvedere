package gcp

import (
	"path/filepath"
	"testing"

	"github.com/codahale/belvedere/internal/assert"
)

func TestSDKConfig(t *testing.T) {
	defer func(f func() (string, error)) { sdkPath = f }(sdkPath)

	sdkPath = func() (string, error) {
		return filepath.Abs("./fixtures")
	}

	got, err := SDKConfig()
	if err != nil {
		t.Fatal(err)
	}

	want := map[string]map[string]string{
		"":      {"bare": "1"},
		"core":  {"project": "boop"},
		"other": {"yes": "no"},
	}

	assert.Equal(t, "SDKConfig()", want, got)
}
