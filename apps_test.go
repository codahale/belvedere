package main

import (
	"context"
	"testing"

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
