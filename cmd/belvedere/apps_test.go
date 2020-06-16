package main

import (
	"bytes"
	"testing"
	"time"

	"github.com/codahale/belvedere/cmd/belvedere/internal/cli"
	"github.com/codahale/belvedere/cmd/belvedere/internal/mocks"
	"github.com/codahale/belvedere/pkg/belvedere"
	"github.com/codahale/belvedere/pkg/belvedere/cfg"
	"github.com/golang/mock/gomock"
	"github.com/spf13/afero"
)

func TestAppsList(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	project, output, fs, pf, of := mockFactories(ctrl)

	list := []belvedere.App{
		{
			Name: "one",
		},
	}

	apps := mocks.NewMockAppService(ctrl)
	apps.EXPECT().
		List(gomock.Any()).
		Return(list, nil)

	project.EXPECT().Apps().Return(apps)

	output.EXPECT().
		Print(list)

	cmd := newRootCmd("test").ToCobra(pf, of, fs)
	cmd.SetOut(bytes.NewBuffer(nil))
	cmd.SetArgs([]string{
		"apps",
		"list",
	})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
}

func TestAppsCreate(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	project, _, fs, pf, of := mockFactories(ctrl)

	config := cfg.Config{
		NumReplicas: 10,
	}

	if err := afero.WriteFile(fs, cli.StdIn, []byte(`{"numReplicas":10}`), 0644); err != nil {
		t.Fatal(err)
	}

	apps := mocks.NewMockAppService(ctrl)
	apps.EXPECT().
		Create(gomock.Any(), "us-west1", "my-app", &config, true, 10*time.Minute).
		Return(nil)

	project.EXPECT().Apps().Return(apps)

	cmd := newRootCmd("test").ToCobra(pf, of, fs)
	cmd.SetOut(bytes.NewBuffer(nil))
	cmd.SetArgs([]string{
		"apps",
		"create",
		"us-west1",
		"my-app",
		"--dry-run",
		"--interval=10m",
	})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
}

func TestAppsCreate_WithFilename(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	project, _, fs, pf, of := mockFactories(ctrl)

	config := cfg.Config{
		NumReplicas: 10,
	}

	if err := afero.WriteFile(fs, "app.yaml", []byte(`{"numReplicas":10}`), 0644); err != nil {
		t.Fatal(err)
	}

	apps := mocks.NewMockAppService(ctrl)
	apps.EXPECT().
		Create(gomock.Any(), "us-west1", "my-app", &config, true, 10*time.Minute).
		Return(nil)

	project.EXPECT().Apps().Return(apps)

	cmd := newRootCmd("test").ToCobra(pf, of, fs)
	cmd.SetOut(bytes.NewBuffer(nil))
	cmd.SetArgs([]string{
		"apps",
		"create",
		"us-west1",
		"my-app",
		"app.yaml",
		"--dry-run",
		"--interval=10m",
	})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
}

func TestAppsUpdate(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	project, _, fs, pf, of := mockFactories(ctrl)

	config := cfg.Config{
		NumReplicas: 10,
	}

	if err := afero.WriteFile(fs, cli.StdIn, []byte(`{"numReplicas":10}`), 0644); err != nil {
		t.Fatal(err)
	}

	apps := mocks.NewMockAppService(ctrl)
	apps.EXPECT().
		Update(gomock.Any(), "my-app", &config, true, 10*time.Minute).
		Return(nil)

	project.EXPECT().Apps().Return(apps)

	cmd := newRootCmd("test").ToCobra(pf, of, fs)
	cmd.SetOut(bytes.NewBuffer(nil))
	cmd.SetArgs([]string{
		"apps",
		"update",
		"my-app",
		"--dry-run",
		"--interval=10m",
	})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
}

func TestAppsUpdate_WithFilename(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	project, _, fs, pf, of := mockFactories(ctrl)

	config := cfg.Config{
		NumReplicas: 10,
	}

	if err := afero.WriteFile(fs, "app.yaml", []byte(`{"numReplicas":10}`), 0644); err != nil {
		t.Fatal(err)
	}

	apps := mocks.NewMockAppService(ctrl)
	apps.EXPECT().
		Update(gomock.Any(), "my-app", &config, true, 10*time.Minute).
		Return(nil)

	project.EXPECT().Apps().Return(apps)

	cmd := newRootCmd("test").ToCobra(pf, of, fs)
	cmd.SetOut(bytes.NewBuffer(nil))
	cmd.SetArgs([]string{
		"apps",
		"update",
		"my-app",
		"app.yaml",
		"--dry-run",
		"--interval=10m",
	})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
}

func TestAppsDelete(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	project, _, fs, pf, of := mockFactories(ctrl)

	apps := mocks.NewMockAppService(ctrl)
	apps.EXPECT().
		Delete(gomock.Any(), "my-app", true, true, 10*time.Minute).
		Return(nil)

	project.EXPECT().Apps().Return(apps)

	cmd := newRootCmd("test").ToCobra(pf, of, fs)
	cmd.SetOut(bytes.NewBuffer(nil))
	cmd.SetArgs([]string{
		"apps",
		"delete",
		"my-app",
		"--async",
		"--dry-run",
		"--interval=10m",
	})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
}
