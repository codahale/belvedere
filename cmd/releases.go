package cmd

import "github.com/spf13/cobra"

var releasesCmd = &cobra.Command{
	Use:   "releases",
	Short: "Commands for managing releases",
}

func init() {
	rootCmd.AddCommand(releasesCmd)
}
