package cmd

import "github.com/spf13/cobra"

var appsCmd = &cobra.Command{
	Use:   "apps",
	Short: "Commands for managing apps",
}

func init() {
	rootCmd.AddCommand(appsCmd)
}
