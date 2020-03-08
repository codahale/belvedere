package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/alessio/shellescape"
	"github.com/codahale/belvedere/cmd/belvedere/internal/appscmd"
	"github.com/codahale/belvedere/cmd/belvedere/internal/cmd"
	"github.com/codahale/belvedere/cmd/belvedere/internal/dnsserverscmd"
	"github.com/codahale/belvedere/cmd/belvedere/internal/instancescmd"
	"github.com/codahale/belvedere/cmd/belvedere/internal/logscmd"
	"github.com/codahale/belvedere/cmd/belvedere/internal/machinetypescmd"
	"github.com/codahale/belvedere/cmd/belvedere/internal/releasescmd"
	"github.com/codahale/belvedere/cmd/belvedere/internal/rootcmd"
	"github.com/codahale/belvedere/cmd/belvedere/internal/secretscmd"
	"github.com/codahale/belvedere/cmd/belvedere/internal/setupcmd"
	"github.com/codahale/belvedere/cmd/belvedere/internal/sshcmd"
	"github.com/codahale/belvedere/cmd/belvedere/internal/teardowncmd"
	"github.com/codahale/belvedere/cmd/belvedere/internal/versioncmd"
	"github.com/codahale/belvedere/pkg/belvedere"
	"github.com/peterbourgon/ff/v2/ffcli"
	"go.opencensus.io/trace"
	"google.golang.org/genproto/googleapis/rpc/code"
)

func main() {
	callback, err := run()
	if err != nil {
		if err == flag.ErrHelp {
			os.Exit(2)
		} else {
			_, _ = fmt.Fprintln(os.Stderr, err)
			os.Exit(-1)
		}
	}

	if callback != nil {
		err = callback()
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(-1)
	}
}

func run() (func() error, error) {
	// Initialize root command.
	rootCommand, rootConfig := rootcmd.New()

	// Populate subcommands.
	rootCommand.Subcommands = []*ffcli.Command{
		setupcmd.New(rootConfig),
		teardowncmd.New(rootConfig),
		dnsserverscmd.New(rootConfig),
		instancescmd.New(rootConfig),
		logscmd.New(rootConfig),
		machinetypescmd.New(rootConfig),
		sshcmd.New(rootConfig),
		appscmd.New(rootConfig),
		releasescmd.New(rootConfig),
		secretscmd.New(rootConfig),
		versioncmd.New(rootConfig, os.Stdout, version, commit, date, builtBy),
	}

	// Parse flags and args.
	if err := rootCommand.Parse(os.Args[1:]); err != nil {
		// TODO handle no-op commands, e.g. belvedere apps
		return nil, err
	}

	// Enable trace logging.
	rootConfig.EnableLogging()

	// Create a root span.
	ctx, cancel, span := rootConfig.RootSpan()
	defer cancel()
	defer span.End()

	// Create a Belvedere project.
	project, err := belvedere.NewProject(ctx, rootConfig.ProjectName)
	if err != nil {
		return nil, err
	}
	span.AddAttributes(
		trace.StringAttribute("project", project.Name()),
		trace.StringAttribute("args", escapeArgs(os.Args[1:])),
	)
	rootConfig.Project = project

	// Initialize helpers.
	rootConfig.Tables = cmd.NewTableWriter(rootConfig.CSV)
	rootConfig.Files = cmd.NewFileReader()

	// Run the actual command.
	if err := rootCommand.Run(ctx); err != nil {
		if err == flag.ErrHelp {
			span.SetStatus(trace.Status{
				Code:    int32(code.Code_INVALID_ARGUMENT),
				Message: "invalid argument",
			})
		} else {
			span.SetStatus(trace.Status{
				Code:    int32(code.Code_INTERNAL),
				Message: err.Error(),
			})
		}
		return nil, err
	}
	return rootConfig.Callback, nil
}

func escapeArgs(args []string) string {
	escaped := make([]string, len(args))
	for i, s := range args {
		escaped[i] = shellescape.Quote(s)
	}
	return strings.Join(escaped, " ")
}

// nolint: gochecknoglobals
var (
	version = "dev"
	commit  = ""
	date    = ""
	builtBy = ""
)
