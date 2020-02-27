package belvedere

import (
	"context"
	"net/http"
	"testing"

	"github.com/codahale/belvedere/pkg/belvedere/internal/it"
	"github.com/google/go-cmp/cmp"
	deploymentmanager "google.golang.org/api/deploymentmanager/v2beta"
	"gopkg.in/h2non/gock.v1"
)

func TestApps(t *testing.T) {
	defer gock.Off()
	it.MockTokenSource()

	gock.New("https://www.googleapis.com/deploymentmanager/v2/projects/my-project/global/deployments?alt=json&prettyPrint=false").
		Reply(http.StatusOK).
		JSON(&deploymentmanager.DeploymentsListResponse{
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
					Name: "random-one",
					Labels: []*deploymentmanager.DeploymentLabelEntry{
						{
							Key:   "alphabet",
							Value: "soup",
						},
					},
				},
				{
					Name: "belvedere-base",
					Labels: []*deploymentmanager.DeploymentLabelEntry{
						{
							Key:   "belvedere-type",
							Value: "app",
						},
						{
							Key:   "belvedere-app",
							Value: "app1",
						},
						{
							Key:   "belvedere-region",
							Value: "us-west1",
						},
					},
				},
			},
		})

	actual, err := Apps(context.TODO(), "my-project")
	if err != nil {
		t.Fatal(err)
	}

	expected := []App{
		{
			Name:    "app1",
			Project: "my-project",
			Region:  "us-west1",
		},
	}

	if !cmp.Equal(expected, actual) {
		t.Fatal(cmp.Diff(expected, actual))
	}
}
