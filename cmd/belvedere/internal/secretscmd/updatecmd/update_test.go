package updatecmd

import (
	"context"
	"testing"

	"github.com/codahale/belvedere/cmd/belvedere/internal/mocks"
	"github.com/codahale/belvedere/cmd/belvedere/internal/rootcmd"
	"github.com/golang/mock/gomock"
)

func TestSecretsCreate(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	files := mocks.NewMockFileReader(ctrl)
	files.EXPECT().
		Read("-").
		Return([]byte(`value`), nil)

	secrets := mocks.NewMockSecretsService(ctrl)
	secrets.EXPECT().
		Update(gomock.Any(), "one", []byte(`value`), false)

	project := mocks.NewMockProject(ctrl)
	project.EXPECT().
		Secrets().
		Return(secrets).
		AnyTimes()

	command := New(&rootcmd.Config{
		Project: project,
		Files:   files,
	})
	if err := command.ParseAndRun(context.Background(), []string{
		"one",
	}); err != nil {
		t.Fatal(err)
	}
}

func TestSecretsCreate_WithFile(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	files := mocks.NewMockFileReader(ctrl)
	files.EXPECT().
		Read("secret.txt").
		Return([]byte(`value`), nil)

	secrets := mocks.NewMockSecretsService(ctrl)
	secrets.EXPECT().
		Update(gomock.Any(), "one", []byte(`value`), false)

	project := mocks.NewMockProject(ctrl)
	project.EXPECT().
		Secrets().
		Return(secrets).
		AnyTimes()

	command := New(&rootcmd.Config{
		Project: project,
		Files:   files,
	})
	if err := command.ParseAndRun(context.Background(), []string{
		"one",
		"secret.txt",
	}); err != nil {
		t.Fatal(err)
	}
}

func TestSecretsCreate_Flags(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	files := mocks.NewMockFileReader(ctrl)
	files.EXPECT().
		Read("secret.txt").
		Return([]byte(`value`), nil)

	secrets := mocks.NewMockSecretsService(ctrl)
	secrets.EXPECT().
		Update(gomock.Any(), "one", []byte(`value`), true)

	project := mocks.NewMockProject(ctrl)
	project.EXPECT().
		Secrets().
		Return(secrets).
		AnyTimes()

	command := New(&rootcmd.Config{
		Project: project,
		Files:   files,
	})
	if err := command.ParseAndRun(context.Background(), []string{
		"-dry-run",
		"one",
		"secret.txt",
	}); err != nil {
		t.Fatal(err)
	}
}
