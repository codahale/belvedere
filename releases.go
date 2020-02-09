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
	App string `optional:"" help:"Limit releases to the given app."`
}

func (cmd *ReleasesListCmd) Run(ctx context.Context, o *Options) error {
	releases, err := belvedere.Releases(ctx, o.Project, cmd.App)
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

func (cmd *ReleasesCreateCmd) Run(ctx context.Context, o *Options) error {
	b, err := readFile(ctx, cmd.Config)
	if err != nil {
		return err
	}

	config, err := belvedere.LoadConfig(ctx, b)
	if err != nil {
		return err
	}

	err = belvedere.CreateRelease(ctx, o.Project, cmd.App, cmd.Release, config, cmd.SHA256, o.DryRun)
	if err != nil {
		return err
	}

	if cmd.Enable {
		err = belvedere.EnableRelease(ctx, o.Project, cmd.App, cmd.Release, o.DryRun)
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

func (cmd *ReleasesEnableCmd) Run(ctx context.Context, o *Options) error {
	return belvedere.EnableRelease(ctx, o.Project, cmd.App, cmd.Release, o.DryRun)
}

type ReleasesDisableCmd struct {
	App     string `arg:"" help:"The app's name."`
	Release string `arg:"" help:"The release's name."`
}

func (cmd *ReleasesDisableCmd) Run(ctx context.Context, o *Options) error {
	return belvedere.DisableRelease(ctx, o.Project, cmd.App, cmd.Release, o.DryRun)
}

type ReleasesDeleteCmd struct {
	App     string `arg:"" help:"The app's name."`
	Release string `arg:"" help:"The release's name."`
	Async   bool   `help:"Return without waiting for successful completion."`
}

func (cmd *ReleasesDeleteCmd) Run(ctx context.Context, o *Options) error {
	return belvedere.DeleteRelease(ctx, o.Project, cmd.App, cmd.Release, o.DryRun, cmd.Async)
}
