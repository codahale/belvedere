package releases

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/codahale/belvedere/cmd/belvedere/internal/cmd"
	"github.com/codahale/belvedere/cmd/belvedere/internal/mocks"
	"github.com/codahale/belvedere/pkg/belvedere"
	"github.com/codahale/belvedere/pkg/belvedere/cfg"
	"github.com/golang/mock/gomock"
)

func TestListCmd_Run(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	rels := []belvedere.Release{
		{
			Release: "1",
		},
	}

	releases := mocks.NewMockReleaseService(ctrl)
	releases.EXPECT().
		List(gomock.Any(), "my-app").
		Return(rels, nil)

	project := mocks.NewMockProject(ctrl)
	project.EXPECT().
		Releases().
		Return(releases)

	tables := mocks.NewMockTableWriter(ctrl)
	tables.EXPECT().
		Print(rels)

	listCmd := &ListCmd{
		App: "my-app",
	}
	if err := listCmd.Run(context.Background(), project, tables); err != nil {
		t.Fatal(err)
	}
}

func TestCreateCmd_Run(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	sha := strings.Repeat("a", 64)
	config := &cfg.Config{
		NumReplicas: 100,
	}

	releases := mocks.NewMockReleaseService(ctrl)
	releases.EXPECT().
		Create(gomock.Any(), "my-app", "v1", config, sha, false, 100*time.Millisecond)

	project := mocks.NewMockProject(ctrl)
	project.EXPECT().
		Releases().
		Return(releases)

	fr := mocks.NewMockFileReader(ctrl)
	fr.EXPECT().
		Read("config.yaml").
		Return([]byte(`{"numReplicas": 100}`), nil)

	createCmd := &CreateCmd{
		App:           "my-app",
		Release:       "v1",
		SHA256:        sha,
		Config:        "config.yaml",
		Enable:        false,
		ModifyOptions: cmd.ModifyOptions{},
		LongRunningOptions: cmd.LongRunningOptions{
			Interval: 100 * time.Millisecond,
		},
	}
	if err := createCmd.Run(context.Background(), project, fr); err != nil {
		t.Fatal(err)
	}
}

func TestCreateCmd_Run_Enable(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	sha := strings.Repeat("a", 64)
	config := &cfg.Config{
		NumReplicas: 100,
	}

	releases := mocks.NewMockReleaseService(ctrl)
	create := releases.EXPECT().
		Create(gomock.Any(), "my-app", "v1", config, sha, false, 100*time.Millisecond)
	releases.EXPECT().
		Enable(gomock.Any(), "my-app", "v1", false, 100*time.Millisecond).
		After(create)

	project := mocks.NewMockProject(ctrl)
	project.EXPECT().
		Releases().
		Return(releases).
		AnyTimes()

	fr := mocks.NewMockFileReader(ctrl)
	fr.EXPECT().
		Read("config.yaml").
		Return([]byte(`{"numReplicas": 100}`), nil)

	createCmd := &CreateCmd{
		App:           "my-app",
		Release:       "v1",
		SHA256:        sha,
		Config:        "config.yaml",
		Enable:        true,
		ModifyOptions: cmd.ModifyOptions{},
		LongRunningOptions: cmd.LongRunningOptions{
			Interval: 100 * time.Millisecond,
		},
	}

	if err := createCmd.Run(context.Background(), project, fr); err != nil {
		t.Fatal(err)
	}
}

func TestEnableCmd_Run(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	releases := mocks.NewMockReleaseService(ctrl)
	releases.EXPECT().
		Enable(gomock.Any(), "my-app", "v1", false, 100*time.Millisecond)

	project := mocks.NewMockProject(ctrl)
	project.EXPECT().
		Releases().
		Return(releases).
		AnyTimes()

	enableCmd := &EnableCmd{
		App:           "my-app",
		Release:       "v1",
		ModifyOptions: cmd.ModifyOptions{},
		LongRunningOptions: cmd.LongRunningOptions{
			Interval: 100 * time.Millisecond,
		},
	}
	if err := enableCmd.Run(context.Background(), project); err != nil {
		t.Fatal(err)
	}
}

func TestDisableCmd_Run(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	releases := mocks.NewMockReleaseService(ctrl)
	releases.EXPECT().
		Disable(gomock.Any(), "my-app", "v1", false, 100*time.Millisecond)

	project := mocks.NewMockProject(ctrl)
	project.EXPECT().
		Releases().
		Return(releases).
		AnyTimes()

	disableCmd := &DisableCmd{
		App:           "my-app",
		Release:       "v1",
		ModifyOptions: cmd.ModifyOptions{},
		LongRunningOptions: cmd.LongRunningOptions{
			Interval: 100 * time.Millisecond,
		},
	}
	if err := disableCmd.Run(context.Background(), project); err != nil {
		t.Fatal(err)
	}
}

func TestDeleteCmd_Run(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	releases := mocks.NewMockReleaseService(ctrl)
	releases.EXPECT().
		Delete(gomock.Any(), "my-app", "v1", false, false, 100*time.Millisecond)

	project := mocks.NewMockProject(ctrl)
	project.EXPECT().
		Releases().
		Return(releases).
		AnyTimes()

	deleteCmd := &DeleteCmd{
		App:           "my-app",
		Release:       "v1",
		ModifyOptions: cmd.ModifyOptions{},
		LongRunningOptions: cmd.LongRunningOptions{
			Interval: 100 * time.Millisecond,
		},
	}
	if err := deleteCmd.Run(context.Background(), project); err != nil {
		t.Fatal(err)
	}
}
