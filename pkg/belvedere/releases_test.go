package belvedere

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/codahale/belvedere/internal/assert"
	"github.com/codahale/belvedere/pkg/belvedere/cfg"
	"github.com/codahale/belvedere/pkg/belvedere/internal/deployments"
	"github.com/golang/mock/gomock"
)

func TestReleaseService_List(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dm := NewDeploymentsManager(ctrl)

	dm.EXPECT().
		List(gomock.Any(), "my-project", `labels.belvedere-type eq "release"`).
		Return([]deployments.Deployment{
			{
				Labels: deployments.Labels{
					Type:    "release",
					App:     "my-app",
					Region:  "us-west1",
					Release: "v1",
					Hash:    "123456",
				},
			},
		}, nil)

	service := &releaseService{
		project: "my-project",
		dm:      dm,
	}
	got, err := service.List(context.Background(), "")
	if err != nil {
		t.Fatal(err)
	}

	want := []Release{
		{
			Project: "my-project",
			App:     "my-app",
			Region:  "us-west1",
			Release: "v1",
			Hash:    "123456",
		},
	}

	assert.Equal(t, "List()", want, got)
}

func TestReleaseService_List_withApp(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dm := NewDeploymentsManager(ctrl)

	dm.EXPECT().
		List(
			gomock.Any(),
			"my-project",
			`labels.belvedere-type eq "release" AND labels.belvedere-app eq "my-app"`,
		).
		Return([]deployments.Deployment{
			{
				Labels: deployments.Labels{
					Type:    "release",
					App:     "my-app",
					Region:  "us-west1",
					Release: "v1",
					Hash:    "123456",
				},
			},
		}, nil)

	service := &releaseService{
		project: "my-project",
		dm:      dm,
	}
	got, err := service.List(context.Background(), "my-app")
	if err != nil {
		t.Fatal(err)
	}

	want := []Release{
		{
			Project: "my-project",
			App:     "my-app",
			Region:  "us-west1",
			Release: "v1",
			Hash:    "123456",
		},
	}

	assert.Equal(t, "List()", want, got)
}

func TestReleaseService_Create(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	res := []deployments.Resource{
		{
			Name: "res",
		},
	}

	config := &cfg.Config{}
	imageSHA256 := strings.Repeat("1", 64)

	dm := NewDeploymentsManager(ctrl)
	dm.EXPECT().
		Insert(gomock.Any(), "my-project", "belvedere-my-app-v1",
			res, deployments.Labels{
				Type:    "release",
				Region:  "us-west1",
				App:     "my-app",
				Release: "v1",
				Hash:    strings.Repeat("1", 32),
			}, false, 10*time.Millisecond)

	apps := NewMockAppService(ctrl)
	apps.EXPECT().
		Get(gomock.Any(), "my-app").
		Return(&App{
			Region: "us-west1",
		}, nil)

	resourceBuilder := NewResourceBuilder(ctrl)
	resourceBuilder.EXPECT().
		Release("my-project", "us-west1", "my-app", "v1", imageSHA256, config).
		Return(res)

	service := &releaseService{
		project:   "my-project",
		dm:        dm,
		resources: resourceBuilder,
		apps:      apps,
	}

	if err := service.Create(
		context.Background(), "my-app", "v1", config, imageSHA256, false, 10*time.Millisecond,
	); err != nil {
		t.Fatal(err)
	}
}

func TestReleaseService_Enable(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	apps := NewMockAppService(ctrl)
	apps.EXPECT().
		Get(gomock.Any(), "my-app").
		Return(&App{
			Region: "us-west1",
		}, nil)

	hc := NewMockHealthChecker(ctrl)
	hc.EXPECT().
		Poll(gomock.Any(), "my-project", "us-west1", "my-app-bes", "my-app-v1-ig", 10*time.Millisecond)

	backendsService := NewBackendsService(ctrl)
	backendsService.EXPECT().
		Add(gomock.Any(), "my-project", "us-west1", "my-app-bes", "my-app-v1-ig", false, 10*time.Millisecond)

	service := &releaseService{
		project:  "my-project",
		apps:     apps,
		backends: backendsService,
		health:   hc,
	}

	if err := service.Enable(
		context.Background(), "my-app", "v1", false, 10*time.Millisecond,
	); err != nil {
		t.Fatal(err)
	}
}

func TestReleaseService_Disable(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	apps := NewMockAppService(ctrl)
	apps.EXPECT().
		Get(gomock.Any(), "my-app").
		Return(&App{
			Region: "us-west1",
		}, nil)

	backendsService := NewBackendsService(ctrl)
	backendsService.EXPECT().
		Remove(gomock.Any(), "my-project", "us-west1", "my-app-bes", "my-app-v1-ig", false, 10*time.Millisecond)

	service := &releaseService{
		project:  "my-project",
		apps:     apps,
		backends: backendsService,
	}

	if err := service.Disable(
		context.Background(), "my-app", "v1", false, 10*time.Millisecond,
	); err != nil {
		t.Fatal(err)
	}
}

func TestReleaseService_Delete(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dm := NewDeploymentsManager(ctrl)

	dm.EXPECT().
		Delete(gomock.Any(), "my-project", "belvedere-my-app-v1", false, false, 10*time.Millisecond)

	service := &releaseService{
		project: "my-project",
		dm:      dm,
	}

	if err := service.Delete(
		context.Background(), "my-app", "v1", false, false, 10*time.Millisecond,
	); err != nil {
		t.Fatal(err)
	}
}
