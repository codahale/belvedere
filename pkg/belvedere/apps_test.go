package belvedere

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/codahale/belvedere/internal/assert"
	"github.com/codahale/belvedere/internal/httpmock"
	"github.com/codahale/belvedere/pkg/belvedere/cfg"
	"github.com/codahale/belvedere/pkg/belvedere/internal/deployments"
	"github.com/golang/mock/gomock"
	"google.golang.org/api/compute/v1"
	"google.golang.org/api/dns/v1"
	"google.golang.org/api/option"
)

func TestAppService_Get(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dm := NewDeploymentsManager(ctrl)
	dm.EXPECT().
		Get(gomock.Any(), "my-project", "belvedere-my-app").
		Return(&deployments.Deployment{
			Labels: deployments.Labels{
				Type:   "app",
				App:    "app1",
				Region: "us-west1",
			},
		}, nil)

	as := &appService{
		project: "my-project",
		dm:      dm,
	}

	got, err := as.Get(context.Background(), "my-app")
	if err != nil {
		t.Fatal(err)
	}

	want := &App{
		Name:    "app1",
		Project: "my-project",
		Region:  "us-west1",
	}

	assert.Equal(t, "Get()", want, got)
}

func TestAppService_List(t *testing.T) {
	t.Parallel()

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

	got, err := as.List(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	want := []App{
		{
			Name:    "app1",
			Project: "my-project",
			Region:  "us-west1",
		},
	}

	assert.Equal(t, "List()", want, got)
}

func TestAppService_Create(t *testing.T) {
	t.Parallel()

	srv := httpmock.NewServer(t)
	defer srv.Finish()

	srv.Expect(`/projects/my-project/regions/us-west1?alt=json&fields=status&prettyPrint=false`,
		httpmock.RespJSON(compute.Region{
			Status: "UP",
		}))

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

	gce, err := compute.NewService(
		context.Background(),
		option.WithEndpoint(srv.URL()),
		option.WithoutAuthentication(),
	)
	if err != nil {
		t.Fatal(err)
	}

	apps := &appService{
		project:   "my-project",
		dm:        dm,
		resources: resourceBuilder,
		setup:     setupService,
		gce:       gce,
	}
	if err := apps.Create(context.Background(), "us-west1", "my-app", config, false, 10*time.Millisecond); err != nil {
		t.Fatal(err)
	}
}

func TestAppService_Create_DownRegion(t *testing.T) {
	t.Parallel()

	srv := httpmock.NewServer(t)
	defer srv.Finish()

	srv.Expect(`/projects/my-project/regions/us-west1?alt=json&fields=status&prettyPrint=false`,
		httpmock.RespJSON(compute.Region{
			Status: "DOWN",
		}))

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	resourceBuilder := NewResourceBuilder(ctrl)
	dm := NewDeploymentsManager(ctrl)
	setupService := NewSetupService(ctrl)

	config := &cfg.Config{}

	gce, err := compute.NewService(
		context.Background(),
		option.WithEndpoint(srv.URL()),
		option.WithoutAuthentication(),
	)
	if err != nil {
		t.Fatal(err)
	}

	apps := &appService{
		project:   "my-project",
		dm:        dm,
		resources: resourceBuilder,
		setup:     setupService,
		gce:       gce,
	}

	err = apps.Create(context.Background(), "us-west1", "my-app", config, false, 10*time.Millisecond)
	if err == nil {
		t.Fatal("no error")
	}
}

func TestAppService_Create_BadRegion(t *testing.T) {
	t.Parallel()

	srv := httpmock.NewServer(t)
	defer srv.Finish()

	srv.Expect(`/projects/my-project/regions/us-west1?alt=json&fields=status&prettyPrint=false`,
		httpmock.Status(http.StatusNotFound))

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	resourceBuilder := NewResourceBuilder(ctrl)
	dm := NewDeploymentsManager(ctrl)
	setupService := NewSetupService(ctrl)

	config := &cfg.Config{}

	gce, err := compute.NewService(
		context.Background(),
		option.WithEndpoint(srv.URL()),
		option.WithoutAuthentication(),
	)
	if err != nil {
		t.Fatal(err)
	}

	apps := &appService{
		project:   "my-project",
		dm:        dm,
		resources: resourceBuilder,
		setup:     setupService,
		gce:       gce,
	}

	err = apps.Create(context.Background(), "us-west1", "my-app", config, false, 10*time.Millisecond)
	if err == nil {
		t.Fatal("no error")
	}
}

func TestAppService_Update(t *testing.T) {
	t.Parallel()

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

	if err := apps.Update(context.Background(), "my-app", config, false, 10*time.Millisecond); err != nil {
		t.Fatal(err)
	}
}

func TestAppService_Delete(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dm := NewDeploymentsManager(ctrl)

	dm.EXPECT().
		Delete(gomock.Any(), "my-project", "belvedere-my-app", false, false, 10*time.Millisecond)

	apps := &appService{
		project: "my-project",
		dm:      dm,
	}

	if err := apps.Delete(context.Background(), "my-app", false, false, 10*time.Millisecond); err != nil {
		t.Fatal(err)
	}
}
