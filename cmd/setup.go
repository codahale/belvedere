package cmd

import (
	"github.com/codahale/belvedere/pkg/belvedere"
	"github.com/spf13/cobra"
)

var setupCmd = &cobra.Command{
	Use:   "setup <dns-zone>",
	Short: "Set up the current GCP project for use with Belvedere",
	Args:  enableUsage(cobra.ExactArgs(1)),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, span := commonContext(cmd)
		defer span.End()

		return belvedere.Setup(ctx, project, args[0], dryRun)
	},
}

func init() {
	rootCmd.AddCommand(setupCmd)
}
