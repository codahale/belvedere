package createcmd

import (
	"context"
	"testing"
	"time"

	"github.com/codahale/belvedere/cmd/belvedere/internal/mocks"
	"github.com/codahale/belvedere/cmd/belvedere/internal/rootcmd"
	"github.com/codahale/belvedere/pkg/belvedere/cfg"
	"github.com/golang/mock/gomock"
)

func TestReleasesCreate(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	config := cfg.Config{
		NumReplicas: 10,
	}

	files := mocks.NewMockFileReader(ctrl)
	files.EXPECT().
		Read("-").
		Return([]byte(`{"numReplicas":10}`), nil)

	releases := mocks.NewMockReleaseService(ctrl)
	releases.EXPECT().
		Create(gomock.Any(), "my-app", "my-rel", &config, "12345", false, 10*time.Second)

	project := mocks.NewMockProject(ctrl)
	project.EXPECT().
		Releases().
		Return(releases).
		AnyTimes()

	command := New(&rootcmd.Config{
		Project: project,
		Files:   files,
	})
	if err := command.ParseAndRun(context.Background(), []string{
		"my-app",
		"my-rel",
		"12345",
	}); err != nil {
		t.Fatal(err)
	}
}

func TestReleasesCreate_AndEnable(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	config := cfg.Config{
		NumReplicas: 10,
	}

	files := mocks.NewMockFileReader(ctrl)
	files.EXPECT().
		Read("-").
		Return([]byte(`{"numReplicas":10}`), nil)

	releases := mocks.NewMockReleaseService(ctrl)
	releases.EXPECT().
		Enable(gomock.Any(), "my-app", "my-rel", false, 10*time.Second).
		After(
			releases.EXPECT().
				Create(gomock.Any(), "my-app", "my-rel", &config, "12345", false, 10*time.Second),
		)

	project := mocks.NewMockProject(ctrl)
	project.EXPECT().
		Releases().
		Return(releases).
		AnyTimes()

	command := New(&rootcmd.Config{
		Project: project,
		Files:   files,
	})
	if err := command.ParseAndRun(context.Background(), []string{
		"-enable",
		"my-app",
		"my-rel",
		"12345",
	}); err != nil {
		t.Fatal(err)
	}
}

func TestReleasesCreate_WithFile(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	config := cfg.Config{
		NumReplicas: 10,
	}

	files := mocks.NewMockFileReader(ctrl)
	files.EXPECT().
		Read("config.yaml").
		Return([]byte(`{"numReplicas":10}`), nil)

	releases := mocks.NewMockReleaseService(ctrl)
	releases.EXPECT().
		Create(gomock.Any(), "my-app", "my-rel", &config, "12345", false, 10*time.Second)

	project := mocks.NewMockProject(ctrl)
	project.EXPECT().
		Releases().
		Return(releases).
		AnyTimes()

	command := New(&rootcmd.Config{
		Project: project,
		Files:   files,
	})
	if err := command.ParseAndRun(context.Background(), []string{
		"my-app",
		"my-rel",
		"12345",
		"config.yaml",
	}); err != nil {
		t.Fatal(err)
	}
}

func TestReleasesCreate_Flags(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	config := cfg.Config{
		NumReplicas: 10,
	}

	files := mocks.NewMockFileReader(ctrl)
	files.EXPECT().
		Read("config.yaml").
		Return([]byte(`{"numReplicas":10}`), nil)

	releases := mocks.NewMockReleaseService(ctrl)
	releases.EXPECT().
		Create(gomock.Any(), "my-app", "my-rel", &config, "12345", true, 10*time.Hour)

	project := mocks.NewMockProject(ctrl)
	project.EXPECT().
		Releases().
		Return(releases).
		AnyTimes()

	command := New(&rootcmd.Config{
		Project: project,
		Files:   files,
	})
	if err := command.ParseAndRun(context.Background(), []string{
		"-dry-run",
		"-interval=10h",
		"my-app",
		"my-rel",
		"12345",
		"config.yaml",
	}); err != nil {
		t.Fatal(err)
	}
}
