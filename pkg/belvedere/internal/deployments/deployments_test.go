package deployments

import (
	"context"
	"testing"
	"time"

	"github.com/codahale/gubbins/assert"
	"github.com/codahale/gubbins/httpmock"
	"google.golang.org/api/compute/v1"
	"google.golang.org/api/deploymentmanager/v2"
	"google.golang.org/api/option"
)

func TestLabelsToEntries(t *testing.T) {
	t.Parallel()

	labels := Labels{
		Type:    "release",
		Region:  "us-west1",
		App:     "my-app",
		Release: "v1",
		Hash:    "12345",
	}
	got := labelsToEntries(&labels)

	want := []*deploymentmanager.DeploymentLabelEntry{
		{
			Key:   "belvedere-app",
			Value: "my-app",
		},
		{
			Key:   "belvedere-hash",
			Value: "12345",
		},
		{
			Key:   "belvedere-region",
			Value: "us-west1",
		},
		{
			Key:   "belvedere-release",
			Value: "v1",
		},
		{
			Key:   "belvedere-type",
			Value: "release",
		},
	}

	assert.Equal(t, "labelsToEntries()", want, got)
}

func TestEntriesToLabels(t *testing.T) {
	t.Parallel()

	got := entriesToLabels([]*deploymentmanager.DeploymentLabelEntry{
		{
			Key:   "belvedere-type",
			Value: "release",
		},
		{
			Key:   "belvedere-app",
			Value: "my-app",
		},
		{
			Key:   "belvedere-region",
			Value: "us-west1",
		},
		{
			Key:   "belvedere-release",
			Value: "v1",
		},
		{
			Key:   "belvedere-hash",
			Value: "12345",
		},
	})

	want := Labels{
		Type:    "release",
		Region:  "us-west1",
		App:     "my-app",
		Release: "v1",
		Hash:    "12345",
	}

	assert.Equal(t, "entriesToLabels()", want, got)
}

func TestManager_Insert(t *testing.T) {
	t.Parallel()

	srv := httpmock.NewServer(t)
	defer srv.Finish()

	srv.Expect(`/deploymentmanager/v2/projects/my-project/global/deployments?alt=json&prettyPrint=false`,
		httpmock.ReqJSON(deploymentmanager.Deployment{
			Name: "my-deployment",
			Labels: []*deploymentmanager.DeploymentLabelEntry{
				{
					Key:   "belvedere-type",
					Value: "base",
				},
			},
			Target: &deploymentmanager.TargetConfiguration{
				Config: &deploymentmanager.ConfigFile{
					Content: `{"resources":[{"name":"my-instance","type":"compute.v1.instance",` +
						`"properties":{"machineType":"n1-standard-1"}}]}`,
				},
			},
		}),
		httpmock.RespJSON(deploymentmanager.Operation{
			Name: "op1",
		}))

	srv.Expect(`/deploymentmanager/v2/projects/my-project/global/operations/op1?`+
		`alt=json&fields=status%2Cerror&prettyPrint=false`,
		httpmock.RespJSON(deploymentmanager.Operation{
			Status: "DONE",
		}))

	dm, err := NewManager(
		context.Background(),
		option.WithEndpoint(srv.URL()),
		option.WithoutAuthentication(),
	)
	if err != nil {
		t.Fatal(err)
	}

	if err := dm.Insert(context.Background(), "my-project", "my-deployment",
		[]Resource{
			{
				Name: "my-instance",
				Type: "compute.v1.instance",
				Properties: &compute.Instance{
					MachineType: "n1-standard-1",
				},
			},
		},
		Labels{
			Type: "base",
		},
		false, 10*time.Millisecond); err != nil {
		t.Fatal(err)
	}
}

func TestManager_Update(t *testing.T) {
	t.Parallel()

	srv := httpmock.NewServer(t)
	defer srv.Finish()

	srv.Expect(`/deploymentmanager/v2/projects/my-project/global/deployments/my-deployment?`+
		`alt=json&prettyPrint=false`,
		httpmock.ReqJSON(deploymentmanager.Deployment{
			Target: &deploymentmanager.TargetConfiguration{
				Config: &deploymentmanager.ConfigFile{
					Content: `{"resources":[{"name":"my-instance","type":"compute.v1.instance",` +
						`"properties":{"machineType":"n1-standard-1"}}]}`,
				},
			},
		}),
		httpmock.RespJSON(deploymentmanager.Operation{
			Name: "op1",
		}))

	srv.Expect(`/deploymentmanager/v2/projects/my-project/global/operations/op1?`+
		`alt=json&fields=status%2Cerror&prettyPrint=false`,
		httpmock.RespJSON(deploymentmanager.Operation{
			Status: "DONE",
		}))

	dm, err := NewManager(
		context.Background(),
		option.WithEndpoint(srv.URL()),
		option.WithoutAuthentication(),
	)
	if err != nil {
		t.Fatal(err)
	}

	if err := dm.Update(context.Background(), "my-project", "my-deployment",
		[]Resource{
			{
				Name: "my-instance",
				Type: "compute.v1.instance",
				Properties: &compute.Instance{
					MachineType: "n1-standard-1",
				},
			},
		},
		false, 10*time.Millisecond); err != nil {
		t.Fatal(err)
	}
}

