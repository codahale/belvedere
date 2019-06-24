package belvedere

import (
	"context"
	"io"
	"io/ioutil"
	"os"

	"go.opencensus.io/trace"
	"gopkg.in/yaml.v2"
)

type Config struct {
	IAMRoles          []string             `yaml:"iam_roles,omitempty"`
	InitialInstances  int                  `yaml:"initialInstances"`
	MachineType       string               `yaml:"machineType"`
	MaxInstances      int                  `yaml:"maxInstances"`
	MinInstances      int                  `yaml:"minInstances"`
	UtilizationTarget float64              `yaml:"utilizationTarget"`
	Container         Container            `yaml:"container"`
	Sidecars          map[string]Container `yaml:"sidecars"`
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
