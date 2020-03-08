package deletecmd

import (
	"context"
	"testing"
	"time"

	"github.com/codahale/belvedere/cmd/belvedere/internal/mocks"
	"github.com/codahale/belvedere/cmd/belvedere/internal/rootcmd"
	"github.com/golang/mock/gomock"
)

func TestAppsDelete(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	apps := mocks.NewMockAppService(ctrl)
	apps.EXPECT().
		Delete(gomock.Any(), "my-app", false, false, 10*time.Second)

	project := mocks.NewMockProject(ctrl)
	project.EXPECT().
		Apps().
		Return(apps).
		AnyTimes()

	command := New(&rootcmd.Config{
		Project: project,
	})
	if err := command.ParseAndRun(context.Background(), []string{
		"my-app",
	}); err != nil {
		t.Fatal(err)
	}
}

func TestAppsDelete_Flags(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	apps := mocks.NewMockAppService(ctrl)
	apps.EXPECT().
		Delete(gomock.Any(), "my-app", true, true, 10*time.Hour)

	project := mocks.NewMockProject(ctrl)
	project.EXPECT().
		Apps().
		Return(apps).
		AnyTimes()

	command := New(&rootcmd.Config{
		Project: project,
	})
	if err := command.ParseAndRun(context.Background(), []string{
		"-dry-run",
		"-async",
		"-interval=10h",
		"my-app",
	}); err != nil {
		t.Fatal(err)
	}
}
