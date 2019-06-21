package belvedere

import (
	"context"
	"errors"
	"io"
	"io/ioutil"
	"os"

	"go.opencensus.io/trace"
	"gopkg.in/yaml.v2"
)

type ReleaseConfig struct {
	MachineType      string            `yaml:"machineType"`
	MinInstances     int               `yaml:"minInstances"`
	MaxInstances     int               `yaml:"maxInstances"`
	InitialInstances int               `yaml:"initialInstances"`
	TargetCapacity   float64           `yaml:"targetCapacity"`
	Env              map[string]string `yaml:"env"`
}

func LoadReleaseConfig(ctx context.Context, path string) (*ReleaseConfig, error) {
	ctx, span := trace.StartSpan(ctx, "belvedere.LoadReleaseConfig")
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

	var config ReleaseConfig
	if err := yaml.Unmarshal(b, &config); err != nil {
		return nil, err
	}
	return &config, nil
}

func ListReleases(ctx context.Context, project, appName string) error {
	ctx, span := trace.StartSpan(ctx, "belvedere.ListReleases")
	span.AddAttributes(
		trace.StringAttribute("project", project),
		trace.StringAttribute("app", appName),
	)
	defer span.End()

	// TODO list deployments with filter `labels.type eq belvedere-release`
	// TODO return list of releases

	return errUnimplemented
}

func CreateRelease(ctx context.Context, project, appName, relName string, config *ReleaseConfig, image string) error {
	ctx, span := trace.StartSpan(ctx, "belvedere.CreateRelease")
	span.AddAttributes(
		trace.StringAttribute("project", project),
		trace.StringAttribute("app", appName),
		trace.StringAttribute("release", relName),
		trace.StringAttribute("image", image),
	)
	defer span.End()

	// TODO create deployment w/ template, group manager, and autoscaler

	return errUnimplemented
}

func EnableRelease(ctx context.Context, project, appName, relName string) error {
	ctx, span := trace.StartSpan(ctx, "belvedere.EnableRelease")
	span.AddAttributes(
		trace.StringAttribute("project", project),
		trace.StringAttribute("app", appName),
		trace.StringAttribute("release", relName),
	)
	defer span.End()

	// TODO patch backend service w/ release IGM

	return errUnimplemented
}

func DisableRelease(ctx context.Context, project, appName, relName string) error {
	ctx, span := trace.StartSpan(ctx, "belvedere.DisableRelease")
	span.AddAttributes(
		trace.StringAttribute("project", project),
		trace.StringAttribute("app", appName),
		trace.StringAttribute("release", relName),
	)
	defer span.End()

	// TODO patch backend service w/o release IGM

	return errUnimplemented
}

func DestroyRelease(ctx context.Context, project, appName, relName string) error {
	ctx, span := trace.StartSpan(ctx, "belvedere.DestroyRelease")
	span.AddAttributes(
		trace.StringAttribute("project", project),
		trace.StringAttribute("app", appName),
		trace.StringAttribute("release", relName),
	)
	defer span.End()

	// TODO delete release deployment

	return errUnimplemented
}

func openPath(path string) (io.ReadCloser, error) {
	if path == "-" {
		return os.Stdin, nil
	}
	return os.Open(path)
}

var errUnimplemented = errors.New("unimplemented")
