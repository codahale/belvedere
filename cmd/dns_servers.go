package cmd

import (
	"github.com/codahale/belvedere/cmd/internal"
	"github.com/codahale/belvedere/pkg/belvedere"
	"github.com/spf13/cobra"
)

var dnsServersCmd = &cobra.Command{
	Use:   "dns-servers",
	Short: "Print a list of the DNS servers for the project's managed zone",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, span := commonContext(cmd)
		defer span.End()

		servers, err := belvedere.DNSServers(ctx, project)
		if err != nil {
			return err
		}

		var rows [][]string
		for _, s := range servers {
			rows = append(rows, []string{s})
		}

		return internal.PrintTable(cmd.OutOrStdout(), rows, "Server")
	},
}

func init() {
	rootCmd.AddCommand(dnsServersCmd)
}
