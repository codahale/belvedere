package cmd

import (
	"github.com/codahale/belvedere/pkg/belvedere"
	"github.com/spf13/cobra"
)

var releasesDisableCmd = &cobra.Command{
	Use:   "disable <app> <release>",
	Short: "Remove a release from service",
	Args:  enableUsage(cobra.ExactArgs(2)),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, span := commonContext(cmd)
		defer span.End()

		app := args[0]
		release := args[1]

		return belvedere.DisableRelease(ctx, project, app, release, dryRun)
	},
}

func init() {
	releasesCmd.AddCommand(releasesDisableCmd)
}
