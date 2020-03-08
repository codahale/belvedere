package updatecmd

import (
	"context"
	"flag"

	"github.com/codahale/belvedere/cmd/belvedere/internal/cmd"
	"github.com/codahale/belvedere/cmd/belvedere/internal/rootcmd"
	bcfg "github.com/codahale/belvedere/pkg/belvedere/cfg"
	"github.com/peterbourgon/ff/v2/ffcli"
)

type Config struct {
	root *rootcmd.Config
	cmd.ModifyOptions
	cmd.LongRunningOptions
}

func New(root *rootcmd.Config) *ffcli.Command {
	cfg := Config{root: root}

	fs := flag.NewFlagSet("belvedere apps update", flag.ExitOnError)
	root.RegisterFlags(fs)
	cfg.ModifyOptions.RegisterFlags(fs)
	cfg.LongRunningOptions.RegisterFlags(fs)

	return &ffcli.Command{
		Name:       "update",
		ShortUsage: "belvedere apps update <name> [<config-file>] [flags]",
		ShortHelp:  "Update an application.",
		LongHelp: cmd.Wrap(`Update an application.

If config-file is not specified (or is specified as '-'), the configuration file is read from STDIN
instead.`),
		FlagSet: fs,
		Exec:    cfg.Exec,
	}
}

func (c *Config) Exec(ctx context.Context, args []string) error {
	if len(args) < 1 || len(args) > 2 {
		return flag.ErrHelp
	}
	name := args[0]
	path := "-"
	if len(args) > 1 {
		path = args[1]
	}

	b, err := c.root.Files.Read(path)
	if err != nil {
		return err
	}

	config, err := bcfg.Parse(b)
	if err != nil {
		return err
	}

	return c.root.Project.Apps().Update(ctx, name, config, c.DryRun, c.Interval)
}
