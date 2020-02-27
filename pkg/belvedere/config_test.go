package belvedere

import (
	"io/ioutil"
	"testing"

	"github.com/codahale/belvedere/pkg/belvedere/internal/fixtures"
	"github.com/google/go-cmp/cmp"
	compute "google.golang.org/api/compute/v0.beta"
)

func TestLoadConfig(t *testing.T) {
	b, err := ioutil.ReadFile("config-example.yaml")
	if err != nil {
		t.Fatal(err)
	}

	actual, err := ParseConfig(b)
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

func TestDockerArgs(t *testing.T) {
	container := Container{
		Image:   "gcr.io/example/example",
		Command: "/usr/bin/example",
		Args:    []string{"-h", "-y"},
		Env: map[string]string{
			"TWO": "2",
			"ONE": "1",
		},
		DockerOptions: []string{"--privileged"},
	}

	expected := []string{
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
	actual := container.dockerArgs("my-example", "v3", "123456",
		map[string]string{
			"env":      "qa",
			"alphabet": "latin",
		})

	if !cmp.Equal(expected, actual) {
		t.Fatal(cmp.Diff(expected, actual))
	}
}

func TestDockerArgsEmptyCommand(t *testing.T) {
	container := Container{
		Image: "gcr.io/example/example",
		Args:  []string{"-h", "-y"},
		Env: map[string]string{
			"TWO": "2",
			"ONE": "1",
		},
		DockerOptions: []string{"--privileged"},
	}

	expected := []string{
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
	actual := container.dockerArgs("my-example", "v3", "123456",
		map[string]string{
			"env":      "qa",
			"alphabet": "latin",
		})

	if !cmp.Equal(expected, actual) {
		t.Fatal(cmp.Diff(expected, actual))
	}
}

func TestCloudConfig(t *testing.T) {
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

	actual := config.cloudConfig("my-app", "v43", "abcdef0123456789")
	fixtures.Compare(t, "cloudconfig.yaml", []byte(actual))
}
