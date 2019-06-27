package belvedere

import (
	"context"
	"io"
	"io/ioutil"
	"os"

	"github.com/ghodss/yaml"
	"go.opencensus.io/trace"
	"golang.org/x/xerrors"
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

// LoadConfig loads the YAML configuration at the given path. If path is `-`, STDIN is used.
func LoadConfig(ctx context.Context, name string) (*Config, error) {
	ctx, span := trace.StartSpan(ctx, "belvedere.LoadConfig")
	span.AddAttributes(
		trace.StringAttribute("name", name),
	)
	defer span.End()

	// Either open the file or use STDIN.
	var r io.ReadCloser
	if name == "-" {
		r = os.Stdin
	} else {
		f, err := os.Open(name)
		if err != nil {
			return nil, xerrors.Errorf("error opening %s: %w", name, err)
		}

		r = f
	}
	defer func() { _ = r.Close() }()

	// Read the entire input.
	b, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, xerrors.Errorf("error reading from %s: %w", name, err)
	}

	// Unmarshal from YAML using the YAML->JSON route. This allows us to embed GCP API structs in
	// our Config struct.
	var config Config
	if err := yaml.Unmarshal(b, &config); err != nil {
		return nil, xerrors.Errorf("error parsing %s: %w", err)
	}
	return &config, nil
}
