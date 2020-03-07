package apps

import (
	"context"

	"github.com/codahale/belvedere/cmd/belvedere/internal/cmd"
	"github.com/codahale/belvedere/pkg/belvedere"
	"github.com/codahale/belvedere/pkg/belvedere/cfg"
)

type RootCmd struct {
	List   ListCmd   `cmd:"" help:"List all apps."`
	Create CreateCmd `cmd:"" help:"Create an application."`
	Update UpdateCmd `cmd:"" help:"Update an application."`
	Delete DeleteCmd `cmd:"" help:"Delete an application."`
}

type ListCmd struct {
}

func (ListCmd) Run(ctx context.Context, project belvedere.Project, tables cmd.TableWriter) error {
	apps, err := project.Apps().List(ctx)
	if err != nil {
		return err
	}
	return tables.Print(apps)
}

type CreateCmd struct {
	App                    string `arg:"" help:"The app's name."`
	Region                 string `arg:"" help:"The app's region."`
	Config                 string `arg:"" optional:"" default:"-" help:"The path to the app's configuration file. Reads from STDIN if not specified."`
	cmd.ModifyOptions      `embed:""`
	cmd.LongRunningOptions `embed:""`
}

func (cmd *CreateCmd) Run(ctx context.Context, project belvedere.Project, fr cmd.FileReader) error {
	b, err := fr.Read(cmd.Config)
	if err != nil {
		return err
	}

	config, err := cfg.Parse(b)
	if err != nil {
		return err
	}

	return project.Apps().Create(ctx, cmd.Region, cmd.App, config, cmd.DryRun, cmd.Interval)
}

type UpdateCmd struct {
	App                    string `arg:"" help:"The app's name."`
	Config                 string `arg:"" optional:"" default:"-" help:"The path to the app's configuration file. Reads from STDIN if not specified."`
	cmd.ModifyOptions      `embed:""`
	cmd.LongRunningOptions `embed:""`
}

func (cmd *UpdateCmd) Run(ctx context.Context, project belvedere.Project, fr cmd.FileReader) error {
	b, err := fr.Read(cmd.Config)
	if err != nil {
		return err
	}

	config, err := cfg.Parse(b)
	if err != nil {
		return err
	}

	return project.Apps().Update(ctx, cmd.App, config, cmd.DryRun, cmd.Interval)
}

type DeleteCmd struct {
	App                    string `arg:"" help:"The app's name."`
	Async                  bool   `help:"Return without waiting for successful completion."`
	cmd.ModifyOptions      `embed:""`
	cmd.LongRunningOptions `embed:""`
}

func (cmd *DeleteCmd) Run(ctx context.Context, project belvedere.Project) error {
	return project.Apps().Delete(ctx, cmd.App, cmd.DryRun, cmd.Async, cmd.Interval)
}
