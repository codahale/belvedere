package main

//go:generate bash ./version.sh

import (
	"context"
	"fmt"
	"os"
	"os/user"
	"time"

	"github.com/codahale/belvedere/pkg/belvedere"
	"go.opencensus.io/examples/exporter"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/trace"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	version = "unknown" // version is injected via the go:generate statement

	// Application and common flags.
	app      = kingpin.New("belvedere", "A fine place from which to survey your estate.")
	debug    = app.Flag("debug", "Enable debug logging.").Bool()
	quiet    = app.Flag("quiet", "Disable all logging.").Short('q').Bool()
	dryRun   = app.Flag("dry-run", "Print modifications instead of performing them.").Bool()
	printCSV = app.Flag("csv", "Print tables as CSV").Bool()
	interval = app.Flag("interval", "Specify a polling interval for long-running operations.").
			Default("10s").Duration()
	timeout = app.Flag("timeout", "Specify a timeout for long-running operations.").
		Default("10m").Duration()
	project = app.Flag("project", "Specify a GCP project ID. Defaults to using gcloud.").String()

	// belvedere setup <dns-zone>
	setupCmd     = app.Command("setup", "Initialize a GCP project for use with Belvedere.")
	setupDnsZone = setupCmd.Arg("dns-zone", "The DNS zone to be managed by this project.").
			Required().String()

	// belvedere teardown [--async]
	teardownCmd   = app.Command("teardown", "Remove all Belvedere resources from this project.")
	teardownAsync = teardownCmd.Flag("async", "Return without waiting for successful completion.").Bool()

	// belvedere dns-servers
	dnsServersCmd = app.Command("dns-servers", "List the DNS servers for this project.")

	// belvedere machine-types
	machineTypesCmd    = app.Command("machine-types", "List available GCE machine types.")
	machineTypesRegion = machineTypesCmd.Arg("region", "Limit types to those available in the given region.").String()

	// belvedere instances [<app>] [<release>]
	instancesCmd     = app.Command("instances", "List running instances.")
	instancesApp     = instancesCmd.Arg("app", "Limit instances to those running the given app.").String()
	instancesRelease = instancesCmd.Arg("release", "Limit instances to those running the given release.").String()

	// belvedere ssh <instance> [-- <args>...]
	sshCmd      = app.Command("ssh", "SSH to an instance over IAP.")
	sshInstance = sshCmd.Arg("instance", "The instance name.").Required().String()
	sshArgs     = sshCmd.Arg("args", "Additional SSH arguments.").Strings()

	// belvedere logs <app> [<release>] [<instance>] [--filter=FILTER...] [--freshness=5m]
	logsCmd       = app.Command("logs", "Display application logs.")
	logsApp       = logsCmd.Arg("app", "Limit logs to the given app.").Required().String()
	logsRelease   = logsCmd.Arg("release", "Limit logs to the given release.").String()
	logsInstance  = logsCmd.Arg("instance", "Limit logs to the given instance.").String()
	logsFilters   = logsCmd.Flag("filter", "Limit logs with the given Stackdriver Logging filter.").Strings()
	logsFreshness = logsCmd.Flag("freshness", "Limit logs to the last period of time.").
			Default("5m").Duration()

	// belvedere apps
	appsCmd = app.Command("apps", "Commands for managing apps.")

	// belvedere apps list
	appsListCmd = appsCmd.Command("list", "List all apps.")

	// belvedere apps create <app> <region> <config>
	appsCreateCmd    = appsCmd.Command("create", "Create an application.")
	appsCreateApp    = appsCreateCmd.Arg("app", "The app's name.").Required().String()
	appsCreateRegion = appsCreateCmd.Arg("region", "The app's region.").Required().String()
	appsCreateConfig = appsCreateCmd.Arg("config", "The app's configuration.").Required().String()

	// belvedere apps update <app> <config>
	appsUpdateCmd    = appsCmd.Command("update", "Update an application.")
	appsUpdateApp    = appsUpdateCmd.Arg("app", "The app's name.").Required().String()
	appsUpdateConfig = appsUpdateCmd.Arg("config", "The app's configuration.").Required().String()

	// belvedere apps delete <app> [--async]
	appsDeleteCmd   = appsCmd.Command("delete", "Delete an application.")
	appsDeleteApp   = appsDeleteCmd.Arg("app", "The app's name.").Required().String()
	appsDeleteAsync = appsDeleteCmd.Flag("async", "Return without waiting for successful completion.").Bool()

	// belvedere releases
	relCmd = app.Command("releases", "Commands for managing releases.")

	// belvedere releases list [<app>]
	relListCmd = relCmd.Command("list", "List all releases.")
	relListApp = relListCmd.Arg("app", "Limit releases to the given app.").String()

	// belvedere releases create <app> <release> <config> <sha256> [--enable]
	relCreateCmd     = relCmd.Command("create", "Create a release.")
	relCreateApp     = relCreateCmd.Arg("app", "The app's name.").Required().String()
	relCreateRelease = relCreateCmd.Arg("release", "The release's name.").Required().String()
	relCreateConfig  = relCreateCmd.Arg("config", "The app's config.").Required().String()
	relCreateHash    = relCreateCmd.Arg("sha256", "The app container's SHA256 hash.").Required().String()
	relCreateEnable  = relCreateCmd.Flag("enable", "Put release into service once created.").Bool()

	// belvedere releases enable <app> <release>
	relEnableCmd     = relCmd.Command("enable", "Put a release into service.")
	relEnableApp     = relEnableCmd.Arg("app", "The app's name.").Required().String()
	relEnableRelease = relEnableCmd.Arg("release", "The release's name.").Required().String()

	// belvedere releases disable <app> <release>
	relDisableCmd     = relCmd.Command("disable", "Remove a release from service.")
	relDisableApp     = relDisableCmd.Arg("app", "The app's name.").Required().String()
	relDisableRelease = relDisableCmd.Arg("release", "The release's name.").Required().String()

	// belvedere releases delete <app> <release> [--async]
	relDeleteCmd     = relCmd.Command("delete", "Delete a release.")
	relDeleteApp     = relDeleteCmd.Arg("app", "The app's name.").Required().String()
	relDeleteRelease = relDeleteCmd.Arg("release", "The releases's name.").Required().String()
	relDeleteAsync   = relDeleteCmd.Flag("async", "Return without waiting for successful completion.").Bool()
)

