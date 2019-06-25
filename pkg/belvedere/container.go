package belvedere

import (
	"fmt"
	"sort"
	"strings"
)

type Container struct {
	Image         string            `yaml:"image"`
	Command       string            `yaml:"command"`
	Args          []string          `yaml:"args"`
	Env           map[string]string `yaml:"env"`
	DockerOptions []string          `yaml:"dockerOptions"`
}

func (c *Container) DockerArgs(app, release, sha256 string, labels map[string]string) []string {
	var labelNames []string
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

	var envNames []string
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
	args = append(args, c.Command)
	args = append(args, c.Args...)

	return args
}
