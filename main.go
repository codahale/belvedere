package main

//go:generate bash ./version.sh

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"text/tabwriter"
	"time"

	"github.com/codahale/belvedere/pkg/belvedere"
	"github.com/docopt/docopt-go"
	"go.opencensus.io/examples/exporter"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/trace"
)

var (
	version = "unknown"
)

func main() {
	usage := `
Belvedere: A fine place from which to survey your estate.

Usage:
  belvedere setup <dns zone> [options]
  belvedere teardown [--async] [options]
  belvedere dns-servers [options]
  belvedere machine-types [<region>] [options]
  belvedere instances [<app>] [<release>] [options]
  belvedere ssh <instance> [options]
  belvedere logs <app> [<release>] [<instance>] [--filter=<s>...] [--freshness=<duration>] [options]
  belvedere apps list [options]
  belvedere apps create <app> <region> <config> [options]
  belvedere apps update <app> <config> [options]
  belvedere apps delete <app> [--async] [options] 
  belvedere releases list [<app>] [options]
  belvedere releases create <app> <release> <config> <sha256> [--enable] [options]
  belvedere releases enable <app> <release> [options]
  belvedere releases disable <app> <release> [options]
  belvedere releases delete <app> <release> [--async] [options]
  belvedere -h | --help
  belvedere --version

Options:
  -h --help              Show this screen.
  --version              Show version.
  --debug                Enable debug output.
  --quiet                Disable all log output.
  --dry-run              Print modifications instead of performing them.
  --project=<project-id> Specify a project ID. Defaults to using gcloud's config.
  --interval=<duration>  Specify a polling interval for long-running operations [default: 10s].
  --timeout=<duration>   Specify a timeout for long-running operations [default: 5m].
`

	// Parse arguments and options.
	opts, err := docopt.ParseArgs(usage, nil, version)
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

	ctx := pollingContext(context.Background(), opts)

	// Run commands.
	if err := run(ctx, opts); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(-1)
	}

	if callback != nil {
		if err := callback(); err != nil {
			_, _ = fmt.Fprintln(os.Stderr, err)
			os.Exit(-1)
		}
	}
}

