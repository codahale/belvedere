package belvedere

import (
	"context"
	"fmt"
	"strings"

	"github.com/codahale/belvedere/pkg/belvedere/internal/backends"
	"github.com/codahale/belvedere/pkg/belvedere/internal/check"
	"github.com/codahale/belvedere/pkg/belvedere/internal/cloudinit"
	"github.com/codahale/belvedere/pkg/belvedere/internal/deployments"
	"github.com/codahale/belvedere/pkg/belvedere/internal/waiter"
	"go.opencensus.io/trace"
	"google.golang.org/api/compute/v0.beta"
	"gopkg.in/alessio/shellescape.v1"
)

type Release struct {
	Project string
	Region  string
	App     string
	Release string
	Hash    string
}

func ListReleases(ctx context.Context, project, app string) ([]Release, error) {
	ctx, span := trace.StartSpan(ctx, "belvedere.ListReleases")
	span.AddAttributes(
		trace.StringAttribute("project", project),
	)
	defer span.End()

	if app != "" {
		span.AddAttributes(trace.StringAttribute("app", app))
	}

	list, err := deployments.List(ctx, project)
	if err != nil {
		return nil, err
	}

	var releases []Release
	for _, labels := range list {
		if (labels["belvedere-type"] == "release") && (app == "" || labels["belvedere-app"] == app) {
			releases = append(releases, Release{
				Project: project,
				Region:  labels["belvedere-region"],
				App:     labels["belvedere-app"],
				Release: labels["belvedere-release"],
				Hash:    labels["belvedere-hash"],
			})
		}
	}
	return releases, nil
}

func CreateRelease(ctx context.Context, project, app, release string, config *Config, imageSHA256 string, dryRun bool) error {
	ctx, span := trace.StartSpan(ctx, "belvedere.CreateRelease")
	span.AddAttributes(
		trace.StringAttribute("project", project),
		trace.StringAttribute("app", app),
		trace.StringAttribute("release", release),
		trace.StringAttribute("image_url", imageSHA256),
		trace.BoolAttribute("dry_run", dryRun),
	)
	defer span.End()

	if err := validateRFC1035(release); err != nil {
		return err
	}

	region, err := findRegion(ctx, project, app)
	if err != nil {
		return err
	}

	resources := releaseResources(project, region, app, release, imageSHA256, config)

	name := fmt.Sprintf("belvedere-%s-%s", app, release)
	return deployments.Create(ctx, project, name, resources, map[string]string{
		"belvedere-type":    "release",
		"belvedere-app":     app,
		"belvedere-release": release,
		"belvedere-region":  region,
		"belvedere-hash":    imageSHA256[:32],
	}, dryRun)
}

func EnableRelease(ctx context.Context, project, app, release string, dryRun bool) error {
	ctx, span := trace.StartSpan(ctx, "belvedere.EnableRelease")
	span.AddAttributes(
		trace.StringAttribute("project", project),
		trace.StringAttribute("app", app),
		trace.StringAttribute("release", release),
		trace.BoolAttribute("dry_run", dryRun),
	)
	defer span.End()

	region, err := findRegion(ctx, project, app)
	if err != nil {
		return err
	}

	backendService := fmt.Sprintf("%s-bes", app)
	instanceGroup := fmt.Sprintf("%s-%s-ig", app, release)
	if err := backends.Add(ctx, project, region, backendService, instanceGroup, dryRun); err != nil {
		return err
	}

	return waiter.Poll(ctx, check.Health(ctx, project, region, backendService, instanceGroup))
}

func DisableRelease(ctx context.Context, project, app, release string, dryRun bool) error {
	ctx, span := trace.StartSpan(ctx, "belvedere.DisableRelease")
	span.AddAttributes(
		trace.StringAttribute("project", project),
		trace.StringAttribute("app", app),
		trace.StringAttribute("release", release),
		trace.BoolAttribute("dry_run", dryRun),
	)
	defer span.End()

	region, err := findRegion(ctx, project, app)
	if err != nil {
		return err
	}

	backendService := fmt.Sprintf("%s-bes", app)
	instanceGroup := fmt.Sprintf("%s-%s-ig", app, release)
	return backends.Remove(ctx, project, region, backendService, instanceGroup, dryRun)
}

func DeleteRelease(ctx context.Context, project, app, release string, dryRun, async bool) error {
	ctx, span := trace.StartSpan(ctx, "belvedere.DeleteRelease")
	span.AddAttributes(
		trace.StringAttribute("project", project),
		trace.StringAttribute("app", app),
		trace.StringAttribute("release", release),
		trace.BoolAttribute("dry_run", dryRun),
		trace.BoolAttribute("async", async),
	)
	defer span.End()

	return deployments.Delete(ctx, project, fmt.Sprintf("belvedere-%s-%s", app, release), dryRun, async)
}

func releaseResources(project string, region string, app string, release string, imageSHA256 string, config *Config) []deployments.Resource {
	instanceTemplate := fmt.Sprintf("%s-%s-it", app, release)
	instanceGroupManager := fmt.Sprintf("%s-%s-ig", app, release)
	autoscaler := fmt.Sprintf("%s-%s-as", app, release)
	resources := []deployments.Resource{
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
						"belvedere-app":     app,
						"belvedere-release": release,
					},
					MachineType: config.MachineType,
					Metadata: &compute.Metadata{
						Items: []*compute.MetadataItems{
							metaData("disable-legacy-endpoints", "true"),
							metaData("enable-os-login", "true"),
							metaData("google-logging-enable", "true"),
							metaData(
								"user-data",
								cloudConfig(app, release, config, imageSHA256),
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
							Email: fmt.Sprintf("app-%s@%s.iam.gserviceaccount.com", app, project),
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
							fmt.Sprintf("belvedere-%s", app),
						},
					},
				},
			},
		},
		{
			Name: instanceGroupManager,
			Type: "compute.beta.regionInstanceGroupManager",
			Properties: compute.InstanceGroupManager{
				BaseInstanceName: fmt.Sprintf("%s-%s", app, release),
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
				Name: fmt.Sprintf("%s-%s", app, release),
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
	}
	return resources
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

func cloudConfig(app, release string, config *Config, imageSHA256 string) string {
	cc := cloudinit.CloudConfig{
		WriteFiles: []cloudinit.File{
			{
				Content: systemdService(app,
					config.Container.DockerArgs(app, release, imageSHA256,
						map[string]string{
							"app":     app,
							"release": release,
						}),
				),
				Owner:       "root",
				Path:        fmt.Sprintf("/etc/systemd/system/docker-%s.service", app),
				Permissions: "0644",
			},
		},
		RunCommands: []string{
			"iptables -w -A INPUT -p tcp --dport 8443 -j ACCEPT",
			"systemctl daemon-reload",
			fmt.Sprintf("systemctl start docker-%s.service", app),
		},
	}

	for name, sidecar := range config.Sidecars {
		cc.WriteFiles = append(cc.WriteFiles,
			cloudinit.File{
				Content: systemdService(name,
					sidecar.DockerArgs(name, "", "",
						map[string]string{
							"app":     app,
							"release": release,
							"sidecar": name,
						}),
				),
				Owner:       "root",
				Path:        fmt.Sprintf("/etc/systemd/system/docker-%s.service", name),
				Permissions: "0644",
			})
		cc.RunCommands = append(cc.RunCommands, fmt.Sprintf("systemctl start docker-%s.service", name))
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
	args := make([]string, len(dockerArgs))
	for i, s := range dockerArgs {
		args[i] = shellescape.Quote(s)
	}
	return fmt.Sprintf(jobTemplate, name, strings.Join(args, " "), name, name)
}
