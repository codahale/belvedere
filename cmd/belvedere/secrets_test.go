package main

import (
	"bytes"
	"testing"

	"github.com/codahale/belvedere/pkg/belvedere"
	"github.com/golang/mock/gomock"
)

func TestSecretsList(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	project, output, pf, of := mockFactories(ctrl)

	list := []belvedere.Secret{
		{
			Name: "one",
		},
	}

	secrets := NewMockSecretsService(ctrl)
	secrets.EXPECT().
		List(gomock.Any()).
		Return(list, nil)

	project.EXPECT().Secrets().Return(secrets)

	output.EXPECT().
		Print(list)

	cmd := newRootCmd("test").ToCobra(pf, of)
	cmd.SetOut(bytes.NewBuffer(nil))
	cmd.SetErr(bytes.NewBuffer(nil))
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

	project, _, pf, of := mockFactories(ctrl)

	value := []byte(`secret`)

	secrets := NewMockSecretsService(ctrl)
	secrets.EXPECT().
		Create(gomock.Any(), "my-secret", value, true)

	project.EXPECT().Secrets().Return(secrets)

	cmd := newRootCmd("test").ToCobra(pf, of)
	cmd.SetIn(bytes.NewBuffer(value))
	cmd.SetOut(bytes.NewBuffer(nil))
	cmd.SetErr(bytes.NewBuffer(nil))
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

	project, _, pf, of := mockFactories(ctrl)

	value := []byte("secret\n")

	secrets := NewMockSecretsService(ctrl)
	secrets.EXPECT().
		Create(gomock.Any(), "my-secret", value, true)

	project.EXPECT().Secrets().Return(secrets)

	cmd := newRootCmd("test").ToCobra(pf, of)
	cmd.SetOut(bytes.NewBuffer(nil))
	cmd.SetErr(bytes.NewBuffer(nil))
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

	project, _, pf, of := mockFactories(ctrl)

	value := []byte(`secret`)

	secrets := NewMockSecretsService(ctrl)
	secrets.EXPECT().
		Update(gomock.Any(), "my-secret", value, true)

	project.EXPECT().Secrets().Return(secrets)

	cmd := newRootCmd("test").ToCobra(pf, of)
	cmd.SetIn(bytes.NewBuffer(value))
	cmd.SetOut(bytes.NewBuffer(nil))
	cmd.SetErr(bytes.NewBuffer(nil))
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

	project, _, pf, of := mockFactories(ctrl)

	value := []byte("secret\n")

	secrets := NewMockSecretsService(ctrl)
	secrets.EXPECT().
		Update(gomock.Any(), "my-secret", value, true)

	project.EXPECT().Secrets().Return(secrets)

	cmd := newRootCmd("test").ToCobra(pf, of)
	cmd.SetOut(bytes.NewBuffer(nil))
	cmd.SetErr(bytes.NewBuffer(nil))
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

	project, _, pf, of := mockFactories(ctrl)

	secrets := NewMockSecretsService(ctrl)
	secrets.EXPECT().
		Grant(gomock.Any(), "my-secret", "my-app", true)

	project.EXPECT().Secrets().Return(secrets)

	cmd := newRootCmd("test").ToCobra(pf, of)
	cmd.SetOut(bytes.NewBuffer(nil))
	cmd.SetErr(bytes.NewBuffer(nil))
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

	project, _, pf, of := mockFactories(ctrl)

	secrets := NewMockSecretsService(ctrl)
	secrets.EXPECT().
		Revoke(gomock.Any(), "my-secret", "my-app", true)

	project.EXPECT().Secrets().Return(secrets)

	cmd := newRootCmd("test").ToCobra(pf, of)
	cmd.SetOut(bytes.NewBuffer(nil))
	cmd.SetErr(bytes.NewBuffer(nil))
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

	project, _, pf, of := mockFactories(ctrl)

	secrets := NewMockSecretsService(ctrl)
	secrets.EXPECT().
		Delete(gomock.Any(), "my-secret", true)

	project.EXPECT().Secrets().Return(secrets)

	cmd := newRootCmd("test").ToCobra(pf, of)
	cmd.SetOut(bytes.NewBuffer(nil))
	cmd.SetErr(bytes.NewBuffer(nil))
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
