package backends

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/codahale/belvedere/pkg/belvedere/internal/it"
	compute "google.golang.org/api/compute/v0.beta"
	"gopkg.in/h2non/gock.v1"
)

func TestAdd(t *testing.T) {
	defer gock.Off()
	it.MockTokenSource()

	gock.New("https://www.googleapis.com/compute/beta/projects/my-project/global/backendServices/bes-1?alt=json&fields=backends%2Cfingerprint&prettyPrint=false").
		Reply(200).
		JSON(compute.BackendService{
			Backends: []*compute.Backend{
				{
					Group: "http://ig-2",
				},
			},
			Fingerprint: "fp",
		})

	gock.New("https://www.googleapis.com/compute/beta/projects/my-project/regions/us-central1/instanceGroups/ig-1?alt=json&fields=selfLink&prettyPrint=false").
		Reply(200).
		JSON(compute.InstanceGroup{
			SelfLink: "http://ig-1",
		})

	gock.New("https://www.googleapis.com/compute/beta/projects/my-project/global/backendServices/bes-1?alt=json&prettyPrint=false").
		JSON(compute.BackendService{
			Backends: []*compute.Backend{
				{
					Group: "http://ig-2",
				},
				{
					Group: "http://ig-1",
				},
			},
			Fingerprint: "fp",
		}).
		Reply(200).
		JSON(compute.Operation{
			Name: "op1",
		})

	gock.New("https://www.googleapis.com/compute/beta/projects/my-project/global/operations/op1?alt=json&fields=status%2Cerror&prettyPrint=false").
		Reply(200).
		JSON(compute.Operation{
			Status: "DONE",
		})

	gce, err := compute.NewService(context.TODO())
	if err != nil {
		t.Fatal(err)
	}

	op, err := Add(context.TODO(), gce, "my-project", "us-central1", "bes-1", "ig-1", false)
	if err != nil {
		t.Fatal(err)
	}

	done, err := op()
	if err != nil {
		t.Fatal(err)
	}

	if !done {
		t.Fatal("should have been done")
	}
}

func TestAddExisting(t *testing.T) {
	defer gock.Off()
	it.MockTokenSource()

	gock.New("https://www.googleapis.com/compute/beta/projects/my-project/global/backendServices/bes-1?alt=json&fields=backends%2Cfingerprint&prettyPrint=false").
		Reply(200).
		JSON(compute.BackendService{
			Backends: []*compute.Backend{
				{
					Group: "http://ig-1",
				},
			},
			Fingerprint: "fp",
		})

	gock.New("https://www.googleapis.com/compute/beta/projects/my-project/regions/us-central1/instanceGroups/ig-1?alt=json&fields=selfLink&prettyPrint=false").
		Reply(200).
		JSON(compute.InstanceGroup{
			SelfLink: "http://ig-1",
		})

	gce, err := compute.NewService(context.TODO())
	if err != nil {
		t.Fatal(err)
	}

	op, err := Add(context.TODO(), gce, "my-project", "us-central1", "bes-1", "ig-1", false)
	if err != nil {
		t.Fatal(err)
	}

	done, err := op()
	if err != nil {
		t.Fatal(err)
	}

	if !done {
		t.Fatal("should have been done")
	}
}

func TestAddDryRun(t *testing.T) {
	defer gock.Off()
	it.MockTokenSource()

	gock.New("https://www.googleapis.com/compute/beta/projects/my-project/global/backendServices/bes-1?alt=json&fields=backends%2Cfingerprint&prettyPrint=false").
		Reply(200).
		JSON(compute.BackendService{
			Backends: []*compute.Backend{
				{
					Group: "http://ig-2",
				},
			},
			Fingerprint: "fp",
		})

	gock.New("https://www.googleapis.com/compute/beta/projects/my-project/regions/us-central1/instanceGroups/ig-1?alt=json&fields=selfLink&prettyPrint=false").
		Reply(200).
		JSON(compute.InstanceGroup{
			SelfLink: "http://ig-1",
		})

	gce, err := compute.NewService(context.TODO())
	if err != nil {
		t.Fatal(err)
	}

	op, err := Add(context.TODO(), gce, "my-project", "us-central1", "bes-1", "ig-1", true)
	if err != nil {
		t.Fatal(err)
	}

	done, err := op()
	if err != nil {
		t.Fatal(err)
	}

	if !done {
		t.Fatal("should have been done")
	}
}

