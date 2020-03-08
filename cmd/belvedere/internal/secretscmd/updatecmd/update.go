package updatecmd

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

	fs := flag.NewFlagSet("belvedere secrets update", flag.ExitOnError)
	root.RegisterFlags(fs)
	cfg.ModifyOptions.RegisterFlags(fs)

	return &ffcli.Command{
		Name:       "update",
		ShortUsage: "belvedere secrets update <name> [<data-file>] [flags]",
		ShortHelp:  "Update a secret.",
		LongHelp: cmd.Wrap(`Update a secret.

Updates the secret's value to be the contents of data-file, read as a bytestring.

If data-file is not specified (or is specified as '-'), the secret's value is read from STDIN
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
		path = args[2]
	}

	b, err := c.root.Files.Read(path)
	if err != nil {
		return err
	}

	return c.root.Project.Secrets().Update(ctx, name, b, c.DryRun)
}
