package main

import (
	"bytes"
	"testing"
	"time"

	"github.com/codahale/belvedere/pkg/belvedere"
	"github.com/golang/mock/gomock"
)

func TestLogs(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	project, output, pf, of := mockFactories(ctrl)

	entries := []belvedere.LogEntry{
		{
			Instance: "one",
		},
	}

	logs := NewMockLogService(ctrl)
	logs.EXPECT().
		List(gomock.Any(), "my-app", "", "", 15*time.Minute, []string{"woo"}).
		Return(entries, nil)

	project.EXPECT().Logs().Return(logs)

	output.EXPECT().
		Print(entries)

	cmd := newRootCmd("test").ToCobra(pf, of)
	cmd.SetOut(bytes.NewBuffer(nil))
	cmd.SetErr(bytes.NewBuffer(nil))
	cmd.SetArgs([]string{
		"logs",
		"my-app",
		"--max-age=15m",
		"--filter=woo",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
}

func TestLogs_WithRelease(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	project, output, pf, of := mockFactories(ctrl)

	entries := []belvedere.LogEntry{
		{
			Instance: "one",
		},
	}

	logs := NewMockLogService(ctrl)
	logs.EXPECT().
		List(gomock.Any(), "my-app", "my-release", "", 15*time.Minute, []string{"woo"}).
		Return(entries, nil)

	project.EXPECT().Logs().Return(logs)

	output.EXPECT().
		Print(entries)

	cmd := newRootCmd("test").ToCobra(pf, of)
	cmd.SetOut(bytes.NewBuffer(nil))
	cmd.SetErr(bytes.NewBuffer(nil))
	cmd.SetArgs([]string{
		"logs",
		"my-app",
		"my-release",
		"--max-age=15m",
		"--filter=woo",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
}

func TestLogs_WithReleaseAndInstance(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	project, output, pf, of := mockFactories(ctrl)

	entries := []belvedere.LogEntry{
		{
			Instance: "one",
		},
	}

	logs := NewMockLogService(ctrl)
	logs.EXPECT().
		List(gomock.Any(), "my-app", "my-release", "my-instance", 15*time.Minute, []string{"woo"}).
		Return(entries, nil)

	project.EXPECT().Logs().Return(logs)

	output.EXPECT().
		Print(entries)

	cmd := newRootCmd("test").ToCobra(pf, of)
	cmd.SetOut(bytes.NewBuffer(nil))
	cmd.SetErr(bytes.NewBuffer(nil))
	cmd.SetArgs([]string{
		"logs",
		"my-app",
		"my-release",
		"my-instance",
		"--max-age=15m",
		"--filter=woo",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
}
