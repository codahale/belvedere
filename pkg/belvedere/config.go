package belvedere

import (
	"context"
	"fmt"

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
