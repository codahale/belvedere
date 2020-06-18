package main

import (
	"context"

	"github.com/codahale/belvedere/cmd/belvedere/internal/cli"
	"github.com/codahale/belvedere/pkg/belvedere"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func newSetupCmd() *cli.Command {
	var mf cli.ModifyFlags
	var lrf cli.LongRunningFlags
	return &cli.Command{
		UI: cobra.Command{
			Use:     `setup <dns-zone>`,
			Example: `belvedere setup example.com.`,
			Short:   `Initialize a GCP project for use with Belvedere`,
			Long: `Initialize a GCP project for use with Belvedere.

Enables all required GCP APIs, grants Deployment Manager access to manage IAM permissions, and
creates a Deployment Manager deployment with the base resources and configuration required to
create, deploy, and manage applications with Belvedere.`,
			Args: cobra.ExactArgs(1),
		},
		Flags: func(fs *pflag.FlagSet) {
			mf.Register(fs)
			lrf.Register(fs)
		},
		Run: func(ctx context.Context, project belvedere.Project, args cli.Args, out cli.Output) error {
			dnsZone := args.String(0)
			return project.Setup(ctx, dnsZone, mf.DryRun, lrf.Interval)
		},
	}
}
