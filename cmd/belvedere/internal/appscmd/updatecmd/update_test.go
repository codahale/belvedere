package updatecmd

import (
	"context"
	"testing"
	"time"

	"github.com/codahale/belvedere/cmd/belvedere/internal/mocks"
	"github.com/codahale/belvedere/cmd/belvedere/internal/rootcmd"
	"github.com/codahale/belvedere/pkg/belvedere/cfg"
	"github.com/golang/mock/gomock"
)

func TestAppsUpdate(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	config := cfg.Config{
		NumReplicas: 10,
	}

	files := mocks.NewMockFileReader(ctrl)
	files.EXPECT().
		Read("-").
		Return([]byte(`{"numReplicas":10}`), nil)

	apps := mocks.NewMockAppService(ctrl)
	apps.EXPECT().
		Update(gomock.Any(), "my-app", &config, false, 10*time.Second)

	project := mocks.NewMockProject(ctrl)
	project.EXPECT().
		Apps().
		Return(apps).
		AnyTimes()

	command := New(&rootcmd.Config{
		Project: project,
		Files:   files,
	})
	if err := command.ParseAndRun(context.Background(), []string{
		"my-app",
	}); err != nil {
		t.Fatal(err)
	}
}

func TestAppsUpdate_WithFile(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	config := cfg.Config{
		NumReplicas: 10,
	}

	files := mocks.NewMockFileReader(ctrl)
	files.EXPECT().
		Read("config.yaml").
		Return([]byte(`{"numReplicas":10}`), nil)

	apps := mocks.NewMockAppService(ctrl)
	apps.EXPECT().
		Update(gomock.Any(), "my-app", &config, false, 10*time.Second)

	project := mocks.NewMockProject(ctrl)
	project.EXPECT().
		Apps().
		Return(apps).
		AnyTimes()

	command := New(&rootcmd.Config{
		Project: project,
		Files:   files,
	})
	if err := command.ParseAndRun(context.Background(), []string{
		"my-app",
		"config.yaml",
	}); err != nil {
		t.Fatal(err)
	}
}

func TestAppsUpdate_Flags(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	config := cfg.Config{
		NumReplicas: 10,
	}

	files := mocks.NewMockFileReader(ctrl)
	files.EXPECT().
		Read("config.yaml").
		Return([]byte(`{"numReplicas":10}`), nil)

	apps := mocks.NewMockAppService(ctrl)
	apps.EXPECT().
		Update(gomock.Any(), "my-app", &config, true, 10*time.Hour)

	project := mocks.NewMockProject(ctrl)
	project.EXPECT().
		Apps().
		Return(apps).
		AnyTimes()

	command := New(&rootcmd.Config{
		Project: project,
		Files:   files,
	})
	if err := command.ParseAndRun(context.Background(), []string{
		"-dry-run",
		"-interval=10h",
		"my-app",
		"config.yaml",
	}); err != nil {
		t.Fatal(err)
	}
}
