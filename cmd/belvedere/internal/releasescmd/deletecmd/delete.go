package deletecmd

import (
	"context"
	"flag"

	"github.com/codahale/belvedere/cmd/belvedere/internal/cmd"
	"github.com/codahale/belvedere/cmd/belvedere/internal/rootcmd"
	"github.com/peterbourgon/ff/v2/ffcli"
)

type Config struct {
	root *rootcmd.Config
	cmd.ModifyOptions
	cmd.LongRunningOptions
	cmd.AsyncOptions
}

func New(root *rootcmd.Config) *ffcli.Command {
	cfg := Config{root: root}

	fs := flag.NewFlagSet("belvedere releases delete", flag.ExitOnError)
	root.RegisterFlags(fs)
	cfg.ModifyOptions.RegisterFlags(fs)
	cfg.LongRunningOptions.RegisterFlags(fs)
	cfg.AsyncOptions.RegisterFlags(fs)

	return &ffcli.Command{
		Name:       "delete",
		ShortUsage: "belvedere releases delete <app> <name> [flags]",
		ShortHelp:  "Delete a release.",
		LongHelp: cmd.Wrap(`Delete a release.

This deletes the managed instance group running the given release. Releases must be disabled before
they can be deleted.`),
		FlagSet: fs,
		Exec:    cfg.Exec,
	}
}

func (c *Config) Exec(ctx context.Context, args []string) error {
	if len(args) != 2 {
		return flag.ErrHelp
	}
	return c.root.Project.Releases().Delete(ctx, args[0], args[1], c.DryRun, c.Async, c.Interval)
}
