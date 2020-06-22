package belvedere

import (
	"context"
	"testing"
	"time"

	"github.com/codahale/belvedere/internal/it"
	"github.com/codahale/belvedere/pkg/belvedere/internal/deployments"
	"github.com/golang/mock/gomock"
	"gopkg.in/h2non/gock.v1"
)

func TestProject_Setup(t *testing.T) {
	defer gock.Off()
	it.MockTokenSource()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	r := NewResourceBuilder(ctrl)
	dm := NewDeploymentsManager(ctrl)
	s := NewSetupService(ctrl)
	res := []deployments.Resource{
		{
			Name: "example",
		},
	}

	s.EXPECT().
		EnableAPIs(gomock.Any(), "my-project", false, 10*time.Millisecond)
	s.EXPECT().
		SetDMPerms(gomock.Any(), "my-project", false)
	r.EXPECT().
		Base("dns.").
		Return(res)
	dm.EXPECT().
		Insert(gomock.Any(), "my-project", "belvedere", res, deployments.Labels{Type: "base"}, false, 10*time.Millisecond)

	p := &project{
		name:      "my-project",
		setup:     s,
		resources: r,
		dm:        dm,
	}

	if err := p.Setup(context.Background(), "dns", false, 10*time.Millisecond); err != nil {
		t.Fatal(err)
	}
}

func TestProject_Teardown(t *testing.T) {
	defer gock.Off()
	it.MockTokenSource()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dm := NewDeploymentsManager(ctrl)

	dm.EXPECT().
		Delete(gomock.Any(), "my-project", "belvedere", false, false, 10*time.Millisecond)

	p := &project{
		name: "my-project",
		dm:   dm,
	}

	if err := p.Teardown(context.Background(), false, false, 10*time.Millisecond); err != nil {
		t.Fatal(err)
	}
}
