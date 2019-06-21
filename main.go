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
  belvedere initialize <project> <dns zone> [options]
  belvedere apps list <project> [options]
  belvedere apps create <project> <app> <config> [options]
  belvedere apps destroy <project> <app> [options] 
  belvedere releases list <project> <app> [options]
  belvedere releases create <project> <app> <release> <config> <image> [options]
  belvedere releases enable <project> <app> <release> [options]
  belvedere releases disable <project> <app> <release> [options]
  belvedere releases destroy <project> <app> <release> [options]
  belvedere -h | --help
  belvedere --version

Options:
  -h --help     Show this screen.
  --version     Show version.
  --debug       Enable debug output.
  --quiet       Disable all log output.
`

	// Parse arguments and options.
	opts, err := docopt.ParseArgs(usage, nil, buildVersion)
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(-1)
	}

	// Export all traces.
	trace.ApplyConfig(trace.Config{
		DefaultSampler: trace.AlwaysSample(),
	})

	if debug, err := opts.Bool("--debug"); err == nil && debug {
		// Use the print exporter for debugging, as it prints everything.
		pe := &exporter.PrintExporter{}
		trace.RegisterExporter(pe)
		view.RegisterExporter(pe)
	} else if quiet, err := opts.Bool("--quiet"); err != nil || !quiet {
		// Unless we're quiet, use the trace logger for more practical logging.
		trace.RegisterExporter(&traceLogger{})
	}

	// Run commands.
	if err := run(context.Background(), opts); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(-1)
	}
}

func run(ctx context.Context, opts docopt.Opts) error {
	// Create a root span with some common attributes.
	ctx, span := trace.StartSpan(ctx, "belvedere")
	if hostname, err := os.Hostname(); err == nil {
		span.AddAttributes(trace.StringAttribute("hostname", hostname))
	}
	if u, err := user.Current(); err == nil {
		span.AddAttributes(trace.StringAttribute("user", u.Username))
	}
	defer span.End()

	projectID, _ := opts.String("<project>")
	switch {
	case isCmd(opts, "initialize"):
		dnsZone, _ := opts.String("<dns zone>")
		return belvedere.Initialize(ctx, projectID, dnsZone)
	case isCmd(opts, "apps", "list"):
		// TODO implement app printing
		return belvedere.ListApps(ctx, projectID)
	case isCmd(opts, "apps", "create"):
		appName, _ := opts.String("<app>")
		configPath, _ := opts.String("<config>")
		config, err := belvedere.LoadAppConfig(configPath)
		if err != nil {
			return err
		}
		return belvedere.CreateApp(ctx, projectID, appName, config)
	case isCmd(opts, "apps", "destroy"):
		appName, _ := opts.String("<app>")
		return belvedere.DestroyApp(ctx, projectID, appName)
	case isCmd(opts, "releases", "list"):
		appName, _ := opts.String("<app>")
		// TODO implement release printing
		return belvedere.ListReleases(ctx, projectID, appName)
	case isCmd(opts, "releases", "create"):
		appName, _ := opts.String("<app>")
		relName, _ := opts.String("<release>")
		image, _ := opts.String("<image>")
		configPath, _ := opts.String("<config>")
		config, err := belvedere.LoadReleaseConfig(configPath)
		if err != nil {
			return err
		}
		return belvedere.CreateRelease(ctx, projectID, appName, relName, config, image)
	case isCmd(opts, "releases", "enable"):
		appName, _ := opts.String("<app>")
		relName, _ := opts.String("<release>")
		return belvedere.EnableRelease(ctx, projectID, appName, relName)
	case isCmd(opts, "releases", "disable"):
		appName, _ := opts.String("<app>")
		relName, _ := opts.String("<release>")
		return belvedere.DisableRelease(ctx, projectID, appName, relName)
	case isCmd(opts, "releases", "destroy"):
		appName, _ := opts.String("<app>")
		relName, _ := opts.String("<release>")
		return belvedere.DestroyRelease(ctx, projectID, appName, relName)
	}
	return fmt.Errorf("unimplemented: %v", opts)
}

func isCmd(opts docopt.Opts, commands ...string) bool {
	for _, s := range commands {
		if ok, _ := opts.Bool(s); !ok {
			return false
		}
	}
	return true
}
