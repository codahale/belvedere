package teardowncmd

import (
	"context"
	"flag"

	"github.com/codahale/belvedere/cmd/belvedere/internal/cmd"
	"github.com/codahale/belvedere/cmd/belvedere/internal/rootcmd"
	"github.com/peterbourgon/ff/v2/ffcli"
)

type Config struct {
	root *rootcmd.Config
	cmd.LongRunningOptions
	cmd.ModifyOptions
	cmd.AsyncOptions
}

func New(root *rootcmd.Config) *ffcli.Command {
	config := Config{root: root}

	fs := flag.NewFlagSet("belvedere teardown", flag.ExitOnError)
	root.RegisterFlags(fs)
	config.ModifyOptions.RegisterFlags(fs)
	config.LongRunningOptions.RegisterFlags(fs)
	config.AsyncOptions.RegisterFlags(fs)

	return &ffcli.Command{
		Name:       "teardown",
		ShortUsage: "belvedere teardown [flags]",
		ShortHelp:  "Remove all Belvedere resources from this project.",
		LongHelp: cmd.Wrap(`Remove all Belvedere resources from this project.

Deletes the base Deployment Manager deployment.`,
		),
		FlagSet: fs,
		Exec:    config.Exec,
	}
}

func (c *Config) Exec(ctx context.Context, args []string) error {
	if len(args) != 0 {
		return flag.ErrHelp
	}
	return c.root.Project.Teardown(ctx, c.DryRun, c.Async, c.Interval)
}
