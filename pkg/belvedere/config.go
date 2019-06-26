package belvedere

import (
	"context"
	"io"
	"io/ioutil"
	"os"

	"github.com/ghodss/yaml"
	"go.opencensus.io/trace"
	compute "google.golang.org/api/compute/v0.beta"
)

type Config struct {
	IAMRoles          []string                   `json:"iamRoles,omitempty"`
	InitialInstances  int                        `json:"initialInstances"`
	MachineType       string                     `json:"machineType"`
	MaxInstances      int                        `json:"maxInstances"`
	MinInstances      int                        `json:"minInstances"`
	UtilizationTarget float64                    `json:"utilizationTarget"`
	Container         Container                  `json:"container"`
	Sidecars          map[string]Container       `json:"sidecars"`
	IAP               *compute.BackendServiceIAP `json:"identityAwareProxy"`
}

// LoadConfig loads the YAML configuration at the given path. If path is `-`, STDIN is used.
func LoadConfig(ctx context.Context, path string) (*Config, error) {
	ctx, span := trace.StartSpan(ctx, "belvedere.LoadConfig")
	span.AddAttributes(
		trace.StringAttribute("path", path),
	)
	defer span.End()

	var r io.ReadCloser
	if path == "-" {
		r = os.Stdin
	} else {
		f, err := os.Open(path)
		if err != nil {
			return nil, err
		}

		r = f
	}
	defer func() { _ = r.Close() }()

	b, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := yaml.Unmarshal(b, &config); err != nil {
		return nil, err
	}
	return &config, nil
}
