package main

import (
	"bytes"
	"testing"

	"github.com/codahale/belvedere/pkg/belvedere"
	"github.com/golang/mock/gomock"
)

func TestInstances(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	project, output, pf, of := mockFactories(ctrl)

	instances := []belvedere.Instance{
		{
			Name: "one",
		},
	}

	project.EXPECT().
		Instances(gomock.Any(), "", "").
		Return(instances, nil)

	output.EXPECT().
		Print(instances)

	cmd := newRootCmd("test").ToCobra(pf, of)
	cmd.SetOut(bytes.NewBuffer(nil))
	cmd.SetErr(bytes.NewBuffer(nil))
	cmd.SetArgs([]string{
		"instances",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
}

func TestInstances_WithApp(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	project, output, pf, of := mockFactories(ctrl)

	instances := []belvedere.Instance{
		{
			Name: "one",
		},
	}

	project.EXPECT().
		Instances(gomock.Any(), "my-app", "").
		Return(instances, nil)

	output.EXPECT().
		Print(instances)

	cmd := newRootCmd("test").ToCobra(pf, of)
	cmd.SetOut(bytes.NewBuffer(nil))
	cmd.SetErr(bytes.NewBuffer(nil))
	cmd.SetArgs([]string{
		"instances",
		"my-app",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
}

func TestInstances_WithAppAndRelease(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	project, output, pf, of := mockFactories(ctrl)

	instances := []belvedere.Instance{
		{
			Name: "one",
		},
	}

	project.EXPECT().
		Instances(gomock.Any(), "my-app", "my-release").
		Return(instances, nil)

	output.EXPECT().
		Print(instances)

	cmd := newRootCmd("test").ToCobra(pf, of)
	cmd.SetOut(bytes.NewBuffer(nil))
	cmd.SetErr(bytes.NewBuffer(nil))
	cmd.SetArgs([]string{
		"instances",
		"my-app",
		"my-release",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
}
