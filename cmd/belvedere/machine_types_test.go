package main

import (
	"bytes"
	"testing"

	"github.com/codahale/belvedere/pkg/belvedere"
	"github.com/golang/mock/gomock"
)

func TestMachineTypes(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	project, output, fs, pf, of := mockFactories(ctrl)

	machineTypes := []belvedere.MachineType{
		{
			Name: "one",
		},
	}

	project.EXPECT().
		MachineTypes(gomock.Any(), "").
		Return(machineTypes, nil)

	output.EXPECT().
		Print(machineTypes)

	cmd := newRootCmd("test").ToCobra(pf, of, fs)
	cmd.SetOut(bytes.NewBuffer(nil))
	cmd.SetArgs([]string{
		"machine-types",
	})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
}

func TestMachineTypes_WithRegion(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	project, output, fs, pf, of := mockFactories(ctrl)

	machineTypes := []belvedere.MachineType{
		{
			Name: "one",
		},
	}

	project.EXPECT().
		MachineTypes(gomock.Any(), "us-west1").
		Return(machineTypes, nil)

	output.EXPECT().
		Print(machineTypes)

	cmd := newRootCmd("test").ToCobra(pf, of, fs)
	cmd.SetOut(bytes.NewBuffer(nil))
	cmd.SetArgs([]string{
		"machine-types",
		"us-west1",
	})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
}
