package main

import (
	"context"
	"testing"

	"github.com/codahale/belvedere/pkg/belvedere"
	"github.com/golang/mock/gomock"
)

func TestSecretsListCmd_Run(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cmd := &SecretsListCmd{}

	secs := []belvedere.Secret{
		{
			Name: "woo",
		},
	}

	secrets := NewMockSecretsService(ctrl)
	secrets.EXPECT().
		List(gomock.Any()).
		Return(secs, nil)

	project := NewMockProject(ctrl)
	project.EXPECT().
		Secrets().
		Return(secrets)

	tables := NewMockTableWriter(ctrl)
	tables.EXPECT().
		Print(secs)

	if err := cmd.Run(context.Background(), project, tables); err != nil {
		t.Fatal(err)
	}
}

func TestSecretsCreateCmd_Run(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cmd := &SecretsCreateCmd{
		Secret:   "one",
		DataFile: FileContentFlag(`value`),
	}

	secrets := NewMockSecretsService(ctrl)
	secrets.EXPECT().
		Create(gomock.Any(), "one", []byte(`value`), false)

	project := NewMockProject(ctrl)
	project.EXPECT().
		Secrets().
		Return(secrets)

	if err := cmd.Run(context.Background(), project); err != nil {
		t.Fatal(err)
	}
}

func TestSecretsUpdateCmd_Run(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cmd := &SecretsUpdateCmd{
		Secret:   "one",
		DataFile: FileContentFlag(`value`),
	}

	secrets := NewMockSecretsService(ctrl)
	secrets.EXPECT().
		Update(gomock.Any(), "one", []byte(`value`), false)

	project := NewMockProject(ctrl)
	project.EXPECT().
		Secrets().
		Return(secrets)

	if err := cmd.Run(context.Background(), project); err != nil {
		t.Fatal(err)
	}
}

func TestSecretsGrantCmd_Run(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cmd := &SecretsGrantCmd{
		Secret: "one",
		App:    "app",
	}

	secrets := NewMockSecretsService(ctrl)
	secrets.EXPECT().
		Grant(gomock.Any(), "one", "app", false)

	project := NewMockProject(ctrl)
	project.EXPECT().
		Secrets().
		Return(secrets)

	if err := cmd.Run(context.Background(), project); err != nil {
		t.Fatal(err)
	}
}

func TestSecretsRevokeCmd_Run(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cmd := &SecretsRevokeCmd{
		Secret: "one",
		App:    "app",
	}

	secrets := NewMockSecretsService(ctrl)
	secrets.EXPECT().
		Revoke(gomock.Any(), "one", "app", false)

	project := NewMockProject(ctrl)
	project.EXPECT().
		Secrets().
		Return(secrets)

	if err := cmd.Run(context.Background(), project); err != nil {
		t.Fatal(err)
	}
}

func TestSecretsDeleteCmd_Run(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cmd := &SecretsDeleteCmd{
		Secret: "one",
	}

	secrets := NewMockSecretsService(ctrl)
	secrets.EXPECT().
		Delete(gomock.Any(), "one", false)

	project := NewMockProject(ctrl)
	project.EXPECT().
		Secrets().
		Return(secrets)

	if err := cmd.Run(context.Background(), project); err != nil {
		t.Fatal(err)
	}
}
