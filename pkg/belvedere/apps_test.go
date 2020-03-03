package belvedere

import (
	"context"
	"testing"
	"time"

	"github.com/codahale/belvedere/pkg/belvedere/cfg"
	"github.com/codahale/belvedere/pkg/belvedere/internal/deployments"
	"github.com/codahale/belvedere/pkg/belvedere/internal/it"
	"github.com/golang/mock/gomock"
	"github.com/google/go-cmp/cmp"
	"google.golang.org/api/dns/v1"
	"gopkg.in/h2non/gock.v1"
)

func TestAppService_List(t *testing.T) {
	defer gock.Off()
	it.MockTokenSource()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dm := NewDeploymentsManager(ctrl)
	dm.EXPECT().
		List(gomock.Any(), "my-project", `labels.belvedere-type eq "app"`).
		Return([]deployments.Deployment{
			{
				Labels: deployments.Labels{
					Type:   "app",
					App:    "app1",
					Region: "us-west1",
				},
			},
		}, nil)

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

	resourceBuilder := NewResourceBuilder(ctrl)
	dm := NewDeploymentsManager(ctrl)
	setupService := NewSetupService(ctrl)

	mz := &dns.ManagedZone{}
	res := []deployments.Resource{
		{
			Name: "res",
		},
	}
	config := &cfg.Config{}

	setupService.EXPECT().
		ManagedZone(gomock.Any(), "my-project").
		Return(mz, nil)

	resourceBuilder.EXPECT().
		App("my-project", "my-app", mz, config).
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
	if err := apps.Create(context.TODO(), "us-west1", "my-app", config, false, 10*time.Millisecond); err != nil {
		t.Fatal(err)
	}
}

func TestAppService_Update(t *testing.T) {
	defer gock.Off()
	it.MockTokenSource()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	resourceBuilder := NewResourceBuilder(ctrl)
	dm := NewDeploymentsManager(ctrl)
	setupService := NewSetupService(ctrl)

	mz := &dns.ManagedZone{}
	res := []deployments.Resource{
		{
			Name: "res",
		},
	}
	config := &cfg.Config{}

	setupService.EXPECT().
		ManagedZone(gomock.Any(), "my-project").
		Return(mz, nil)

	resourceBuilder.EXPECT().
		App("my-project", "my-app", mz, config).
		Return(res)

	dm.EXPECT().
		Update(gomock.Any(), "my-project", "belvedere-my-app", res, false, 10*time.Millisecond)

	apps := &appService{
		project:   "my-project",
		dm:        dm,
		resources: resourceBuilder,
		setup:     setupService,
	}
	if err := apps.Update(context.TODO(), "my-app", config, false, 10*time.Millisecond); err != nil {
		t.Fatal(err)
	}
}

func TestAppService_Delete(t *testing.T) {
	defer gock.Off()
	it.MockTokenSource()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dm := NewDeploymentsManager(ctrl)

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
