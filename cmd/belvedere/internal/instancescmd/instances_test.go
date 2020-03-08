package instancescmd

import (
	"context"
	"testing"

	"github.com/codahale/belvedere/cmd/belvedere/internal/mocks"
	"github.com/codahale/belvedere/cmd/belvedere/internal/rootcmd"
	"github.com/codahale/belvedere/pkg/belvedere"
	"github.com/golang/mock/gomock"
)

func TestInstances(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	list := []belvedere.Instance{
		{
			Name: "one",
		},
	}

	tables := mocks.NewMockTableWriter(ctrl)
	tables.EXPECT().
		Print(list)

	project := mocks.NewMockProject(ctrl)
	project.EXPECT().
		Instances(gomock.Any(), "", "").
		Return(list, nil)

	command := New(&rootcmd.Config{
		Project: project,
		Tables:  tables,
	})
	if err := command.ParseAndRun(context.Background(), []string{}); err != nil {
		t.Fatal(err)
	}
}

func TestInstances_WithApp(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	list := []belvedere.Instance{
		{
			Name: "one",
		},
	}

	tables := mocks.NewMockTableWriter(ctrl)
	tables.EXPECT().
		Print(list)

	project := mocks.NewMockProject(ctrl)
	project.EXPECT().
		Instances(gomock.Any(), "my-app", "").
		Return(list, nil)

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

func TestInstances_WithAppRel(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	list := []belvedere.Instance{
		{
			Name: "one",
		},
	}

	tables := mocks.NewMockTableWriter(ctrl)
	tables.EXPECT().
		Print(list)

	project := mocks.NewMockProject(ctrl)
	project.EXPECT().
		Instances(gomock.Any(), "my-app", "my-rel").
		Return(list, nil)

	command := New(&rootcmd.Config{
		Project: project,
		Tables:  tables,
	})
	if err := command.ParseAndRun(context.Background(), []string{
		"my-app",
		"my-rel",
	}); err != nil {
		t.Fatal(err)
	}
}
