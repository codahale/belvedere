package apps

import (
	"context"
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

	list := []belvedere.App{
		{
			Name: "my-app",
		},
	}

	apps := mocks.NewMockAppService(ctrl)
	apps.EXPECT().
		List(gomock.Any()).
		Return(list, nil)

	tables := mocks.NewMockTableWriter(ctrl)
	tables.EXPECT().
		Print(list)

	project := mocks.NewMockProject(ctrl)
	project.EXPECT().
		Apps().
		Return(apps)

	listCmd := &ListCmd{}
	if err := listCmd.Run(context.Background(), project, tables); err != nil {
		t.Fatal(err)
	}
}

func TestCreateCmd_Run(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	config := &cfg.Config{
		NumReplicas: 100,
	}

	apps := mocks.NewMockAppService(ctrl)
	apps.EXPECT().
		Create(gomock.Any(), "us-west1", "my-app", config, false, 10*time.Millisecond)

	project := mocks.NewMockProject(ctrl)
	project.EXPECT().
		Apps().
		Return(apps)

	fr := mocks.NewMockFileReader(ctrl)
	fr.EXPECT().
		Read("config.yaml").
		Return([]byte(`{"numReplicas": 100}`), nil)

	createCmd := &CreateCmd{
		App:    "my-app",
		Region: "us-west1",
		Config: "config.yaml",
		LongRunningOptions: cmd.LongRunningOptions{
			Interval: 10 * time.Millisecond,
		},
	}
	if err := createCmd.Run(context.Background(), project, fr); err != nil {
		t.Fatal(err)
	}
}

func TestUpdateCmd_Run(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	config := &cfg.Config{
		NumReplicas: 100,
	}

	apps := mocks.NewMockAppService(ctrl)
	apps.EXPECT().
		Update(gomock.Any(), "my-app", config, false, 10*time.Millisecond)

	project := mocks.NewMockProject(ctrl)
	project.EXPECT().
		Apps().
		Return(apps)

	fr := mocks.NewMockFileReader(ctrl)
	fr.EXPECT().
		Read("config.yaml").
		Return([]byte(`{"numReplicas": 100}`), nil)

	updateCmd := &UpdateCmd{
		App:    "my-app",
		Config: "config.yaml",
		LongRunningOptions: cmd.LongRunningOptions{
			Interval: 10 * time.Millisecond,
		},
	}
	if err := updateCmd.Run(context.Background(), project, fr); err != nil {
		t.Fatal(err)
	}
}

func TestDeleteCmd_Run(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	apps := mocks.NewMockAppService(ctrl)
	apps.EXPECT().
		Delete(gomock.Any(), "my-app", false, false, 10*time.Millisecond)

	project := mocks.NewMockProject(ctrl)
	project.EXPECT().
		Apps().
		Return(apps)

	deleteCmd := &DeleteCmd{
		App:   "my-app",
		Async: false,
		LongRunningOptions: cmd.LongRunningOptions{
			Interval: 10 * time.Millisecond,
		},
	}
	if err := deleteCmd.Run(context.Background(), project); err != nil {
		t.Fatal(err)
	}
}
