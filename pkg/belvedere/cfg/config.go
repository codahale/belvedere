// Package cfg provides common data structure for Belvedere app configuration.
package cfg

import (
	"fmt"
	"io"

	"github.com/ghodss/yaml"
	"google.golang.org/api/compute/v1"

	// Make the YAML lib a direct dependency so we can get dependabot updates for it.
	_ "gopkg.in/yaml.v2"
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
	WAFRules          []*compute.SecurityPolicyRule    `json:"wafRules"`
	SessionAffinity   string                           `json:"sessionAffinity"`
}

type InvalidSessionAffinityError struct {
	Value string
}

func (e *InvalidSessionAffinityError) Error() string {
	return fmt.Sprintf("invalid session affinity: %s", e.Value)
}

const (
	SessionAffinityNone   = "none"
	SessionAffinityIP     = "ip"
	SessionAffinityCookie = "cookie"
)

// Parse loads the given bytes as a YAML configuration.
func Parse(r io.Reader) (*Config, error) {
	// Read the configuration.
	b, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("error reading config: %w", err)
	}

	// Unmarshal from YAML using the YAML->JSON route. This allows us to embed GCP API structs in
	// our Config struct.
	var config Config
	if err := yaml.Unmarshal(b, &config); err != nil {
		return nil, fmt.Errorf("error parsing config: %w", err)
	}

	// Validate properties.
	switch config.SessionAffinity {
	case "", SessionAffinityCookie, SessionAffinityIP, SessionAffinityNone:
	default:
		return nil, &InvalidSessionAffinityError{Value: config.SessionAffinity}
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
