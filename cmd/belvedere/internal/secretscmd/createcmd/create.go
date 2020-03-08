package createcmd

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

	fs := flag.NewFlagSet("belvedere secrets create", flag.ExitOnError)
	root.RegisterFlags(fs)
	config.ModifyOptions.RegisterFlags(fs)

	return &ffcli.Command{
		Name:       "create",
		ShortUsage: "belvedere secrets create <name> [<data-file>] [flags]",
		ShortHelp:  "Create a secret.",
		LongHelp: cmd.Wrap(`Create a secret.

Creates a new secret with a value that is the contents of data-file, read as a bytestring.

If data-file is not specified (or is specified as '-'), the secret's value is read from STDIN
instead.`),
		FlagSet: fs,
		Exec:    config.Exec,
	}
}

func (c *Config) Exec(ctx context.Context, args []string) error {
	if len(args) < 1 || len(args) > 2 {
		return flag.ErrHelp
	}
	name := args[0]
	path := "-"
	if len(args) > 1 {
		path = args[1]
	}

	b, err := c.root.Files.Read(path)
	if err != nil {
		return err
	}

	return c.root.Project.Secrets().Create(ctx, name, b, c.DryRun)
}