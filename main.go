package main

import (
	"context"

	"github.com/codahale/belvedere/pkg/belvedere"
	"go.opencensus.io/examples/exporter"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/trace"
)

func main() {
	pe := &exporter.PrintExporter{}
	trace.RegisterExporter(pe)
	trace.ApplyConfig(trace.Config{
		DefaultSampler: trace.AlwaysSample(),
	})
	view.RegisterExporter(pe)

	if err := belvedere.EnableServices(context.Background(), "codahale-com"); err != nil {
		panic(err)
	}
}
