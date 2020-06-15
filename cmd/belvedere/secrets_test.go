package main

import (
	"bytes"
	"testing"

	"github.com/codahale/belvedere/cmd/belvedere/internal/mocks"
	"github.com/codahale/belvedere/pkg/belvedere"
	"github.com/golang/mock/gomock"
	"github.com/spf13/afero"
)

func TestSecretsList(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	project, output, fs, pf, of := mockFactories(ctrl)

	list := []belvedere.Secret{
		{
			Name: "one",
		},
	}

	secrets := mocks.NewMockSecretsService(ctrl)
	secrets.EXPECT().
		List(gomock.Any()).
		Return(list, nil)

	project.EXPECT().Secrets().Return(secrets)

	output.EXPECT().
		Print(list)

	cmd := newRootCmd("test").ToCobra(pf, of, fs)
	cmd.SetOut(bytes.NewBuffer(nil))
	cmd.SetArgs([]string{
		"secrets",
		"list",
	})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
}

func TestSecretsCreate(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	project, _, fs, pf, of := mockFactories(ctrl)

	value := []byte(`secret`)
	if err := afero.WriteFile(fs, "-", value, 0644); err != nil {
		t.Fatal(err)
	}

	secrets := mocks.NewMockSecretsService(ctrl)
	secrets.EXPECT().
		Create(gomock.Any(), "my-secret", value, true)

	project.EXPECT().Secrets().Return(secrets)

	cmd := newRootCmd("test").ToCobra(pf, of, fs)
	cmd.SetOut(bytes.NewBuffer(nil))
	cmd.SetArgs([]string{
		"secrets",
		"create",
		"my-secret",
		"--dry-run",
	})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
}

func TestSecretsCreate_WithFilename(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	project, _, fs, pf, of := mockFactories(ctrl)

	value := []byte(`secret`)
	if err := afero.WriteFile(fs, "secret.txt", value, 0644); err != nil {
		t.Fatal(err)
	}

	secrets := mocks.NewMockSecretsService(ctrl)
	secrets.EXPECT().
		Create(gomock.Any(), "my-secret", value, true)

	project.EXPECT().Secrets().Return(secrets)

	cmd := newRootCmd("test").ToCobra(pf, of, fs)
	cmd.SetOut(bytes.NewBuffer(nil))
	cmd.SetArgs([]string{
		"secrets",
		"create",
		"my-secret",
		"secret.txt",
		"--dry-run",
	})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
}

func TestSecretsUpdate(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	project, _, fs, pf, of := mockFactories(ctrl)

	value := []byte(`secret`)
	if err := afero.WriteFile(fs, "-", value, 0644); err != nil {
		t.Fatal(err)
	}

	secrets := mocks.NewMockSecretsService(ctrl)
	secrets.EXPECT().
		Update(gomock.Any(), "my-secret", value, true)

	project.EXPECT().Secrets().Return(secrets)

	cmd := newRootCmd("test").ToCobra(pf, of, fs)
	cmd.SetOut(bytes.NewBuffer(nil))
	cmd.SetArgs([]string{
		"secrets",
		"update",
		"my-secret",
		"--dry-run",
	})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
}

func TestSecretsUpdate_WithFilename(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	project, _, fs, pf, of := mockFactories(ctrl)

	value := []byte(`secret`)
	if err := afero.WriteFile(fs, "secret.txt", value, 0644); err != nil {
		t.Fatal(err)
	}

	secrets := mocks.NewMockSecretsService(ctrl)
	secrets.EXPECT().
		Update(gomock.Any(), "my-secret", value, true)

	project.EXPECT().Secrets().Return(secrets)

	cmd := newRootCmd("test").ToCobra(pf, of, fs)
	cmd.SetOut(bytes.NewBuffer(nil))
	cmd.SetArgs([]string{
		"secrets",
		"update",
		"my-secret",
		"secret.txt",
		"--dry-run",
	})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
}

func TestSecretsGrant(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	project, _, fs, pf, of := mockFactories(ctrl)

	secrets := mocks.NewMockSecretsService(ctrl)
	secrets.EXPECT().
		Grant(gomock.Any(), "my-secret", "my-app", true)

	project.EXPECT().Secrets().Return(secrets)

	cmd := newRootCmd("test").ToCobra(pf, of, fs)
	cmd.SetOut(bytes.NewBuffer(nil))
	cmd.SetArgs([]string{
		"secrets",
		"grant",
		"my-secret",
		"my-app",
		"--dry-run",
	})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
}

func TestSecretsRevoke(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	project, _, fs, pf, of := mockFactories(ctrl)

	secrets := mocks.NewMockSecretsService(ctrl)
	secrets.EXPECT().
		Revoke(gomock.Any(), "my-secret", "my-app", true)

	project.EXPECT().Secrets().Return(secrets)

	cmd := newRootCmd("test").ToCobra(pf, of, fs)
	cmd.SetOut(bytes.NewBuffer(nil))
	cmd.SetArgs([]string{
		"secrets",
		"revoke",
		"my-secret",
		"my-app",
		"--dry-run",
	})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
}

func TestSecretsDelete(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	project, _, fs, pf, of := mockFactories(ctrl)

	secrets := mocks.NewMockSecretsService(ctrl)
	secrets.EXPECT().
		Delete(gomock.Any(), "my-secret", true)

	project.EXPECT().Secrets().Return(secrets)

	cmd := newRootCmd("test").ToCobra(pf, of, fs)
	cmd.SetOut(bytes.NewBuffer(nil))
	cmd.SetArgs([]string{
		"secrets",
		"delete",
		"my-secret",
		"--dry-run",
	})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
}
