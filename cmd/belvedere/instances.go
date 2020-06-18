package main

import (
	"context"

	"github.com/codahale/belvedere/cmd/belvedere/internal/cli"
	"github.com/codahale/belvedere/pkg/belvedere"
	"github.com/spf13/cobra"
)

func newInstancesCmd() *cli.Command {
	return &cli.Command{
		UI: cobra.Command{
			Use:     `instances [<app>] [<release>]`,
			Example: `belvedere my-app v43`,
			Short:   `List running application instances`,
			Long: `List running application instances.

Instances can be filtered by application name and release name.`,
			Args: cobra.RangeArgs(0, 2),
		},
		Run: func(ctx context.Context, project belvedere.Project, in cli.Input, out cli.Output, args []string) error {
			var app string
			if len(args) > 0 {
				app = args[0]
			}

			var release string
			if len(args) > 1 {
				release = args[1]
			}

			instances, err := project.Instances(ctx, app, release)
			if err != nil {
				return err
			}
			return out.Print(instances)
		},
	}
}
