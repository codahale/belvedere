package main

import (
	"bytes"
	"testing"
	"time"

	"github.com/codahale/belvedere/cmd/belvedere/internal/mocks"
	"github.com/codahale/belvedere/pkg/belvedere"
	"github.com/codahale/belvedere/pkg/belvedere/cfg"
	"github.com/golang/mock/gomock"
	"github.com/spf13/afero"
)

func TestReleasesList(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	project, tables, fs, pf, tf := mockFactories(ctrl)

	list := []belvedere.Release{
		{
			Release: "one",
		},
	}

	releases := mocks.NewMockReleaseService(ctrl)
	releases.EXPECT().
		List(gomock.Any(), "my-app").
		Return(list, nil)

	project.EXPECT().Releases().Return(releases)

	tables.EXPECT().
		Print(list)

	cmd := newRootCmd("test").ToCobra(pf, tf, fs)
	cmd.SetOut(bytes.NewBuffer(nil))
	cmd.SetArgs([]string{
		"releases",
		"list",
		"my-app",
	})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
}

func TestReleasesCreate(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	project, _, fs, pf, tf := mockFactories(ctrl)

	config := cfg.Config{
		NumReplicas: 10,
	}

	if err := afero.WriteFile(fs, "-", []byte(`{"numReplicas":10}`), 0644); err != nil {
		t.Fatal(err)
	}

	releases := mocks.NewMockReleaseService(ctrl)
	releases.EXPECT().
		Create(gomock.Any(), "my-app", "my-release", &config, "12345", true, 5*time.Minute)

	project.EXPECT().Releases().Return(releases)

	cmd := newRootCmd("test").ToCobra(pf, tf, fs)
	cmd.SetOut(bytes.NewBuffer(nil))
	cmd.SetArgs([]string{
		"releases",
		"create",
		"my-app",
		"my-release",
		"12345",
		"--dry-run",
		"--interval=5m",
	})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
}

func TestReleasesCreate_AndEnable(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	project, _, fs, pf, tf := mockFactories(ctrl)

	config := cfg.Config{
		NumReplicas: 10,
	}

	if err := afero.WriteFile(fs, "-", []byte(`{"numReplicas":10}`), 0644); err != nil {
		t.Fatal(err)
	}

	releases := mocks.NewMockReleaseService(ctrl)
	releases.EXPECT().
		Enable(gomock.Any(), "my-app", "my-release", true, 5*time.Minute).
		After(
			releases.EXPECT().
				Create(gomock.Any(), "my-app", "my-release", &config, "12345", true, 5*time.Minute),
		)

	project.EXPECT().Releases().Return(releases).AnyTimes()

	cmd := newRootCmd("test").ToCobra(pf, tf, fs)
	cmd.SetOut(bytes.NewBuffer(nil))
	cmd.SetArgs([]string{
		"releases",
		"create",
		"my-app",
		"my-release",
		"12345",
		"--enable",
		"--dry-run",
		"--interval=5m",
	})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
}

func TestReleasesCreate_WithFilename(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	project, _, fs, pf, tf := mockFactories(ctrl)

	config := cfg.Config{
		NumReplicas: 10,
	}

	if err := afero.WriteFile(fs, "app.yaml", []byte(`{"numReplicas":10}`), 0644); err != nil {
		t.Fatal(err)
	}

	releases := mocks.NewMockReleaseService(ctrl)
	releases.EXPECT().
		Create(gomock.Any(), "my-app", "my-release", &config, "12345", true, 5*time.Minute)

	project.EXPECT().Releases().Return(releases)

	cmd := newRootCmd("test").ToCobra(pf, tf, fs)
	cmd.SetOut(bytes.NewBuffer(nil))
	cmd.SetArgs([]string{
		"releases",
		"create",
		"my-app",
		"my-release",
		"12345",
		"app.yaml",
		"--dry-run",
		"--interval=5m",
	})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
}

func TestReleasesEnable(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	project, _, fs, pf, tf := mockFactories(ctrl)

	releases := mocks.NewMockReleaseService(ctrl)
	releases.EXPECT().
		Enable(gomock.Any(), "my-app", "my-release", true, 10*time.Minute)

	project.EXPECT().Releases().Return(releases)

	cmd := newRootCmd("test").ToCobra(pf, tf, fs)
	cmd.SetOut(bytes.NewBuffer(nil))
	cmd.SetArgs([]string{
		"releases",
		"enable",
		"my-app",
		"my-release",
		"--dry-run",
		"--interval=10m",
	})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
}

func TestReleasesDisable(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	project, _, fs, pf, tf := mockFactories(ctrl)

	releases := mocks.NewMockReleaseService(ctrl)
	releases.EXPECT().
		Disable(gomock.Any(), "my-app", "my-release", true, 10*time.Minute)

	project.EXPECT().Releases().Return(releases)

	cmd := newRootCmd("test").ToCobra(pf, tf, fs)
	cmd.SetOut(bytes.NewBuffer(nil))
	cmd.SetArgs([]string{
		"releases",
		"disable",
		"my-app",
		"my-release",
		"--dry-run",
		"--interval=10m",
	})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
}

func TestReleasesDelete(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	project, _, fs, pf, tf := mockFactories(ctrl)

	releases := mocks.NewMockReleaseService(ctrl)
	releases.EXPECT().
		Delete(gomock.Any(), "my-app", "my-release", true, true, 10*time.Minute)

	project.EXPECT().Releases().Return(releases)

	cmd := newRootCmd("test").ToCobra(pf, tf, fs)
	cmd.SetOut(bytes.NewBuffer(nil))
	cmd.SetArgs([]string{
		"releases",
		"delete",
		"my-app",
		"my-release",
		"--dry-run",
		"--async",
		"--interval=10m",
	})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
}