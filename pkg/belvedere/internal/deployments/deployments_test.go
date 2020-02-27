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

func TestLabelsToEntries(t *testing.T) {
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

func TestLabelsFromEntries(t *testing.T) {
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

func TestInsert(t *testing.T) {
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

	if err := Insert(context.TODO(), "my-project", "my-deployment",
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

func TestUpdate(t *testing.T) {
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

	if err := Update(context.TODO(), "my-project", "my-deployment",
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

func TestDelete(t *testing.T) {
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

	if err := Delete(context.TODO(), "my-project", "my-deployment", false, false, 10*time.Millisecond); err != nil {
		t.Fatal(err)
	}
}

func TestList(t *testing.T) {
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

	actual, err := List(context.TODO(), "my-project", "bobs eq 1")
	if err != nil {
		t.Fatal(err)
	}

	expected := map[string]Labels{
		"belvedere-base": {
			Type: "base",
		},
		"belvedere-my-app": {
			Type:   "app",
			Region: "us-west1",
			App:    "my-app",
		},
	}

	if !cmp.Equal(expected, actual) {
		t.Error(cmp.Diff(expected, actual))
	}
}
