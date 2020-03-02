package resources

import (
	"fmt"

	"github.com/codahale/belvedere/pkg/belvedere/cfg"
	"github.com/codahale/belvedere/pkg/belvedere/internal/deployments"
	compute "google.golang.org/api/compute/v0.beta"
)

func (*builder) Release(project, region, app, release, imageSHA256 string, config *cfg.Config) []deployments.Resource {
	instanceTemplate := fmt.Sprintf("%s-%s-it", app, release)
	instanceGroupManager := fmt.Sprintf("%s-%s-ig", app, release)
	autoscaler := fmt.Sprintf("%s-%s-as", app, release)
	if config.Network == "" {
		config.Network = defaultNetwork
	}
	dep := []deployments.Resource{
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
							metaData("user-data", config.CloudConfig(app, release, imageSHA256)),
						},
					},
					// Enable outbound internet access for the instances.
					NetworkInterfaces: []*compute.NetworkInterface{
						{
							Network:    config.Network,
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
							"belvedere", Name(app),
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
		dep = append(dep, deployments.Resource{
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
