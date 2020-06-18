package main

import (
	"context"
	"io"

	"github.com/codahale/belvedere/cmd/belvedere/internal/cli"
	"github.com/codahale/belvedere/cmd/belvedere/internal/mocks"
	"github.com/codahale/belvedere/pkg/belvedere"
	"github.com/golang/mock/gomock"
	"google.golang.org/api/option"
)

func mockFactories(ctrl *gomock.Controller) (*mocks.MockProject, *mocks.MockOutput, cli.ProjectFactory, cli.OutputFactory) {
	project := mocks.NewMockProject(ctrl)
	project.EXPECT().Name().Return("my-project").AnyTimes()
	output := mocks.NewMockOutput(ctrl)
	return project, output,
		func(ctx context.Context, name string, opts ...option.ClientOption) (belvedere.Project, error) {
			return project, nil
		},
		func(w io.Writer, format string) (cli.Output, error) {
			return output, nil
		}
}
