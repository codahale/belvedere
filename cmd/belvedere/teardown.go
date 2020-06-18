package main

import (
	"context"

	"github.com/codahale/belvedere/cmd/belvedere/internal/cli"
	"github.com/codahale/belvedere/pkg/belvedere"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func newTeardownCmd() *cli.Command {
	var mf cli.ModifyFlags
	var lrf cli.LongRunningFlags
	var af cli.AsyncFlags
	return &cli.Command{
		UI: cobra.Command{
			Use:     `teardown`,
			Example: `belvedere teardown`,
			Short:   `Remove all Belvedere resources from this project`,
			Long: `Remove all Belvedere resources from this project.

Deletes the base Deployment Manager deployment.`,
			Args: cobra.NoArgs,
		},
		Flags: func(fs *pflag.FlagSet) {
			mf.Register(fs)
			lrf.Register(fs)
			af.Register(fs)
		},
		Run: func(ctx context.Context, project belvedere.Project, in cli.Input, out cli.Output, args []string) error {
			return project.Teardown(ctx, mf.DryRun, af.Async, lrf.Interval)
		},
	}
}
