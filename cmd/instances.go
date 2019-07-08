package cmd

import (
	"github.com/codahale/belvedere/cmd/internal"
	"github.com/codahale/belvedere/pkg/belvedere"
	"github.com/spf13/cobra"
)

var instancesCmd = &cobra.Command{
	Use:   "instances",
	Short: "List all instances",
	Args:  enableUsage(cobra.NoArgs),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, span := commonContext(cmd)
		defer span.End()

		app := cmd.Flag("app").Value.String()
		release := cmd.Flag("release").Value.String()

		instances, err := belvedere.ListInstances(ctx, project, app, release)
		if err != nil {
			return err
		}

		var rows [][]string
		for _, i := range instances {
			rows = append(rows, []string{i.Name, i.MachineType, i.Zone, i.Status})
		}

		return internal.PrintTable(cmd.OutOrStdout(), rows, "Name", "Machine Type", "Zone", "Status")
	},
}

func init() {
	instancesCmd.Flags().String("app", "", "Limit instances to an app")
	instancesCmd.Flags().String("release", "", "Limit instances to a release")

	rootCmd.AddCommand(instancesCmd)
}
