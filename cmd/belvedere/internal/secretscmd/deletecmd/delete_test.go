package deletecmd

import (
	"context"
	"testing"

	"github.com/codahale/belvedere/cmd/belvedere/internal/mocks"
	"github.com/codahale/belvedere/cmd/belvedere/internal/rootcmd"
	"github.com/golang/mock/gomock"
)

func TestSecretsDelete(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	secrets := mocks.NewMockSecretsService(ctrl)
	secrets.EXPECT().
		Delete(gomock.Any(), "one", false)

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
	}); err != nil {
		t.Fatal(err)
	}
}

func TestSecretsDelete_Flags(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	secrets := mocks.NewMockSecretsService(ctrl)
	secrets.EXPECT().
		Delete(gomock.Any(), "one", true)

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
	}); err != nil {
		t.Fatal(err)
	}
}
