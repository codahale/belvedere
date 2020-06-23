package backends

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	compute "google.golang.org/api/compute/v1"
	"google.golang.org/api/option"
	"gopkg.in/h2non/gock.v1"
)

func TestService_Add(t *testing.T) {
	defer gock.Off()

	gock.New("https://compute.googleapis.com/compute/v1/projects/my-project/global/backendServices/bes-1?alt=json&fields=backends%2Cfingerprint&prettyPrint=false").
		Reply(http.StatusOK).
		JSON(compute.BackendService{
			Backends: []*compute.Backend{
				{
					Group: "http://ig-2",
				},
			},
			Fingerprint: "fp",
		})

	gock.New("https://compute.googleapis.com/compute/v1/projects/my-project/regions/us-central1/instanceGroups/ig-1?alt=json&fields=selfLink&prettyPrint=false").
		Reply(http.StatusOK).
		JSON(compute.InstanceGroup{
			SelfLink: "http://ig-1",
		})

	gock.New("https://compute.googleapis.com/compute/v1/projects/my-project/global/backendServices/bes-1?alt=json&prettyPrint=false").
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

	gock.New("https://compute.googleapis.com/compute/v1/projects/my-project/global/operations/op1?alt=json&fields=status%2Cerror&prettyPrint=false").
		Reply(http.StatusOK).
		JSON(compute.Operation{
			Status: "DONE",
		})

	gce, err := compute.NewService(
		context.Background(),
		option.WithHTTPClient(http.DefaultClient),
		option.WithoutAuthentication(),
	)
	if err != nil {
		t.Fatal(err)
	}

	s := NewService(gce)

	if err := s.Add(context.Background(), "my-project", "us-central1", "bes-1", "ig-1", false, 10*time.Millisecond); err != nil {
		t.Fatal(err)
	}
}

func TestService_AddExisting(t *testing.T) {
	defer gock.Off()

	gock.New("https://compute.googleapis.com/compute/v1/projects/my-project/global/backendServices/bes-1?alt=json&fields=backends%2Cfingerprint&prettyPrint=false").
		Reply(http.StatusOK).
		JSON(compute.BackendService{
			Backends: []*compute.Backend{
				{
					Group: "http://ig-1",
				},
			},
			Fingerprint: "fp",
		})

	gock.New("https://compute.googleapis.com/compute/v1/projects/my-project/regions/us-central1/instanceGroups/ig-1?alt=json&fields=selfLink&prettyPrint=false").
		Reply(http.StatusOK).
		JSON(compute.InstanceGroup{
			SelfLink: "http://ig-1",
		})

	gce, err := compute.NewService(
		context.Background(),
		option.WithHTTPClient(http.DefaultClient),
		option.WithoutAuthentication(),
	)
	if err != nil {
		t.Fatal(err)
	}

	s := NewService(gce)

	if err := s.Add(context.Background(), "my-project", "us-central1", "bes-1", "ig-1", false, 10*time.Millisecond); err != nil {
		t.Fatal(err)
	}
}

func TestService_AddDryRun(t *testing.T) {
	defer gock.Off()

	gock.New("https://compute.googleapis.com/compute/v1/projects/my-project/global/backendServices/bes-1?alt=json&fields=backends%2Cfingerprint&prettyPrint=false").
		Reply(http.StatusOK).
		JSON(compute.BackendService{
			Backends: []*compute.Backend{
				{
					Group: "http://ig-2",
				},
			},
			Fingerprint: "fp",
		})

	gock.New("https://compute.googleapis.com/compute/v1/projects/my-project/regions/us-central1/instanceGroups/ig-1?alt=json&fields=selfLink&prettyPrint=false").
		Reply(http.StatusOK).
		JSON(compute.InstanceGroup{
			SelfLink: "http://ig-1",
		})

	gce, err := compute.NewService(
		context.Background(),
		option.WithHTTPClient(http.DefaultClient),
		option.WithoutAuthentication(),
	)
	if err != nil {
		t.Fatal(err)
	}

	s := NewService(gce)

	if err := s.Add(context.Background(), "my-project", "us-central1", "bes-1", "ig-1", true, 10*time.Millisecond); err != nil {
		t.Fatal(err)
	}
}

