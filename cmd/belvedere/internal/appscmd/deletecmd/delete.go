package deletecmd

import (
	"context"
	"flag"

	"github.com/codahale/belvedere/cmd/belvedere/internal/cmd"
	"github.com/codahale/belvedere/cmd/belvedere/internal/rootcmd"
	"github.com/peterbourgon/ff/v3/ffcli"
)

type Config struct {
	root *rootcmd.Config
	cmd.LongRunningOptions
	cmd.ModifyOptions
	cmd.AsyncOptions
}

func New(root *rootcmd.Config) *ffcli.Command {
	config := Config{root: root}

	fs := flag.NewFlagSet("belvedere apps delete", flag.ExitOnError)
	root.RegisterFlags(fs)
	config.ModifyOptions.RegisterFlags(fs)
	config.LongRunningOptions.RegisterFlags(fs)
	config.AsyncOptions.RegisterFlags(fs)

	return &ffcli.Command{
		Name:       "delete",
		ShortUsage: "belvedere apps delete <name> [flags]",
		ShortHelp:  "Delete an application.",
		LongHelp: cmd.Wrap(`Delete an application.

An application must not have any releases before being deleted.`),
		FlagSet: fs,
		Exec:    config.Exec,
	}
}

func (c *Config) Exec(ctx context.Context, args []string) error {
	if len(args) < 1 {
		return flag.ErrHelp
	}
	name := args[0]

	return c.root.Project.Apps().Delete(ctx, name, c.DryRun, c.Async, c.Interval)
}
