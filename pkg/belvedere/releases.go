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
	"github.com/codahale/belvedere/pkg/belvedere/internal/waiter"
	"go.opencensus.io/trace"
	compute "google.golang.org/api/compute/v0.beta"
	"gopkg.in/alessio/shellescape.v1"
)

// A Release describes a specific release of an app.
type Release struct {
	Project string
	Region  string
	App     string
	Release string
	Hash    string
}

// Releases returns a list of releases in the given project for the given app, if any is passed.
func Releases(ctx context.Context, project, app string) ([]Release, error) {
	ctx, span := trace.StartSpan(ctx, "belvedere.Releases")
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

// CreateRelease creates a deployment containing release resources for the given app.
func CreateRelease(ctx context.Context, project, app, release string, config *Config, imageSHA256 string, dryRun bool, interval time.Duration) error {
	ctx, span := trace.StartSpan(ctx, "belvedere.CreateRelease")
	span.AddAttributes(
		trace.StringAttribute("project", project),
		trace.StringAttribute("app", app),
		trace.StringAttribute("release", release),
		trace.StringAttribute("image_sha256", imageSHA256),
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

	name := fmt.Sprintf("belvedere-%s-%s", app, release)
	return deployments.Insert(ctx, project, name,
		releaseResources(project, region, app, release, imageSHA256, config),
		map[string]string{
			"belvedere-type":    "release",
			"belvedere-app":     app,
			"belvedere-release": release,
			"belvedere-region":  region,
			"belvedere-hash":    imageSHA256[:32],
		}, dryRun, interval)
}

// EnableRelease adds the release's instance group to the app's backend service and waits for the
// instances to go fully into service.
func EnableRelease(ctx context.Context, project, app, release string, dryRun bool, interval time.Duration) error {
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
	if err := backends.Add(ctx, project, region, backendService, instanceGroup, dryRun, interval); err != nil {
		return err
	}

	return waiter.Poll(ctx, interval, check.Health(ctx, project, region, backendService, instanceGroup))
}

// DisableRelease removes the release's instance group from the app's backend service.
func DisableRelease(ctx context.Context, project, app, release string, dryRun bool, interval time.Duration) error {
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
	return backends.Remove(ctx, project, region, backendService, instanceGroup, dryRun, interval)
}

// DeleteRelease deletes the release's deployment and waits for all underlying resources to be
// deleted.
func DeleteRelease(ctx context.Context, project, app, release string, dryRun, async bool, interval time.Duration) error {
	ctx, span := trace.StartSpan(ctx, "belvedere.DeleteRelease")
	span.AddAttributes(
		trace.StringAttribute("project", project),
		trace.StringAttribute("app", app),
		trace.StringAttribute("release", release),
		trace.BoolAttribute("dry_run", dryRun),
		trace.BoolAttribute("async", async),
	)
	defer span.End()

	return deployments.Delete(ctx, project, fmt.Sprintf("belvedere-%s-%s", app, release), dryRun, async, interval)
}

// releaseResources returns a list of Deployment Manager resources for the given release.
func releaseResources(project string, region string, app string, release string, imageSHA256 string, config *Config) []deployments.Resource {
	instanceTemplate := fmt.Sprintf("%s-%s-it", app, release)
	instanceGroupManager := fmt.Sprintf("%s-%s-ig", app, release)
	autoscaler := fmt.Sprintf("%s-%s-as", app, release)
	network := "global/networks/default"
	if config.Network != "" {
		network = config.Network
	}
	resources := []deployments.Resource{
		// An instance template for creating release instances.
		{
			Name: instanceTemplate,
			Type: "compute.beta.instanceTemplate",
			Properties: &compute.InstanceTemplate{
				Properties: &compute.InstanceProperties{
					// Use Google Container-Optimized OS with a default disk size.
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
							// https://cloud.google.com/compute/docs/storing-retrieving-metadata#querying
							metaData("disable-legacy-endpoints", "true"),
							// https://cloud.google.com/compute/docs/instances/managing-instance-access
							metaData("enable-os-login", "true"),
							// Enable the Stackdriver Logging Agent for the instance.
							metaData("google-logging-enable", "true"),
							// Inject the cloud-init metadata.
							metaData("user-data", cloudConfig(app, release, config, imageSHA256)),
						},
					},
					// Enable outbound internet access for the instances.
					NetworkInterfaces: []*compute.NetworkInterface{
						{
							Network:    network,
							Subnetwork: config.Subnetwork,
							AccessConfigs: []*compute.AccessConfig{
								{
									Name: "External NAT",
									Type: "ONE_TO_ONE_NAT",
								},
							},
						},
					},
					// Bind the instances to the app's service account and use IAM roles to handle
					// permissions.
					ServiceAccounts: []*compute.ServiceAccount{
						{
							Email: fmt.Sprintf("app-%s@%s.iam.gserviceaccount.com", app, project),
							Scopes: []string{
								compute.CloudPlatformScope,
							},
						},
					},
					// TODO move to v1 when shielded VMs goes GA
					// Enable all Shielded VM options.
					ShieldedVmConfig: &compute.ShieldedVmConfig{
						EnableIntegrityMonitoring: true,
						EnableSecureBoot:          true,
						EnableVtpm:                true,
					},
					// Tag the instance to disable SSH access and enable IAP tunneling.
					Tags: &compute.Tags{
						Items: []string{
							"belvedere",
							fmt.Sprintf("belvedere-%s", app),
						},
					},
				},
			},
		},
		// An instance manager to start and stop instances as needed.
		{
			Name: instanceGroupManager,
			Type: "compute.beta.regionInstanceGroupManager",
			Properties: &compute.InstanceGroupManager{
				BaseInstanceName: fmt.Sprintf("%s-%s", app, release),
				InstanceTemplate: deployments.SelfLink(instanceTemplate),
				Region:           region,
				NamedPorts: []*compute.NamedPort{
					{
						Name: "svc-https",
						Port: 8443,
					},
				},
				TargetSize: int64(config.NumReplicas),
			},
		},
	}

	// An optional autoscaler.
	if config.AutoscalingPolicy != nil {
		resources = append(resources, deployments.Resource{
			Name: autoscaler,
			Type: "compute.beta.regionAutoscaler",
			Properties: &compute.Autoscaler{
				Name:              fmt.Sprintf("%s-%s", app, release),
				AutoscalingPolicy: config.AutoscalingPolicy,
				Region:            region,
				Target:            deployments.SelfLink(instanceGroupManager),
			},
		})
	}

	return resources
}

