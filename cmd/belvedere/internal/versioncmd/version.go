package versioncmd

import (
	"context"
	"flag"
	"fmt"
	"io"

	"github.com/codahale/belvedere/cmd/belvedere/internal/cmd"
	"github.com/codahale/belvedere/cmd/belvedere/internal/rootcmd"
	"github.com/peterbourgon/ff/v3/ffcli"
)

func New(root *rootcmd.Config, w io.Writer, version, commit, date, builtBy string) *ffcli.Command {
	fs := flag.NewFlagSet("belvedere version", flag.ExitOnError)
	root.RegisterFlags(fs)

	return &ffcli.Command{
		Name:       "version",
		ShortUsage: "belvedere version",
		ShortHelp:  "Print version information for Belvedere.",
		LongHelp: cmd.Wrap(`Print version information for Belvedere.

For released binaries, this includes the commit hash, the build date, and the builder.`),
		FlagSet: fs,
		Exec: func(ctx context.Context, args []string) error {
			_, err := fmt.Fprintln(w, buildVersion(version, commit, date, builtBy))
			return err
		},
	}
}

func buildVersion(version, commit, date, builtBy string) string {
	var result = fmt.Sprintf("version: %s", version)
	if commit != "" {
		result = fmt.Sprintf("%s\ncommit: %s", result, commit)
	}
	if date != "" {
		result = fmt.Sprintf("%s\nbuilt at: %s", result, date)
	}
	if builtBy != "" {
		result = fmt.Sprintf("%s\nbuilt by: %s", result, builtBy)
	}
	return result
}
