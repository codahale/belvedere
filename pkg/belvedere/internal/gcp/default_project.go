package gcp

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"

	"gopkg.in/ini.v1"
)

// DefaultProject returns the project that the Google Cloud SDK is configured to use.
func DefaultProject() (string, error) {
	// Find the SDK config path.
	sdkPath, err := sdkPath()
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

var errNoHomeDir = errors.New("unable to get current user home directory: os/user lookup failed; $HOME is empty")

//nolint:gochecknoglobals // keep this mutable for testing
var sdkPath = func() (string, error) {
	if runtime.GOOS == "windows" {
		return filepath.Join(os.Getenv("APPDATA"), "gcloud"), nil
	}
	homeDir := guessUnixHomeDir()
	if homeDir == "" {
		return "", errNoHomeDir
	}
	return filepath.Join(homeDir, ".config", "gcloud"), nil
}

func guessUnixHomeDir() string {
	// Prefer $HOME over user.Current due to glibc bug: golang.org/issue/13470
	if v := os.Getenv("HOME"); v != "" {
		return v
	}

	// Else, fall back to user.Current:
	if u, err := user.Current(); err == nil {
		return u.HomeDir
	}

	return ""
}