func main() {
	app.HelpFlag.Short('h')
	app.Version(version)

	// Parse command line.
	cmd, err := app.Parse(os.Args[1:])
	if err != nil {
		die(err)
	}

	// Set up trace logging.
	enableLogging()

	// Create a root span.
	ctx, cancel, span := rootSpan()
	defer cancel()
	defer span.End()

	// If a project was not explicitly specified, detect one.
	if err := detectProject(ctx); err != nil {
		die(err)
	}

	// Run command.
	var exit func() error
	if err := runCmd(ctx, cmd, &exit); err != nil {
		die(err)
	}

	// Run exit handlers, if any.
	if exit != nil {
		if err := exit(); err != nil {
			die(err)
		}
	}
}

func detectProject(ctx context.Context) error {
	if *project == "" {
		p, err := belvedere.DefaultProject(ctx)
		if err != nil {
			return err
		}
		project = &p
	}
	return nil
}

func rootSpan() (context.Context, context.CancelFunc, *trace.Span) {
	// Initialize a context with a timeout and an interval.
	ctx := belvedere.WithInterval(context.Background(), *interval)
	ctx, cancel := context.WithTimeout(ctx, *timeout)
	// Create a root span.
	ctx, span := trace.StartSpan(ctx, "belvedere.run")
	if hostname, err := os.Hostname(); err == nil {
		span.AddAttributes(trace.StringAttribute("hostname", hostname))
	}
	if u, err := user.Current(); err == nil {
		span.AddAttributes(trace.StringAttribute("username", u.Username))
		span.AddAttributes(trace.StringAttribute("uid", u.Uid))
	}
	return ctx, cancel, span
}

