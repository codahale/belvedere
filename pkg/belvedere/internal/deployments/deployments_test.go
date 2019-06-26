package deployments

import (
	"context"
	"testing"
	"time"

	"github.com/codahale/belvedere/pkg/belvedere/internal/it"
	"github.com/codahale/belvedere/pkg/belvedere/internal/waiter"
	"github.com/google/go-cmp/cmp"
	compute "google.golang.org/api/compute/v0.beta"
	"google.golang.org/api/deploymentmanager/v2"
	"gopkg.in/h2non/gock.v1"
)

func TestCreate(t *testing.T) {
	defer gock.Off()
	it.MockTokenSource()

	gock.New("https://www.googleapis.com/deploymentmanager/v2/projects/my-project/global/deployments?alt=json&prettyPrint=false").
		JSON(deploymentmanager.Deployment{
			Name: "my-deployment",
			Labels: []*deploymentmanager.DeploymentLabelEntry{
				{
					Key:   "one",
					Value: "two",
				},
			},
			Target: &deploymentmanager.TargetConfiguration{
				Config: &deploymentmanager.ConfigFile{
					Content: `{"resources":[{"name":"my-instance","type":"compute.beta.instance","properties":{"machineType":"n1-standard-1"}}]}`,
				},
			},
		}).
		Reply(200).
		JSON(deploymentmanager.Operation{
			Name: "op1",
		})

	gock.New("https://www.googleapis.com/deploymentmanager/v2/projects/my-project/global/operations/op1?alt=json&fields=status%2Cerror&prettyPrint=false").
		Reply(200).
		JSON(deploymentmanager.Operation{
			Status: "DONE",
		})

	ctx := waiter.WithInterval(context.TODO(), 10*time.Millisecond)
	if err := Create(ctx, "my-project", "my-deployment",
		[]Resource{
			{
				Name: "my-instance",
				Type: "compute.beta.instance",
				Properties: compute.Instance{
					MachineType: "n1-standard-1",
				},
			},
		},
		map[string]string{
			"one": "two",
		},
		false); err != nil {
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
		Reply(200).
		JSON(deploymentmanager.Operation{
			Name: "op1",
		})

	gock.New("https://www.googleapis.com/deploymentmanager/v2/projects/my-project/global/operations/op1?alt=json&fields=status%2Cerror&prettyPrint=false").
		Reply(200).
		JSON(deploymentmanager.Operation{
			Status: "DONE",
		})

	ctx := waiter.WithInterval(context.TODO(), 10*time.Millisecond)
	if err := Update(ctx, "my-project", "my-deployment",
		[]Resource{
			{
				Name: "my-instance",
				Type: "compute.beta.instance",
				Properties: compute.Instance{
					MachineType: "n1-standard-1",
				},
			},
		},
		false); err != nil {
		t.Fatal(err)
	}
}

func TestDelete(t *testing.T) {
	defer gock.Off()
	it.MockTokenSource()

	gock.New("https://www.googleapis.com/deploymentmanager/v2/projects/my-project/global/deployments/my-deployment?alt=json&prettyPrint=false").
		Reply(200).
		JSON(deploymentmanager.Operation{
			Name: "op1",
		})

	gock.New("https://www.googleapis.com/deploymentmanager/v2/projects/my-project/global/operations/op1?alt=json&fields=status%2Cerror&prettyPrint=false").
		Reply(200).
		JSON(deploymentmanager.Operation{
			Status: "DONE",
		})

	ctx := waiter.WithInterval(context.TODO(), 10*time.Millisecond)
	if err := Delete(ctx, "my-project", "my-deployment", false, false); err != nil {
		t.Fatal(err)
	}
}

func TestList(t *testing.T) {
	defer gock.Off()
	it.MockTokenSource()

	gock.New("https://www.googleapis.com/deploymentmanager/v2/projects/my-project/global/deployments?alt=json&prettyPrint=false").
		Reply(200).
		JSON(deploymentmanager.DeploymentsListResponse{
			Deployments: []*deploymentmanager.Deployment{
				{
					Name: "one",
					Labels: []*deploymentmanager.DeploymentLabelEntry{
						{
							Key:   "one-a",
							Value: "a",
						},
						{
							Key:   "one-b",
							Value: "b",
						},
					},
				},
				{
					Name: "two",
					Labels: []*deploymentmanager.DeploymentLabelEntry{
						{
							Key:   "two-a",
							Value: "a",
						},
						{
							Key:   "two-b",
							Value: "b",
						},
					},
				},
			},
		})

	actual, err := List(context.TODO(), "my-project")
	if err != nil {
		t.Fatal(err)
	}

	expected := []map[string]string{
		{"name": "one", "one-a": "a", "one-b": "b"},
		{"name": "two", "two-a": "a", "two-b": "b"},
	}

	if !cmp.Equal(expected, actual) {
		t.Error(cmp.Diff(expected, actual))
	}
}
