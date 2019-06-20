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

func main() {
	usage := `Belvedere: A fine place from which to survey your estate.

Usage:
  belvedere enable [--debug]
  belvedere -h | --help
  belvedere --version

Options:
  -h --help     Show this screen.
  --version     Show version.
  --debug       Enable debug output.
`

	opts, err := docopt.ParseDoc(usage)
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
		trace.RegisterExporter(&belvedere.TraceLogger{})
	}

	if enable, _ := opts.Bool("enable"); enable {
		if err := belvedere.EnableServices(context.Background(), "codahale-com"); err != nil {
			panic(err)
		}
	}
}
