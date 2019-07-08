package cmd

import (
	"github.com/codahale/belvedere/pkg/belvedere"
	"github.com/spf13/cobra"
)

var sshCmd = &cobra.Command{
	Use:   "ssh <instance-name>",
	Short: "Connect to an instance using SSH over IAP",
	Args:  enableUsage(cobra.ExactArgs(1)),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, span := commonContext(cmd)
		defer span.End()

		ssh, err := belvedere.SSH(ctx, project, args[0])
		if err != nil {
			return err
		}

		exitHandlers = append(exitHandlers, ssh)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(sshCmd)
}
