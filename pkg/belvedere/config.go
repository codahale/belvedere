package belvedere

import (
	"fmt"
	"sort"
	"strings"

	"github.com/alessio/shellescape"
	"github.com/codahale/belvedere/pkg/belvedere/internal/cloudinit"
	"github.com/ghodss/yaml"
	compute "google.golang.org/api/compute/v0.beta"
)

// Config contains all the mutable parameters of an app's configuration.
type Config struct {
	IAMRoles          []string                         `json:"iamRoles,omitempty"`
	NumReplicas       int                              `json:"numReplicas"`
	MachineType       string                           `json:"machineType"`
	Container         Container                        `json:"container"`
	Sidecars          map[string]Container             `json:"sidecars"`
	IAP               *compute.BackendServiceIAP       `json:"identityAwareProxy"`
	AutoscalingPolicy *compute.AutoscalingPolicy       `json:"autoscalingPolicy"`
	CDNPolicy         *compute.BackendServiceCdnPolicy `json:"cdnPolicy"`
	Network           string                           `json:"network"`
	Subnetwork        string                           `json:"subnetwork"`
}

// ParseConfig loads the given bytes as a YAML configuration.
func ParseConfig(b []byte) (*Config, error) {
	// Unmarshal from YAML using the YAML->JSON route. This allows us to embed GCP API structs in
	// our Config struct.
	var config Config
	if err := yaml.Unmarshal(b, &config); err != nil {
		return nil, fmt.Errorf("error parsing config: %w", err)
	}
	return &config, nil
}

// CloudConfig returns a cloud-config manifest for the given release.
func (c *Config) CloudConfig(app, release string, imageSHA256 string) string {
	cc := cloudinit.CloudConfig{
		WriteFiles: []cloudinit.File{
			// Write a systemd service for running the app's container in Docker.
			{
				Content: systemdService(app,
					c.Container.dockerArgs(app, release, imageSHA256,
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
			cloudinit.File{
				Content: systemdService(name,
					sidecar.dockerArgs(name, "", "",
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
	args := make([]string, 0, len(dockerArgs))
	for _, s := range dockerArgs {
		args = append(args, shellescape.Quote(s))
	}
	return fmt.Sprintf(systemdTemplate, name, strings.Join(args, " "), name, name)
}

// A Container describes all the elements of an app or sidecar container.
type Container struct {
	Image         string            `json:"image"`
	Command       string            `json:"command"`
	Args          []string          `json:"args"`
	Env           map[string]string `json:"env"`
	DockerOptions []string          `json:"dockerOptions"`
}

// dockerArgs returns a list of arguments to `docker run` for running the given container.
func (c *Container) dockerArgs(app, release, sha256 string, labels map[string]string) []string {
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
