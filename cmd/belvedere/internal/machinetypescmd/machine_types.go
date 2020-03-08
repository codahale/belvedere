package machinetypescmd

import (
	"context"
	"flag"

	"github.com/codahale/belvedere/cmd/belvedere/internal/cmd"
	"github.com/codahale/belvedere/cmd/belvedere/internal/rootcmd"
	"github.com/peterbourgon/ff/v2/ffcli"
)

type Config struct {
	root *rootcmd.Config
}

func New(root *rootcmd.Config) *ffcli.Command {
	config := Config{root: root}

	fs := flag.NewFlagSet("belvedere machine-types", flag.ExitOnError)
	root.RegisterFlags(fs)

	return &ffcli.Command{
		Name:       "machine-types",
		ShortUsage: "belvedere machine-types [<region>] [flags]",
		ShortHelp:  "List available virtual machine types.",
		LongHelp: cmd.Wrap(`List available virtual machine types.

Machine types can be filtered by region. For more information on pricing and billing models, see
https://cloud.google.com/compute/vm-instance-pricing.`),
		FlagSet: fs,
		Exec:    config.Exec,
	}
}

func (c *Config) Exec(ctx context.Context, args []string) error {
	if len(args) > 1 {
		return flag.ErrHelp
	}

	var region string
	if len(args) > 0 {
		region = args[0]
	}

	machineTypes, err := c.root.Project.MachineTypes(ctx, region)
	if err != nil {
		return err
	}
	return c.root.Tables.Print(machineTypes)
}
