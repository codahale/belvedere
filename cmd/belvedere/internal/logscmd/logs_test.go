package logscmd

import (
	"context"
	"testing"
	"time"

	"github.com/codahale/belvedere/cmd/belvedere/internal/mocks"
	"github.com/codahale/belvedere/cmd/belvedere/internal/rootcmd"
	"github.com/codahale/belvedere/pkg/belvedere"
	"github.com/golang/mock/gomock"
)

func TestLogs(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	list := []belvedere.LogEntry{
		{
			Instance: "one",
		},
	}

	tables := mocks.NewMockTableWriter(ctrl)
	tables.EXPECT().
		Print(list)

	logs := mocks.NewMockLogService(ctrl)
	logs.EXPECT().
		List(gomock.Any(), "my-app", "", "", 10*time.Minute, []string{}).
		Return(list, nil)

	project := mocks.NewMockProject(ctrl)
	project.EXPECT().
		Logs().
		Return(logs).
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

func TestLogs_WithRel(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	list := []belvedere.LogEntry{
		{
			Instance: "one",
		},
	}

	tables := mocks.NewMockTableWriter(ctrl)
	tables.EXPECT().
		Print(list)

	logs := mocks.NewMockLogService(ctrl)
	logs.EXPECT().
		List(gomock.Any(), "my-app", "my-rel", "", 10*time.Minute, []string{}).
		Return(list, nil)

	project := mocks.NewMockProject(ctrl)
	project.EXPECT().
		Logs().
		Return(logs).
		AnyTimes()

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

func TestLogs_WithRelAndInstance(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	list := []belvedere.LogEntry{
		{
			Instance: "one",
		},
	}

	tables := mocks.NewMockTableWriter(ctrl)
	tables.EXPECT().
		Print(list)

	logs := mocks.NewMockLogService(ctrl)
	logs.EXPECT().
		List(gomock.Any(), "my-app", "my-rel", "my-in", 10*time.Minute, []string{}).
		Return(list, nil)

	project := mocks.NewMockProject(ctrl)
	project.EXPECT().
		Logs().
		Return(logs).
		AnyTimes()

	command := New(&rootcmd.Config{
		Project: project,
		Tables:  tables,
	})
	if err := command.ParseAndRun(context.Background(), []string{
		"my-app",
		"my-rel",
		"my-in",
	}); err != nil {
		t.Fatal(err)
	}
}

func TestLogs_Flags(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	list := []belvedere.LogEntry{
		{
			Instance: "one",
		},
	}

	tables := mocks.NewMockTableWriter(ctrl)
	tables.EXPECT().
		Print(list)

	logs := mocks.NewMockLogService(ctrl)
	logs.EXPECT().
		List(gomock.Any(), "my-app", "my-rel", "my-in", 10*time.Hour,
			[]string{"one", "two"}).
		Return(list, nil)

	project := mocks.NewMockProject(ctrl)
	project.EXPECT().
		Logs().
		Return(logs).
		AnyTimes()

	command := New(&rootcmd.Config{
		Project: project,
		Tables:  tables,
	})
	if err := command.ParseAndRun(context.Background(), []string{
		"-max-age=10h",
		"-filter=one",
		"-filter=two",
		"my-app",
		"my-rel",
		"my-in",
	}); err != nil {
		t.Fatal(err)
	}
}
