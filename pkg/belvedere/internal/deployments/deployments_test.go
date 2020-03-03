package deployments

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/codahale/belvedere/pkg/belvedere/internal/it"
	"github.com/google/go-cmp/cmp"
	compute "google.golang.org/api/compute/v0.beta"
	"google.golang.org/api/deploymentmanager/v2"
	"gopkg.in/h2non/gock.v1"
)

func TestLabels_toEntries(t *testing.T) {
	expected := []*deploymentmanager.DeploymentLabelEntry{
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

	labels := Labels{
		Type:    "release",
		Region:  "us-west1",
		App:     "my-app",
		Release: "v1",
		Hash:    "12345",
	}
	actual := labels.toEntries()

	if !cmp.Equal(expected, actual) {
		t.Fatal(cmp.Diff(expected, actual))
	}
}

func TestLabels_fromEntries(t *testing.T) {
	var actual Labels
	actual.fromEntries([]*deploymentmanager.DeploymentLabelEntry{
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

	expected := Labels{
		Type:    "release",
		Region:  "us-west1",
		App:     "my-app",
		Release: "v1",
		Hash:    "12345",
	}

	if !cmp.Equal(expected, actual) {
		t.Fatal(cmp.Diff(expected, actual))
	}
}

func TestManager_Insert(t *testing.T) {
	defer gock.Off()
	it.MockTokenSource()

	gock.New("https://www.googleapis.com/deploymentmanager/v2/projects/my-project/global/deployments?alt=json&prettyPrint=false").
		JSON(deploymentmanager.Deployment{
			Name: "my-deployment",
			Labels: []*deploymentmanager.DeploymentLabelEntry{
				{
					Key:   "belvedere-type",
					Value: "base",
				},
			},
			Target: &deploymentmanager.TargetConfiguration{
				Config: &deploymentmanager.ConfigFile{
					Content: `{"resources":[{"name":"my-instance","type":"compute.beta.instance","properties":{"machineType":"n1-standard-1"}}]}`,
				},
			},
		}).
		Reply(http.StatusOK).
		JSON(deploymentmanager.Operation{
			Name: "op1",
		})

	gock.New("https://www.googleapis.com/deploymentmanager/v2/projects/my-project/global/operations/op1?alt=json&fields=status%2Cerror&prettyPrint=false").
		Reply(http.StatusOK).
		JSON(deploymentmanager.Operation{
			Status: "DONE",
		})

	dm, err := NewManager(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	if err := dm.Insert(context.Background(), "my-project", "my-deployment",
		[]Resource{
			{
				Name: "my-instance",
				Type: "compute.beta.instance",
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
	defer gock.Off()
	it.MockTokenSource()

	gock.New("https://www.googleapis.com/deploymentmanager/v2/projects/my-project/global/deployments/my-deployment?alt=json&prettyPrint=false").
		JSON(deploymentmanager.Deployment{
			Target: &deploymentmanager.TargetConfiguration{
				Config: &deploymentmanager.ConfigFile{
					Content: `{"resources":[{"name":"my-instance","type":"compute.beta.instance","properties":{"machineType":"n1-standard-1"}}]}`,
				},
			},
		}).
		Reply(http.StatusOK).
		JSON(deploymentmanager.Operation{
			Name: "op1",
		})

	gock.New("https://www.googleapis.com/deploymentmanager/v2/projects/my-project/global/operations/op1?alt=json&fields=status%2Cerror&prettyPrint=false").
		Reply(http.StatusOK).
		JSON(deploymentmanager.Operation{
			Status: "DONE",
		})

	dm, err := NewManager(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	if err := dm.Update(context.Background(), "my-project", "my-deployment",
		[]Resource{
			{
				Name: "my-instance",
				Type: "compute.beta.instance",
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
	defer gock.Off()
	it.MockTokenSource()

	gock.New("https://www.googleapis.com/deploymentmanager/v2/projects/my-project/global/deployments/my-deployment?alt=json&prettyPrint=false").
		Reply(http.StatusOK).
		JSON(deploymentmanager.Operation{
			Name: "op1",
		})

	gock.New("https://www.googleapis.com/deploymentmanager/v2/projects/my-project/global/operations/op1?alt=json&fields=status%2Cerror&prettyPrint=false").
		Reply(http.StatusOK).
		JSON(deploymentmanager.Operation{
			Status: "DONE",
		})

	dm, err := NewManager(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	if err := dm.Delete(context.Background(), "my-project", "my-deployment", false, false, 10*time.Millisecond); err != nil {
		t.Fatal(err)
	}
}

func TestManager_List(t *testing.T) {
	defer gock.Off()
	it.MockTokenSource()

	gock.New("https://www.googleapis.com/deploymentmanager/v2/projects/my-project/global/deployments?alt=json&filter=bobs+eq+1&prettyPrint=false").
		Reply(http.StatusOK).
		JSON(deploymentmanager.DeploymentsListResponse{
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
		})

	dm, err := NewManager(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	actual, err := dm.List(context.Background(), "my-project", "bobs eq 1")
	if err != nil {
		t.Fatal(err)
	}

	expected := []Deployment{
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

	if !cmp.Equal(expected, actual) {
		t.Error(cmp.Diff(expected, actual))
	}
}

func TestManager_Get(t *testing.T) {
	defer gock.Off()
	it.MockTokenSource()

	gock.New("https://www.googleapis.com/deploymentmanager/v2/projects/my-project/global/deployments/belvedere-base?alt=json&fields=&prettyPrint=false").
		Reply(http.StatusOK).
		JSON(&deploymentmanager.Deployment{
			Name: "belvedere-base",
			Labels: []*deploymentmanager.DeploymentLabelEntry{
				{
					Key:   "belvedere-type",
					Value: "base",
				},
			},
		})

	dm, err := NewManager(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	actual, err := dm.Get(context.Background(), "my-project", "belvedere-base")
	if err != nil {
		t.Fatal(err)
	}

	expected := &Deployment{
		Name: "belvedere-base",
		Labels: Labels{
			Type: "base",
		},
	}

	if !cmp.Equal(expected, actual) {
		t.Error(cmp.Diff(expected, actual))
	}
}
