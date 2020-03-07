package releases

import (
	"context"

	"github.com/codahale/belvedere/cmd/belvedere/internal/cmd"
	"github.com/codahale/belvedere/pkg/belvedere"
	"github.com/codahale/belvedere/pkg/belvedere/cfg"
)

type RootCmd struct {
	List    ListCmd    `cmd:"" help:"List all releases."`
	Create  CreateCmd  `cmd:"" help:"Create a release."`
	Enable  EnableCmd  `cmd:"" help:"Put a release into service."`
	Disable DisableCmd `cmd:"" help:"Remove a release from service."`
	Delete  DeleteCmd  `cmd:"" help:"Delete a release."`
}

type ListCmd struct {
	App string `arg:"" optional:"" help:"Limit releases to the given app."`
}

func (cmd *ListCmd) Run(ctx context.Context, project belvedere.Project, tables cmd.TableWriter) error {
	releases, err := project.Releases().List(ctx, cmd.App)
	if err != nil {
		return err
	}
	return tables.Print(releases)
}

type CreateCmd struct {
	App                    string `arg:"" help:"The app's name."`
	Release                string `arg:"" help:"The release's name."`
	SHA256                 string `arg:"" help:"The app container's SHA256 hash."`
	Config                 string `arg:"" optional:"" default:"-" help:"The path to the app's configuration file. Reads from STDIN if not specified."`
	Enable                 bool   `help:"Put release into service once created."`
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

	err = project.Releases().Create(ctx, cmd.App, cmd.Release, config, cmd.SHA256, cmd.DryRun, cmd.Interval)
	if err != nil {
		return err
	}

	if cmd.Enable {
		err = project.Releases().Enable(ctx, cmd.App, cmd.Release, cmd.DryRun, cmd.Interval)
		if err != nil {
			return err
		}
	}
	return nil
}

type EnableCmd struct {
	App                    string `arg:"" help:"The app's name."`
	Release                string `arg:"" help:"The release's name."`
	cmd.ModifyOptions      `embed:""`
	cmd.LongRunningOptions `embed:""`
}

func (cmd *EnableCmd) Run(ctx context.Context, project belvedere.Project) error {
	return project.Releases().Enable(ctx, cmd.App, cmd.Release, cmd.DryRun, cmd.Interval)
}

type DisableCmd struct {
	App                    string `arg:"" help:"The app's name."`
	Release                string `arg:"" help:"The release's name."`
	cmd.ModifyOptions      `embed:""`
	cmd.LongRunningOptions `embed:""`
}

func (cmd *DisableCmd) Run(ctx context.Context, project belvedere.Project) error {
	return project.Releases().Disable(ctx, cmd.App, cmd.Release, cmd.DryRun, cmd.Interval)
}

type DeleteCmd struct {
	App                    string `arg:"" help:"The app's name."`
	Release                string `arg:"" help:"The release's name."`
	Async                  bool   `help:"Return without waiting for successful completion."`
	cmd.ModifyOptions      `embed:""`
	cmd.LongRunningOptions `embed:""`
}

func (cmd *DeleteCmd) Run(ctx context.Context, project belvedere.Project) error {
	return project.Releases().Delete(ctx, cmd.App, cmd.Release, cmd.DryRun, cmd.Async, cmd.Interval)
}
