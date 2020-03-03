package main

import (
	"context"
	"testing"

	"github.com/codahale/belvedere/pkg/belvedere"
	"github.com/golang/mock/gomock"
)

func TestReleasesListCmd_Run(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cmd := &ReleasesListCmd{
		App: "my-app",
	}

	rels := []belvedere.Release{
		{
			Release: "1",
		},
	}

	releases := NewMockReleaseService(ctrl)
	releases.EXPECT().
		List(gomock.Any(), "my-app").
		Return(rels, nil)

	project := NewMockProject(ctrl)
	project.EXPECT().
		Releases().
		Return(releases)

	tables := NewMockTableWriter(ctrl)
	tables.EXPECT().
		Print(rels)

	if err := cmd.Run(context.TODO(), project, tables); err != nil {
		t.Fatal(err)
	}
}
