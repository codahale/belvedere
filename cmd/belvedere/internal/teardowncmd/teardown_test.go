package teardowncmd

import (
	"context"
	"testing"
	"time"

	"github.com/codahale/belvedere/cmd/belvedere/internal/mocks"
	"github.com/codahale/belvedere/cmd/belvedere/internal/rootcmd"
	"github.com/golang/mock/gomock"
)

func TestTeardown(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	project := mocks.NewMockProject(ctrl)
	project.EXPECT().
		Teardown(gomock.Any(), false, false, 10*time.Second)

	command := New(&rootcmd.Config{
		Project: project,
	})
	if err := command.ParseAndRun(context.Background(), []string{}); err != nil {
		t.Fatal(err)
	}
}

func TestTeardown_Flags(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	project := mocks.NewMockProject(ctrl)
	project.EXPECT().
		Teardown(gomock.Any(), true, true, 10*time.Minute)

	command := New(&rootcmd.Config{
		Project: project,
	})
	if err := command.ParseAndRun(context.Background(), []string{
		"-async",
		"-dry-run",
		"-interval",
		"10m",
	}); err != nil {
		t.Fatal(err)
	}
}
