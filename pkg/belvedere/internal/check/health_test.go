package check

import (
	"context"
	"net/http"
	"testing"

	"github.com/codahale/belvedere/internal/assert"
	compute "google.golang.org/api/compute/v1"
	"google.golang.org/api/option"
	"gopkg.in/h2non/gock.v1"
)

func TestHealthNotStable(t *testing.T) {
	defer gock.Off()

	gock.New("https://compute.googleapis.com/compute/v1/projects/my-project/regions/us-central1/instanceGroupManagers/ig-1?alt=json&fields=status&prettyPrint=false").
		Reply(http.StatusOK).
		JSON(compute.InstanceGroupManager{
			Status: &compute.InstanceGroupManagerStatus{
				IsStable: false,
			},
		})

	gce, err := compute.NewService(
		context.Background(),
		option.WithHTTPClient(http.DefaultClient),
		option.WithoutAuthentication(),
	)
	if err != nil {
		t.Fatal(err)
	}

	done, err := Health(context.Background(), gce, "my-project", "us-central1", "bes-1", "ig-1")()
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, "Health()", false, done)
}

func TestHealthNotRegistered(t *testing.T) {
	defer gock.Off()

	gock.New("https://compute.googleapis.com/compute/v1/projects/my-project/regions/us-central1/instanceGroupManagers/ig-1?alt=json&fields=status&prettyPrint=false").
		Reply(http.StatusOK).
		JSON(compute.InstanceGroupManager{
			Status: &compute.InstanceGroupManagerStatus{
				IsStable: true,
			},
		})

	gock.New("https://compute.googleapis.com/compute/v1/projects/my-project/regions/us-central1/instanceGroups/ig-1?alt=json&fields=selfLink%2Csize&prettyPrint=false").
		Reply(http.StatusOK).
		JSON(compute.InstanceGroup{
			SelfLink: "https://self-link/",
			Size:     2,
		})

	gock.New("https://compute.googleapis.com/compute/v1/projects/my-project/global/backendServices/bes-1/getHealth?alt=json&prettyPrint=false").
		Reply(http.StatusOK).
		JSON(compute.BackendServiceGroupHealth{
			HealthStatus: []*compute.HealthStatus{},
		})

	gce, err := compute.NewService(
		context.Background(),
		option.WithHTTPClient(http.DefaultClient),
		option.WithoutAuthentication(),
	)
	if err != nil {
		t.Fatal(err)
	}

	done, err := Health(context.Background(), gce, "my-project", "us-central1", "bes-1", "ig-1")()
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, "Health()", false, done)
}

func TestHealthNotHealthy(t *testing.T) {
	defer gock.Off()

	gock.New("https://compute.googleapis.com/compute/v1/projects/my-project/regions/us-central1/instanceGroupManagers/ig-1?alt=json&fields=status&prettyPrint=false").
		Reply(http.StatusOK).
		JSON(compute.InstanceGroupManager{
			Status: &compute.InstanceGroupManagerStatus{
				IsStable: true,
			},
		})

	gock.New("https://compute.googleapis.com/compute/v1/projects/my-project/regions/us-central1/instanceGroups/ig-1?alt=json&fields=selfLink%2Csize&prettyPrint=false").
		Reply(http.StatusOK).
		JSON(compute.InstanceGroup{
			SelfLink: "https://self-link/",
			Size:     2,
		})

	gock.New("https://compute.googleapis.com/compute/v1/projects/my-project/global/backendServices/bes-1/getHealth?alt=json&prettyPrint=false").
		Reply(http.StatusOK).
		JSON(compute.BackendServiceGroupHealth{
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
		})

	gce, err := compute.NewService(
		context.Background(),
		option.WithHTTPClient(http.DefaultClient),
		option.WithoutAuthentication(),
	)
	if err != nil {
		t.Fatal(err)
	}

	done, err := Health(context.Background(), gce, "my-project", "us-central1", "bes-1", "ig-1")()
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, "Health()", false, done)
}

func TestHealthDone(t *testing.T) {
	defer gock.Off()

	gock.New("https://compute.googleapis.com/compute/v1/projects/my-project/regions/us-central1/instanceGroupManagers/ig-1?alt=json&fields=status&prettyPrint=false").
		Reply(http.StatusOK).
		JSON(compute.InstanceGroupManager{
			Status: &compute.InstanceGroupManagerStatus{
				IsStable: true,
			},
		})

	gock.New("https://compute.googleapis.com/compute/v1/projects/my-project/regions/us-central1/instanceGroups/ig-1?alt=json&fields=selfLink%2Csize&prettyPrint=false").
		Reply(http.StatusOK).
		JSON(compute.InstanceGroup{
			SelfLink: "https://self-link/",
			Size:     2,
		})

	gock.New("https://compute.googleapis.com/compute/v1/projects/my-project/global/backendServices/bes-1/getHealth?alt=json&prettyPrint=false").
		Reply(http.StatusOK).
		JSON(compute.BackendServiceGroupHealth{
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
		})

	gce, err := compute.NewService(
		context.Background(),
		option.WithHTTPClient(http.DefaultClient),
		option.WithoutAuthentication(),
	)
	if err != nil {
		t.Fatal(err)
	}

	done, err := Health(context.Background(), gce, "my-project", "us-central1", "bes-1", "ig-1")()
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, "Health()", true, done)
}
