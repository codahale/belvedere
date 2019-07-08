package cmd

import (
	"github.com/codahale/belvedere/cmd/internal"
	"github.com/codahale/belvedere/pkg/belvedere"
	"github.com/spf13/cobra"
)

var releasesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List releases",
	Args:  enableUsage(cobra.NoArgs),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, span := commonContext(cmd)
		defer span.End()

		app := cmd.Flag("app").Value.String()

		releases, err := belvedere.ListReleases(ctx, project, app)
		if err != nil {
			return err
		}

		var rows [][]string
		for _, p := range releases {
			rows = append(rows, []string{p.Project, p.Region, p.App, p.Release, p.Hash})
		}
		return internal.PrintTable(cmd.OutOrStdout(), rows, "Project", "Region", "App", "Release", "Hash")
	},
}

func init() {
	releasesListCmd.Flags().String("app", "", "Limit releases to an app")
	releasesCmd.AddCommand(releasesListCmd)
}
