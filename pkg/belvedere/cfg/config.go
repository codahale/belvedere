// Package cfg provides common data structure for Belvedere app configuration.
package cfg

import (
	"fmt"

	"github.com/ghodss/yaml"
	compute "google.golang.org/api/compute/v1"
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

// Parse loads the given bytes as a YAML configuration.
func Parse(b []byte) (*Config, error) {
	// Unmarshal from YAML using the YAML->JSON route. This allows us to embed GCP API structs in
	// our Config struct.
	var config Config
	if err := yaml.Unmarshal(b, &config); err != nil {
		return nil, fmt.Errorf("error parsing config: %w", err)
	}
	return &config, nil
}

// A Container describes all the elements of an app or sidecar container.
type Container struct {
	Image         string            `json:"image"`
	Command       string            `json:"command"`
	Args          []string          `json:"args"`
	Env           map[string]string `json:"env"`
	DockerOptions []string          `json:"dockerOptions"`
}
