package main

import (
	"context"

	"github.com/codahale/belvedere/cmd/belvedere/internal/cli"
	"github.com/codahale/belvedere/pkg/belvedere"
	"github.com/codahale/belvedere/pkg/belvedere/cfg"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func newReleasesCmd() *cli.Command {
	return &cli.Command{
		UI: cobra.Command{
			Use:   "releases",
			Short: "Commands for managing releases",
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
			Use:   "list [<app>]",
			Short: "List releases",
			Long: `List releases.

The list of releases can by filtered by application.`,
			Args: cobra.MinimumNArgs(1),
		},
		Run: func(ctx context.Context, project belvedere.Project, output cli.Output, fs afero.Fs, args []string) error {
			var app string
			if len(args) > 0 {
				app = args[0]
			}
			releases, err := project.Releases().List(ctx, app)
			if err != nil {
				return err
			}
			return output.Print(releases)
		},
	}
}

func newReleasesCreateCmd() *cli.Command {
	var mf cli.ModifyFlags
	var lrf cli.LongRunningFlags
	var enable bool
	return &cli.Command{
		UI: cobra.Command{
			Use:   "create <app> <name> <sha-256> [<config-file>]",
			Short: "Create a release",
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
		Run: func(ctx context.Context, project belvedere.Project, output cli.Output, fs afero.Fs, args []string) error {
			app := args[0]
			name := args[1]
			digest := args[2]

			path := "-"
			if len(args) > 3 {
				path = args[3]
			}

			b, err := afero.ReadFile(fs, path)
			if err != nil {
				return err
			}

			config, err := cfg.Parse(b)
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
	var mf cli.ModifyFlags
	var lrf cli.LongRunningFlags
	return &cli.Command{
		UI: cobra.Command{
			Use:   "enable <app> <name>",
			Short: "Enable a release",
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
		Run: func(ctx context.Context, project belvedere.Project, output cli.Output, fs afero.Fs, args []string) error {
			app := args[0]
			name := args[1]
			return project.Releases().Enable(ctx, app, name, mf.DryRun, lrf.Interval)
		},
	}
}

func newReleasesDisableCmd() *cli.Command {
	var mf cli.ModifyFlags
	var lrf cli.LongRunningFlags
	return &cli.Command{
		UI: cobra.Command{
			Use:   "disable <app> <name>",
			Short: "Disable a release",
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
		Run: func(ctx context.Context, project belvedere.Project, output cli.Output, fs afero.Fs, args []string) error {
			app := args[0]
			name := args[1]
			return project.Releases().Disable(ctx, app, name, mf.DryRun, lrf.Interval)
		},
	}
}

func newReleasesDeleteCmd() *cli.Command {
	var mf cli.ModifyFlags
	var lrf cli.LongRunningFlags
	var af cli.AsyncFlags
	return &cli.Command{
		UI: cobra.Command{
			Use:   "delete <app> <name>",
			Short: "Delete a release",
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
		Run: func(ctx context.Context, project belvedere.Project, output cli.Output, fs afero.Fs, args []string) error {
			app := args[0]
			name := args[1]
			return project.Releases().Delete(ctx, app, name, mf.DryRun, af.Async, lrf.Interval)
		},
	}
}
