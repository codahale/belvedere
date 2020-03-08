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
	config := Config{root: root}

	fs := flag.NewFlagSet("belvedere apps list", flag.ExitOnError)
	root.RegisterFlags(fs)

	return &ffcli.Command{
		Name:       "list",
		ShortUsage: "belvedere apps list",
		ShortHelp:  "List applications.",
		LongHelp: cmd.Wrap(`List applications.

Prints a table of provisioned applications in the current project.`),
		FlagSet: fs,
		Exec:    config.Exec,
	}
}

func (c *Config) Exec(ctx context.Context, _ []string) error {
	apps, err := c.root.Project.Apps().List(ctx)
	if err != nil {
		return err
	}
	return c.root.Tables.Print(apps)
}
