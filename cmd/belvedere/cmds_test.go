package main

import (
	"context"
	"testing"
	"time"

	"github.com/codahale/belvedere/cmd/belvedere/internal/cmd"
	"github.com/codahale/belvedere/cmd/belvedere/internal/mocks"
	"github.com/codahale/belvedere/pkg/belvedere"
	"github.com/golang/mock/gomock"
)

func TestSetupCmd_Run(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	project := mocks.NewMockProject(ctrl)
	project.EXPECT().
		Setup(gomock.Any(), "cornbread.club", false, 10*time.Millisecond)

	setupCmd := &SetupCmd{
		DNSZone: "cornbread.club",
		LongRunningOptions: cmd.LongRunningOptions{
			Interval: 10 * time.Millisecond,
		},
	}
	if err := setupCmd.Run(context.Background(), project); err != nil {
		t.Fatal(err)
	}
}

func TestTeardownCmd_Run(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	project := mocks.NewMockProject(ctrl)
	project.EXPECT().
		Teardown(gomock.Any(), false, false, 10*time.Millisecond)

	teardownCmd := &TeardownCmd{
		Async: false,
		LongRunningOptions: cmd.LongRunningOptions{
			Interval: 10 * time.Millisecond,
		},
	}
	if err := teardownCmd.Run(context.Background(), project); err != nil {
		t.Fatal(err)
	}
}

func TestInstancesCmd_Run(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ins := []belvedere.Instance{
		{
			Name: "woo",
		},
	}

	project := mocks.NewMockProject(ctrl)
	project.EXPECT().
		Instances(gomock.Any(), "my-app", "v1").
		Return(ins, nil)

	tables := mocks.NewMockTableWriter(ctrl)
	tables.EXPECT().
		Print(ins)

	instancesCmd := &InstancesCmd{
		App:     "my-app",
		Release: "v1",
	}
	if err := instancesCmd.Run(context.Background(), project, tables); err != nil {
		t.Fatal(err)
	}
}

func TestMachineTypesCmd_Run(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mt := []belvedere.MachineType{
		{
			Name: "woo",
		},
	}

	project := mocks.NewMockProject(ctrl)
	project.EXPECT().
		MachineTypes(gomock.Any(), "us-west1").
		Return(mt, nil)

	tables := mocks.NewMockTableWriter(ctrl)
	tables.EXPECT().
		Print(mt)

	machineTypesCmd := &MachineTypesCmd{
		Region: "us-west1",
	}
	if err := machineTypesCmd.Run(context.Background(), project, tables); err != nil {
		t.Fatal(err)
	}
}

func TestDNSServersCmd_Run(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	servers := []belvedere.DNSServer{
		{
			Hostname: "woo",
		},
	}

	project := mocks.NewMockProject(ctrl)
	project.EXPECT().
		DNSServers(gomock.Any()).
		Return(servers, nil)

	tables := mocks.NewMockTableWriter(ctrl)
	tables.EXPECT().
		Print(servers)

	dnsServersCmd := &DNSServersCmd{}
	if err := dnsServersCmd.Run(context.Background(), project, tables); err != nil {
		t.Fatal(err)
	}
}

func TestLogsCmd_Run(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	entries := []belvedere.LogEntry{
		{
			Message: "yay",
		},
	}

	logs := mocks.NewMockLogService(ctrl)
	logs.EXPECT().
		List(gomock.Any(), "my-app", "v1", "my-app-v1-abc", 40*time.Minute, []string{"one eq 1"}).
		Return(entries, nil)

	project := mocks.NewMockProject(ctrl)
	project.EXPECT().
		Logs().
		Return(logs).
		AnyTimes()

	tables := mocks.NewMockTableWriter(ctrl)
	tables.EXPECT().
		Print(entries)

	logsCmd := &LogsCmd{
		App:       "my-app",
		Release:   "v1",
		Instance:  "my-app-v1-abc",
		Filters:   []string{"one eq 1"},
		Freshness: 40 * time.Minute,
	}
	if err := logsCmd.Run(context.Background(), project, tables); err != nil {
		t.Fatal(err)
	}
}
