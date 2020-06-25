package gcp

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/mitchellh/go-homedir"
	"gopkg.in/ini.v1"
)

// DefaultProject returns the project that the Google Cloud SDK is configured to use.
func DefaultProject(sdkPathFunc func() (string, error)) (string, error) {
	// Find the SDK config path.
	sdkPath, err := sdkPathFunc()
	if err != nil {
		return "", fmt.Errorf("error getting SDK config path: %w", err)
	}

	// Read the active config.
	configName, err := ioutil.ReadFile(filepath.Join(sdkPath, "active_config"))
	if err != nil {
		return "", fmt.Errorf("error reading active config name: %w", err)
	}

	// Find the default config file.
	configPath := filepath.Join(sdkPath, "configurations",
		fmt.Sprintf("config_%s", strings.TrimSpace(string(configName))))

	// Read and parse it.
	cfg, err := ini.Load(configPath)
	if err != nil {
		return "", fmt.Errorf("failed to load active config: %w", err)
	}

	// Find core.project, if any.
	key, err := cfg.Section("core").GetKey("project")
	if err != nil {
		return "", fmt.Errorf("error reading core.project: %w", err)
	}

	// Return core.project, if it exists.
	return key.String(), nil
}

func SDKPath() (string, error) {
	if runtime.GOOS == "windows" {
		return filepath.Join(os.Getenv("APPDATA"), "gcloud"), nil
	}

	home, err := homedir.Dir()
	if err != nil {
		return "", fmt.Errorf("unable to get SDK config: %w", err)
	}

	return filepath.Join(home, ".config", "gcloud"), nil
}
