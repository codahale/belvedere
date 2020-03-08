package enablecmd

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

	fs := flag.NewFlagSet("belvedere releases enable", flag.ExitOnError)
	root.RegisterFlags(fs)
	cfg.ModifyOptions.RegisterFlags(fs)
	cfg.LongRunningOptions.RegisterFlags(fs)

	return &ffcli.Command{
		Name:       "enable",
		ShortUsage: "belvedere releases enable <app> <name> [flags]",
		ShortHelp:  "Enable a release.",
		LongHelp: cmd.Wrap(`Enable a release.

Enabling a release registers the release's managed instance group with the application's load
balancer and waits for the instances to pass health checks and go into service. Use the -timeout
flag to bound the amount of time allowed for health checks to pass.`),
		FlagSet: fs,
		Exec:    cfg.Exec,
	}
}

func (c *Config) Exec(ctx context.Context, args []string) error {
	if len(args) != 2 {
		return flag.ErrHelp
	}
	return c.root.Project.Releases().Enable(ctx, args[0], args[1], c.DryRun, c.Interval)
}
