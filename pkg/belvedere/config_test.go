package belvedere

import (
	"context"
	"io/ioutil"
	"testing"

	"github.com/google/go-cmp/cmp"
	compute "google.golang.org/api/compute/v0.beta"
)

func TestLoadConfig(t *testing.T) {
	b, err := ioutil.ReadFile("config-example.yaml")
	if err != nil {
		t.Fatal(err)
	}

	actual, err := LoadConfig(context.TODO(), b)
	if err != nil {
		t.Fatal(err)
	}

	expected := &Config{
		MachineType: "n1-standard-1",
		NumReplicas: 2,
		Container: Container{
			Image:         "gcr.io/cloudslap/helloworld",
			Command:       "ls",
			Args:          []string{"-al"},
			Env:           map[string]string{"ONE": "1"},
			DockerOptions: []string{"--verbose"},
		},
		Sidecars: map[string]Container{
			"nginx-frontend": {
				Image: "gcr.io/cloudslap/nginx-frontend",
			},
		},
		AutoscalingPolicy: &compute.AutoscalingPolicy{
			MinNumReplicas:    1,
			MaxNumReplicas:    10,
			CoolDownPeriodSec: 60,
			CpuUtilization: &compute.AutoscalingPolicyCpuUtilization{
				UtilizationTarget: 0.6,
			},
			CustomMetricUtilizations: []*compute.AutoscalingPolicyCustomMetricUtilization{
				{
					Metric:                "www.googleapis.com/compute/instance/network/received_bytes_count",
					UtilizationTargetType: "GAUGE",
					UtilizationTarget:     200,
				},
			},
			LoadBalancingUtilization: &compute.AutoscalingPolicyLoadBalancingUtilization{
				UtilizationTarget: 0.6,
			},
		},
		IAP: &compute.BackendServiceIAP{
			Enabled:            true,
			Oauth2ClientId:     "client-id",
			Oauth2ClientSecret: "secret-id",
		},
		CDNPolicy: &compute.BackendServiceCdnPolicy{
			CacheKeyPolicy: &compute.CacheKeyPolicy{
				IncludeProtocol:      true,
				IncludeHost:          true,
				IncludeQueryString:   false,
				QueryStringWhitelist: []string{"q"},
				QueryStringBlacklist: []string{"id"},
			},
			SignedUrlKeyNames:       []string{"one"},
			SignedUrlCacheMaxAgeSec: 200,
		},
		IAMRoles:   []string{"roles/cloudkms.cryptoKeyDecrypter"},
		Network:    "projects/project/global/networks/network",
		Subnetwork: "regions/region/subnetworks/subnetwork",
	}

	if !cmp.Equal(expected, actual) {
		t.Error(cmp.Diff(expected, actual))

	}
}
