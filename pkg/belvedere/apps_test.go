package belvedere

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/codahale/belvedere/pkg/belvedere/internal/deployments"
	"github.com/codahale/belvedere/pkg/belvedere/internal/it"
	"github.com/codahale/belvedere/pkg/belvedere/internal/mocks"
	"github.com/golang/mock/gomock"
	"github.com/google/go-cmp/cmp"
	deploymentmanager "google.golang.org/api/deploymentmanager/v2beta"
	"google.golang.org/api/dns/v1"
	"gopkg.in/h2non/gock.v1"
)

func TestAppService_List(t *testing.T) {
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

	dm, err := deployments.NewManager(context.TODO())
	if err != nil {
		t.Fatal(err)
	}

	as := &appService{
		project: "my-project",
		dm:      dm,
	}

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

func TestAppService_Create(t *testing.T) {
	defer gock.Off()
	it.MockTokenSource()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	resourceBuilder := mocks.NewResourceBuilder(ctrl)
	dm := mocks.NewDeploymentsManager(ctrl)
	setupService := mocks.NewSetupService(ctrl)

	mz := &dns.ManagedZone{}
	res := []deployments.Resource{
		{
			Name: "res",
		},
	}
	config := Config{}

	setupService.EXPECT().
		ManagedZone(gomock.Any(), "my-project").
		Return(mz, nil)

	resourceBuilder.EXPECT().
		App("my-project", "my-app", mz, config.CDNPolicy, config.IAP, config.IAMRoles).
		Return(res)

	dm.EXPECT().
		Insert(gomock.Any(), "my-project", "belvedere-my-app", res,
			deployments.Labels{
				Type:   "app",
				App:    "my-app",
				Region: "us-west1",
			},
			false, 10*time.Millisecond)

	apps := &appService{
		project:   "my-project",
		dm:        dm,
		resources: resourceBuilder,
		setup:     setupService,
	}
	if err := apps.Create(context.TODO(), "us-west1", "my-app", &config, false, 10*time.Millisecond); err != nil {
		t.Fatal(err)
	}
}

func TestAppService_Update(t *testing.T) {
	defer gock.Off()
	it.MockTokenSource()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	resourceBuilder := mocks.NewResourceBuilder(ctrl)
	dm := mocks.NewDeploymentsManager(ctrl)
	setupService := mocks.NewSetupService(ctrl)

	mz := &dns.ManagedZone{}
	res := []deployments.Resource{
		{
			Name: "res",
		},
	}
	config := Config{}

	setupService.EXPECT().
		ManagedZone(gomock.Any(), "my-project").
		Return(mz, nil)

	resourceBuilder.EXPECT().
		App("my-project", "my-app", mz, config.CDNPolicy, config.IAP, config.IAMRoles).
		Return(res)

	dm.EXPECT().
		Update(gomock.Any(), "my-project", "belvedere-my-app", res, false, 10*time.Millisecond)

	apps := &appService{
		project:   "my-project",
		dm:        dm,
		resources: resourceBuilder,
		setup:     setupService,
	}
	if err := apps.Update(context.TODO(), "my-app", &config, false, 10*time.Millisecond); err != nil {
		t.Fatal(err)
	}
}

func TestAppService_Delete(t *testing.T) {
	defer gock.Off()
	it.MockTokenSource()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dm := mocks.NewDeploymentsManager(ctrl)

	dm.EXPECT().
		Delete(gomock.Any(), "my-project", "belvedere-my-app", false, false, 10*time.Millisecond)

	apps := &appService{
		project: "my-project",
		dm:      dm,
	}
	if err := apps.Delete(context.TODO(), "my-app", false, false, 10*time.Millisecond); err != nil {
		t.Fatal(err)
	}
}
