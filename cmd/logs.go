package cmd

import (
	"time"

	"github.com/codahale/belvedere/cmd/internal"
	"github.com/codahale/belvedere/pkg/belvedere"
	"github.com/spf13/cobra"
)

var (
	freshness time.Duration
	filters   []string
	logsCmd   = &cobra.Command{
		Use:   "logs <app>",
		Short: "Show log entries for an app",
		Args:  enableUsage(cobra.ExactArgs(1)),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, span := commonContext(cmd)
			defer span.End()

			app := args[0]
			release := cmd.Flag("release").Value.String()
			instance := cmd.Flag("instance").Value.String()
			minTimestamp := time.Now().Add(-freshness)

			logs, err := belvedere.Logs(ctx, project, app, release, instance, minTimestamp, filters)
			if err != nil {
				return err
			}

			var rows [][]string
			for _, log := range logs {
				rows = append(rows, []string{
					log.Timestamp.Format(time.Stamp),
					log.Release,
					log.Instance,
					log.Container,
					log.Message,
				})
			}

			return internal.PrintTable(cmd.OutOrStdout(), rows, "Timestamp", "Release", "Instance", "Container", "Message")
		},
	}
)

func init() {
	logsCmd.Flags().String("release", "", "Limit log entries to a release")
	logsCmd.Flags().String("instance", "", "Limit log entries to an instance")
	logsCmd.Flags().StringSliceVar(&filters, "filter", nil, "Include arbitrary log filter")
	logsCmd.Flags().DurationVar(&freshness, "freshness", 5*time.Minute, "Limit log entries to the given period of time")

	rootCmd.AddCommand(logsCmd)
}
