package backends

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/codahale/belvedere/pkg/belvedere/internal/it"
	compute "google.golang.org/api/compute/v0.beta"
	"gopkg.in/h2non/gock.v1"
)

func TestAdd(t *testing.T) {
	defer gock.Off()
	it.MockTokenSource()

	gock.New("https://compute.googleapis.com/compute/beta/projects/my-project/global/backendServices/bes-1?alt=json&fields=backends%2Cfingerprint&prettyPrint=false").
		Reply(http.StatusOK).
		JSON(compute.BackendService{
			Backends: []*compute.Backend{
				{
					Group: "http://ig-2",
				},
			},
			Fingerprint: "fp",
		})

	gock.New("https://compute.googleapis.com/compute/beta/projects/my-project/regions/us-central1/instanceGroups/ig-1?alt=json&fields=selfLink&prettyPrint=false").
		Reply(http.StatusOK).
		JSON(compute.InstanceGroup{
			SelfLink: "http://ig-1",
		})

	gock.New("https://compute.googleapis.com/compute/beta/projects/my-project/global/backendServices/bes-1?alt=json&prettyPrint=false").
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
		Reply(http.StatusOK).
		JSON(compute.Operation{
			Name: "op1",
		})

	gock.New("https://compute.googleapis.com/compute/beta/projects/my-project/global/operations/op1?alt=json&fields=status%2Cerror&prettyPrint=false").
		Reply(http.StatusOK).
		JSON(compute.Operation{
			Status: "DONE",
		})

	if err := Add(context.TODO(), "my-project", "us-central1", "bes-1", "ig-1", false, 10*time.Millisecond); err != nil {
		t.Fatal(err)
	}
}

func TestAddExisting(t *testing.T) {
	defer gock.Off()
	it.MockTokenSource()

	gock.New("https://compute.googleapis.com/compute/beta/projects/my-project/global/backendServices/bes-1?alt=json&fields=backends%2Cfingerprint&prettyPrint=false").
		Reply(http.StatusOK).
		JSON(compute.BackendService{
			Backends: []*compute.Backend{
				{
					Group: "http://ig-1",
				},
			},
			Fingerprint: "fp",
		})

	gock.New("https://compute.googleapis.com/compute/beta/projects/my-project/regions/us-central1/instanceGroups/ig-1?alt=json&fields=selfLink&prettyPrint=false").
		Reply(http.StatusOK).
		JSON(compute.InstanceGroup{
			SelfLink: "http://ig-1",
		})

	if err := Add(context.TODO(), "my-project", "us-central1", "bes-1", "ig-1", false, 10*time.Millisecond); err != nil {
		t.Fatal(err)
	}
}

func TestAddDryRun(t *testing.T) {
	defer gock.Off()
	it.MockTokenSource()

	gock.New("https://compute.googleapis.com/compute/beta/projects/my-project/global/backendServices/bes-1?alt=json&fields=backends%2Cfingerprint&prettyPrint=false").
		Reply(http.StatusOK).
		JSON(compute.BackendService{
			Backends: []*compute.Backend{
				{
					Group: "http://ig-2",
				},
			},
			Fingerprint: "fp",
		})

	gock.New("https://compute.googleapis.com/compute/beta/projects/my-project/regions/us-central1/instanceGroups/ig-1?alt=json&fields=selfLink&prettyPrint=false").
		Reply(http.StatusOK).
		JSON(compute.InstanceGroup{
			SelfLink: "http://ig-1",
		})

	if err := Add(context.TODO(), "my-project", "us-central1", "bes-1", "ig-1", true, 10*time.Millisecond); err != nil {
		t.Fatal(err)
	}
}

