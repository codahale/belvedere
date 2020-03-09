package releasescmd

import (
	"flag"

	"github.com/codahale/belvedere/cmd/belvedere/internal/cmd"
	"github.com/codahale/belvedere/cmd/belvedere/internal/releasescmd/createcmd"
	"github.com/codahale/belvedere/cmd/belvedere/internal/releasescmd/deletecmd"
	"github.com/codahale/belvedere/cmd/belvedere/internal/releasescmd/disablecmd"
	"github.com/codahale/belvedere/cmd/belvedere/internal/releasescmd/enablecmd"
	"github.com/codahale/belvedere/cmd/belvedere/internal/releasescmd/listcmd"
	"github.com/codahale/belvedere/cmd/belvedere/internal/rootcmd"
	"github.com/peterbourgon/ff/v3/ffcli"
)

func New(root *rootcmd.Config) *ffcli.Command {
	fs := flag.NewFlagSet("belvedere releases", flag.ExitOnError)
	root.RegisterFlags(fs)

	return &ffcli.Command{
		Name:       "releases",
		ShortUsage: "belvedere releases <subcommand> [<arg>...] [flags]",
		ShortHelp:  "Commands for managing releases.",
		LongHelp: cmd.Wrap(`Commands for managing releases.

A release is a specific Docker image (identified by its SHA-256 digest) and a managed instance group
of GCE instances running the Docker image.`),
		FlagSet: fs,
		Subcommands: []*ffcli.Command{
			listcmd.New(root),
			createcmd.New(root),
			enablecmd.New(root),
			disablecmd.New(root),
			deletecmd.New(root),
		},
	}
}
