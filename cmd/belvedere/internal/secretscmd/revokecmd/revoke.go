package revokecmd

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
	cfg := Config{root: root}

	fs := flag.NewFlagSet("belvedere secrets revoke", flag.ExitOnError)
	root.RegisterFlags(fs)
	cfg.ModifyOptions.RegisterFlags(fs)

	return &ffcli.Command{
		Name:       "grant",
		ShortUsage: "belvedere secrets revoke <name> <app> [flags]",
		ShortHelp:  "Revoke an application's access to a secret.",
		LongHelp: cmd.Wrap(`Revoke an application access to a secret.

This modifies the secret's IAM policy to disallow the application's service account access to the
secrets' value.`),
		FlagSet: fs,
		Exec:    cfg.Exec,
	}
}

func (c *Config) Exec(ctx context.Context, args []string) error {
	if len(args) != 2 {
		return flag.ErrHelp
	}
	return c.root.Project.Secrets().Revoke(ctx, args[0], args[1], c.DryRun)
}
