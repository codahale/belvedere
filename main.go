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

func main() {
	app.HelpFlag.Short('h')
	app.Version(version)

	// Parse command line and execute the action associated with the command.
	_, err := app.Parse(os.Args[1:])
	if err != nil {
		die(err)
	}
}

// Define the CLI app, commands, and flags.
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

	// belvedere apps grant-secret <app> <secret>
	appsGrantSecretCmd    = appsCmd.Command("grant-secret", "Grant access to a secret for an application.")
	appsGrantSecretApp    = appsGrantSecretCmd.Arg("app", "The app's name.").Required().String()
	appsGrantSecretSecret = appsGrantSecretCmd.Arg("secret", "The secrets's name.").Required().String()

	// belvedere apps revoke-secret <app> <secret>
	appsRevokeSecretCmd    = appsCmd.Command("revoke-secret", "Revoke access to a secret for an application.")
	appsRevokeSecretApp    = appsRevokeSecretCmd.Arg("app", "The app's name.").Required().String()
	appsRevokeSecretSecret = appsRevokeSecretCmd.Arg("secret", "The secrets's name.").Required().String()

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

// Map all commands to actions. This has to happen outside the `var` clauses because command actions
// refer to command args and flags which themselves refer to commands. Go doesn't do cycles.
func init() {
	setupCmd.Action(contextAction(runSetup))
	teardownCmd.Action(contextAction(runTeardown))
	dnsServersCmd.Action(contextAction(runDnsServers))
	machineTypesCmd.Action(contextAction(runMachineTypes))
	instancesCmd.Action(contextAction(runInstances))
	sshCmd.Action(contextAction(runSSH))
	logsCmd.Action(contextAction(runLogs))
	appsListCmd.Action(contextAction(runAppsList))
	appsCreateCmd.Action(contextAction(runAppsCreate))
	appsUpdateCmd.Action(contextAction(runAppsUpdate))
	appsDeleteCmd.Action(contextAction(runAppsDelete))
	appsGrantSecretCmd.Action(contextAction(runAppsGrantSecret))
	appsRevokeSecretCmd.Action(contextAction(runAppsRevokeSecret))
	relListCmd.Action(contextAction(runRelList))
	relCreateCmd.Action(contextAction(runRelCreate))
	relEnableCmd.Action(contextAction(runRelEnable))
	relDisableCmd.Action(contextAction(runRelDisable))
	relDeleteCmd.Action(contextAction(runRelDelete))
}

func runSetup(ctx context.Context, _ *func() error) error {
	return belvedere.Setup(ctx, *project, *setupDnsZone, *dryRun)
}

func runTeardown(ctx context.Context, _ *func() error) error {
	return belvedere.Teardown(ctx, *project, *dryRun, *teardownAsync)
}

func runDnsServers(ctx context.Context, _ *func() error) error {
	servers, err := belvedere.DNSServers(ctx, *project)
	if err != nil {
		return err
	}
	return printTable(servers)
}

func runMachineTypes(ctx context.Context, _ *func() error) error {
	machineTypes, err := belvedere.MachineTypes(ctx, *project, *machineTypesRegion)
	if err != nil {
		return err
	}
	return printTable(machineTypes)
}

func runInstances(ctx context.Context, _ *func() error) error {
	instances, err := belvedere.Instances(ctx, *project, *instancesApp, *instancesRelease)
	if err != nil {
		return err
	}
	return printTable(instances)
}

func runSSH(ctx context.Context, exit *func() error) error {
	ssh, err := belvedere.SSH(ctx, *project, *sshInstance, *sshArgs)
	if err != nil {
		return err
	}
	*exit = ssh
	return nil
}

func runLogs(ctx context.Context, _ *func() error) error {
	t := time.Now().Add(-*logsFreshness)
	logs, err := belvedere.Logs(ctx, *project, *logsApp, *logsRelease, *logsInstance, t, *logsFilters)
	if err != nil {
		return err
	}
	return printTable(logs)
}

func runAppsList(ctx context.Context, _ *func() error) error {
	apps, err := belvedere.Apps(ctx, *project)
	if err != nil {
		return err
	}
	return printTable(apps)
}

func runAppsCreate(ctx context.Context, _ *func() error) error {
	config, err := belvedere.LoadConfig(ctx, *appsCreateConfig)
	if err != nil {
		return err
	}
	return belvedere.CreateApp(ctx, *project, *appsCreateRegion, *appsCreateApp, config, *dryRun)
}

func runAppsUpdate(ctx context.Context, _ *func() error) error {
	config, err := belvedere.LoadConfig(ctx, *appsUpdateConfig)
	if err != nil {
		return err
	}
	return belvedere.UpdateApp(ctx, *project, *appsUpdateApp, config, *dryRun)
}

func runAppsDelete(ctx context.Context, _ *func() error) error {
	return belvedere.DeleteApp(ctx, *project, *appsDeleteApp, *dryRun, *appsDeleteAsync)
}

func runAppsGrantSecret(ctx context.Context, _ *func() error) error {
	return belvedere.GrantAppSecret(ctx, *project, *appsGrantSecretApp, *appsGrantSecretSecret, *dryRun)
}

func runAppsRevokeSecret(ctx context.Context, _ *func() error) error {
	return belvedere.RevokeAppSecret(ctx, *project, *appsRevokeSecretApp, *appsRevokeSecretSecret, *dryRun)
}

func runRelList(ctx context.Context, _ *func() error) error {
	releases, err := belvedere.Releases(ctx, *project, *relListApp)
	if err != nil {
		return err
	}
	return printTable(releases)
}

func runRelCreate(ctx context.Context, _ *func() error) error {
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
}

func runRelEnable(ctx context.Context, _ *func() error) error {
	return belvedere.EnableRelease(ctx, *project, *relEnableApp, *relEnableRelease, *dryRun)
}

func runRelDisable(ctx context.Context, _ *func() error) error {
	return belvedere.DisableRelease(ctx, *project, *relDisableApp, *relDisableRelease, *dryRun)
}

func runRelDelete(ctx context.Context, _ *func() error) error {
	if err := belvedere.DisableRelease(ctx, *project, *relDeleteApp, *relDeleteRelease, false); err != nil {
		return err
	}
	return belvedere.DeleteRelease(ctx, *project, *relDeleteApp, *relDeleteRelease, *dryRun, *relDeleteAsync)
}

func contextAction(f func(ctx context.Context, exit *func() error) error) kingpin.Action {
	return func(parseContext *kingpin.ParseContext) error {
		// Enable trace logging.
		enableLogging()

		// Create a root span.
		ctx, cancel, span := rootSpan()
		defer cancel()
		defer span.End()

		// If a project was not explicitly specified, detect one.
		if err := detectProject(ctx); err != nil {
			die(err)
		}

		var exit func() error
		if err := f(ctx, &exit); err != nil {
			return err
		}

		if exit != nil {
			// Manually end the root span and execute the exit handler.
			span.End()
			return exit()
		}

		return nil
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

func die(err error) {
	_, _ = fmt.Fprintln(os.Stderr, err)
	os.Exit(-1)
}
