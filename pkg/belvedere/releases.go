package belvedere

import (
	"context"
	"fmt"
	"io/ioutil"

	"github.com/codahale/belvedere/pkg/belvedere/internal/deployments"
	"go.opencensus.io/trace"
	"google.golang.org/api/compute/v0.beta"
	"google.golang.org/api/deploymentmanager/v2"
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

func ListReleases(ctx context.Context, project, appName string) ([]string, error) {
	ctx, span := trace.StartSpan(ctx, "belvedere.ListReleases")
	span.AddAttributes(
		trace.StringAttribute("project", project),
		trace.StringAttribute("app", appName),
	)
	defer span.End()

	dm, err := deploymentmanager.NewService(ctx)
	if err != nil {
		return nil, err
	}

	resp, err := dm.Deployments.List(project).Do()
	if err != nil {
		return nil, err
	}

	var names []string
	for _, d := range resp.Deployments {
		var app bool
		var name string
		for _, l := range d.Labels {
			if l.Key == "belvedere-release" {
				name = l.Value
			} else if l.Key == "belvedere-type" && l.Value == "release" {
				app = true
			}
		}
		if app {
			names = append(names, name)
		}
	}
	return names, nil
}

const (
	cosStable = "https://www.googleapis.com/compute/v1/projects/gce-uefi-images/global/images/family/cos-stable"
)

func metaData(key, value string) *compute.MetadataItems {
	return &compute.MetadataItems{
		Key:   key,
		Value: &value,
	}
}

func CreateRelease(ctx context.Context, project, appName, relName string, release *ReleaseConfig, imageURL string) error {
	ctx, span := trace.StartSpan(ctx, "belvedere.CreateRelease")
	span.AddAttributes(
		trace.StringAttribute("project", project),
		trace.StringAttribute("app", appName),
		trace.StringAttribute("release", relName),
		trace.StringAttribute("image_url", imageURL),
	)
	defer span.End()

	config := &deployments.Config{
		Resources: []deployments.Resource{
			{
				Name: "instance-template",
				Type: "compute.beta.instanceTemplate",
				Properties: compute.InstanceTemplate{
					Name: fmt.Sprintf("belvedere-%s-%s", appName, relName),
					Properties: &compute.InstanceProperties{
						Disks: []*compute.AttachedDisk{
							{
								AutoDelete: true,
								Boot:       true,
								DeviceName: "boot",
								Type:       "PERSISTENT",
								InitializeParams: &compute.AttachedDiskInitializeParams{
									DiskName:    "pd-standard",
									SourceImage: cosStable,
								},
							},
						},
						Labels: map[string]string{
							"belvedere-app":     appName,
							"belvedere-release": relName,
						},
						MachineType: release.MachineType,
						Metadata: &compute.Metadata{
							Items: []*compute.MetadataItems{
								metaData("disable-legacy-endpoints", "true"),
								metaData("enable-os-login", "true"),
								metaData("google-logging-enable", "true"),
								// TODO inject cloud-init script
							},
						},
						NetworkInterfaces: []*compute.NetworkInterface{
							{
								Network: "global/networks/default",
								AccessConfigs: []*compute.AccessConfig{
									{
										Name: "External NAT",
										Type: "ONE_TO_ONE_NAT",
									},
								},
							},
						},
						ShieldedVmConfig: &compute.ShieldedVmConfig{
							EnableIntegrityMonitoring: true,
							EnableSecureBoot:          true,
							EnableVtpm:                true,
						},
						Tags: &compute.Tags{
							Items: []string{
								"belvedere",
								fmt.Sprintf("belvedere-%s", appName),
							},
						},
					},
				},
			},
			// TODO region autoscaler
			// TODO region instance group manager
		},
	}

	name := fmt.Sprintf("belvedere-%s-%s", appName, relName)
	return deployments.Insert(ctx, project, name, config, map[string]string{
		"belvedere-type":      "release",
		"belvedere-app":       appName,
		"belvedere-release":   relName,
		"belvedere-image-url": imageURL,
	})
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

	return deployments.Delete(ctx, project, fmt.Sprintf("belvedere-%s-%s", appName, relName))
}
