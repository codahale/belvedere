package backends

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/codahale/belvedere/internal/httpmock"
	"google.golang.org/api/compute/v1"
	"google.golang.org/api/option"
)

func TestService_Add(t *testing.T) {
	t.Parallel()

	srv := httpmock.NewServer(t)
	defer srv.Finish()

	srv.Expect(`/projects/my-project/global/backendServices/bes-1?`+
		`alt=json&fields=backends%2Cfingerprint&prettyPrint=false`,
		httpmock.RespJSON(compute.BackendService{
			Backends: []*compute.Backend{
				{
					Group: "http://ig-2",
				},
			},
			Fingerprint: "fp",
		}))

	srv.Expect(`/projects/my-project/regions/us-central1/instanceGroups/ig-1?`+
		`alt=json&fields=selfLink&prettyPrint=false`,
		httpmock.RespJSON(compute.InstanceGroup{
			SelfLink: "http://ig-1",
		}))

	srv.Expect(`/projects/my-project/global/backendServices/bes-1?alt=json&prettyPrint=false`,
		httpmock.ReqJSON(compute.BackendService{
			Backends: []*compute.Backend{
				{
					Group: "http://ig-2",
				},
				{
					Group: "http://ig-1",
				},
			},
			Fingerprint: "fp",
		}),
		httpmock.RespJSON(compute.Operation{
			Name: "op1",
		}))

	srv.Expect(`/projects/my-project/global/operations/op1?alt=json&fields=status%2Cerror&prettyPrint=false`,
		httpmock.RespJSON(compute.Operation{
			Status: "DONE",
		}))

	gce, err := compute.NewService(
		context.Background(),
		option.WithEndpoint(srv.URL()),
		option.WithoutAuthentication(),
	)
	if err != nil {
		t.Fatal(err)
	}

	s := NewService(gce)

	if err := s.Add(
		context.Background(), "my-project", "us-central1", "bes-1",
		"ig-1", false, 10*time.Millisecond,
	); err != nil {
		t.Fatal(err)
	}
}

func TestService_AddExisting(t *testing.T) {
	t.Parallel()

	srv := httpmock.NewServer(t)
	defer srv.Finish()

	srv.Expect(`/projects/my-project/global/backendServices/bes-1?`+
		`alt=json&fields=backends%2Cfingerprint&prettyPrint=false`,
		httpmock.RespJSON(compute.BackendService{
			Backends: []*compute.Backend{
				{
					Group: "http://ig-1",
				},
			},
			Fingerprint: "fp",
		}))

	srv.Expect(`/projects/my-project/regions/us-central1/instanceGroups/ig-1?`+
		`alt=json&fields=selfLink&prettyPrint=false`,
		httpmock.RespJSON(compute.InstanceGroup{
			SelfLink: "http://ig-1",
		}))

	gce, err := compute.NewService(
		context.Background(),
		option.WithEndpoint(srv.URL()),
		option.WithoutAuthentication(),
	)
	if err != nil {
		t.Fatal(err)
	}

	s := NewService(gce)

	if err := s.Add(
		context.Background(), "my-project", "us-central1", "bes-1",
		"ig-1", false, 10*time.Millisecond,
	); err != nil {
		t.Fatal(err)
	}
}

func TestService_AddDryRun(t *testing.T) {
	t.Parallel()

	srv := httpmock.NewServer(t)
	defer srv.Finish()

	srv.Expect(`/projects/my-project/global/backendServices/bes-1?`+
		`alt=json&fields=backends%2Cfingerprint&prettyPrint=false`,
		httpmock.RespJSON(compute.BackendService{
			Backends: []*compute.Backend{
				{
					Group: "http://ig-2",
				},
			},
			Fingerprint: "fp",
		}))

	srv.Expect(`/projects/my-project/regions/us-central1/instanceGroups/ig-1?`+
		`alt=json&fields=selfLink&prettyPrint=false`,
		httpmock.RespJSON(compute.InstanceGroup{
			SelfLink: "http://ig-1",
		}))

	gce, err := compute.NewService(
		context.Background(),
		option.WithEndpoint(srv.URL()),
		option.WithoutAuthentication(),
	)
	if err != nil {
		t.Fatal(err)
	}

	s := NewService(gce)

	if err := s.Add(
		context.Background(), "my-project", "us-central1", "bes-1",
		"ig-1", true, 10*time.Millisecond,
	); err != nil {
		t.Fatal(err)
	}
}