var callback func() error

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
		async, _ := opts.Bool("--async")
		return belvedere.Teardown(ctx, project, dryRun, async)
	case isCmd(opts, "dns-servers"):
		servers, err := belvedere.DNSServers(ctx, project)
		if err != nil {
			return err
		}
		for _, s := range servers {
			fmt.Println(s)
		}
		return nil
	case isCmd(opts, "machine-types"):
		region, _ := opts.String("<region>")
		machineTypes, err := belvedere.MachineTypes(ctx, project, region)
		if err != nil {
			return err
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		_, _ = fmt.Fprintln(w, "Name\tvCPUs\tMemory (MiB)")
		for _, mt := range machineTypes {
			_, _ = fmt.Fprintf(w, "%s\t%4d\t%10d\n", mt.Name, mt.CPU, mt.Memory)
		}
		return w.Flush()
	case isCmd(opts, "instances"):
		app, _ := opts.String("<app>")
		release, _ := opts.String("<release>")

		instances, err := belvedere.ListInstances(ctx, project, app, release)
		if err != nil {
			return err
		}

		for _, app := range instances {
			fmt.Println(app)
		}
		return nil
	case isCmd(opts, "ssh"):
		instance, _ := opts.String("<instance>")
		ssh, err := belvedere.SSH(ctx, project, instance)
		if err != nil {
			return err
		}
		callback = ssh
		return nil
	case isCmd(opts, "logs"):
		app, _ := opts.String("<app>")
		release, _ := opts.String("<release>")
		instance, _ := opts.String("<instance>")
		freshness, _ := opts.String("--freshness")
		if freshness == "" {
			freshness = "5m"
		}

		d, err := time.ParseDuration(freshness)
		if err != nil {
			return err
		}

		filters := opts["--filter"].([]string)

		logs, err := belvedere.Logs(ctx, project, app, release, instance, time.Now().Add(-d), filters)
		if err != nil {
			return err
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		_, _ = fmt.Fprintln(w, "Timestamp\tRelease\tInstance\tContainer\tMessage")
		for _, log := range logs {
			_, _ = fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
				log.Timestamp.Format(time.Stamp), log.Release, log.Instance, log.Container, log.Message)
		}
		return w.Flush()
	case isCmd(opts, "apps", "list"):
		apps, err := belvedere.ListApps(ctx, project)
		if err != nil {
			return err
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		_, _ = fmt.Fprintln(w, "Project\tRegion\tApp")
		for _, app := range apps {
			_, _ = fmt.Fprintf(w, "%s\t%s\t%s\n", app.Project, app.Region, app.Name)
		}
		return w.Flush()
	case isCmd(opts, "apps", "create"):
		app, _ := opts.String("<app>")
		region, _ := opts.String("<region>")
		path, _ := opts.String("<config>")
		config, err := belvedere.LoadConfig(ctx, path)
		if err != nil {
			return err
		}
		return belvedere.CreateApp(ctx, project, region, app, config, dryRun)
	case isCmd(opts, "apps", "update"):
		app, _ := opts.String("<app>")
		path, _ := opts.String("<config>")
		config, err := belvedere.LoadConfig(ctx, path)
		if err != nil {
			return err
		}
		return belvedere.UpdateApp(ctx, project, app, config, dryRun)
	case isCmd(opts, "apps", "delete"):
		app, _ := opts.String("<app>")
		async, _ := opts.Bool("--async")
		return belvedere.DeleteApp(ctx, project, app, dryRun, async)
	case isCmd(opts, "releases", "list"):
		app, _ := opts.String("<app>")
		releases, err := belvedere.ListReleases(ctx, project, app)
		if err != nil {
			return err
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		_, _ = fmt.Fprintln(w, "Project\tRegion\tApp\tRelease\tHash")
		for _, release := range releases {
			_, _ = fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
				release.Project, release.Region, release.App, release.Release, release.Hash)
		}
		return w.Flush()
	case isCmd(opts, "releases", "create"):
		app, _ := opts.String("<app>")
		release, _ := opts.String("<release>")
		imageSHA256, _ := opts.String("<sha256>")
		path, _ := opts.String("<config>")
		enable, _ := opts.Bool("--enable")
		config, err := belvedere.LoadConfig(ctx, path)
		if err != nil {
			return err
		}

		if err := belvedere.CreateRelease(ctx, project, app, release, config, imageSHA256, dryRun); err != nil {
			return err
		}

		if enable {
			return belvedere.EnableRelease(ctx, project, app, release, dryRun)
		}

		return nil
	case isCmd(opts, "releases", "enable"):
		app, _ := opts.String("<app>")
		release, _ := opts.String("<release>")
		return belvedere.EnableRelease(ctx, project, app, release, dryRun)
	case isCmd(opts, "releases", "disable"):
		app, _ := opts.String("<app>")
		release, _ := opts.String("<release>")
		return belvedere.DisableRelease(ctx, project, app, release, dryRun)
	case isCmd(opts, "releases", "delete"):
		app, _ := opts.String("<app>")
		release, _ := opts.String("<release>")
		async, _ := opts.Bool("--async")

		if err := belvedere.DisableRelease(ctx, project, app, release, dryRun); err != nil {
			return err
		}
		return belvedere.DeleteRelease(ctx, project, app, release, dryRun, async)
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

func pollingContext(ctx context.Context, opts docopt.Opts) context.Context {
	interval, err := time.ParseDuration(opts["--interval"].(string))
	if err != nil {
		panic(err)
	}
	ctx = belvedere.WithInterval(ctx, interval)

	timeout, err := time.ParseDuration(opts["--timeout"].(string))
	if err != nil {
		panic(err)
	}
	ctx, _ = context.WithTimeout(ctx, timeout)

	return ctx
}
