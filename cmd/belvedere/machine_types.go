package main

import (
	"context"

	"github.com/codahale/belvedere/cmd/belvedere/internal/cli"
	"github.com/codahale/belvedere/pkg/belvedere"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

func newMachineTypesCmd() *cli.Command {
	return &cli.Command{
		UI: cobra.Command{
			Use:     `machine-types [<region>]`,
			Example: `belvedere machine-types us-west1`,
			Short:   `List available virtual machine types`,
			Long: `List available virtual machine types.

Machine types can be filtered by region. For more information on pricing and billing models, see
https://cloud.google.com/compute/vm-instance-pricing.`,
			Args: cobra.MaximumNArgs(1),
		},
		Run: func(ctx context.Context, project belvedere.Project, output cli.Output, fs afero.Fs, args []string) error {
			var region string
			if len(args) > 0 {
				region = args[0]
			}

			machineTypes, err := project.MachineTypes(ctx, region)
			if err != nil {
				return err
			}
			return output.Print(machineTypes)
		},
	}
}
