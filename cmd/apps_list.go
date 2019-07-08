package cmd

import (
	"github.com/codahale/belvedere/cmd/internal"
	"github.com/codahale/belvedere/pkg/belvedere"
	"github.com/spf13/cobra"
)

var appsList = &cobra.Command{
	Use:   "list",
	Short: "List apps",
	Args:  enableUsage(cobra.NoArgs),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, span := commonContext(cmd)
		defer span.End()

		apps, err := belvedere.ListApps(ctx, project)
		if err != nil {
			return err
		}

		var rows [][]string
		for _, app := range apps {
			rows = append(rows, []string{app.Project, app.Region, app.Name})
		}

		return internal.PrintTable(cmd.OutOrStdout(), rows, "Project", "Region", "Name")
	},
}

func init() {
	appsCmd.AddCommand(appsList)
}