func TestService_Remove(t *testing.T) {
	t.Parallel()

	srv := httpmock.NewServer(t)
	defer srv.Finish()

	srv.Expect(`/projects/my-project/global/backendServices/bes-1?`+
		`alt=json&fields=backends%2Cfingerprint&prettyPrint=false`,
		httpmock.RespJSON(compute.BackendService{
			Backends: []*compute.Backend{
				{
					Group: "http://ig-1",
				},
				{
					Group: "http://ig-2",
				},
			},
			Fingerprint: "fp",
		}))

	srv.Expect(`/projects/my-project/regions/us-central1/instanceGroups/ig-1?`+
		`alt=json&fields=selfLink&prettyPrint=false`,
		httpmock.RespJSON(compute.InstanceGroup{
			SelfLink: "http://ig-1",
		}))

	srv.Expect(`/projects/my-project/global/backendServices/bes-1?alt=json&prettyPrint=false`,
		httpmock.ReqJSON(compute.BackendService{
			Backends: []*compute.Backend{
				{
					Group: "http://ig-2",
				},
			},
			Fingerprint: "fp",
		}),
		httpmock.RespJSON(compute.Operation{
			Name: "op1",
		}))

	srv.Expect(`/projects/my-project/global/operations/op1?alt=json&fields=status%2Cerror&prettyPrint=false`,
		httpmock.RespJSON(compute.Operation{
			Status: "DONE",
		}))

	gce, err := compute.NewService(
		context.Background(),
		option.WithEndpoint(srv.URL()),
		option.WithoutAuthentication(),
	)
	if err != nil {
		t.Fatal(err)
	}

	s := NewService(gce)

	if err := s.Remove(
		context.Background(), "my-project", "us-central1", "bes-1",
		"ig-1", false, 10*time.Millisecond,
	); err != nil {
		t.Fatal(err)
	}
}

func TestSetup_RemoveLast(t *testing.T) {
	t.Parallel()

	srv := httpmock.NewServer(t)
	defer srv.Finish()

	srv.Expect(`/projects/my-project/global/backendServices/bes-1?`+
		`alt=json&fields=backends%2Cfingerprint&prettyPrint=false`,
		httpmock.RespJSON(compute.BackendService{
			Backends: []*compute.Backend{
				{
					Group: "http://ig-1",
				},
			},
			Fingerprint: "fp",
		}))

	srv.Expect(`/projects/my-project/regions/us-central1/instanceGroups/ig-1?`+
		`alt=json&fields=selfLink&prettyPrint=false`,
		httpmock.RespJSON(compute.InstanceGroup{
			SelfLink: "http://ig-1",
		}))

	srv.Expect(`/projects/my-project/global/backendServices/bes-1?alt=json&prettyPrint=false`,
		httpmock.ReqJSON(json.RawMessage(`{"backends":[],"fingerprint":"fp"}`)),
		httpmock.RespJSON(compute.Operation{
			Name: "op1",
		}))

	srv.Expect(`/projects/my-project/global/operations/op1?alt=json&fields=status%2Cerror&prettyPrint=false`,
		httpmock.RespJSON(compute.Operation{
			Status: "DONE",
		}))

	gce, err := compute.NewService(
		context.Background(),
		option.WithEndpoint(srv.URL()),
		option.WithoutAuthentication(),
	)
	if err != nil {
		t.Fatal(err)
	}

	s := NewService(gce)

	if err := s.Remove(
		context.Background(), "my-project", "us-central1", "bes-1",
		"ig-1", false, 10*time.Millisecond,
	); err != nil {
		t.Fatal(err)
	}
}

func TestSetup_RemoveMissing(t *testing.T) {
	t.Parallel()

	srv := httpmock.NewServer(t)
	defer srv.Finish()

	srv.Expect(`/projects/my-project/global/backendServices/bes-1?`+
		`alt=json&fields=backends%2Cfingerprint&prettyPrint=false`,
		httpmock.RespJSON(compute.BackendService{
			Backends: []*compute.Backend{
				{
					Group: "http://ig-2",
				},
			},
			Fingerprint: "fp",
		}))

	srv.Expect(`/projects/my-project/regions/us-central1/instanceGroups/ig-1?`+
		`alt=json&fields=selfLink&prettyPrint=false`,
		httpmock.RespJSON(compute.InstanceGroup{
			SelfLink: "http://ig-1",
		}))

	gce, err := compute.NewService(
		context.Background(),
		option.WithEndpoint(srv.URL()),
		option.WithoutAuthentication(),
	)
	if err != nil {
		t.Fatal(err)
	}

	s := NewService(gce)

	if err := s.Remove(
		context.Background(), "my-project", "us-central1", "bes-1",
		"ig-1", false, 10*time.Millisecond,
	); err != nil {
		t.Fatal(err)
	}
}

func TestSetup_RemoveDryRun(t *testing.T) {
	t.Parallel()

	srv := httpmock.NewServer(t)
	defer srv.Finish()

	srv.Expect(`/projects/my-project/global/backendServices/bes-1?`+
		`alt=json&fields=backends%2Cfingerprint&prettyPrint=false`,
		httpmock.RespJSON(compute.BackendService{
			Backends: []*compute.Backend{
				{
					Group: "http://ig-1",
				},
				{
					Group: "http://ig-2",
				},
			},
			Fingerprint: "fp",
		}))

	srv.Expect(`/projects/my-project/regions/us-central1/instanceGroups/ig-1?`+
		`alt=json&fields=selfLink&prettyPrint=false`,
		httpmock.RespJSON(compute.InstanceGroup{
			SelfLink: "http://ig-1",
		}))

	gce, err := compute.NewService(
		context.Background(),
		option.WithEndpoint(srv.URL()),
		option.WithoutAuthentication(),
	)
	if err != nil {
		t.Fatal(err)
	}

	s := NewService(gce)

	if err := s.Remove(
		context.Background(), "my-project", "us-central1", "bes-1",
		"ig-1", true, 10*time.Millisecond,
	); err != nil {
		t.Fatal(err)
	}
}
