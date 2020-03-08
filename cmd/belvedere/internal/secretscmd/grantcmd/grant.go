package grantcmd

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
}

func New(root *rootcmd.Config) *ffcli.Command {
	config := Config{root: root}

	fs := flag.NewFlagSet("belvedere secrets grant", flag.ExitOnError)
	root.RegisterFlags(fs)
	config.ModifyOptions.RegisterFlags(fs)

	return &ffcli.Command{
		Name:       "grant",
		ShortUsage: "belvedere secrets grant <name> <app> [flags]",
		ShortHelp:  "Grant an application access to a secret.",
		LongHelp: cmd.Wrap(`Grant an application access to a secret.

This modifies the secret's IAM policy to allow the application's service account access to the
secrets' value.`),
		FlagSet: fs,
		Exec:    config.Exec,
	}
}

func (c *Config) Exec(ctx context.Context, args []string) error {
	if len(args) != 2 {
		return flag.ErrHelp
	}
	return c.root.Project.Secrets().Grant(ctx, args[0], args[1], c.DryRun)
}
