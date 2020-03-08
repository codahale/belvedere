package createcmd

import (
	"context"
	"flag"

	"github.com/codahale/belvedere/cmd/belvedere/internal/cmd"
	"github.com/codahale/belvedere/cmd/belvedere/internal/rootcmd"
	"github.com/codahale/belvedere/pkg/belvedere/cfg"
	"github.com/peterbourgon/ff/v2/ffcli"
)

type Config struct {
	root *rootcmd.Config
	cmd.ModifyOptions
	cmd.LongRunningOptions
	enable bool
}

func New(root *rootcmd.Config) *ffcli.Command {
	config := Config{root: root}

	fs := flag.NewFlagSet("belvedere releases create", flag.ExitOnError)
	root.RegisterFlags(fs)
	config.ModifyOptions.RegisterFlags(fs)
	config.LongRunningOptions.RegisterFlags(fs)
	config.RegisterFlags(fs)

	return &ffcli.Command{
		Name:       "create",
		ShortUsage: "belvedere releases create <app> <name> <sha-256> [<config-file>] [flags]",
		ShortHelp:  "Create a release.",
		LongHelp: cmd.Wrap(`Create a release.

This requires the application name, a release name, the SHA-256 digest of the Docker image to
deploy, and the application configuration.

If config-file is not specified (or is specified as '-'), the configuration file is read from STDIN
instead.`),
		FlagSet: fs,
		Exec:    config.Exec,
	}
}

func (c *Config) RegisterFlags(fs *flag.FlagSet) {
	fs.BoolVar(&c.enable, "enable", false, "enable the release after its successful creation")
}

func (c *Config) Exec(ctx context.Context, args []string) error {
	if len(args) < 3 || len(args) > 4 {
		return flag.ErrHelp
	}
	app := args[0]
	name := args[1]
	digest := args[3]

	path := "-"
	if len(args) > 3 {
		path = args[4]
	}

	b, err := c.root.Files.Read(path)
	if err != nil {
		return err
	}

	config, err := cfg.Parse(b)
	if err != nil {
		return err
	}

	if err := c.root.Project.Releases().Create(ctx, app, name, config, digest, c.DryRun, c.Interval); err != nil {
		return err
	}

	if c.enable {
		return c.root.Project.Releases().Enable(ctx, app, name, c.DryRun, c.Interval)
	}
	return nil
}
