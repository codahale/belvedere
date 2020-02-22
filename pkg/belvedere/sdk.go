package belvedere

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"

	"go.opencensus.io/trace"
)

// ActiveProject returns the project, if any, which the Google Cloud SDK is configured to use.
func ActiveProject(ctx context.Context) (string, error) {
	_, span := trace.StartSpan(ctx, "belvedere.ActiveProject")
	defer span.End()

	// Find the SDK config path.
	sdkPath, err := sdkPath()
	if err != nil {
		return "", fmt.Errorf("error getting SDK config path: %w", err)
	}

	// Read the active config.
	activeConfig, err := ioutil.ReadFile(filepath.Join(sdkPath, "active_config"))
	if err != nil {
		return "", fmt.Errorf("error reading active config name: %w", err)
	}

	// Open the default config file.
	configPath := filepath.Join(sdkPath, "configurations", fmt.Sprintf("config_%s", activeConfig))
	f, err := os.Open(configPath)
	if err != nil {
		return "", fmt.Errorf("failed to load active config: %w", err)
	}
	defer func() { _ = f.Close() }()

	// Parse as an INI file.
	config, err := parseINI(f)
	if err != nil {
		return "", fmt.Errorf("failed to parse config: %w", err)
	}

	// Return core.project, if it exists.
	if core, ok := config["core"]; ok {
		if project, ok := core["project"]; ok {
			return project, nil
		}
	}

	// Complain if core.project doesn't exist.
	return "", fmt.Errorf("core.project not found")
}

func parseINI(ini io.Reader) (map[string]map[string]string, error) {
	result := map[string]map[string]string{
		"": {}, // root section
	}
	scanner := bufio.NewScanner(ini)
	currentSection := ""
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, ";") {
			// comment.
			continue
		}
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			currentSection = strings.TrimSpace(line[1 : len(line)-1])
			result[currentSection] = map[string]string{}
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 && parts[0] != "" {
			result[currentSection][strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error scanning ini: %v", err)
	}
	return result, nil
}

var sdkPath = func() (string, error) {
	if runtime.GOOS == "windows" {
		return filepath.Join(os.Getenv("APPDATA"), "gcloud"), nil
	}
	homeDir := guessUnixHomeDir()
	if homeDir == "" {
		return "", errors.New("unable to get current user home directory: os/user lookup failed; $HOME is empty")
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
