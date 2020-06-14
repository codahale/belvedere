package main

import (
	"bytes"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
)

func TestTeardown(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	project, _, fs, pf, tf := mockFactories(ctrl)

	project.EXPECT().
		Teardown(gomock.Any(), true, true, 1*time.Minute)

	cmd := newRootCmd("test").ToCobra(pf, tf, fs)
	cmd.SetOut(bytes.NewBuffer(nil))
	cmd.SetArgs([]string{
		"teardown",
		"--dry-run",
		"--async",
		"--interval=1m",
	})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
}
