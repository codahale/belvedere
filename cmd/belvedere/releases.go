package main

import (
	"bytes"
	"context"

	"github.com/codahale/belvedere/cmd/belvedere/internal/cli"
	"github.com/codahale/belvedere/pkg/belvedere"
	"github.com/codahale/belvedere/pkg/belvedere/cfg"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func newReleasesCmd() *cli.Command {
	return &cli.Command{
		UI: cobra.Command{
			Use:   `releases`,
			Short: `Commands for managing releases`,
			Long: `Commands for managing releases.

A release is a specific Docker image (identified by its SHA-256 digest) and a managed instance group
of GCE instances running the Docker image.`,
		},
		Subcommands: []*cli.Command{
			newReleasesListCmd(),
			newReleasesCreateCmd(),
			newReleasesEnableCmd(),
			newReleasesDisableCmd(),
			newReleasesDeleteCmd(),
		},
	}
}

func newReleasesListCmd() *cli.Command {
	return &cli.Command{
		UI: cobra.Command{
			Use:     `list [<app>]`,
			Example: `belvedere releases list my-app`,
			Short:   `List releases`,
			Long: `List releases.

The list of releases can by filtered by application.`,
			Args: cobra.MinimumNArgs(1),
		},
		Run: func(ctx context.Context, project belvedere.Project, args cli.Args, out cli.Output) error {
			app := args.String(0)

			releases, err := project.Releases().List(ctx, app)
			if err != nil {
				return err
			}

			return out.Print(releases)
		},
	}
}

func newReleasesCreateCmd() *cli.Command {
	var (
		mf     cli.ModifyFlags
		lrf    cli.LongRunningFlags
		enable bool
	)

	return &cli.Command{
		UI: cobra.Command{
			Use: `create <app> <name> <sha-256> [<config-file>]`,
			Example: `belvedere releases create my-app v1 ` +
				`5fb4ba1a651bae8057ec6b5cdafc93fa7e0b7d944d6f02a4b751de4e15464def my-app.yaml`,
			Short: `Create a release`,
			Long: `Create a release.

This requires the application name, a release name, the SHA-256 digest of the Docker image to
deploy, and the application configuration.

If config-file is not specified (or is specified as '-'), the configuration file is read from STDIN
instead.`,
			Args: cobra.RangeArgs(3, 4),
		},
		Flags: func(fs *pflag.FlagSet) {
			mf.Register(fs)
			lrf.Register(fs)
			fs.BoolVar(&enable, "enable", false, "enable the release after its successful creation")
		},
		Run: func(ctx context.Context, project belvedere.Project, args cli.Args, out cli.Output) error {
			app := args.String(0)
			name := args.String(1)
			digest := args.String(2)
			b, err := args.File(3)
			if err != nil {
				return err
			}

			config, err := cfg.Parse(bytes.NewReader(b))
			if err != nil {
				return err
			}

			if err := project.Releases().Create(ctx, app, name, config, digest, mf.DryRun, lrf.Interval); err != nil {
				return err
			}

			if enable {
				return project.Releases().Enable(ctx, app, name, mf.DryRun, lrf.Interval)
			}
			return nil
		},
	}
}

func newReleasesEnableCmd() *cli.Command {
	var (
		mf  cli.ModifyFlags
		lrf cli.LongRunningFlags
	)

	return &cli.Command{
		UI: cobra.Command{
			Use:     `enable <app> <name>`,
			Example: `belvedere releases enable my-app v1`,
			Short:   `Enable a release`,
			Long: `Enable a release.

Enabling a release registers the release's managed instance group with the application's load
balancer and waits for the instances to pass health checks and go into service. Use the -timeout
flag to bound the amount of time allowed for health checks to pass.`,
			Args: cobra.ExactArgs(2),
		},
		Flags: func(fs *pflag.FlagSet) {
			mf.Register(fs)
			lrf.Register(fs)
		},
		Run: func(ctx context.Context, project belvedere.Project, args cli.Args, out cli.Output) error {
			app := args.String(0)
			name := args.String(1)
			return project.Releases().Enable(ctx, app, name, mf.DryRun, lrf.Interval)
		},
	}
}

func newReleasesDisableCmd() *cli.Command {
	var (
		mf  cli.ModifyFlags
		lrf cli.LongRunningFlags
	)

	return &cli.Command{
		UI: cobra.Command{
			Use:     `disable <app> <name>`,
			Example: `belvedere releases disable my-app v1`,
			Short:   `Disable a release`,
			Long: `Disable a release.

Disabling a release unregisters the release's managed instance group from the application's load
balancer, removing it from service. This is used on old releases to roll a deploy forward or on new
releases which did not pass health checks in order to roll a deploy back.`,
			Args: cobra.ExactArgs(2),
		},
		Flags: func(fs *pflag.FlagSet) {
			mf.Register(fs)
			lrf.Register(fs)
		},
		Run: func(ctx context.Context, project belvedere.Project, args cli.Args, out cli.Output) error {
			app := args.String(0)
			name := args.String(1)
			return project.Releases().Disable(ctx, app, name, mf.DryRun, lrf.Interval)
		},
	}
}

func newReleasesDeleteCmd() *cli.Command {
	var (
		mf  cli.ModifyFlags
		lrf cli.LongRunningFlags
		af  cli.AsyncFlags
	)

	return &cli.Command{
		UI: cobra.Command{
			Use:     "delete <app> <name>",
			Example: "belvedere releases delete my-app v1",
			Short:   "Delete a release",
			Long: `Delete a release.

This deletes the managed instance group running the given release. Releases must be disabled before
they can be deleted.`,
			Args: cobra.ExactArgs(2),
		},
		Flags: func(fs *pflag.FlagSet) {
			mf.Register(fs)
			lrf.Register(fs)
			af.Register(fs)
		},
		Run: func(ctx context.Context, project belvedere.Project, args cli.Args, out cli.Output) error {
			app := args.String(0)
			name := args.String(1)
			return project.Releases().Delete(ctx, app, name, mf.DryRun, af.Async, lrf.Interval)
		},
	}
}
