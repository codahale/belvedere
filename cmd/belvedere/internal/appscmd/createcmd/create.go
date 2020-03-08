package createcmd

import (
	"context"
	"flag"

	"github.com/codahale/belvedere/cmd/belvedere/internal/cmd"
	"github.com/codahale/belvedere/cmd/belvedere/internal/rootcmd"
	bcfg "github.com/codahale/belvedere/pkg/belvedere/cfg"
	"github.com/peterbourgon/ff/v2/ffcli"
)

type Config struct {
	root *rootcmd.Config
	cmd.ModifyOptions
	cmd.LongRunningOptions
}

func New(root *rootcmd.Config) *ffcli.Command {
	cfg := Config{root: root}

	fs := flag.NewFlagSet("belvedere apps create", flag.ExitOnError)
	root.RegisterFlags(fs)
	cfg.ModifyOptions.RegisterFlags(fs)
	cfg.LongRunningOptions.RegisterFlags(fs)

	return &ffcli.Command{
		Name:       "create",
		ShortUsage: "belvedere apps create <region> <name> [<config-file>] [flags]",
		ShortHelp:  "Create an application.",
		LongHelp: cmd.Wrap(`Create an application.

The resources which run an application are provisioned inside a GCP region (e.g. us-west1), and this
property cannot be changed once the application is created.

If config-file is not specified (or is specified as '-'), the configuration file is read from STDIN
instead.`),
		FlagSet: fs,
		Exec:    cfg.Exec,
	}
}

func (c *Config) Exec(ctx context.Context, args []string) error {
	if len(args) < 2 || len(args) > 3 {
		return flag.ErrHelp
	}
	region := args[0]
	name := args[1]
	path := "-"
	if len(args) > 2 {
		path = args[2]
	}

	b, err := c.root.Files.Read(path)
	if err != nil {
		return err
	}

	config, err := bcfg.Parse(b)
	if err != nil {
		return err
	}

	return c.root.Project.Apps().Create(ctx, region, name, config, c.DryRun, c.Interval)
}
