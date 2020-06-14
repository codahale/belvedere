package main

import (
	"context"
	"io"

	"github.com/codahale/belvedere/cmd/belvedere/internal/cli"
	"github.com/codahale/belvedere/cmd/belvedere/internal/mocks"
	"github.com/codahale/belvedere/pkg/belvedere"
	"github.com/golang/mock/gomock"
	"github.com/spf13/afero"
	"google.golang.org/api/option"
)

func mockFactories(ctrl *gomock.Controller) (*mocks.MockProject, *mocks.MockTableWriter, afero.Fs, cli.ProjectFactory, cli.TableWriterFactory) {
	project := mocks.NewMockProject(ctrl)
	project.EXPECT().Name().Return("my-project").AnyTimes()
	tables := mocks.NewMockTableWriter(ctrl)
	fs := afero.NewMemMapFs()
	return project, tables, fs,
		func(ctx context.Context, name string, opts ...option.ClientOption) (belvedere.Project, error) {
			return project, nil
		},
		func(w io.Writer, csv bool) cli.TableWriter {
			return tables
		}
}
