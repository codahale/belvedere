package secrets

import (
	"context"
	"testing"

	"github.com/codahale/belvedere/cmd/belvedere/internal/mocks"
	"github.com/codahale/belvedere/pkg/belvedere"
	"github.com/golang/mock/gomock"
)

func TestListCmd_Run(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	secs := []belvedere.Secret{
		{
			Name: "woo",
		},
	}

	secrets := mocks.NewMockSecretsService(ctrl)
	secrets.EXPECT().
		List(gomock.Any()).
		Return(secs, nil)

	project := mocks.NewMockProject(ctrl)
	project.EXPECT().
		Secrets().
		Return(secrets)

	tables := mocks.NewMockTableWriter(ctrl)
	tables.EXPECT().
		Print(secs)

	listCmd := &ListCmd{}
	if err := listCmd.Run(context.Background(), project, tables); err != nil {
		t.Fatal(err)
	}
}

func TestCreateCmd_Run(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	secrets := mocks.NewMockSecretsService(ctrl)
	secrets.EXPECT().
		Create(gomock.Any(), "one", []byte(`value`), false)

	project := mocks.NewMockProject(ctrl)
	project.EXPECT().
		Secrets().
		Return(secrets)

	fr := mocks.NewMockFileReader(ctrl)
	fr.EXPECT().
		Read("secret.txt").
		Return([]byte(`value`), nil)

	createCmd := &CreateCmd{
		Secret:   "one",
		DataFile: "secret.txt",
	}
	if err := createCmd.Run(context.Background(), project, fr); err != nil {
		t.Fatal(err)
	}
}

func TestUpdateCmd_Run(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	secrets := mocks.NewMockSecretsService(ctrl)
	secrets.EXPECT().
		Update(gomock.Any(), "one", []byte(`value`), false)

	project := mocks.NewMockProject(ctrl)
	project.EXPECT().
		Secrets().
		Return(secrets)

	fr := mocks.NewMockFileReader(ctrl)
	fr.EXPECT().
		Read("secret.txt").
		Return([]byte(`value`), nil)

	updateCmd := &UpdateCmd{
		Secret:   "one",
		DataFile: "secret.txt",
	}
	if err := updateCmd.Run(context.Background(), project, fr); err != nil {
		t.Fatal(err)
	}
}

func TestGrantCmd_Run(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	secrets := mocks.NewMockSecretsService(ctrl)
	secrets.EXPECT().
		Grant(gomock.Any(), "one", "app", false)

	project := mocks.NewMockProject(ctrl)
	project.EXPECT().
		Secrets().
		Return(secrets)

	grantCmd := &GrantCmd{
		Secret: "one",
		App:    "app",
	}
	if err := grantCmd.Run(context.Background(), project); err != nil {
		t.Fatal(err)
	}
}

func TestRevokeCmd_Run(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	secrets := mocks.NewMockSecretsService(ctrl)
	secrets.EXPECT().
		Revoke(gomock.Any(), "one", "app", false)

	project := mocks.NewMockProject(ctrl)
	project.EXPECT().
		Secrets().
		Return(secrets)

	revokeCmd := &RevokeCmd{
		Secret: "one",
		App:    "app",
	}
	if err := revokeCmd.Run(context.Background(), project); err != nil {
		t.Fatal(err)
	}
}

func TestDeleteCmd_Run(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	secrets := mocks.NewMockSecretsService(ctrl)
	secrets.EXPECT().
		Delete(gomock.Any(), "one", false)

	project := mocks.NewMockProject(ctrl)
	project.EXPECT().
		Secrets().
		Return(secrets)

	deleteCmd := &DeleteCmd{
		Secret: "one",
	}
	if err := deleteCmd.Run(context.Background(), project); err != nil {
		t.Fatal(err)
	}
}
