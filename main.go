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
  belvedere teardown [options]
  belvedere dns-servers [options]
  belvedere instances [<app>] [<release>] [options]
  belvedere ssh <instance> [options]
  belvedere apps list [options]
  belvedere apps create <app> <region> <config> [options]
  belvedere apps update <app> <config> [options]
  belvedere apps destroy <app> [options] 
  belvedere releases list <app> [options]
  belvedere releases create <app> <release> <config> <sha256> [--enable] [options]
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
  --dry-run              Print modifications instead of performing them.
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

	project, err := config(ctx, opts)
	if err != nil {
		return err
	}
	dryRun, _ := opts.Bool("--dry-run")
	span.AddAttributes(trace.BoolAttribute("dry_run", dryRun))

	switch {
	case isCmd(opts, "setup"):
		dnsZone, _ := opts.String("<dns zone>")
		return belvedere.Setup(ctx, project, dnsZone, dryRun)
	case isCmd(opts, "teardown"):
		return belvedere.Teardown(ctx, project, dryRun)
	case isCmd(opts, "dns-servers"):
		servers, err := belvedere.DNSServers(ctx, project)
		if err != nil {
			return err
		}
		for _, s := range servers {
			fmt.Println(s)
		}
		return nil
	case isCmd(opts, "instances"):
		appName, _ := opts.String("<app>")
		relName, _ := opts.String("<release>")

		instances, err := belvedere.ListInstances(ctx, project, appName, relName)
		if err != nil {
			return err
		}

		for _, app := range instances {
			fmt.Println(app)
		}
		return nil
	case isCmd(opts, "ssh"):
		instance, _ := opts.String("<instance>")
		return belvedere.SSH(ctx, project, instance)
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
		region, _ := opts.String("<region>")
		path, _ := opts.String("<config>")
		config, err := belvedere.LoadConfig(ctx, path)
		if err != nil {
			return err
		}
		return belvedere.CreateApp(ctx, project, region, appName, config, dryRun)
	case isCmd(opts, "apps", "update"):
		appName, _ := opts.String("<app>")
		path, _ := opts.String("<config>")
		config, err := belvedere.LoadConfig(ctx, path)
		if err != nil {
			return err
		}
		return belvedere.UpdateApp(ctx, project, appName, config, dryRun)
	case isCmd(opts, "apps", "destroy"):
		appName, _ := opts.String("<app>")
		return belvedere.DestroyApp(ctx, project, appName, dryRun)
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
		imageSHA256, _ := opts.String("<sha256>")
		path, _ := opts.String("<config>")
		enable, _ := opts.Bool("--enable")
		config, err := belvedere.LoadConfig(ctx, path)
		if err != nil {
			return err
		}

		if err := belvedere.CreateRelease(ctx, project, appName, relName, config, imageSHA256, dryRun); err != nil {
			return err
		}

		if enable {
			return belvedere.EnableRelease(ctx, project, appName, relName, dryRun)
		}

		return nil
	case isCmd(opts, "releases", "enable"):
		appName, _ := opts.String("<app>")
		relName, _ := opts.String("<release>")
		return belvedere.EnableRelease(ctx, project, appName, relName, dryRun)
	case isCmd(opts, "releases", "disable"):
		appName, _ := opts.String("<app>")
		relName, _ := opts.String("<release>")
		return belvedere.DisableRelease(ctx, project, appName, relName, dryRun)
	case isCmd(opts, "releases", "destroy"):
		appName, _ := opts.String("<app>")
		relName, _ := opts.String("<release>")

		if err := belvedere.DisableRelease(ctx, project, appName, relName, dryRun); err != nil {
			return err
		}
		return belvedere.DestroyRelease(ctx, project, appName, relName, dryRun)
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

func config(ctx context.Context, opts docopt.Opts) (project string, err error) {
	ctx, span := trace.StartSpan(ctx, "belvedere.config")
	defer func() {
		span.AddAttributes(
			trace.StringAttribute("project", project),
		)
	}()
	defer span.End()

	project, _ = opts.String("--project")

	if project != "" {
		return project, nil
	}

	cmd := exec.Command("gcloud", "config", "config-helper", "--format=json")
	b, err := cmd.Output()
	if err != nil {
		return "", err
	}

	var config struct {
		Configuration struct {
			Properties struct {
				Core struct {
					Project string `json:"project"`
				} `json:"core"`
			} `json:"properties"`
		} `json:"configuration"`
	}
	if err = json.Unmarshal(b, &config); err != nil {
		return "", err
	}

	if config.Configuration.Properties.Core.Project != "" {
		return config.Configuration.Properties.Core.Project, nil
	}

	return "", errors.New("project not found")
}
