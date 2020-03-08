package disablecmd

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
}

func New(root *rootcmd.Config) *ffcli.Command {
	cfg := Config{root: root}

	fs := flag.NewFlagSet("belvedere releases disable", flag.ExitOnError)
	root.RegisterFlags(fs)
	cfg.ModifyOptions.RegisterFlags(fs)
	cfg.LongRunningOptions.RegisterFlags(fs)

	return &ffcli.Command{
		Name:       "disable",
		ShortUsage: "belvedere releases disable <app> <name> [flags]",
		ShortHelp:  "Disable a release.",
		LongHelp: cmd.Wrap(`Disable a release.

Disabling a release unregisters the release's managed instance group from the application's load
balancer, removing it from service. This is used on old releases to roll a deploy forward or on new
releases which did not pass health checks in order to roll a deploy back.`),
		FlagSet: fs,
		Exec:    cfg.Exec,
	}
}

func (c *Config) Exec(ctx context.Context, args []string) error {
	if len(args) != 2 {
		return flag.ErrHelp
	}
	return c.root.Project.Releases().Disable(ctx, args[0], args[1], c.DryRun, c.Interval)
}
