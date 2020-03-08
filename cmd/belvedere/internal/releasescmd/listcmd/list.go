package listcmd

import (
	"context"
	"flag"

	"github.com/codahale/belvedere/cmd/belvedere/internal/cmd"
	"github.com/codahale/belvedere/cmd/belvedere/internal/rootcmd"
	"github.com/peterbourgon/ff/v2/ffcli"
)

type Config struct {
	root *rootcmd.Config
}

func New(root *rootcmd.Config) *ffcli.Command {
	cfg := Config{root: root}

	fs := flag.NewFlagSet("belvedere releases list", flag.ExitOnError)
	root.RegisterFlags(fs)

	return &ffcli.Command{
		Name:       "list",
		ShortUsage: "belvedere releases list [<app>] [flags]",
		ShortHelp:  "List releases.",
		LongHelp: cmd.Wrap(`List releases.

The list of releases can by filtered by application.`),
		FlagSet: fs,
		Exec:    cfg.Exec,
	}
}

func (c *Config) Exec(ctx context.Context, args []string) error {
	if len(args) > 1 {
		return flag.ErrHelp
	}
	var app string
	if len(args) > 0 {
		app = args[0]
	}
	releases, err := c.root.Project.Releases().List(ctx, app)
	if err != nil {
		return err
	}
	return c.root.Tables.Print(releases)
}
