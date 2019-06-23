package belvedere

import (
	"fmt"
	"strings"
)

type Container struct {
	Image         string            `yaml:"image"`
	Command       []string          `yaml:"command"`
	Args          []string          `yaml:"args"`
	Env           map[string]string `yaml:"env"`
	DockerOptions []string          `yaml:"dockerOptions"`
}

func (c *Container) DockerArgs(name, sha256 string, labels map[string]string) []string {
	var labelNames []string
	for k := range labels {
		labelNames = append(labelNames, k)
	}

	args := []string{
		// TODO sort out logging options
		// https://docs.docker.com/config/containers/logging/gcplogs/
		"--log-driver", "gcplogs",
		"--log-opt", fmt.Sprintf("labels=%s", strings.Join(labelNames, ",")),
		"--name", name,
		"--network", "host",
		"--oom-kill-disable",
	}

	for k, v := range labels {
		args = append(args, []string{
			"--label", fmt.Sprintf("%s=%s", k, v),
		}...)
	}

	for k, v := range c.Env {
		args = append(args, []string{
			"--env", fmt.Sprintf("%s=%s", k, v),
		}...)
	}

	args = append(args, c.DockerOptions...)
	url := c.Image
	if sha256 != "" {
		url = fmt.Sprintf("%s@sha256:%s", url, sha256)
	}
	args = append(args, url)
	args = append(args, c.Command...)
	args = append(args, c.Args...)

	return args
}
