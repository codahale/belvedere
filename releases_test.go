package main

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/codahale/belvedere/pkg/belvedere"
	"github.com/codahale/belvedere/pkg/belvedere/cfg"
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

func TestReleasesCreateCmd_Run(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	sha := strings.Repeat("a", 64)
	cmd := &ReleasesCreateCmd{
		App:           "my-app",
		Release:       "v1",
		SHA256:        sha,
		Config:        FileContentFlag(`{"numReplicas":100}`),
		Enable:        false,
		ModifyOptions: ModifyOptions{},
		LongRunningOptions: LongRunningOptions{
			Interval: 100 * time.Millisecond,
		},
	}

	config := &cfg.Config{
		NumReplicas: 100,
	}

	releases := NewMockReleaseService(ctrl)
	releases.EXPECT().
		Create(gomock.Any(), "my-app", "v1", config, sha, false, 100*time.Millisecond)

	project := NewMockProject(ctrl)
	project.EXPECT().
		Releases().
		Return(releases)

	if err := cmd.Run(context.TODO(), project); err != nil {
		t.Fatal(err)
	}
}

func TestReleasesCreateCmd_Run_Enable(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	sha := strings.Repeat("a", 64)
	cmd := &ReleasesCreateCmd{
		App:           "my-app",
		Release:       "v1",
		SHA256:        sha,
		Config:        FileContentFlag(`{"numReplicas":100}`),
		Enable:        true,
		ModifyOptions: ModifyOptions{},
		LongRunningOptions: LongRunningOptions{
			Interval: 100 * time.Millisecond,
		},
	}

	config := &cfg.Config{
		NumReplicas: 100,
	}

	releases := NewMockReleaseService(ctrl)
	create := releases.EXPECT().
		Create(gomock.Any(), "my-app", "v1", config, sha, false, 100*time.Millisecond)
	releases.EXPECT().
		Enable(gomock.Any(), "my-app", "v1", false, 100*time.Millisecond).
		After(create)

	project := NewMockProject(ctrl)
	project.EXPECT().
		Releases().
		Return(releases).
		AnyTimes()

	if err := cmd.Run(context.TODO(), project); err != nil {
		t.Fatal(err)
	}
}

func TestReleasesEnableCmd_Run(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cmd := &ReleasesEnableCmd{
		App:           "my-app",
		Release:       "v1",
		ModifyOptions: ModifyOptions{},
		LongRunningOptions: LongRunningOptions{
			Interval: 100 * time.Millisecond,
		},
	}

	releases := NewMockReleaseService(ctrl)
	releases.EXPECT().
		Enable(gomock.Any(), "my-app", "v1", false, 100*time.Millisecond)

	project := NewMockProject(ctrl)
	project.EXPECT().
		Releases().
		Return(releases).
		AnyTimes()

	if err := cmd.Run(context.TODO(), project); err != nil {
		t.Fatal(err)
	}
}

func TestReleasesDisableCmd_Run(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cmd := &ReleasesDisableCmd{
		App:           "my-app",
		Release:       "v1",
		ModifyOptions: ModifyOptions{},
		LongRunningOptions: LongRunningOptions{
			Interval: 100 * time.Millisecond,
		},
	}

	releases := NewMockReleaseService(ctrl)
	releases.EXPECT().
		Disable(gomock.Any(), "my-app", "v1", false, 100*time.Millisecond)

	project := NewMockProject(ctrl)
	project.EXPECT().
		Releases().
		Return(releases).
		AnyTimes()

	if err := cmd.Run(context.TODO(), project); err != nil {
		t.Fatal(err)
	}
}

func TestReleasesDeleteCmd_Run(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cmd := &ReleasesDeleteCmd{
		App:           "my-app",
		Release:       "v1",
		ModifyOptions: ModifyOptions{},
		LongRunningOptions: LongRunningOptions{
			Interval: 100 * time.Millisecond,
		},
	}

	releases := NewMockReleaseService(ctrl)
	releases.EXPECT().
		Delete(gomock.Any(), "my-app", "v1", false, false, 100*time.Millisecond)

	project := NewMockProject(ctrl)
	project.EXPECT().
		Releases().
		Return(releases).
		AnyTimes()

	if err := cmd.Run(context.TODO(), project); err != nil {
		t.Fatal(err)
	}
}
