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

	gock.New("https://www.googleapis.com/deploymentmanager/v2/projects/my-project/global/deployments?alt=json&filter=labels.belvedere-type+eq+%22app%22&prettyPrint=false").
		Reply(http.StatusOK).
		JSON(&deploymentmanager.DeploymentsListResponse{
			Deployments: []*deploymentmanager.Deployment{
				{
					Name: "belvedere-app1",
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

	as := &appService{project: "my-project"}

	actual, err := as.List(context.TODO())
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