const (
	// https://cloud.google.com/container-optimized-os/docs/
	cosStable = "https://www.googleapis.com/compute/v1/projects/gce-uefi-images/global/images/family/cos-stable"
)

// metaData returns a GCE metadata item with the given key and value.
func metaData(key, value string) *compute.MetadataItems {
	return &compute.MetadataItems{
		Key:   key,
		Value: &value,
	}
}

// cloudConfig returns a cloud-config manifest for the given release.
func cloudConfig(app, release string, config *Config, imageSHA256 string) string {
	cc := cloudinit.CloudConfig{
		WriteFiles: []cloudinit.File{
			// Write a systemd service for running the app's container in Docker.
			{
				Content: systemdService(app,
					dockerArgs(&config.Container, app, release, imageSHA256,
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
			// Enable service traffic through the host firewall.
			"iptables -w -A INPUT -p tcp --dport 8443 -j ACCEPT",
			// Load all new systemd services.
			"systemctl daemon-reload",
			// Start the app's systemd service.
			fmt.Sprintf("systemctl start docker-%s.service", app),
		},
	}

	for name, sidecar := range config.Sidecars {
		sidecar := sidecar
		// Add a systemd service for running the sidecar in Docker.
		cc.WriteFiles = append(cc.WriteFiles,
			cloudinit.File{
				Content: systemdService(name,
					dockerArgs(&sidecar, name, "", "",
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
		// Start the sidecar's systemd service.
		cc.RunCommands = append(cc.RunCommands, fmt.Sprintf("systemctl start docker-%s.service", name))
	}

	return cc.String()
}

// systemdTemplate is a template for starting a container in Docker. It includes authenticating
// Docker with GCR, which needs to be done here b/c the credentials are not preserved across
// instance reboots.
const systemdTemplate = `[Unit]
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

// systemdService returns a systemd service file with the given Docker arguments. All Docker
// arguments are escaped, if necessary.
func systemdService(name string, dockerArgs []string) string {
	var args []string
	for _, s := range dockerArgs {
		args = append(args, shellescape.Quote(s))
	}
	return fmt.Sprintf(systemdTemplate, name, strings.Join(args, " "), name, name)
}
