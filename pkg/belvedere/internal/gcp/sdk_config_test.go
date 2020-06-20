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

	got, err := SDKConfig()
	if err != nil {
		t.Fatal(err)
	}

	want := map[string]map[string]string{
		"":      {"bare": "1"},
		"core":  {"project": "boop"},
		"other": {"yes": "no"},
	}

	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("SDKConfig() mismatch (-want +got):\n%s", diff)
	}
}
