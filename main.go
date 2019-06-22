package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
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
  belvedere setup <dns zone> [options]
  belvedere apps list [options]
  belvedere apps create <app> <region> <config> [options]
  belvedere apps destroy <app> [options] 
  belvedere releases list <app> [options]
  belvedere releases create <app> <release> <config> <image> [options]
  belvedere releases enable <app> <release> [options]
  belvedere releases disable <app> <release> [options]
  belvedere releases destroy <app> <release> [options]
  belvedere -h | --help
  belvedere --version

Options:
  -h --help              Show this screen.
  --version              Show version.
  --debug                Enable debug output.
  --quiet                Disable all log output.
  --project=<project-id> Specify a project ID. Defaults to using gcloud's config.
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

	project, region, err := config(ctx, opts)
	if err != nil {
		return err
	}

	switch {
	case isCmd(opts, "setup"):
		dnsZone, _ := opts.String("<dns zone>")
		return belvedere.Setup(ctx, project, dnsZone)
	case isCmd(opts, "apps", "list"):
		apps, err := belvedere.ListApps(ctx, project)
		if err != nil {
			return err
		}

		for _, app := range apps {
			fmt.Println(app)
		}
		return nil
	case isCmd(opts, "apps", "create"):
		appName, _ := opts.String("<app>")
		region, _ := opts.String("<app>")
		path, _ := opts.String("<config>")
		config, err := belvedere.LoadAppConfig(ctx, path)
		if err != nil {
			return err
		}
		return belvedere.CreateApp(ctx, project, region, appName, config)
	case isCmd(opts, "apps", "destroy"):
		appName, _ := opts.String("<app>")
		return belvedere.DestroyApp(ctx, project, appName)
	case isCmd(opts, "releases", "list"):
		appName, _ := opts.String("<app>")
		releases, err := belvedere.ListReleases(ctx, project, appName)
		if err != nil {
			return err
		}
		for _, release := range releases {
			fmt.Println(release)
		}
		return nil
	case isCmd(opts, "releases", "create"):
		appName, _ := opts.String("<app>")
		relName, _ := opts.String("<release>")
		image, _ := opts.String("<image>")
		path, _ := opts.String("<config>")
		config, err := belvedere.LoadReleaseConfig(ctx, path)
		if err != nil {
			return err
		}
		return belvedere.CreateRelease(ctx, project, region, appName, relName, config, image)
	case isCmd(opts, "releases", "enable"):
		appName, _ := opts.String("<app>")
		relName, _ := opts.String("<release>")
		return belvedere.EnableRelease(ctx, project, region, appName, relName)
	case isCmd(opts, "releases", "disable"):
		appName, _ := opts.String("<app>")
		relName, _ := opts.String("<release>")
		return belvedere.DisableRelease(ctx, project, region, appName, relName)
	case isCmd(opts, "releases", "destroy"):
		appName, _ := opts.String("<app>")
		relName, _ := opts.String("<release>")
		return belvedere.DestroyRelease(ctx, project, appName, relName)
	default:
		return fmt.Errorf("unimplemented: %v", opts)
	}
}

func isCmd(opts docopt.Opts, commands ...string) bool {
	for _, s := range commands {
		if ok, _ := opts.Bool(s); !ok {
			return false
		}
	}
	return true
}

func config(ctx context.Context, opts docopt.Opts) (project, region string, err error) {
	ctx, span := trace.StartSpan(ctx, "belvedere.config")
	defer func() {
		span.AddAttributes(
			trace.StringAttribute("project", project),
			trace.StringAttribute("region", region),
		)
	}()
	defer span.End()

	project, _ = opts.String("--project")
	region, _ = opts.String("--region")

	if project != "" && region != "" {
		return
	}

	cmd := exec.Command("gcloud", "config", "config-helper", "--format=json")
	b, err := cmd.Output()
	if err != nil {
		return
	}

	var config struct {
		Configuration struct {
			Properties struct {
				Core struct {
					Project string `json:"project"`
				} `json:"core"`
				Compute struct {
					Region string `json:"region"`
				} `json:"compute"`
			} `json:"properties"`
		} `json:"configuration"`
	}
	if err = json.Unmarshal(b, &config); err != nil {
		return
	}

	if project == "" {
		project = config.Configuration.Properties.Core.Project
	}

	if region == "" {
		region = config.Configuration.Properties.Compute.Region
	}

	if project == "" {
		return "", "", errors.New("project not found")
	}

	if region == "" {
		return "", "", errors.New("region not found")
	}

	return
}
