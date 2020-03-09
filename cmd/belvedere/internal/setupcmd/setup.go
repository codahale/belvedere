package setupcmd

import (
	"context"
	"flag"

	"github.com/codahale/belvedere/cmd/belvedere/internal/cmd"
	"github.com/codahale/belvedere/cmd/belvedere/internal/rootcmd"
	"github.com/peterbourgon/ff/v3/ffcli"
)

type Config struct {
	root *rootcmd.Config
	cmd.ModifyOptions
	cmd.LongRunningOptions
}

func New(root *rootcmd.Config) *ffcli.Command {
	config := Config{root: root}

	fs := flag.NewFlagSet("belvedere setup", flag.ExitOnError)
	root.RegisterFlags(fs)
	config.ModifyOptions.RegisterFlags(fs)
	config.LongRunningOptions.RegisterFlags(fs)

	return &ffcli.Command{
		Name:       "setup",
		ShortUsage: "belvedere setup <dns-name> [flags]",
		ShortHelp:  "Initialize a GCP project for use with Belvedere.",
		LongHelp: cmd.Wrap(`Initialize a GCP project for use with Belvedere.

Enables all required GCP APIs, grants Deployment Manager access to manage IAM permissions, and
creates a Deployment Manager deployment with the base resources and configuration required to
create, deploy, and manage applications with Belvedere.`,
		),
		FlagSet: fs,
		Exec:    config.Exec,
	}
}

func (c *Config) Exec(ctx context.Context, args []string) error {
	if len(args) != 1 {
		return flag.ErrHelp
	}
	return c.root.Project.Setup(ctx, args[0], c.DryRun, c.Interval)
}
