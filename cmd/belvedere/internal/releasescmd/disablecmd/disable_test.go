package disablecmd

import (
	"context"
	"testing"
	"time"

	"github.com/codahale/belvedere/cmd/belvedere/internal/mocks"
	"github.com/codahale/belvedere/cmd/belvedere/internal/rootcmd"
	"github.com/golang/mock/gomock"
)

func TestReleasesDisable(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	releases := mocks.NewMockReleaseService(ctrl)
	releases.EXPECT().
		Disable(gomock.Any(), "my-app", "my-rel", false, 10*time.Second)

	project := mocks.NewMockProject(ctrl)
	project.EXPECT().
		Releases().
		Return(releases).
		AnyTimes()

	command := New(&rootcmd.Config{
		Project: project,
	})
	if err := command.ParseAndRun(context.Background(), []string{
		"my-app",
		"my-rel",
	}); err != nil {
		t.Fatal(err)
	}
}

func TestReleasesDisable_Flags(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	releases := mocks.NewMockReleaseService(ctrl)
	releases.EXPECT().
		Disable(gomock.Any(), "my-app", "my-rel", true, 10*time.Hour)

	project := mocks.NewMockProject(ctrl)
	project.EXPECT().
		Releases().
		Return(releases).
		AnyTimes()

	command := New(&rootcmd.Config{
		Project: project,
	})
	if err := command.ParseAndRun(context.Background(), []string{
		"-dry-run",
		"-interval=10h",
		"my-app",
		"my-rel",
	}); err != nil {
		t.Fatal(err)
	}
}
