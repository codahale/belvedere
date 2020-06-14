package main

import (
	"context"
	"time"

	"github.com/codahale/belvedere/cmd/belvedere/internal/cli"
	"github.com/codahale/belvedere/pkg/belvedere"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func newLogsCmd() *cli.Command {
	var filters []string
	var maxAge time.Duration
	return &cli.Command{
		UI: cobra.Command{
			Use:   "logs <app> [<release>] [<instance>] [--filter=<filter>...]",
			Short: "Display application logs",
			Long: `Display application logs.

Log entries are bounded by the -max-age parameter and filtered by the application name. They can
also be filtered by the release name, the instance name, and any additional Google Cloud Logging
filters. For more information on filter syntax, see
https://cloud.google.com/logging/docs/view/advanced-queries#advanced_logs_query_syntax.`,
			Args: cobra.RangeArgs(1, 3),
		},
		Flags: func(fs *pflag.FlagSet) {
			fs.StringSliceVar(&filters, "filter", nil, "limit entries to the given filter")
			fs.DurationVar(&maxAge, "max-age", 10*time.Minute, "limit entries by maximum age")
		},
		Run: func(ctx context.Context, project belvedere.Project, tables cli.TableWriter, fs afero.Fs, args []string) error {
			app := args[0]

			var release string
			if len(args) > 1 {
				release = args[1]
			}

			var instance string
			if len(args) > 2 {
				instance = args[2]
			}

			entries, err := project.Logs().List(ctx, app, release, instance, maxAge, filters)
			if err != nil {
				return err
			}
			return tables.Print(entries)
		},
	}
}
