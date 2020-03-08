package listcmd

import (
	"context"
	"testing"

	"github.com/codahale/belvedere/cmd/belvedere/internal/mocks"
	"github.com/codahale/belvedere/cmd/belvedere/internal/rootcmd"
	"github.com/codahale/belvedere/pkg/belvedere"
	"github.com/golang/mock/gomock"
)

func TestSecretsList(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	list := []belvedere.Secret{
		{
			Name: "one",
		},
	}

	secrets := mocks.NewMockSecretsService(ctrl)
	secrets.EXPECT().
		List(gomock.Any()).
		Return(list, nil)

	tables := mocks.NewMockTableWriter(ctrl)
	tables.EXPECT().
		Print(list)

	project := mocks.NewMockProject(ctrl)
	project.EXPECT().
		Secrets().
		Return(secrets).
		AnyTimes()

	command := New(&rootcmd.Config{
		Project: project,
		Tables:  tables,
	})
	if err := command.ParseAndRun(context.Background(), []string{
		"one",
		"two",
	}); err != nil {
		t.Fatal(err)
	}
}
