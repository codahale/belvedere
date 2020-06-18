package main

import (
	"bytes"
	"testing"

	"github.com/golang/mock/gomock"
)

func TestSSH(t *testing.T) {
	t.Skip("no real way to test SSH command")

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	_, _, pf, of := mockFactories(ctrl)

	cmd := newRootCmd("test").ToCobra(pf, of)
	cmd.SetOut(bytes.NewBuffer(nil))
	cmd.SetErr(bytes.NewBuffer(nil))
	cmd.SetArgs([]string{
		"ssh",
	})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
}
