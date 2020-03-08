package machinetypescmd

import (
	"context"
	"testing"

	"github.com/codahale/belvedere/cmd/belvedere/internal/mocks"
	"github.com/codahale/belvedere/cmd/belvedere/internal/rootcmd"
	"github.com/codahale/belvedere/pkg/belvedere"
	"github.com/golang/mock/gomock"
)

func TestMachineTypes(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	list := []belvedere.MachineType{
		{
			Name: "one",
		},
	}

	tables := mocks.NewMockTableWriter(ctrl)
	tables.EXPECT().
		Print(list)

	project := mocks.NewMockProject(ctrl)
	project.EXPECT().
		MachineTypes(gomock.Any(), "").
		Return(list, nil)

	command := New(&rootcmd.Config{
		Project: project,
		Tables:  tables,
	})
	if err := command.ParseAndRun(context.Background(), []string{}); err != nil {
		t.Fatal(err)
	}
}

func TestMachineTypes_WithRegion(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	list := []belvedere.MachineType{
		{
			Name: "one",
		},
	}

	tables := mocks.NewMockTableWriter(ctrl)
	tables.EXPECT().
		Print(list)

	project := mocks.NewMockProject(ctrl)
	project.EXPECT().
		MachineTypes(gomock.Any(), "us-west1").
		Return(list, nil)

	command := New(&rootcmd.Config{
		Project: project,
		Tables:  tables,
	})
	if err := command.ParseAndRun(context.Background(), []string{
		"us-west1",
	}); err != nil {
		t.Fatal(err)
	}
}
