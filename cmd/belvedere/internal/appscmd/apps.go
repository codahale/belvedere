package appscmd

import (
	"flag"

	"github.com/codahale/belvedere/cmd/belvedere/internal/appscmd/createcmd"
	"github.com/codahale/belvedere/cmd/belvedere/internal/appscmd/deletecmd"
	"github.com/codahale/belvedere/cmd/belvedere/internal/appscmd/listcmd"
	"github.com/codahale/belvedere/cmd/belvedere/internal/appscmd/updatecmd"
	"github.com/codahale/belvedere/cmd/belvedere/internal/cmd"
	"github.com/codahale/belvedere/cmd/belvedere/internal/rootcmd"
	"github.com/peterbourgon/ff/v3/ffcli"
)

func New(root *rootcmd.Config) *ffcli.Command {
	fs := flag.NewFlagSet("belvedere apps", flag.ExitOnError)
	root.RegisterFlags(fs)

	return &ffcli.Command{
		Name:       "apps",
		ShortUsage: "belvedere apps <subcommand> [<arg>...] [flags]",
		ShortHelp:  "Commands for managing applications.",
		LongHelp: cmd.Wrap(`Commands for managing applications.

A Belvedere application is anything in a Docker container which accepts HTTP2 connections.`),
		FlagSet: fs,
		Subcommands: []*ffcli.Command{
			listcmd.New(root),
			createcmd.New(root),
			updatecmd.New(root),
			deletecmd.New(root),
		},
	}
}
