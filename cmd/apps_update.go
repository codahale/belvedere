package cmd

import (
	"github.com/codahale/belvedere/pkg/belvedere"
	"github.com/spf13/cobra"
)

var appsUpdateCmd = &cobra.Command{
	Use:   "update <app> <config>",
	Short: "Update an app's configuration",
	Args:  enableUsage(cobra.ExactArgs(2)),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, span := commonContext(cmd)
		defer span.End()

		app := args[0]
		path := args[1]

		config, err := belvedere.LoadConfig(ctx, path)
		if err != nil {
			return err
		}

		return belvedere.UpdateApp(ctx, project, app, config, dryRun)
	},
}

func init() {
	appsCmd.AddCommand(appsUpdateCmd)
}
