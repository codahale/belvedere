package belvedere

import (
	"context"
	"fmt"
	"time"

	"github.com/GoogleCloudPlatform/konlet/gce-containers-startup/types"
	"github.com/codahale/belvedere/pkg/belvedere/internal/backends"
	"github.com/codahale/belvedere/pkg/belvedere/internal/check"
	"github.com/codahale/belvedere/pkg/belvedere/internal/deployments"
	"go.opencensus.io/trace"
	"google.golang.org/api/compute/v0.beta"
	"google.golang.org/api/deploymentmanager/v2"
	"gopkg.in/yaml.v2"
	"k8s.io/apimachinery/pkg/util/wait"
)

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

func CreateRelease(ctx context.Context, project, appName, relName string, config *Config, imageSHA256 string) error {
	ctx, span := trace.StartSpan(ctx, "belvedere.CreateRelease")
	span.AddAttributes(
		trace.StringAttribute("project", project),
		trace.StringAttribute("app", appName),
		trace.StringAttribute("release", relName),
		trace.StringAttribute("image_url", imageSHA256),
	)
	defer span.End()

	region, err := findRegion(ctx, project, appName)
	if err != nil {
		return err
	}

	instanceTemplate := fmt.Sprintf("%s-%s-it", appName, relName)
	instanceGroupManager := fmt.Sprintf("%s-%s-ig", appName, relName)
	autoscaler := fmt.Sprintf("%s-%s-as", appName, relName)
	dmConfig := &deployments.Config{
		Resources: []deployments.Resource{
			{
				Name: instanceTemplate,
				Type: "compute.beta.instanceTemplate",
				Properties: compute.InstanceTemplate{
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
						MachineType: config.MachineType,
						Metadata: &compute.Metadata{
							Items: []*compute.MetadataItems{
								metaData("disable-legacy-endpoints", "true"),
								metaData("enable-os-login", "true"),
								metaData("google-logging-enable", "true"),
								metaData(
									"gce-container-declaration",
									containerDeclaration(appName, relName, config, imageSHA256),
								),
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
						ServiceAccounts: []*compute.ServiceAccount{
							{
								Email: fmt.Sprintf("app-%s@%s.iam.gserviceaccount.com", appName, project),
								Scopes: []string{
									compute.CloudPlatformScope,
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
			{
				Name: instanceGroupManager,
				Type: "compute.beta.regionInstanceGroupManager",
				Properties: compute.InstanceGroupManager{
					BaseInstanceName: fmt.Sprintf("%s-%s", appName, relName),
					InstanceTemplate: deployments.SelfLink(instanceTemplate),
					Region:           region,
					NamedPorts: []*compute.NamedPort{
						{
							Name: "svc-https",
							Port: 8443,
						},
					},
					TargetSize: int64(config.InitialInstances),
				},
			},
			{
				Name: autoscaler,
				Type: "compute.beta.regionAutoscaler",
				Properties: compute.Autoscaler{
					Name: fmt.Sprintf("%s-%s", appName, relName),
					AutoscalingPolicy: &compute.AutoscalingPolicy{
						LoadBalancingUtilization: &compute.AutoscalingPolicyLoadBalancingUtilization{
							UtilizationTarget: config.UtilizationTarget,
						},
						MaxNumReplicas: int64(config.MaxInstances),
						MinNumReplicas: int64(config.MinInstances),
					},
					Region: region,
					Target: deployments.SelfLink(instanceGroupManager),
				},
			},
		},
	}

	name := fmt.Sprintf("belvedere-%s-%s", appName, relName)
	return deployments.Insert(ctx, project, name, dmConfig, map[string]string{
		"belvedere-type":    "release",
		"belvedere-app":     appName,
		"belvedere-release": relName,
		"belvedere-region":  region,
		"belvedere-hash":    imageSHA256[:32],
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

	region, err := findRegion(ctx, project, appName)
	if err != nil {
		return err
	}

	gce, err := compute.NewService(ctx)
	if err != nil {
		return err
	}

	backendService := fmt.Sprintf("%s-bes", appName)
	instanceGroup := fmt.Sprintf("%s-%s-ig", appName, relName)
	if err := backends.Add(ctx, gce, project, region, backendService, instanceGroup); err != nil {
		return err
	}

	f := check.Health(ctx, gce, project, region, backendService, instanceGroup)
	return wait.Poll(10*time.Second, 5*time.Minute, f)
}

func DisableRelease(ctx context.Context, project, appName, relName string) error {
	ctx, span := trace.StartSpan(ctx, "belvedere.DisableRelease")
	span.AddAttributes(
		trace.StringAttribute("project", project),
		trace.StringAttribute("app", appName),
		trace.StringAttribute("release", relName),
	)
	defer span.End()

	region, err := findRegion(ctx, project, appName)
	if err != nil {
		return err
	}

	gce, err := compute.NewService(ctx)
	if err != nil {
		return err
	}

	backendService := fmt.Sprintf("%s-bes", appName)
	instanceGroup := fmt.Sprintf("%s-%s-ig", appName, relName)
	return backends.Remove(ctx, gce, project, region, backendService, instanceGroup)
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

const (
	cosStable = "https://www.googleapis.com/compute/v1/projects/gce-uefi-images/global/images/family/cos-stable"
)

func metaData(key, value string) *compute.MetadataItems {
	return &compute.MetadataItems{
		Key:   key,
		Value: &value,
	}
}

func containerDeclaration(appName, relName string, config *Config, imageSHA256 string) string {
	var env []struct {
		Name  string
		Value string
	}

	for k, v := range config.Env {
		env = append(env, struct {
			Name  string
			Value string
		}{Name: k, Value: v})
	}

	template := types.ContainerSpec{
		Spec: types.ContainerSpecStruct{
			Containers: []types.Container{
				{
					Name:    fmt.Sprintf("%s-%s", appName, relName),
					Image:   fmt.Sprintf("%s@sha256:%s", config.ImageURL, imageSHA256),
					Command: []string{},
					Env:     env,
				},
			},
		},
	}
	y, _ := yaml.Marshal(template)
	return string(y)
}
