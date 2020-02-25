package resources

import (
	"encoding/json"
	"testing"

	"github.com/codahale/belvedere/pkg/belvedere/internal/fixtures"
	compute "google.golang.org/api/compute/v0.beta"
)

func TestReleaseResources(t *testing.T) {
	resources := Release(
		"my-project", "us-central1", "my-app", "v43",
		"", "", "n1-standard-1", "echo woo",
		20,
		&compute.AutoscalingPolicy{
			MinNumReplicas: 10,
			MaxNumReplicas: 100,
			LoadBalancingUtilization: &compute.AutoscalingPolicyLoadBalancingUtilization{
				UtilizationTarget: 0.6,
			},
		},
	)

	actual, err := json.MarshalIndent(map[string]interface{}{
		"resources": resources,
	}, "", "  ")
	if err != nil {
		t.Fatal(err)
	}

	fixtures.Compare(t, "release.json", actual)
}
