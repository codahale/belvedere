package main

import (
	"bytes"
	"testing"
	"time"

	"github.com/codahale/belvedere/pkg/belvedere"
	"github.com/codahale/belvedere/pkg/belvedere/cfg"
	"github.com/golang/mock/gomock"
)

func TestReleasesList(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	project, output, pf, of := mockFactories(ctrl)

	list := []belvedere.Release{
		{
			Release: "one",
		},
	}

	releases := NewMockReleaseService(ctrl)
	releases.EXPECT().
		List(gomock.Any(), "my-app").
		Return(list, nil)

	project.EXPECT().Releases().Return(releases)

	output.EXPECT().
		Print(list)

	cmd := newRootCmd("test").ToCobra(pf, of)
	cmd.SetOut(bytes.NewBuffer(nil))
	cmd.SetErr(bytes.NewBuffer(nil))
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

	project, _, pf, of := mockFactories(ctrl)

	config := cfg.Config{
		NumReplicas: 10,
	}

	releases := NewMockReleaseService(ctrl)
	releases.EXPECT().
		Create(gomock.Any(), "my-app", "my-release", &config, "12345", true, 5*time.Minute)

	project.EXPECT().Releases().Return(releases)

	cmd := newRootCmd("test").ToCobra(pf, of)
	cmd.SetIn(bytes.NewBufferString(`numReplicas: 10`))
	cmd.SetOut(bytes.NewBuffer(nil))
	cmd.SetErr(bytes.NewBuffer(nil))
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

	project, _, pf, of := mockFactories(ctrl)

	config := cfg.Config{
		NumReplicas: 10,
	}

	releases := NewMockReleaseService(ctrl)
	releases.EXPECT().
		Enable(gomock.Any(), "my-app", "my-release", true, 5*time.Minute).
		After(
			releases.EXPECT().
				Create(gomock.Any(), "my-app", "my-release", &config, "12345", true, 5*time.Minute),
		)

	project.EXPECT().Releases().Return(releases).AnyTimes()

	cmd := newRootCmd("test").ToCobra(pf, of)
	cmd.SetIn(bytes.NewBufferString(`numReplicas: 10`))
	cmd.SetOut(bytes.NewBuffer(nil))
	cmd.SetErr(bytes.NewBuffer(nil))
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

	project, _, pf, of := mockFactories(ctrl)

	config := cfg.Config{
		NumReplicas: 10,
	}

	releases := NewMockReleaseService(ctrl)
	releases.EXPECT().
		Create(gomock.Any(), "my-app", "my-release", &config, "12345", true, 5*time.Minute)

	project.EXPECT().Releases().Return(releases)

	cmd := newRootCmd("test").ToCobra(pf, of)
	cmd.SetOut(bytes.NewBuffer(nil))
	cmd.SetErr(bytes.NewBuffer(nil))
	cmd.SetArgs([]string{
		"releases",
		"create",
		"my-app",
		"my-release",
		"12345",
		"example.yaml",
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

	project, _, pf, of := mockFactories(ctrl)

	releases := NewMockReleaseService(ctrl)
	releases.EXPECT().
		Enable(gomock.Any(), "my-app", "my-release", true, 10*time.Minute)

	project.EXPECT().Releases().Return(releases)

	cmd := newRootCmd("test").ToCobra(pf, of)
	cmd.SetOut(bytes.NewBuffer(nil))
	cmd.SetErr(bytes.NewBuffer(nil))
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

	project, _, pf, of := mockFactories(ctrl)

	releases := NewMockReleaseService(ctrl)
	releases.EXPECT().
		Disable(gomock.Any(), "my-app", "my-release", true, 10*time.Minute)

	project.EXPECT().Releases().Return(releases)

	cmd := newRootCmd("test").ToCobra(pf, of)
	cmd.SetOut(bytes.NewBuffer(nil))
	cmd.SetErr(bytes.NewBuffer(nil))
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

	project, _, pf, of := mockFactories(ctrl)

	releases := NewMockReleaseService(ctrl)
	releases.EXPECT().
		Delete(gomock.Any(), "my-app", "my-release", true, true, 10*time.Minute)

	project.EXPECT().Releases().Return(releases)

	cmd := newRootCmd("test").ToCobra(pf, of)
	cmd.SetOut(bytes.NewBuffer(nil))
	cmd.SetErr(bytes.NewBuffer(nil))
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
