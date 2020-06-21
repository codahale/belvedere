package cfg

import (
	"io/ioutil"
	"testing"

	"github.com/codahale/belvedere/internal/assert"
	compute "google.golang.org/api/compute/v1"
)

func TestParse(t *testing.T) {
	b, err := ioutil.ReadFile("config-example.yaml")
	if err != nil {
		t.Fatal(err)
	}

	got, err := Parse(b)
	if err != nil {
		t.Fatal(err)
	}

	want := &Config{
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
		IAMRoles: []string{"roles/cloudkms.cryptoKeyDecrypter"},
		WAFRules: []*compute.SecurityPolicyRule{
			{
				Action:      "deny(403)",
				Description: "Prevent XSS attacks.",
				Match: &compute.SecurityPolicyRuleMatcher{
					Expr: &compute.Expr{
						Expression: "evaluatePreconfiguredExpr('xss-stable')",
					},
				},
				Priority: 1,
			},
		},
		Network:    "projects/project/global/networks/network",
		Subnetwork: "regions/region/subnetworks/subnetwork",
	}

	assert.Equal(t, "Parse()", want, got)
}
