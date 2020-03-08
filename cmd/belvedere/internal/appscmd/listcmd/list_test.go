package listcmd

import (
	"context"
	"testing"

	"github.com/codahale/belvedere/cmd/belvedere/internal/mocks"
	"github.com/codahale/belvedere/cmd/belvedere/internal/rootcmd"
	"github.com/codahale/belvedere/pkg/belvedere"
	"github.com/golang/mock/gomock"
)

func TestAppsList(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	list := []belvedere.App{
		{
			Name: "one",
		},
	}

	tables := mocks.NewMockTableWriter(ctrl)
	tables.EXPECT().
		Print(list)

	apps := mocks.NewMockAppService(ctrl)
	apps.EXPECT().
		List(gomock.Any()).
		Return(list, nil)

	project := mocks.NewMockProject(ctrl)
	project.EXPECT().
		Apps().
		Return(apps).
		AnyTimes()

	command := New(&rootcmd.Config{
		Project: project,
		Tables:  tables,
	})
	if err := command.ParseAndRun(context.Background(), []string{
		"my-app",
	}); err != nil {
		t.Fatal(err)
	}
}