func TestService_Remove(t *testing.T) {
	defer gock.Off()

	gock.New("https://compute.googleapis.com/compute/v1/projects/my-project/global/backendServices/bes-1?alt=json&fields=backends%2Cfingerprint&prettyPrint=false").
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

	gock.New("https://compute.googleapis.com/compute/v1/projects/my-project/regions/us-central1/instanceGroups/ig-1?alt=json&fields=selfLink&prettyPrint=false").
		Reply(http.StatusOK).
		JSON(compute.InstanceGroup{
			SelfLink: "http://ig-1",
		})

	gock.New("https://compute.googleapis.com/compute/v1/projects/my-project/global/backendServices/bes-1?alt=json&prettyPrint=false").
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

	gock.New("https://compute.googleapis.com/compute/v1/projects/my-project/global/operations/op1?alt=json&fields=status%2Cerror&prettyPrint=false").
		Reply(http.StatusOK).
		JSON(compute.Operation{
			Status: "DONE",
		})

	gce, err := compute.NewService(
		context.Background(),
		option.WithHTTPClient(http.DefaultClient),
		option.WithoutAuthentication(),
	)
	if err != nil {
		t.Fatal(err)
	}

	s := NewService(gce)

	if err := s.Remove(context.Background(), "my-project", "us-central1", "bes-1", "ig-1", false, 10*time.Millisecond); err != nil {
		t.Fatal(err)
	}
}

func TestSetup_RemoveLast(t *testing.T) {
	defer gock.Off()

	gock.New("https://compute.googleapis.com/compute/v1/projects/my-project/global/backendServices/bes-1?alt=json&fields=backends%2Cfingerprint&prettyPrint=false").
		Reply(http.StatusOK).
		JSON(compute.BackendService{
			Backends: []*compute.Backend{
				{
					Group: "http://ig-1",
				},
			},
			Fingerprint: "fp",
		})

	gock.New("https://compute.googleapis.com/compute/v1/projects/my-project/regions/us-central1/instanceGroups/ig-1?alt=json&fields=selfLink&prettyPrint=false").
		Reply(http.StatusOK).
		JSON(compute.InstanceGroup{
			SelfLink: "http://ig-1",
		})

	gock.New("https://compute.googleapis.com/compute/v1/projects/my-project/global/backendServices/bes-1?alt=json&prettyPrint=false").
		JSON(json.RawMessage(`{"backends":[],"fingerprint":"fp"}`)).
		Reply(http.StatusOK).
		JSON(compute.Operation{
			Name: "op1",
		})

	gock.New("https://compute.googleapis.com/compute/v1/projects/my-project/global/operations/op1?alt=json&fields=status%2Cerror&prettyPrint=false").
		Reply(http.StatusOK).
		JSON(compute.Operation{
			Status: "DONE",
		})

	gce, err := compute.NewService(
		context.Background(),
		option.WithHTTPClient(http.DefaultClient),
		option.WithoutAuthentication(),
	)
	if err != nil {
		t.Fatal(err)
	}

	s := NewService(gce)

	if err := s.Remove(context.Background(), "my-project", "us-central1", "bes-1", "ig-1", false, 10*time.Millisecond); err != nil {
		t.Fatal(err)
	}
}

func TestSetup_RemoveMissing(t *testing.T) {
	defer gock.Off()

	gock.New("https://compute.googleapis.com/compute/v1/projects/my-project/global/backendServices/bes-1?alt=json&fields=backends%2Cfingerprint&prettyPrint=false").
		Reply(http.StatusOK).
		JSON(compute.BackendService{
			Backends: []*compute.Backend{
				{
					Group: "http://ig-2",
				},
			},
			Fingerprint: "fp",
		})

	gock.New("https://compute.googleapis.com/compute/v1/projects/my-project/regions/us-central1/instanceGroups/ig-1?alt=json&fields=selfLink&prettyPrint=false").
		Reply(http.StatusOK).
		JSON(compute.InstanceGroup{
			SelfLink: "http://ig-1",
		})

	gce, err := compute.NewService(
		context.Background(),
		option.WithHTTPClient(http.DefaultClient),
		option.WithoutAuthentication(),
	)
	if err != nil {
		t.Fatal(err)
	}

	s := NewService(gce)

	if err := s.Remove(context.Background(), "my-project", "us-central1", "bes-1", "ig-1", false, 10*time.Millisecond); err != nil {
		t.Fatal(err)
	}
}

func TestSetup_RemoveDryRun(t *testing.T) {
	defer gock.Off()

	gock.New("https://compute.googleapis.com/compute/v1/projects/my-project/global/backendServices/bes-1?alt=json&fields=backends%2Cfingerprint&prettyPrint=false").
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

	gock.New("https://compute.googleapis.com/compute/v1/projects/my-project/regions/us-central1/instanceGroups/ig-1?alt=json&fields=selfLink&prettyPrint=false").
		Reply(http.StatusOK).
		JSON(compute.InstanceGroup{
			SelfLink: "http://ig-1",
		})

	gce, err := compute.NewService(
		context.Background(),
		option.WithHTTPClient(http.DefaultClient),
		option.WithoutAuthentication(),
	)
	if err != nil {
		t.Fatal(err)
	}

	s := NewService(gce)

	if err := s.Remove(context.Background(), "my-project", "us-central1", "bes-1", "ig-1", true, 10*time.Millisecond); err != nil {
		t.Fatal(err)
	}
}
