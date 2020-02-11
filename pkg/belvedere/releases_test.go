package belvedere

import (
	"encoding/json"
	"testing"

	"github.com/codahale/belvedere/pkg/belvedere/internal/fixtures"
	"google.golang.org/api/compute/v0.beta"
)

func TestReleaseResources(t *testing.T) {
	config := &Config{
		Container: Container{
			Image:   "gcr.io/example/helloworld",
			Command: "/usr/bin/helloworld",
			Args:    []string{"one", "two"},
			Env: map[string]string{
				"ONE": "1 or 2",
				"TWO": "2 or 3",
			},
			DockerOptions: []string{"--turbo"},
		},
		Sidecars: map[string]Container{
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
	resources := releaseResources("my-project", "us-central1", "my-app", "v43", "abcdef0123456789", config)

	actual, err := json.MarshalIndent(map[string]interface{}{
		"resources": resources,
	}, "", "  ")
	if err != nil {
		t.Fatal(err)
	}

	fixtures.Compare(t, "release_fixture.json", actual)
}
