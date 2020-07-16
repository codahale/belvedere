package resources

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/alessio/shellescape"
	"github.com/codahale/belvedere/pkg/belvedere/cfg"
	"github.com/codahale/belvedere/pkg/belvedere/internal/deployments"
	compute "google.golang.org/api/compute/v1"
)

func (*builder) Release(project, region, app, release, imageSHA256 string, config *cfg.Config) []deployments.Resource {
	instanceTemplate := fmt.Sprintf("%s-%s-it", app, release)
	instanceGroupManager := fmt.Sprintf("%s-%s-ig", app, release)
	autoscaler := fmt.Sprintf("%s-%s-as", app, release)

	network := defaultNetwork
	if config.Network != "" {
		network = config.Network
	}

	dep := []deployments.Resource{
		// An instance template for creating release instances.
		{
			Name: instanceTemplate,
			Type: "compute.v1.instanceTemplate",
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
							metaData("user-data", cloudConfig(config, app, release, imageSHA256)),
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
					// Enable all Shielded VM options.
					ShieldedInstanceConfig: &compute.ShieldedInstanceConfig{
						EnableIntegrityMonitoring: true,
						EnableSecureBoot:          true,
						EnableVtpm:                true,
					},
					// Tag the instance to disable SSH access and enable IAP tunneling.
					Tags: &compute.Tags{
						Items: []string{
							"belvedere", Name(app),
						},
					},
				},
			},
		},
		// An instance manager to start and stop instances as needed.
		{
			Name: instanceGroupManager,
			Type: "compute.v1.regionInstanceGroupManager",
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
		dep = append(dep, deployments.Resource{
			Name: autoscaler,
			Type: "compute.v1.regionAutoscaler",
			Properties: &compute.Autoscaler{
				Name:              fmt.Sprintf("%s-%s", app, release),
				AutoscalingPolicy: config.AutoscalingPolicy,
				Region:            region,
				Target:            deployments.SelfLink(instanceGroupManager),
			},
		})
	}

	return dep
}

const (
	defaultNetwork = "global/networks/default"
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
func cloudConfig(c *cfg.Config, app, release, imageSHA256 string) string {
	type file struct {
		Path        string `json:"path,omitempty"`
		Permissions string `json:"permissions,omitempty"`
		Owner       string `json:"owner,omitempty"`
		Content     string `json:"content,omitempty"`
	}

	cc := struct {
		WriteFiles  []file   `json:"write_files,omitempty"`
		RunCommands []string `json:"runcmd,omitempty"`
	}{
		WriteFiles: []file{
			// Write a systemd service for running the app's container in Docker.
			{
				Content: systemdService(app,
					dockerArgs(&c.Container, app, release, imageSHA256,
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

	for name, sidecar := range c.Sidecars {
		sidecar := sidecar
		// Add a systemd service for running the sidecar in Docker.
		cc.WriteFiles = append(cc.WriteFiles,
			file{
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

	b, _ := json.Marshal(cc)

	return fmt.Sprintf("#cloud-config\n\n%s", b)
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
	args := make([]string, 0, len(dockerArgs))
	for _, s := range dockerArgs {
		args = append(args, shellescape.Quote(s))
	}

	return fmt.Sprintf(systemdTemplate, name, strings.Join(args, " "), name, name)
}

// dockerArgs returns a list of arguments to `docker run` for running the given container.
func dockerArgs(c *cfg.Container, app, release, sha256 string, labels map[string]string) []string {
	labelNames := make([]string, 0, len(labels))
	for k := range labels {
		labelNames = append(labelNames, k)
	}

	sort.Stable(sort.StringSlice(labelNames))

	args := []string{
		"--log-driver", "gcplogs",
		"--log-opt", fmt.Sprintf("labels=%s", strings.Join(labelNames, ",")),
		"--name", app,
		"--network", "host",
		"--oom-kill-disable",
	}

	for _, k := range labelNames {
		args = append(args, []string{
			"--label", fmt.Sprintf("%s=%s", k, labels[k]),
		}...)
	}

	if release != "" {
		args = append(args, []string{
			"--env", fmt.Sprintf("RELEASE=%s", release),
		}...)
	}

	envNames := make([]string, 0, len(c.Env))
	for k := range c.Env {
		envNames = append(envNames, k)
	}

	sort.Stable(sort.StringSlice(envNames))

	for _, k := range envNames {
		args = append(args, []string{
			"--env", fmt.Sprintf("%s=%s", k, c.Env[k]),
		}...)
	}

	args = append(args, c.DockerOptions...)
	url := c.Image

	if sha256 != "" {
		url = fmt.Sprintf("%s@sha256:%s", url, sha256)
	}

	args = append(args, url)
	if c.Command != "" {
		args = append(args, c.Command)
	}

	args = append(args, c.Args...)

	return args
}
