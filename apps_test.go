package main

import (
	"context"
	"testing"
	"time"

	"github.com/codahale/belvedere/pkg/belvedere"
	"github.com/codahale/belvedere/pkg/belvedere/cfg"
	"github.com/golang/mock/gomock"
)

func TestAppsListCmd_Run(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	list := []belvedere.App{
		{
			Name: "my-app",
		},
	}

	apps := NewMockAppService(ctrl)
	apps.EXPECT().
		List(gomock.Any()).
		Return(list, nil)

	tables := NewMockTableWriter(ctrl)
	tables.EXPECT().
		Print(list)

	project := NewMockProject(ctrl)
	project.EXPECT().
		Apps().
		Return(apps)

	cmd := &AppsListCmd{}

	if err := cmd.Run(context.Background(), project, tables); err != nil {
		t.Fatal(err)
	}
}

func TestAppsCreateCmd_Run(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	config := &cfg.Config{
		NumReplicas: 100,
	}

	apps := NewMockAppService(ctrl)
	apps.EXPECT().
		Create(gomock.Any(), "us-west1", "my-app", config, false, 10*time.Millisecond)

	project := NewMockProject(ctrl)
	project.EXPECT().
		Apps().
		Return(apps)

	cmd := &AppsCreateCmd{
		App:    "my-app",
		Region: "us-west1",
		Config: FileContentFlag(`{"numReplicas": 100}`),
		LongRunningOptions: LongRunningOptions{
			Interval: 10 * time.Millisecond,
		},
	}

	if err := cmd.Run(context.Background(), project); err != nil {
		t.Fatal(err)
	}
}

func TestAppsUpdateCmd_Run(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	config := &cfg.Config{
		NumReplicas: 100,
	}

	apps := NewMockAppService(ctrl)
	apps.EXPECT().
		Update(gomock.Any(), "my-app", config, false, 10*time.Millisecond)

	project := NewMockProject(ctrl)
	project.EXPECT().
		Apps().
		Return(apps)

	cmd := &AppsUpdateCmd{
		App:    "my-app",
		Config: FileContentFlag(`{"numReplicas": 100}`),
		LongRunningOptions: LongRunningOptions{
			Interval: 10 * time.Millisecond,
		},
	}

	if err := cmd.Run(context.Background(), project); err != nil {
		t.Fatal(err)
	}
}

func TestAppsDeleteCmd_Run(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	apps := NewMockAppService(ctrl)
	apps.EXPECT().
		Delete(gomock.Any(), "my-app", false, false, 10*time.Millisecond)

	project := NewMockProject(ctrl)
	project.EXPECT().
		Apps().
		Return(apps)

	cmd := &AppsDeleteCmd{
		App:   "my-app",
		Async: false,
		LongRunningOptions: LongRunningOptions{
			Interval: 10 * time.Millisecond,
		},
	}

	if err := cmd.Run(context.Background(), project); err != nil {
		t.Fatal(err)
	}
}
