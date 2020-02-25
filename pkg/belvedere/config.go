package belvedere

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/ghodss/yaml"
	"go.opencensus.io/trace"
	compute "google.golang.org/api/compute/v0.beta"
)

// Config contains all the mutable parameters of an app's configuration.
type Config struct {
	IAMRoles          []string                         `json:"iamRoles,omitempty"`
	NumReplicas       int                              `json:"numReplicas"`
	MachineType       string                           `json:"machineType"`
	Container         Container                        `json:"container"`
	Sidecars          map[string]Container             `json:"sidecars"`
	IAP               *compute.BackendServiceIAP       `json:"identityAwareProxy"`
	AutoscalingPolicy *compute.AutoscalingPolicy       `json:"autoscalingPolicy"`
	CDNPolicy         *compute.BackendServiceCdnPolicy `json:"cdnPolicy"`
	Network           string                           `json:"network"`
	Subnetwork        string                           `json:"subnetwork"`
}

// A Container describes all the elements of an app or sidecar container.
type Container struct {
	Image         string            `json:"image"`
	Command       string            `json:"command"`
	Args          []string          `json:"args"`
	Env           map[string]string `json:"env"`
	DockerOptions []string          `json:"dockerOptions"`
}

// dockerArgs returns a list of arguments to `docker run` for running the given container.
func (c *Container) dockerArgs(app, release, sha256 string, labels map[string]string) []string {
	labelNames := make([]string, 0, len(labels))
	for k := range labels {
		labelNames = append(labelNames, k)
	}
	sort.Stable(sort.StringSlice(labelNames))

	args := []string{
		"--log-driver", "gcplogs",
		"--log-opt", fmt.Sprintf("labels=%s", strings.Join(labelNames, ",")),
		"--name", app,
		"--network", "host",
		"--oom-kill-disable",
	}

	for _, k := range labelNames {
		args = append(args, []string{
			"--label", fmt.Sprintf("%s=%s", k, labels[k]),
		}...)
	}

	if release != "" {
		args = append(args, []string{
			"--env", fmt.Sprintf("RELEASE=%s", release),
		}...)
	}

	envNames := make([]string, 0, len(c.Env))
	for k := range c.Env {
		envNames = append(envNames, k)
	}
	sort.Stable(sort.StringSlice(envNames))

	for _, k := range envNames {
		args = append(args, []string{
			"--env", fmt.Sprintf("%s=%s", k, c.Env[k]),
		}...)
	}

	args = append(args, c.DockerOptions...)
	url := c.Image
	if sha256 != "" {
		url = fmt.Sprintf("%s@sha256:%s", url, sha256)
	}
	args = append(args, url)
	if c.Command != "" {
		args = append(args, c.Command)
	}
	args = append(args, c.Args...)

	return args
}

// LoadConfig loads the given bytes as a YAML configuration.
func LoadConfig(ctx context.Context, b []byte) (*Config, error) {
	_, span := trace.StartSpan(ctx, "belvedere.LoadConfig")
	defer span.End()

	// Unmarshal from YAML using the YAML->JSON route. This allows us to embed GCP API structs in
	// our Config struct.
	var config Config
	if err := yaml.Unmarshal(b, &config); err != nil {
		return nil, fmt.Errorf("error parsing config: %w", err)
	}
	return &config, nil
}
