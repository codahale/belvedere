package cmd

import (
	"github.com/codahale/belvedere/pkg/belvedere"
	"github.com/spf13/cobra"
)

var appsCreateCmd = &cobra.Command{
	Use:   "create <app> <region> <config>",
	Short: "Create a new app",
	Args:  enableUsage(cobra.ExactArgs(3)),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, span := commonContext(cmd)
		defer span.End()

		app := args[0]
		region := args[1]
		path := args[2]

		config, err := belvedere.LoadConfig(ctx, path)
		if err != nil {
			return err
		}

		return belvedere.CreateApp(ctx, project, region, app, config, dryRun)
	},
}

func init() {
	appsCmd.AddCommand(appsCreateCmd)
}
