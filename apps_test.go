package main

import (
	"context"
	"testing"
	"time"

	"github.com/codahale/belvedere/pkg/belvedere"
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

	if err := cmd.Run(context.TODO(), project, tables); err != nil {
		t.Fatal(err)
	}
}

func TestAppsCreateCmd_Run(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	apps := NewMockAppService(ctrl)
	apps.EXPECT().
		Create(gomock.Any(), "us-west1", "my-app", gomock.Any(), false, 10*time.Millisecond)

	project := NewMockProject(ctrl)
	project.EXPECT().
		Apps().
		Return(apps)

	cmd := &AppsCreateCmd{
		App:    "my-app",
		Region: "us-west1",
		Config: "pkg/belvedere/cfg/config-example.yaml",
	}

	options := &Options{
		Interval: 10 * time.Millisecond,
		Timeout:  1 * time.Second,
	}

	if err := cmd.Run(context.TODO(), project, options); err != nil {
		t.Fatal(err)
	}
}

func TestAppsUpdateCmd_Run(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	apps := NewMockAppService(ctrl)
	apps.EXPECT().
		Update(gomock.Any(), "my-app", gomock.Any(), false, 10*time.Millisecond)

	project := NewMockProject(ctrl)
	project.EXPECT().
		Apps().
		Return(apps)

	cmd := &AppsUpdateCmd{
		App:    "my-app",
		Config: "pkg/belvedere/cfg/config-example.yaml",
	}

	options := &Options{
		Interval: 10 * time.Millisecond,
		Timeout:  1 * time.Second,
	}

	if err := cmd.Run(context.TODO(), project, options); err != nil {
		t.Fatal(err)
	}
}
