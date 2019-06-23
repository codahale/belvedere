package belvedere

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/codahale/belvedere/pkg/belvedere/internal/backends"
	"github.com/codahale/belvedere/pkg/belvedere/internal/check"
	"github.com/codahale/belvedere/pkg/belvedere/internal/cloudinit"
	"github.com/codahale/belvedere/pkg/belvedere/internal/deployments"
	"go.opencensus.io/trace"
	"google.golang.org/api/compute/v0.beta"
	"google.golang.org/api/deploymentmanager/v2"
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

func CreateRelease(ctx context.Context, project, appName, relName string, config *Config, imageSHA256 string, dryRun bool) error {
	ctx, span := trace.StartSpan(ctx, "belvedere.CreateRelease")
	span.AddAttributes(
		trace.StringAttribute("project", project),
		trace.StringAttribute("app", appName),
		trace.StringAttribute("release", relName),
		trace.StringAttribute("image_url", imageSHA256),
		trace.BoolAttribute("dry_run", dryRun),
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
									"user-data",
									cloudConfig(appName, relName, config, imageSHA256),
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
	}, dryRun)
}

func EnableRelease(ctx context.Context, project, appName, relName string, dryRun bool) error {
	ctx, span := trace.StartSpan(ctx, "belvedere.EnableRelease")
	span.AddAttributes(
		trace.StringAttribute("project", project),
		trace.StringAttribute("app", appName),
		trace.StringAttribute("release", relName),
		trace.BoolAttribute("dry_run", dryRun),
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
	if err := backends.Add(ctx, gce, project, region, backendService, instanceGroup, dryRun); err != nil {
		return err
	}

	f := check.Health(ctx, gce, project, region, backendService, instanceGroup)
	return wait.Poll(10*time.Second, 5*time.Minute, f)
}

func DisableRelease(ctx context.Context, project, appName, relName string, dryRun bool) error {
	ctx, span := trace.StartSpan(ctx, "belvedere.DisableRelease")
	span.AddAttributes(
		trace.StringAttribute("project", project),
		trace.StringAttribute("app", appName),
		trace.StringAttribute("release", relName),
		trace.BoolAttribute("dry_run", dryRun),
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
	return backends.Remove(ctx, gce, project, region, backendService, instanceGroup, dryRun)
}

func DestroyRelease(ctx context.Context, project, appName, relName string, dryRun bool) error {
	ctx, span := trace.StartSpan(ctx, "belvedere.DestroyRelease")
	span.AddAttributes(
		trace.StringAttribute("project", project),
		trace.StringAttribute("app", appName),
		trace.StringAttribute("release", relName),
		trace.BoolAttribute("dry_run", dryRun),
	)
	defer span.End()

	return deployments.Delete(ctx, project, fmt.Sprintf("belvedere-%s-%s", appName, relName), dryRun)
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

func cloudConfig(appName, relName string, config *Config, imageSHA256 string) string {
	cc := cloudinit.CloudConfig{
		Files: []cloudinit.File{
			{
				Content: systemdService(appName,
					config.Container.DockerArgs(appName, relName, imageSHA256,
						map[string]string{
							"app":     appName,
							"release": relName,
						}),
				),
				Owner:       "root",
				Path:        fmt.Sprintf("/etc/systemd/system/docker-%s.service", appName),
				Permissions: "0644",
			},
		},
		Commands: []string{
			"systemctl daemon-reload",
			fmt.Sprintf("systemctl start docker-%s.service", appName),
		},
	}

	for name, sidecar := range config.Sidecars {
		cc.Files = append(cc.Files,
			cloudinit.File{
				Content: systemdService(name,
					sidecar.DockerArgs(name, "", "",
						map[string]string{
							"app":     appName,
							"release": relName,
							"sidecar": name,
						}),
				),
				Owner:       "root",
				Path:        fmt.Sprintf("/etc/systemd/system/docker-%s.service", name),
				Permissions: "0644",
			})
		cc.Commands = append(cc.Commands, fmt.Sprintf("systemctl start docker-%s.service", name))
	}

	return cc.String()
}

const jobTemplate = `[Unit]
Description=Start the %s container
Wants=gcr-online.target
After=gcr-online.target

[Service]
Environment="HOME=/var/lib/docker"
ExecStartPre=/usr/bin/docker-credential-gcr configure-docker
ExecStart=/usr/bin/docker run --rm %s
ExecStop=/usr/bin/docker stop %s
ExecStopPost=/usr/bin/docker rm %s
`

func systemdService(name string, dockerArgs []string) string {
	return fmt.Sprintf(jobTemplate, name, strings.Join(dockerArgs, " "), name, name)
}
