package gcp

import (
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestSDKConfig(t *testing.T) {
	oldSdkPath := sdkPath
	sdkPath = func() (string, error) {
		return filepath.Abs("./fixtures")
	}
	defer func() { sdkPath = oldSdkPath }()

	actual, err := SDKConfig()
	if err != nil {
		t.Fatal(err)
	}

	expected := map[string]map[string]string{
		"":      {"bare": "1"},
		"core":  {"project": "boop"},
		"other": {"yes": "no"},
	}

	if !cmp.Equal(expected, actual) {
		t.Fatal(cmp.Diff(expected, actual))
	}
}