func enableLogging() {
	// Export all traces.
	trace.ApplyConfig(trace.Config{
		DefaultSampler: trace.AlwaysSample(),
	})

	if *debug {
		// Use the print exporter for debugging, as it prints everything.
		pe := &exporter.PrintExporter{}
		trace.RegisterExporter(pe)
		view.RegisterExporter(pe)
	} else if !*quiet {
		// Unless we're quiet, use the trace logger for more practical logging.
		trace.RegisterExporter(&traceLogger{})
	}
}

func runCmd(ctx context.Context, cmd string, exit *func() error) error {
	switch cmd {
	case setupCmd.FullCommand():
		return belvedere.Setup(ctx, *project, *setupDnsZone, *dryRun)
	case teardownCmd.FullCommand():
		return belvedere.Teardown(ctx, *project, *dryRun, *teardownAsync)
	case dnsServersCmd.FullCommand():
		servers, err := belvedere.DNSServers(ctx, *project)
		if err != nil {
			return err
		}
		return printTable(servers)
	case machineTypesCmd.FullCommand():
		machineTypes, err := belvedere.MachineTypes(ctx, *project, *machineTypesRegion)
		if err != nil {
			return err
		}
		return printTable(machineTypes)
	case instancesCmd.FullCommand():
		instances, err := belvedere.Instances(ctx, *project, *instancesApp, *instancesRelease)
		if err != nil {
			return err
		}
		return printTable(instances)
	case sshCmd.FullCommand():
		ssh, err := belvedere.SSH(ctx, *project, *sshInstance, *sshArgs)
		if err != nil {
			return err
		}
		*exit = ssh
		return nil
	case logsCmd.FullCommand():
		t := time.Now().Add(-*logsFreshness)
		logs, err := belvedere.Logs(ctx, *project, *logsApp, *logsRelease, *logsInstance, t, *logsFilters)
		if err != nil {
			return err
		}
		return printTable(logs)
	case appsListCmd.FullCommand():
		apps, err := belvedere.Apps(ctx, *project)
		if err != nil {
			return err
		}
		return printTable(apps)
	case appsCreateCmd.FullCommand():
		config, err := belvedere.LoadConfig(ctx, *appsCreateConfig)
		if err != nil {
			return err
		}
		return belvedere.CreateApp(ctx, *project, *appsCreateRegion, *appsCreateApp, config, *dryRun)
	case appsUpdateCmd.FullCommand():
		config, err := belvedere.LoadConfig(ctx, *appsUpdateConfig)
		if err != nil {
			return err
		}
		return belvedere.UpdateApp(ctx, *project, *appsUpdateApp, config, *dryRun)
	case appsDeleteCmd.FullCommand():
		return belvedere.DeleteApp(ctx, *project, *appsDeleteApp, *dryRun, *appsDeleteAsync)
	case relListCmd.FullCommand():
		releases, err := belvedere.Releases(ctx, *project, *relListApp)
		if err != nil {
			return err
		}
		return printTable(releases)
	case relCreateCmd.FullCommand():
		config, err := belvedere.LoadConfig(ctx, *relCreateConfig)
		if err != nil {
			return err
		}
		if err := belvedere.CreateRelease(
			ctx, *project, *relCreateApp, *relCreateRelease, config, *relCreateHash, *dryRun,
		); err != nil {
			return err
		}
		if *relCreateEnable {
			if err := belvedere.EnableRelease(ctx, *project, *relCreateApp, *relCreateRelease, *dryRun); err != nil {
				return err
			}
		}
		return nil
	case relEnableCmd.FullCommand():
		return belvedere.EnableRelease(ctx, *project, *relEnableApp, *relEnableRelease, *dryRun)
	case relDisableCmd.FullCommand():
		return belvedere.DisableRelease(ctx, *project, *relDisableApp, *relDisableRelease, *dryRun)
	case relDeleteCmd.FullCommand():
		if err := belvedere.DisableRelease(ctx, *project, *relDeleteApp, *relDeleteRelease, false); err != nil {
			return err
		}
		return belvedere.DeleteRelease(ctx, *project, *relDeleteApp, *relDeleteRelease, *dryRun, *relDeleteAsync)
	default:
		return fmt.Errorf("command not implemented: %s", cmd)
	}
}

func die(err error) {
	_, _ = fmt.Fprintln(os.Stderr, err)
	os.Exit(-1)
}
