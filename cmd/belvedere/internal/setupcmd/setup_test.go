package setupcmd

import (
	"context"
	"testing"
	"time"

	"github.com/codahale/belvedere/cmd/belvedere/internal/mocks"
	"github.com/codahale/belvedere/cmd/belvedere/internal/rootcmd"
	"github.com/golang/mock/gomock"
)

func TestSetup(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	project := mocks.NewMockProject(ctrl)
	project.EXPECT().
		Setup(gomock.Any(), "cloudslap.club.", false, 10*time.Second)

	command := New(&rootcmd.Config{
		Project: project,
	})
	if err := command.ParseAndRun(context.Background(),
		[]string{
			"cloudslap.club.",
		},
	); err != nil {
		t.Fatal(err)
	}
}
