package cmd

import (
	"github.com/codahale/belvedere/pkg/belvedere"
	"github.com/spf13/cobra"
)

var (
	enableRelease     bool
	releasesCreateCmd = &cobra.Command{
		Use:   "create <app> <release> <config> <sha256>",
		Short: "Create a new release",
		Args:  enableUsage(cobra.ExactArgs(4)),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, span := commonContext(cmd)
			defer span.End()

			app := args[0]
			release := args[1]
			path := args[2]
			imageSHA256 := args[3]

			config, err := belvedere.LoadConfig(ctx, path)
			if err != nil {
				return err
			}

			if err := belvedere.CreateRelease(ctx, project, app, release, config, imageSHA256, dryRun); err != nil {
				return err
			}

			if enableRelease {
				return belvedere.EnableRelease(ctx, project, app, release, dryRun)
			}

			return nil
		},
	}
)

func init() {
	releasesCreateCmd.Flags().BoolVar(&enableRelease, "enable", false, "Enable the release after creating")

	releasesCmd.AddCommand(releasesCreateCmd)
}
