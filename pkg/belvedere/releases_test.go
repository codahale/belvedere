package belvedere

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/codahale/belvedere/pkg/belvedere/cfg"
	"github.com/codahale/belvedere/pkg/belvedere/internal/deployments"
	"github.com/codahale/belvedere/pkg/belvedere/internal/it"
	"github.com/golang/mock/gomock"
	"github.com/google/go-cmp/cmp"
	"gopkg.in/h2non/gock.v1"
)

func TestReleaseService_List(t *testing.T) {
	defer gock.Off()
	it.MockTokenSource()

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
	actual, err := service.List(context.TODO(), "")
	if err != nil {
		t.Fatal(err)
	}

	expected := []Release{
		{
			Project: "my-project",
			App:     "my-app",
			Region:  "us-west1",
			Release: "v1",
			Hash:    "123456",
		},
	}

	if !cmp.Equal(expected, actual) {
		t.Fatal(cmp.Diff(expected, actual))
	}
}

func TestReleaseService_List_withApp(t *testing.T) {
	defer gock.Off()
	it.MockTokenSource()

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
	actual, err := service.List(context.TODO(), "my-app")
	if err != nil {
		t.Fatal(err)
	}

	expected := []Release{
		{
			Project: "my-project",
			App:     "my-app",
			Region:  "us-west1",
			Release: "v1",
			Hash:    "123456",
		},
	}

	if !cmp.Equal(expected, actual) {
		t.Fatal(cmp.Diff(expected, actual))
	}
}

func TestReleaseService_Create(t *testing.T) {
	defer gock.Off()
	it.MockTokenSource()

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
		Get(gomock.Any(), "my-project", "belvedere-my-app").
		Return(&deployments.Deployment{
			Labels: deployments.Labels{
				Region: "us-west1",
			},
		}, nil)
	dm.EXPECT().
		Insert(gomock.Any(), "my-project", "belvedere-my-app-v1",
			res, deployments.Labels{
				Type:    "release",
				Region:  "us-west1",
				App:     "my-app",
				Release: "v1",
				Hash:    strings.Repeat("1", 32),
			}, false, 10*time.Millisecond)

	resourceBuilder := NewResourceBuilder(ctrl)
	resourceBuilder.EXPECT().
		Release("my-project", "us-west1", "my-app", "v1", imageSHA256, config).
		Return(res)

	service := &releaseService{
		project:   "my-project",
		dm:        dm,
		resources: resourceBuilder,
	}

	if err := service.Create(
		context.TODO(), "my-app", "v1", config, imageSHA256, false, 10*time.Millisecond,
	); err != nil {
		t.Fatal()
	}
}

func TestReleaseService_Enable(t *testing.T) {
	defer gock.Off()
	it.MockTokenSource()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dm := NewDeploymentsManager(ctrl)
	dm.EXPECT().
		Get(gomock.Any(), "my-project", "belvedere-my-app").
		Return(&deployments.Deployment{
			Labels: deployments.Labels{
				Region: "us-west1",
			},
		}, nil)

	hc := NewMockHealthChecker(ctrl)
	hc.EXPECT().
		Poll(gomock.Any(), "my-project", "us-west1", "my-app-bes", "my-app-v1-ig", 10*time.Millisecond)

	backendsService := NewBackendsService(ctrl)
	backendsService.EXPECT().
		Add(gomock.Any(), "my-project", "us-west1", "my-app-bes", "my-app-v1-ig", false, 10*time.Millisecond)

	service := &releaseService{
		project:       "my-project",
		dm:            dm,
		backends:      backendsService,
		healthChecker: hc,
	}

	if err := service.Enable(
		context.TODO(), "my-app", "v1", false, 10*time.Millisecond,
	); err != nil {
		t.Fatal()
	}
}

func TestReleaseService_Disable(t *testing.T) {
	defer gock.Off()
	it.MockTokenSource()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dm := NewDeploymentsManager(ctrl)
	dm.EXPECT().
		Get(gomock.Any(), "my-project", "belvedere-my-app").
		Return(&deployments.Deployment{
			Labels: deployments.Labels{
				Region: "us-west1",
			},
		}, nil)

	backendsService := NewBackendsService(ctrl)
	backendsService.EXPECT().
		Remove(gomock.Any(), "my-project", "us-west1", "my-app-bes", "my-app-v1-ig", false, 10*time.Millisecond)

	service := &releaseService{
		project:  "my-project",
		dm:       dm,
		backends: backendsService,
	}

	if err := service.Disable(
		context.TODO(), "my-app", "v1", false, 10*time.Millisecond,
	); err != nil {
		t.Fatal()
	}
}

func TestReleaseService_Delete(t *testing.T) {
	defer gock.Off()
	it.MockTokenSource()

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
		context.TODO(), "my-app", "v1", false, false, 10*time.Millisecond,
	); err != nil {
		t.Fatal(err)
	}
}
