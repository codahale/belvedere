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
