package cmd

import (
	"github.com/codahale/belvedere/pkg/belvedere"
	"github.com/spf13/cobra"
)

var releasesDeleteCmd = &cobra.Command{
	Use:   "delete <app> <release>",
	Short: "Delete a release",
	Args:  enableUsage(cobra.ExactArgs(2)),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, span := commonContext(cmd)
		defer span.End()

		app := args[0]
		release := args[1]

		if err := belvedere.DisableRelease(ctx, project, app, release, dryRun); err != nil {
			return err
		}

		return belvedere.DeleteRelease(ctx, project, app, release, dryRun, async)
	},
}

func init() {
	releasesCmd.AddCommand(releasesDeleteCmd)
}