func TestRemove(t *testing.T) {
	defer gock.Off()
	it.MockTokenSource()

	gock.New("https://compute.googleapis.com/compute/beta/projects/my-project/global/backendServices/bes-1?alt=json&fields=backends%2Cfingerprint&prettyPrint=false").
		Reply(http.StatusOK).
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

	gock.New("https://compute.googleapis.com/compute/beta/projects/my-project/regions/us-central1/instanceGroups/ig-1?alt=json&fields=selfLink&prettyPrint=false").
		Reply(http.StatusOK).
		JSON(compute.InstanceGroup{
			SelfLink: "http://ig-1",
		})

	gock.New("https://compute.googleapis.com/compute/beta/projects/my-project/global/backendServices/bes-1?alt=json&prettyPrint=false").
		JSON(compute.BackendService{
			Backends: []*compute.Backend{
				{
					Group: "http://ig-2",
				},
			},
			Fingerprint: "fp",
		}).
		Reply(http.StatusOK).
		JSON(compute.Operation{
			Name: "op1",
		})

	gock.New("https://compute.googleapis.com/compute/beta/projects/my-project/global/operations/op1?alt=json&fields=status%2Cerror&prettyPrint=false").
		Reply(http.StatusOK).
		JSON(compute.Operation{
			Status: "DONE",
		})

	if err := Remove(context.TODO(), "my-project", "us-central1", "bes-1", "ig-1", false, 10*time.Millisecond); err != nil {
		t.Fatal(err)
	}
}

func TestRemoveLast(t *testing.T) {
	defer gock.Off()
	it.MockTokenSource()

	gock.New("https://compute.googleapis.com/compute/beta/projects/my-project/global/backendServices/bes-1?alt=json&fields=backends%2Cfingerprint&prettyPrint=false").
		Reply(http.StatusOK).
		JSON(compute.BackendService{
			Backends: []*compute.Backend{
				{
					Group: "http://ig-1",
				},
			},
			Fingerprint: "fp",
		})

	gock.New("https://compute.googleapis.com/compute/beta/projects/my-project/regions/us-central1/instanceGroups/ig-1?alt=json&fields=selfLink&prettyPrint=false").
		Reply(http.StatusOK).
		JSON(compute.InstanceGroup{
			SelfLink: "http://ig-1",
		})

	gock.New("https://compute.googleapis.com/compute/beta/projects/my-project/global/backendServices/bes-1?alt=json&prettyPrint=false").
		JSON(json.RawMessage(`{"backends":[],"fingerprint":"fp"}`)).
		Reply(http.StatusOK).
		JSON(compute.Operation{
			Name: "op1",
		})

	gock.New("https://compute.googleapis.com/compute/beta/projects/my-project/global/operations/op1?alt=json&fields=status%2Cerror&prettyPrint=false").
		Reply(http.StatusOK).
		JSON(compute.Operation{
			Status: "DONE",
		})

	if err := Remove(context.TODO(), "my-project", "us-central1", "bes-1", "ig-1", false, 10*time.Millisecond); err != nil {
		t.Fatal(err)
	}
}

func TestRemoveMissing(t *testing.T) {
	defer gock.Off()
	it.MockTokenSource()

	gock.New("https://compute.googleapis.com/compute/beta/projects/my-project/global/backendServices/bes-1?alt=json&fields=backends%2Cfingerprint&prettyPrint=false").
		Reply(http.StatusOK).
		JSON(compute.BackendService{
			Backends: []*compute.Backend{
				{
					Group: "http://ig-2",
				},
			},
			Fingerprint: "fp",
		})

	gock.New("https://compute.googleapis.com/compute/beta/projects/my-project/regions/us-central1/instanceGroups/ig-1?alt=json&fields=selfLink&prettyPrint=false").
		Reply(http.StatusOK).
		JSON(compute.InstanceGroup{
			SelfLink: "http://ig-1",
		})

	if err := Remove(context.TODO(), "my-project", "us-central1", "bes-1", "ig-1", false, 10*time.Millisecond); err != nil {
		t.Fatal(err)
	}
}

func TestRemoveDryRun(t *testing.T) {
	defer gock.Off()
	it.MockTokenSource()

	gock.New("https://compute.googleapis.com/compute/beta/projects/my-project/global/backendServices/bes-1?alt=json&fields=backends%2Cfingerprint&prettyPrint=false").
		Reply(http.StatusOK).
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

	gock.New("https://compute.googleapis.com/compute/beta/projects/my-project/regions/us-central1/instanceGroups/ig-1?alt=json&fields=selfLink&prettyPrint=false").
		Reply(http.StatusOK).
		JSON(compute.InstanceGroup{
			SelfLink: "http://ig-1",
		})

	if err := Remove(context.TODO(), "my-project", "us-central1", "bes-1", "ig-1", true, 10*time.Millisecond); err != nil {
		t.Fatal(err)
	}
}
