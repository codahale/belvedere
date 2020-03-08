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
}

func New(root *rootcmd.Config) *ffcli.Command {
	config := Config{root: root}

	fs := flag.NewFlagSet("belvedere secrets delete", flag.ExitOnError)
	root.RegisterFlags(fs)
	config.ModifyOptions.RegisterFlags(fs)

	return &ffcli.Command{
		Name:       "delete",
		ShortUsage: "belvedere secrets delete <name> [flags]",
		ShortHelp:  "Delete a secret.",
		LongHelp: cmd.Wrap(`Delete a secret.

This deletes all versions of the secret as well, and cannot be undone.`),
		FlagSet: fs,
		Exec:    config.Exec,
	}
}

func (c *Config) Exec(ctx context.Context, args []string) error {
	if len(args) < 1 {
		return flag.ErrHelp
	}
	return c.root.Project.Secrets().Delete(ctx, args[0], c.DryRun)
}
