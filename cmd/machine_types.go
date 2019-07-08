package cmd

import (
	"strconv"

	"github.com/codahale/belvedere/cmd/internal"
	"github.com/codahale/belvedere/pkg/belvedere"
	"github.com/spf13/cobra"
)

var machineTypesCmd = &cobra.Command{
	Use:   "machine-types",
	Short: "List all available GCE machine types",
	Args:  enableUsage(cobra.NoArgs),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, span := commonContext(cmd)
		defer span.End()

		region := cmd.Flag("region").Value.String()

		machineTypes, err := belvedere.MachineTypes(ctx, project, region)
		if err != nil {
			return err
		}

		var rows [][]string
		for _, mt := range machineTypes {
			rows = append(rows, []string{mt.Name, strconv.Itoa(mt.CPU), strconv.Itoa(mt.Memory)})
		}
		return internal.PrintTable(cmd.OutOrStdout(), rows, "Name", "vCPUs", "Memory (MiB)")
	},
}

func init() {
	machineTypesCmd.Flags().String("region", "", "Limit machine types to those available in a region")
	rootCmd.AddCommand(machineTypesCmd)
}
