package check

import (
	"context"
	"testing"

	"github.com/codahale/gubbins/assert"
	"github.com/codahale/gubbins/httpmock"
	"google.golang.org/api/compute/v1"
	"google.golang.org/api/option"
)

func TestHealth(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		igm    compute.InstanceGroupManager
		ig     compute.InstanceGroup
		health compute.BackendServiceGroupHealth
		done   bool
		errMsg string
	}{
		{
			name: "not stable",
			igm: compute.InstanceGroupManager{
				Status: &compute.InstanceGroupManagerStatus{
					IsStable: false,
				},
			},
			done: false,
		},
		{
			name: "not registered",
			igm: compute.InstanceGroupManager{
				Status: &compute.InstanceGroupManagerStatus{
					IsStable: true,
				},
			},
			ig: compute.InstanceGroup{
				SelfLink: "https://self-link/",
				Size:     2,
			},
			health: compute.BackendServiceGroupHealth{
				HealthStatus: []*compute.HealthStatus{},
			},
			done: false,
		},
		{
			name: "unhealthy",
			igm: compute.InstanceGroupManager{
				Status: &compute.InstanceGroupManagerStatus{
					IsStable: true,
				},
			},
			ig: compute.InstanceGroup{
				SelfLink: "https://self-link/",
				Size:     2,
			},
			health: compute.BackendServiceGroupHealth{
				HealthStatus: []*compute.HealthStatus{
					{
						Instance:    "instance1",
						HealthState: "UNHEALTHY",
					},
					{
						Instance:    "instance2",
						HealthState: "UNHEALTHY",
					},
				},
			},
			done: false,
		},
		{
			name: "done",
			igm: compute.InstanceGroupManager{
				Status: &compute.InstanceGroupManagerStatus{
					IsStable: true,
				},
			},
			ig: compute.InstanceGroup{
				SelfLink: "https://self-link/",
				Size:     2,
			},
			health: compute.BackendServiceGroupHealth{
				HealthStatus: []*compute.HealthStatus{
					{
						Instance:    "instance1",
						HealthState: "HEALTHY",
					},
					{
						Instance:    "instance2",
						HealthState: "HEALTHY",
					},
				},
			},
			done: true,
		},
	}
	for _, testCase := range tests {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			srv := httpmock.NewServer(t)
			defer srv.Finish()

			srv.Expect(`/projects/my-project/regions/us-central1/instanceGroupManagers/ig-1?`+
				`alt=json&fields=status&prettyPrint=false`,
				httpmock.RespJSON(testCase.igm))

			srv.Expect(`/projects/my-project/regions/us-central1/instanceGroups/ig-1?`+
				`alt=json&fields=selfLink%2Csize&prettyPrint=false`,
				httpmock.RespJSON(testCase.ig),
				httpmock.Optional())

			srv.Expect(`/projects/my-project/global/backendServices/bes-1/getHealth?`+
				`alt=json&prettyPrint=false`,
				httpmock.RespJSON(testCase.health),
				httpmock.Optional())

			gce, err := compute.NewService(
				context.Background(),
				option.WithEndpoint(srv.URL()),
				option.WithoutAuthentication(),
			)
			if err != nil {
				t.Fatal(err)
			}

			done, err := Health(context.Background(), gce, "my-project", "us-central1", "bes-1", "ig-1")()

			assert.Equal(t, "done", testCase.done, done)

			errMsg := ""
			if err != nil {
				errMsg = err.Error()
			}

			assert.Equal(t, "errMsg", testCase.errMsg, errMsg)
		})
	}
}
