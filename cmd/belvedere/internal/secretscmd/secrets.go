package secretscmd

import (
	"flag"

	"github.com/codahale/belvedere/cmd/belvedere/internal/cmd"
	"github.com/codahale/belvedere/cmd/belvedere/internal/rootcmd"
	"github.com/codahale/belvedere/cmd/belvedere/internal/secretscmd/createcmd"
	"github.com/codahale/belvedere/cmd/belvedere/internal/secretscmd/deletecmd"
	"github.com/codahale/belvedere/cmd/belvedere/internal/secretscmd/grantcmd"
	"github.com/codahale/belvedere/cmd/belvedere/internal/secretscmd/listcmd"
	"github.com/codahale/belvedere/cmd/belvedere/internal/secretscmd/revokecmd"
	"github.com/codahale/belvedere/cmd/belvedere/internal/secretscmd/updatecmd"
	"github.com/peterbourgon/ff/v3/ffcli"
)

func New(root *rootcmd.Config) *ffcli.Command {
	fs := flag.NewFlagSet("belvedere secrets", flag.ExitOnError)
	root.RegisterFlags(fs)

	return &ffcli.Command{
		Name:       "secrets",
		ShortUsage: "belvedere secrets <subcommand> [<arg>...] [flags]",
		ShortHelp:  "Commands for managing secrets.",
		LongHelp: cmd.Wrap(`Commands for managing secrets.

Secrets are stored in Google Secret Manager, which provides integrity and confidentiality both at
rest and in flight, strong audit logging, and access control via IAM permissions. Secrets' values
are versioned, allowing for update rollouts and rollbacks.
`),
		FlagSet: fs,
		Subcommands: []*ffcli.Command{
			listcmd.New(root),
			createcmd.New(root),
			updatecmd.New(root),
			deletecmd.New(root),
			grantcmd.New(root),
			revokecmd.New(root),
		},
	}
}
