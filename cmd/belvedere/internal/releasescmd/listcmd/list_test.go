package listcmd

import (
	"context"
	"testing"

	"github.com/codahale/belvedere/cmd/belvedere/internal/mocks"
	"github.com/codahale/belvedere/cmd/belvedere/internal/rootcmd"
	"github.com/codahale/belvedere/pkg/belvedere"
	"github.com/golang/mock/gomock"
)

func TestReleasesList(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	list := []belvedere.Release{
		{
			App: "one",
		},
	}

	tables := mocks.NewMockTableWriter(ctrl)
	tables.EXPECT().
		Print(list)

	releases := mocks.NewMockReleaseService(ctrl)
	releases.EXPECT().
		List(gomock.Any(), "").
		Return(list, nil)

	project := mocks.NewMockProject(ctrl)
	project.EXPECT().
		Releases().
		Return(releases).
		AnyTimes()

	command := New(&rootcmd.Config{
		Project: project,
		Tables:  tables,
	})
	if err := command.ParseAndRun(context.Background(), []string{}); err != nil {
		t.Fatal(err)
	}
}

func TestReleasesList_WithApp(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	list := []belvedere.Release{
		{
			App: "one",
		},
	}

	tables := mocks.NewMockTableWriter(ctrl)
	tables.EXPECT().
		Print(list)

	releases := mocks.NewMockReleaseService(ctrl)
	releases.EXPECT().
		List(gomock.Any(), "my-app").
		Return(list, nil)

	project := mocks.NewMockProject(ctrl)
	project.EXPECT().
		Releases().
		Return(releases).
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
