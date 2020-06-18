package main

import (
	"context"

	"github.com/codahale/belvedere/cmd/belvedere/internal/cli"
	"github.com/codahale/belvedere/pkg/belvedere"
	"github.com/codahale/belvedere/pkg/belvedere/cfg"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func newAppsCmd() *cli.Command {
	return &cli.Command{
		UI: cobra.Command{
			Use:   `apps`,
			Short: `Commands for managing applications`,
			Long: `Commands for managing applications.

A Belvedere application is anything in a Docker container which accepts HTTP2 connections.`,
		},
		Subcommands: []*cli.Command{
			newAppsListCmd(),
			newAppsCreateCmd(),
			newAppsUpdateCmd(),
			newAppsDeleteCmd(),
		},
	}
}

func newAppsListCmd() *cli.Command {
	return &cli.Command{
		UI: cobra.Command{
			Use:     `list`,
			Example: `belvedere apps list`,
			Short:   `List applications`,
			Long: `List applications.

Prints a table of provisioned applications in the current project.`,
			Args: cobra.NoArgs,
		},
		Run: func(ctx context.Context, project belvedere.Project, args cli.Args, out cli.Output) error {
			apps, err := project.Apps().List(ctx)
			if err != nil {
				return err
			}
			return out.Print(apps)
		},
	}
}

func newAppsCreateCmd() *cli.Command {
	var mf cli.ModifyFlags
	var lrf cli.LongRunningFlags

	return &cli.Command{
		UI: cobra.Command{
			Use:     `create <region> <name> [<config-file>]`,
			Example: `belvedere apps create us-west1 my-app my-app.yaml`,
			Short:   `Create an application`,
			Long: `Create an application.

The resources which run an application are provisioned inside a GCP region (e.g. us-west1), and this
property cannot be changed once the application is created.

If config-file is not specified (or is specified as '-'), the configuration file is read from STDIN
instead.`,
			Args: cobra.RangeArgs(2, 3),
		},
		Flags: func(fs *pflag.FlagSet) {
			mf.Register(fs)
			lrf.Register(fs)
		},
		Run: func(ctx context.Context, project belvedere.Project, args cli.Args, out cli.Output) error {
			region := args.String(0)
			name := args.String(1)
			b, err := args.File(2)
			if err != nil {
				return err
			}

			config, err := cfg.Parse(b)
			if err != nil {
				return err
			}

			return project.Apps().Create(ctx, region, name, config, mf.DryRun, lrf.Interval)
		},
	}
}

func newAppsUpdateCmd() *cli.Command {
	var mf cli.ModifyFlags
	var lrf cli.LongRunningFlags

	return &cli.Command{
		UI: cobra.Command{
			Use:     `update <name> [<config-file>]`,
			Example: `belvedere apps update my-app my-app.yaml`,
			Short:   `Update an application`,
			Long: `Update an application.

If config-file is not specified (or is specified as '-'), the configuration file is read from STDIN
instead.`,
			Args: cobra.RangeArgs(1, 2),
		},
		Flags: func(fs *pflag.FlagSet) {
			mf.Register(fs)
			lrf.Register(fs)
		},
		Run: func(ctx context.Context, project belvedere.Project, args cli.Args, out cli.Output) error {
			name := args.String(0)
			b, err := args.File(1)
			if err != nil {
				return err
			}

			config, err := cfg.Parse(b)
			if err != nil {
				return err
			}

			return project.Apps().Update(ctx, name, config, mf.DryRun, lrf.Interval)
		},
	}
}

func newAppsDeleteCmd() *cli.Command {
	var mf cli.ModifyFlags
	var lrf cli.LongRunningFlags
	var af cli.AsyncFlags
	return &cli.Command{
		UI: cobra.Command{
			Use:     `delete <name>`,
			Example: `belvedere apps delete my-app`,
			Short:   `Delete an application`,
			Long: `Delete an application.

An application must not have any releases before being deleted.`,
			Args: cobra.ExactArgs(1),
		},
		Flags: func(fs *pflag.FlagSet) {
			mf.Register(fs)
			lrf.Register(fs)
			af.Register(fs)
		},
		Run: func(ctx context.Context, project belvedere.Project, args cli.Args, out cli.Output) error {
			name := args.String(0)
			return project.Apps().Delete(ctx, name, mf.DryRun, af.Async, lrf.Interval)
		},
	}
}
