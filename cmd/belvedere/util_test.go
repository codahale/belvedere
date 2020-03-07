package main

import (
	"context"
	"testing"
	"time"

	"github.com/codahale/belvedere/pkg/belvedere"
	"github.com/golang/mock/gomock"
)

func TestSetupCmd_Run(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cmd := &SetupCmd{
		DNSZone: "cornbread.club",
		LongRunningOptions: LongRunningOptions{
			Interval: 10 * time.Millisecond,
		},
	}

	project := NewMockProject(ctrl)
	project.EXPECT().
		Setup(gomock.Any(), "cornbread.club", false, 10*time.Millisecond)

	if err := cmd.Run(context.Background(), project); err != nil {
		t.Fatal(err)
	}
}

func TestTeardownCmd_Run(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cmd := &TeardownCmd{
		Async: false,
		LongRunningOptions: LongRunningOptions{
			Interval: 10 * time.Millisecond,
		},
	}

	project := NewMockProject(ctrl)
	project.EXPECT().
		Teardown(gomock.Any(), false, false, 10*time.Millisecond)

	if err := cmd.Run(context.Background(), project); err != nil {
		t.Fatal(err)
	}
}

func TestInstancesCmd_Run(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cmd := &InstancesCmd{
		App:     "my-app",
		Release: "v1",
	}

	ins := []belvedere.Instance{
		{
			Name: "woo",
		},
	}

	project := NewMockProject(ctrl)
	project.EXPECT().
		Instances(gomock.Any(), "my-app", "v1").
		Return(ins, nil)

	tables := NewMockTableWriter(ctrl)
	tables.EXPECT().
		Print(ins)

	if err := cmd.Run(context.Background(), project, tables); err != nil {
		t.Fatal(err)
	}
}

func TestMachineTypesCmd_Run(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cmd := &MachineTypesCmd{
		Region: "us-west1",
	}

	mt := []belvedere.MachineType{
		{
			Name: "woo",
		},
	}

	project := NewMockProject(ctrl)
	project.EXPECT().
		MachineTypes(gomock.Any(), "us-west1").
		Return(mt, nil)

	tables := NewMockTableWriter(ctrl)
	tables.EXPECT().
		Print(mt)

	if err := cmd.Run(context.Background(), project, tables); err != nil {
		t.Fatal(err)
	}
}

func TestDNSServersCmd_Run(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cmd := &DNSServersCmd{}

	servers := []belvedere.DNSServer{
		{
			Hostname: "woo",
		},
	}

	project := NewMockProject(ctrl)
	project.EXPECT().
		DNSServers(gomock.Any()).
		Return(servers, nil)

	tables := NewMockTableWriter(ctrl)
	tables.EXPECT().
		Print(servers)

	if err := cmd.Run(context.Background(), project, tables); err != nil {
		t.Fatal(err)
	}
}

func TestLogsCmd_Run(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cmd := &LogsCmd{
		App:       "my-app",
		Release:   "v1",
		Instance:  "my-app-v1-abc",
		Filters:   []string{"one eq 1"},
		Freshness: 40 * time.Minute,
	}

	entries := []belvedere.LogEntry{
		{
			Message: "yay",
		},
	}

	logs := NewMockLogService(ctrl)
	logs.EXPECT().
		List(gomock.Any(), "my-app", "v1", "my-app-v1-abc", 40*time.Minute, []string{"one eq 1"}).
		Return(entries, nil)

	project := NewMockProject(ctrl)
	project.EXPECT().
		Logs().
		Return(logs).
		AnyTimes()

	tables := NewMockTableWriter(ctrl)
	tables.EXPECT().
		Print(entries)

	if err := cmd.Run(context.Background(), project, tables); err != nil {
		t.Fatal(err)
	}
}
