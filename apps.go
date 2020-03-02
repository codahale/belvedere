package main

import (
	"context"

	"github.com/codahale/belvedere/pkg/belvedere"
	"github.com/codahale/belvedere/pkg/belvedere/cfg"
)

type AppsCmd struct {
	List   AppsListCmd   `cmd:"" help:"List all apps."`
	Create AppsCreateCmd `cmd:"" help:"Create an application."`
	Update AppsUpdateCmd `cmd:"" help:"Update an application."`
	Delete AppsDeleteCmd `cmd:"" help:"Delete an application."`
}

type AppsListCmd struct {
}

func (AppsListCmd) Run(ctx context.Context, project belvedere.Project, tables TableWriter) error {
	apps, err := project.Apps().List(ctx)
	if err != nil {
		return err
	}
	return tables.Print(apps)
}

type AppsCreateCmd struct {
	App    string `arg:"" help:"The app's name."`
	Region string `arg:"" help:"The app's region."`
	Config string `arg:"" optional:"" help:"The path to the app's configuration file. Reads from STDIN if not specified."`
}

func (cmd *AppsCreateCmd) Run(ctx context.Context, project belvedere.Project, o *Options) error {
	b, err := readFile(ctx, cmd.Config)
	if err != nil {
		return err
	}

	config, err := cfg.Parse(b)
	if err != nil {
		return err
	}
	return project.Apps().Create(ctx, cmd.Region, cmd.App, config, o.DryRun, o.Interval)
}

type AppsUpdateCmd struct {
	App    string `arg:"" help:"The app's name."`
	Config string `arg:"" optional:"" help:"The path to the app's configuration file. Reads from STDIN if not specified."`
}

func (cmd *AppsUpdateCmd) Run(ctx context.Context, project belvedere.Project, o *Options) error {
	b, err := readFile(ctx, cmd.Config)
	if err != nil {
		return err
	}

	config, err := cfg.Parse(b)
	if err != nil {
		return err
	}
	return project.Apps().Update(ctx, cmd.App, config, o.DryRun, o.Interval)
}

type AppsDeleteCmd struct {
	App   string `arg:"" help:"The app's name."`
	Async bool   `help:"Return without waiting for successful completion."`
}

func (cmd *AppsDeleteCmd) Run(ctx context.Context, project belvedere.Project, o *Options) error {
	return project.Apps().Delete(ctx, cmd.App, o.DryRun, cmd.Async, o.Interval)
}
