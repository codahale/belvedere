package cmd

import (
	"github.com/codahale/belvedere/pkg/belvedere"
	"github.com/spf13/cobra"
)

var appsDeleteCmd = &cobra.Command{
	Use:   "delete <app>",
	Short: "Delete an app",
	Args:  enableUsage(cobra.ExactArgs(1)),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, span := commonContext(cmd)
		defer span.End()

		app := args[0]

		return belvedere.DeleteApp(ctx, project, app, dryRun, async)
	},
}

func init() {
	appsCmd.AddCommand(appsDeleteCmd)
}
