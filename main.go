package main

import (
	"context"
	"fmt"
	"os"

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
		if err := enable(context.Background(), projectID); err != nil {
			panic(err)
		}
	}
}

func enable(ctx context.Context, projectID string) error {
	ctx, span := trace.StartSpan(context.Background(), "belvedere.enable")
	defer span.End()

	span.AddAttributes(trace.StringAttribute("project_id", projectID))

	if err := belvedere.EnableServices(ctx, projectID); err != nil {
		return err
	}

	return belvedere.EnableDeploymentManagerIAM(ctx, projectID)
}
