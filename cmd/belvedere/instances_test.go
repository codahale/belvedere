package main

import (
	"bytes"
	"testing"

	"github.com/codahale/belvedere/pkg/belvedere"
	"github.com/golang/mock/gomock"
)

func TestInstances(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	project, tables, fs, pf, tf := mockFactories(ctrl)

	instances := []belvedere.Instance{
		{
			Name: "one",
		},
	}

	project.EXPECT().
		Instances(gomock.Any(), "", "").
		Return(instances, nil)

	tables.EXPECT().
		Print(instances)

	cmd := newRootCmd("test").ToCobra(pf, tf, fs)
	cmd.SetOut(bytes.NewBuffer(nil))
	cmd.SetArgs([]string{
		"instances",
	})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
}

func TestInstances_WithApp(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	project, tables, fs, pf, tf := mockFactories(ctrl)

	instances := []belvedere.Instance{
		{
			Name: "one",
		},
	}

	project.EXPECT().
		Instances(gomock.Any(), "my-app", "").
		Return(instances, nil)

	tables.EXPECT().
		Print(instances)

	cmd := newRootCmd("test").ToCobra(pf, tf, fs)
	cmd.SetOut(bytes.NewBuffer(nil))
	cmd.SetArgs([]string{
		"instances",
		"my-app",
	})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
}

func TestInstances_WithAppAndRelease(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	project, tables, fs, pf, tf := mockFactories(ctrl)

	instances := []belvedere.Instance{
		{
			Name: "one",
		},
	}

	project.EXPECT().
		Instances(gomock.Any(), "my-app", "my-release").
		Return(instances, nil)

	tables.EXPECT().
		Print(instances)

	cmd := newRootCmd("test").ToCobra(pf, tf, fs)
	cmd.SetOut(bytes.NewBuffer(nil))
	cmd.SetArgs([]string{
		"instances",
		"my-app",
		"my-release",
	})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
}
