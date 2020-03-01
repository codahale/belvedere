package belvedere

import (
	"context"
	"testing"

	"github.com/codahale/belvedere/pkg/belvedere/internal/deployments"
	"github.com/codahale/belvedere/pkg/belvedere/internal/it"
	"github.com/codahale/belvedere/pkg/belvedere/internal/mocks"
	"github.com/golang/mock/gomock"
	"github.com/google/go-cmp/cmp"
	"gopkg.in/h2non/gock.v1"
)

func TestReleaseService_List(t *testing.T) {
	defer gock.Off()
	it.MockTokenSource()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dm := mocks.NewDeploymentsManager(ctrl)

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

	dm := mocks.NewDeploymentsManager(ctrl)

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
