package check

import (
	"context"
	"net/http"
	"testing"

	"github.com/codahale/belvedere/pkg/belvedere/internal/it"
	compute "google.golang.org/api/compute/v0.beta"
	"gopkg.in/h2non/gock.v1"
)

func TestHealthNotStable(t *testing.T) {
	defer gock.Off()
	it.MockTokenSource()

	gock.New("https://compute.googleapis.com/compute/beta/projects/my-project/regions/us-central1/instanceGroupManagers/ig-1?alt=json&fields=status&prettyPrint=false").
		Reply(http.StatusOK).
		JSON(compute.InstanceGroupManager{
			Status: &compute.InstanceGroupManagerStatus{
				IsStable: false,
			},
		})

	gce, err := compute.NewService(context.TODO())
	if err != nil {
		t.Fatal(err)
	}

	f := Health(context.TODO(), gce, "my-project", "us-central1", "bes-1", "ig-1")
	done, err := f()
	if err != nil {
		t.Fatal(err)
	}

	if done {
		t.Error("shouldn't have been done but was")
	}
}

func TestHealthNotRegistered(t *testing.T) {
	defer gock.Off()
	it.MockTokenSource()

	gock.New("https://compute.googleapis.com/compute/beta/projects/my-project/regions/us-central1/instanceGroupManagers/ig-1?alt=json&fields=status&prettyPrint=false").
		Reply(http.StatusOK).
		JSON(compute.InstanceGroupManager{
			Status: &compute.InstanceGroupManagerStatus{
				IsStable: true,
			},
		})

	gock.New("https://compute.googleapis.com/compute/beta/projects/my-project/regions/us-central1/instanceGroups/ig-1?alt=json&fields=selfLink%2Csize&prettyPrint=false").
		Reply(http.StatusOK).
		JSON(compute.InstanceGroup{
			SelfLink: "https://self-link/",
			Size:     2,
		})

	gock.New("https://compute.googleapis.com/compute/beta/projects/my-project/global/backendServices/bes-1/getHealth?alt=json&prettyPrint=false").
		Reply(http.StatusOK).
		JSON(compute.BackendServiceGroupHealth{
			HealthStatus: []*compute.HealthStatus{},
		})

	gce, err := compute.NewService(context.TODO())
	if err != nil {
		t.Fatal(err)
	}

	f := Health(context.TODO(), gce, "my-project", "us-central1", "bes-1", "ig-1")
	done, err := f()
	if err != nil {
		t.Fatal(err)
	}

	if done {
		t.Error("shouldn't have been done but was")
	}
}

func TestHealthNotHealthy(t *testing.T) {
	defer gock.Off()
	it.MockTokenSource()

	gock.New("https://compute.googleapis.com/compute/beta/projects/my-project/regions/us-central1/instanceGroupManagers/ig-1?alt=json&fields=status&prettyPrint=false").
		Reply(http.StatusOK).
		JSON(compute.InstanceGroupManager{
			Status: &compute.InstanceGroupManagerStatus{
				IsStable: true,
			},
		})

	gock.New("https://compute.googleapis.com/compute/beta/projects/my-project/regions/us-central1/instanceGroups/ig-1?alt=json&fields=selfLink%2Csize&prettyPrint=false").
		Reply(http.StatusOK).
		JSON(compute.InstanceGroup{
			SelfLink: "https://self-link/",
			Size:     2,
		})

	gock.New("https://compute.googleapis.com/compute/beta/projects/my-project/global/backendServices/bes-1/getHealth?alt=json&prettyPrint=false").
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

	gce, err := compute.NewService(context.TODO())
	if err != nil {
		t.Fatal(err)
	}

	f := Health(context.TODO(), gce, "my-project", "us-central1", "bes-1", "ig-1")
	done, err := f()
	if err != nil {
		t.Fatal(err)
	}

	if done {
		t.Error("shouldn't have been done but was")
	}
}

func TestHealthDone(t *testing.T) {
	defer gock.Off()
	it.MockTokenSource()

	gock.New("https://compute.googleapis.com/compute/beta/projects/my-project/regions/us-central1/instanceGroupManagers/ig-1?alt=json&fields=status&prettyPrint=false").
		Reply(http.StatusOK).
		JSON(compute.InstanceGroupManager{
			Status: &compute.InstanceGroupManagerStatus{
				IsStable: true,
			},
		})

	gock.New("https://compute.googleapis.com/compute/beta/projects/my-project/regions/us-central1/instanceGroups/ig-1?alt=json&fields=selfLink%2Csize&prettyPrint=false").
		Reply(http.StatusOK).
		JSON(compute.InstanceGroup{
			SelfLink: "https://self-link/",
			Size:     2,
		})

	gock.New("https://compute.googleapis.com/compute/beta/projects/my-project/global/backendServices/bes-1/getHealth?alt=json&prettyPrint=false").
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

	gce, err := compute.NewService(context.TODO())
	if err != nil {
		t.Fatal(err)
	}

	f := Health(context.TODO(), gce, "my-project", "us-central1", "bes-1", "ig-1")
	done, err := f()
	if err != nil {
		t.Fatal(err)
	}

	if !done {
		t.Error("should have been done but wasn't")
	}
}
