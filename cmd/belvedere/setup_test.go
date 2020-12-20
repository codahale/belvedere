package main

import (
	"bytes"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
)

func TestSetup(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	project, _, pf, of := mockFactories(ctrl)

	project.EXPECT().
		Setup(gomock.Any(), "cloudslap.club.", true, 1*time.Minute)

	cmd := newRootCmd("test").ToCobra(pf, of)
	cmd.SetOut(bytes.NewBuffer(nil))
	cmd.SetErr(bytes.NewBuffer(nil))
	cmd.SetArgs([]string{
		"setup",
		"--dry-run",
		"--interval=1m",
		"cloudslap.club.",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
}

func TestSetup_Missing_DNS_Name(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	_, _, pf, of := mockFactories(ctrl)

	cmd := newRootCmd("test").ToCobra(pf, of)
	cmd.SetOut(bytes.NewBuffer(nil))
	cmd.SetErr(bytes.NewBuffer(nil))
	cmd.SetArgs([]string{
		"setup",
		"--dry-run",
		"--interval=1m",
	})

	if err := cmd.Execute(); err == nil {
		t.Fatal("should have failed")
	}
}
