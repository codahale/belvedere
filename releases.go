package main

import (
	"context"

	"github.com/codahale/belvedere/pkg/belvedere"
)

type ReleasesCmd struct {
	List    ReleasesListCmd    `cmd:"" help:"List all releases."`
	Create  ReleasesCreateCmd  `cmd:"" help:"Create a release."`
	Enable  ReleasesEnableCmd  `cmd:"" help:"Put a release into service."`
	Disable ReleasesDisableCmd `cmd:"" help:"Remove a release from service."`
	Delete  ReleasesDeleteCmd  `cmd:"" help:"Delete a release."`
}

type ReleasesListCmd struct {
	App string `arg:"" optional:"" help:"Limit releases to the given app."`
}

func (cmd *ReleasesListCmd) Run(ctx context.Context, project belvedere.Project, o *Options) error {
	releases, err := project.Releases().List(ctx, cmd.App)
	if err != nil {
		return err
	}
	return o.printTable(releases)
}

type ReleasesCreateCmd struct {
	App     string `arg:"" help:"The app's name."`
	Release string `arg:"" help:"The release's name."`
	SHA256  string `arg:"" help:"The app container's SHA256 hash."`
	Config  string `arg:"" optional:"" help:"The path to the app's configuration file. Reads from STDIN if not specified."`
	Enable  bool   `help:"Put release into service once created."`
}

func (cmd *ReleasesCreateCmd) Run(ctx context.Context, project belvedere.Project, o *Options) error {
	b, err := readFile(ctx, cmd.Config)
	if err != nil {
		return err
	}

	config, err := belvedere.ParseConfig(b)
	if err != nil {
		return err
	}

	err = project.Releases().Create(ctx, cmd.App, cmd.Release, config, cmd.SHA256, o.DryRun, o.Interval)
	if err != nil {
		return err
	}

	if cmd.Enable {
		err = project.Releases().Enable(ctx, cmd.App, cmd.Release, o.DryRun, o.Interval)
		if err != nil {
			return err
		}
	}
	return nil
}

type ReleasesEnableCmd struct {
	App     string `arg:"" help:"The app's name."`
	Release string `arg:"" help:"The release's name."`
}

func (cmd *ReleasesEnableCmd) Run(ctx context.Context, project belvedere.Project, o *Options) error {
	return project.Releases().Enable(ctx, cmd.App, cmd.Release, o.DryRun, o.Interval)
}

type ReleasesDisableCmd struct {
	App     string `arg:"" help:"The app's name."`
	Release string `arg:"" help:"The release's name."`
}

func (cmd *ReleasesDisableCmd) Run(ctx context.Context, project belvedere.Project, o *Options) error {
	return project.Releases().Disable(ctx, cmd.App, cmd.Release, o.DryRun, o.Interval)
}

type ReleasesDeleteCmd struct {
	App     string `arg:"" help:"The app's name."`
	Release string `arg:"" help:"The release's name."`
	Async   bool   `help:"Return without waiting for successful completion."`
}

func (cmd *ReleasesDeleteCmd) Run(ctx context.Context, project belvedere.Project, o *Options) error {
	return project.Releases().Delete(ctx, cmd.App, cmd.Release, o.DryRun, cmd.Async, o.Interval)
}
