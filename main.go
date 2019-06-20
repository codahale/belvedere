package main

import (
	"context"
	"fmt"
	"os"
	"os/user"

	"github.com/codahale/belvedere/pkg/belvedere"
	"github.com/docopt/docopt-go"
	"go.opencensus.io/examples/exporter"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/trace"
)

var (
	buildVersion = "unknown"
)

func main() {
	usage := `Belvedere: A fine place from which to survey your estate.

Usage:
  belvedere enable <project-id> [--debug]
  belvedere -h | --help
  belvedere --version

Options:
  -h --help     Show this screen.
  --version     Show version.
  --debug       Enable debug output.
`

	opts, err := docopt.ParseArgs(usage, nil, buildVersion)
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(-1)
	}

	trace.ApplyConfig(trace.Config{
		DefaultSampler: trace.AlwaysSample(),
	})

	if debug, err := opts.Bool("--debug"); err == nil && debug {
		pe := &exporter.PrintExporter{}
		trace.RegisterExporter(pe)
		view.RegisterExporter(pe)
	} else {
		trace.RegisterExporter(belvedere.NewTraceLogger())
	}

	if ok, _ := opts.Bool("enable"); ok {
		projectID, _ := opts.String("<project-id>")
		if err := enable(projectID); err != nil {
			panic(err)
		}
	}
}

func enable(projectID string) error {
	ctx, span := rootSpan("belvedere.enable")
	span.AddAttributes(trace.StringAttribute("project_id", projectID))
	defer span.End()

	if err := belvedere.EnableServices(ctx, projectID); err != nil {
		return err
	}
	return belvedere.EnableDeploymentManagerIAM(ctx, projectID)
}

func rootSpan(name string) (context.Context, *trace.Span) {
	ctx, span := trace.StartSpan(context.Background(), name)
	if hostname, err := os.Hostname(); err == nil {
		span.AddAttributes(trace.StringAttribute("hostname", hostname))
	}
	if u, err := user.Current(); err == nil {
		span.AddAttributes(trace.StringAttribute("user", u.Username))
	}
	return ctx, span
}
