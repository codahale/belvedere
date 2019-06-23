package belvedere

import (
	"context"
	"io/ioutil"

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
	Env               map[string]string    `yaml:"env"`
	ImageURL          string               `yaml:"imageURL"`
	Container         Container            `yaml:"container"`
	Sidecars          map[string]Container `yaml:"sidecars"`
}

func LoadConfig(ctx context.Context, path string) (*Config, error) {
	ctx, span := trace.StartSpan(ctx, "belvedere.LoadConfig")
	span.AddAttributes(
		trace.StringAttribute("path", path),
	)
	defer span.End()

	r, err := openPath(path)
	if err != nil {
		return nil, err
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
