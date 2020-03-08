package instancescmd

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
	cfg := Config{root: root}

	fs := flag.NewFlagSet("belvedere instances", flag.ExitOnError)
	root.RegisterFlags(fs)

	return &ffcli.Command{
		Name:       "instances",
		ShortUsage: "belvedere instances [<app>] [<release>] [flags]",
		ShortHelp:  "List running application instances.",
		LongHelp: cmd.Wrap(`List running application instances.

Instances can be filtered by application name and release name.`),
		FlagSet: fs,
		Exec:    cfg.Exec,
	}
}

func (c *Config) Exec(ctx context.Context, args []string) error {
	if len(args) > 2 {
		return flag.ErrHelp
	}

	var app string
	if len(args) > 0 {
		app = args[0]
	}

	var release string
	if len(args) > 1 {
		release = args[1]
	}

	instances, err := c.root.Project.Instances(ctx, app, release)
	if err != nil {
		return err
	}
	return c.root.Tables.Print(instances)
}