func TestRemove(t *testing.T) {
	defer gock.Off()
	it.MockTokenSource()

	gock.New("https://www.googleapis.com/compute/beta/projects/my-project/global/backendServices/bes-1?alt=json&fields=backends%2Cfingerprint&prettyPrint=false").
		Reply(200).
		JSON(compute.BackendService{
			Backends: []*compute.Backend{
				{
					Group: "http://ig-1",
				},
				{
					Group: "http://ig-2",
				},
			},
			Fingerprint: "fp",
		})

	gock.New("https://www.googleapis.com/compute/beta/projects/my-project/regions/us-central1/instanceGroups/ig-1?alt=json&fields=selfLink&prettyPrint=false").
		Reply(200).
		JSON(compute.InstanceGroup{
			SelfLink: "http://ig-1",
		})

	gock.New("https://www.googleapis.com/compute/beta/projects/my-project/global/backendServices/bes-1?alt=json&prettyPrint=false").
		JSON(compute.BackendService{
			Backends: []*compute.Backend{
				{
					Group: "http://ig-2",
				},
			},
			Fingerprint: "fp",
		}).
		Reply(200).
		JSON(compute.Operation{
			Name: "op1",
		})

	gock.New("https://www.googleapis.com/compute/beta/projects/my-project/global/operations/op1?alt=json&fields=status%2Cerror&prettyPrint=false").
		Reply(200).
		JSON(compute.Operation{
			Status: "DONE",
		})

	gce, err := compute.NewService(context.TODO())
	if err != nil {
		t.Fatal(err)
	}

	op, err := Remove(context.TODO(), gce, "my-project", "us-central1", "bes-1", "ig-1", false)
	if err != nil {
		t.Fatal(err)
	}

	done, err := op()
	if err != nil {
		t.Fatal(err)
	}

	if !done {
		t.Fatal("should have been done")
	}
}

func TestRemoveLast(t *testing.T) {
	defer gock.Off()
	it.MockTokenSource()

	gock.New("https://www.googleapis.com/compute/beta/projects/my-project/global/backendServices/bes-1?alt=json&fields=backends%2Cfingerprint&prettyPrint=false").
		Reply(200).
		JSON(compute.BackendService{
			Backends: []*compute.Backend{
				{
					Group: "http://ig-1",
				},
			},
			Fingerprint: "fp",
		})

	gock.New("https://www.googleapis.com/compute/beta/projects/my-project/regions/us-central1/instanceGroups/ig-1?alt=json&fields=selfLink&prettyPrint=false").
		Reply(200).
		JSON(compute.InstanceGroup{
			SelfLink: "http://ig-1",
		})

	gock.New("https://www.googleapis.com/compute/beta/projects/my-project/global/backendServices/bes-1?alt=json&prettyPrint=false").
		JSON(json.RawMessage(`{"backends":[],"fingerprint":"fp"}`)).
		Reply(200).
		JSON(compute.Operation{
			Name: "op1",
		})

	gock.New("https://www.googleapis.com/compute/beta/projects/my-project/global/operations/op1?alt=json&fields=status%2Cerror&prettyPrint=false").
		Reply(200).
		JSON(compute.Operation{
			Status: "DONE",
		})

	gce, err := compute.NewService(context.TODO())
	if err != nil {
		t.Fatal(err)
	}

	op, err := Remove(context.TODO(), gce, "my-project", "us-central1", "bes-1", "ig-1", false)
	if err != nil {
		t.Fatal(err)
	}

	done, err := op()
	if err != nil {
		t.Fatal(err)
	}

	if !done {
		t.Fatal("should have been done")
	}
}

func TestRemoveMissing(t *testing.T) {
	defer gock.Off()
	it.MockTokenSource()

	gock.New("https://www.googleapis.com/compute/beta/projects/my-project/global/backendServices/bes-1?alt=json&fields=backends%2Cfingerprint&prettyPrint=false").
		Reply(200).
		JSON(compute.BackendService{
			Backends: []*compute.Backend{
				{
					Group: "http://ig-2",
				},
			},
			Fingerprint: "fp",
		})

	gock.New("https://www.googleapis.com/compute/beta/projects/my-project/regions/us-central1/instanceGroups/ig-1?alt=json&fields=selfLink&prettyPrint=false").
		Reply(200).
		JSON(compute.InstanceGroup{
			SelfLink: "http://ig-1",
		})

	gce, err := compute.NewService(context.TODO())
	if err != nil {
		t.Fatal(err)
	}

	op, err := Remove(context.TODO(), gce, "my-project", "us-central1", "bes-1", "ig-1", false)
	if err != nil {
		t.Fatal(err)
	}

	done, err := op()
	if err != nil {
		t.Fatal(err)
	}

	if !done {
		t.Fatal("should have been done")
	}
}
