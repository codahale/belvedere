package grantcmd

import (
	"context"
	"testing"

	"github.com/codahale/belvedere/cmd/belvedere/internal/mocks"
	"github.com/codahale/belvedere/cmd/belvedere/internal/rootcmd"
	"github.com/golang/mock/gomock"
)

func TestSecretsGrant(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	secrets := mocks.NewMockSecretsService(ctrl)
	secrets.EXPECT().
		Grant(gomock.Any(), "one", "two", false)

	project := mocks.NewMockProject(ctrl)
	project.EXPECT().
		Secrets().
		Return(secrets).
		AnyTimes()

	command := New(&rootcmd.Config{
		Project: project,
	})
	if err := command.ParseAndRun(context.Background(), []string{
		"one",
		"two",
	}); err != nil {
		t.Fatal(err)
	}
}

func TestSecretsGrant_Flags(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	secrets := mocks.NewMockSecretsService(ctrl)
	secrets.EXPECT().
		Grant(gomock.Any(), "one", "two", true)

	project := mocks.NewMockProject(ctrl)
	project.EXPECT().
		Secrets().
		Return(secrets).
		AnyTimes()

	command := New(&rootcmd.Config{
		Project: project,
	})
	if err := command.ParseAndRun(context.Background(), []string{
		"-dry-run",
		"one",
		"two",
	}); err != nil {
		t.Fatal(err)
	}
}