func TestManager_Delete(t *testing.T) {
	t.Parallel()

	srv := httpmock.NewServer(t)
	defer srv.Finish()

	srv.Expect(`/deploymentmanager/v2/projects/my-project/global/deployments/my-deployment?`+
		`alt=json&prettyPrint=false`,
		httpmock.RespJSON(deploymentmanager.Operation{
			Name: "op1",
		}))

	srv.Expect(`/deploymentmanager/v2/projects/my-project/global/operations/op1?`+
		`alt=json&fields=status%2Cerror&prettyPrint=false`,
		httpmock.RespJSON(deploymentmanager.Operation{
			Status: "DONE",
		}))

	dm, err := NewManager(
		context.Background(),
		option.WithEndpoint(srv.URL()),
		option.WithoutAuthentication(),
	)
	if err != nil {
		t.Fatal(err)
	}

	if err := dm.Delete(
		context.Background(), "my-project", "my-deployment", false,
		false, 10*time.Millisecond,
	); err != nil {
		t.Fatal(err)
	}
}

func TestManager_List(t *testing.T) {
	t.Parallel()

	srv := httpmock.NewServer(t)
	defer srv.Finish()

	srv.Expect(`/deploymentmanager/v2/projects/my-project/global/deployments?`+
		`alt=json&filter=bobs+eq+1&prettyPrint=false`,
		httpmock.RespJSON(deploymentmanager.DeploymentsListResponse{
			Deployments: []*deploymentmanager.Deployment{
				{
					Name: "belvedere-base",
					Labels: []*deploymentmanager.DeploymentLabelEntry{
						{
							Key:   "belvedere-type",
							Value: "base",
						},
					},
				},
				{
					Name: "belvedere-my-app",
					Labels: []*deploymentmanager.DeploymentLabelEntry{
						{
							Key:   "belvedere-type",
							Value: "app",
						},
						{
							Key:   "belvedere-region",
							Value: "us-west1",
						},
						{
							Key:   "belvedere-app",
							Value: "my-app",
						},
					},
				},
			},
		}))

	dm, err := NewManager(
		context.Background(),
		option.WithEndpoint(srv.URL()),
		option.WithoutAuthentication(),
	)
	if err != nil {
		t.Fatal(err)
	}

	got, err := dm.List(context.Background(), "my-project", "bobs eq 1")
	if err != nil {
		t.Fatal(err)
	}

	want := []Deployment{
		{
			Name: "belvedere-base",
			Labels: Labels{
				Type: "base",
			},
		},
		{
			Name: "belvedere-my-app",
			Labels: Labels{
				Type:   "app",
				Region: "us-west1",
				App:    "my-app",
			},
		},
	}

	assert.Equal(t, "List()", want, got)
}

func TestManager_Get(t *testing.T) {
	t.Parallel()

	srv := httpmock.NewServer(t)
	defer srv.Finish()

	srv.Expect(`/deploymentmanager/v2/projects/my-project/global/deployments/belvedere-base?`+
		`alt=json&prettyPrint=false`,
		httpmock.RespJSON(&deploymentmanager.Deployment{
			Name: "belvedere-base",
			Labels: []*deploymentmanager.DeploymentLabelEntry{
				{
					Key:   "belvedere-type",
					Value: "base",
				},
			},
		}))

	dm, err := NewManager(
		context.Background(),
		option.WithEndpoint(srv.URL()),
		option.WithoutAuthentication(),
	)
	if err != nil {
		t.Fatal(err)
	}

	got, err := dm.Get(context.Background(), "my-project", "belvedere-base")
	if err != nil {
		t.Fatal(err)
	}

	want := &Deployment{
		Name: "belvedere-base",
		Labels: Labels{
			Type: "base",
		},
	}

	assert.Equal(t, "Get()", want, got)
}
