package listcmd

import (
	"context"
	"flag"

	"github.com/codahale/belvedere/cmd/belvedere/internal/cmd"
	"github.com/codahale/belvedere/cmd/belvedere/internal/rootcmd"
	"github.com/peterbourgon/ff/v3/ffcli"
)

type Config struct {
	root *rootcmd.Config
}

func New(root *rootcmd.Config) *ffcli.Command {
	config := Config{root: root}

	fs := flag.NewFlagSet("belvedere secrets list", flag.ExitOnError)
	root.RegisterFlags(fs)

	return &ffcli.Command{
		Name:       "list",
		ShortUsage: "belvedere secrets list",
		ShortHelp:  "List secrets.",
		LongHelp: cmd.Wrap(`List secrets.

Because applications may share secrets (e.g. two applications both need to use the same API key),
secrets exist in their own namespace.`),
		FlagSet: fs,
		Exec:    config.Exec,
	}
}

func (c *Config) Exec(ctx context.Context, _ []string) error {
	apps, err := c.root.Project.Secrets().List(ctx)
	if err != nil {
		return err
	}
	return c.root.Tables.Print(apps)
}
