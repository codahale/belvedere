package resources

import (
	"encoding/json"
	"testing"

	"github.com/codahale/belvedere/internal/assert"
	"github.com/codahale/belvedere/pkg/belvedere/cfg"
	compute "google.golang.org/api/compute/v1"
)

func TestReleaseResources(t *testing.T) {
	resources := NewBuilder().Release(
		"my-project", "us-central1", "my-app", "v43", "echo woo",
		&cfg.Config{
			Network:     "network",
			Subnetwork:  "subnetwork",
			MachineType: "n1-standard-1",
			NumReplicas: 20,
			AutoscalingPolicy: &compute.AutoscalingPolicy{
				MinNumReplicas: 10,
				MaxNumReplicas: 100,
				LoadBalancingUtilization: &compute.AutoscalingPolicyLoadBalancingUtilization{
					UtilizationTarget: 0.6,
				},
			},
		},
	)

	got, err := json.MarshalIndent(map[string]interface{}{
		"resources": resources,
	}, "", "  ")
	if err != nil {
		t.Fatal(err)
	}

	assert.EqualFixture(t, "Release()", "release.json", got)
}

func TestDockerArgs(t *testing.T) {
	container := &cfg.Container{
		Image:   "gcr.io/example/example",
		Command: "/usr/bin/example",
		Args:    []string{"-h", "-y"},
		Env: map[string]string{
			"TWO": "2",
			"ONE": "1",
		},
		DockerOptions: []string{"--privileged"},
	}

	want := []string{
		"--log-driver", "gcplogs",
		"--log-opt", "labels=alphabet,env",
		"--name", "my-example",
		"--network", "host",
		"--oom-kill-disable",
		"--label", "alphabet=latin",
		"--label", "env=qa",
		"--env", "RELEASE=v3",
		"--env", "ONE=1",
		"--env", "TWO=2",
		"--privileged",
		"gcr.io/example/example@sha256:123456",
		"/usr/bin/example", "-h", "-y",
	}
	got := dockerArgs(container, "my-example", "v3", "123456",
		map[string]string{
			"env":      "qa",
			"alphabet": "latin",
		})

	assert.Equal(t, "dockerArgs()", want, got)
}

func TestDockerArgsEmptyCommand(t *testing.T) {
	container := &cfg.Container{
		Image: "gcr.io/example/example",
		Args:  []string{"-h", "-y"},
		Env: map[string]string{
			"TWO": "2",
			"ONE": "1",
		},
		DockerOptions: []string{"--privileged"},
	}

	want := []string{
		"--log-driver", "gcplogs",
		"--log-opt", "labels=alphabet,env",
		"--name", "my-example",
		"--network", "host",
		"--oom-kill-disable",
		"--label", "alphabet=latin",
		"--label", "env=qa",
		"--env", "RELEASE=v3",
		"--env", "ONE=1",
		"--env", "TWO=2",
		"--privileged",
		"gcr.io/example/example@sha256:123456",
		"-h", "-y",
	}
	got := dockerArgs(container, "my-example", "v3", "123456",
		map[string]string{
			"env":      "qa",
			"alphabet": "latin",
		})

	assert.Equal(t, "dockerArgs()", want, got)
}

func TestCloudConfig(t *testing.T) {
	config := &cfg.Config{
		Container: cfg.Container{
			Image:   "gcr.io/example/helloworld",
			Command: "/usr/bin/helloworld",
			Args:    []string{"one", "two"},
			Env: map[string]string{
				"ONE": "1 or 2",
				"TWO": "2 or 3",
			},
			DockerOptions: []string{"--turbo"},
		},
		Sidecars: map[string]cfg.Container{
			"nginx": {
				Image: "gcr.io/example/nginx",
				Args:  []string{"three", "four"},
				Env: map[string]string{
					"THREE": "3 or 4",
					"FOUR":  "4 or 5",
				},
				DockerOptions: []string{"--slow"},
			},
		},
		NumReplicas: 20,
		MachineType: "n1-standard-1",
		AutoscalingPolicy: &compute.AutoscalingPolicy{
			MinNumReplicas: 10,
			MaxNumReplicas: 100,
			LoadBalancingUtilization: &compute.AutoscalingPolicyLoadBalancingUtilization{
				UtilizationTarget: 0.6,
			},
		},
	}

	got := cloudConfig(config, "my-app", "v43", "abcdef0123456789")
	assert.EqualFixture(t, "cloudConfig()", "cloudconfig.yaml", []byte(got))
}
