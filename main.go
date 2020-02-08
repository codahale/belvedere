package main

//go:generate bash ./version.sh

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/user"
	"time"

	"github.com/alecthomas/kong"
	"github.com/codahale/belvedere/pkg/belvedere"
	"go.opencensus.io/examples/exporter"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/trace"
)

func main() {
	var opts Options
	cli := kong.Parse(&opts,
		kong.Name("belvedere"), kong.UsageOnError(),
		kong.Vars{
			"version": version,
		},
	)

	// Enable trace logging.
	opts.enableLogging()

	// Create a root span.
	ctx, cancel, span := opts.rootSpan()
	defer cancel()
	defer span.End()

	// If a project was not explicitly specified, detect one.
	cli.FatalIfErrorf(opts.detectProject(ctx))

	// Run the given command.
	cli.BindTo(ctx, (*context.Context)(nil))
	cli.FatalIfErrorf(cli.Run(&opts))

	// Run any post-command hook.
	if opts.exit != nil {
		cli.FatalIfErrorf(opts.exit())
	}
}

var (
	version = "unknown" // version is injected via the go:generate statement
)

type SetupCmd struct {
	DNSZone string `arg:"" required:"" help:"The DNS zone to be managed by this project."`
}

func (cmd *SetupCmd) Run(ctx context.Context, o *Options) error {
	return belvedere.Setup(ctx, o.Project, cmd.DNSZone, o.DryRun)
}

type TeardownCmd struct {
	Async bool `help:"Return without waiting for successful completion."`
}

func (cmd *TeardownCmd) Run(ctx context.Context, o *Options) error {
	return belvedere.Teardown(ctx, o.Project, o.DryRun, cmd.Async)
}

type DNSServersCmd struct {
}

func (cmd *DNSServersCmd) Run(ctx context.Context, o *Options) error {
	servers, err := belvedere.DNSServers(ctx, o.Project)
	if err != nil {
		return err
	}
	return o.printTable(servers)
}

type MachineTypesCmd struct {
	Region string `help:"Limit types to those available in the given region."`
}

func (cmd *MachineTypesCmd) Run(ctx context.Context, o *Options) error {
	machineTypes, err := belvedere.MachineTypes(ctx, o.Project, cmd.Region)
	if err != nil {
		return err
	}
	return o.printTable(machineTypes)
}

type InstancesCmd struct {
	App     string `arg:"" optional:"" help:"Limit instances to those running the given app."`
	Release string `arg:"" optional:"" help:"Limit instances to those running the given release."`
}

func (cmd *InstancesCmd) Run(ctx context.Context, o *Options) error {
	instances, err := belvedere.Instances(ctx, o.Project, cmd.App, cmd.Release)
	if err != nil {
		return err
	}
	return o.printTable(instances)
}

type SSHCmd struct {
	Instance string   `arg:"" required:"" help:"The instance name."`
	Args     []string `arg:"" help:"Additional SSH arguments."`
}

func (cmd *SSHCmd) Run(ctx context.Context, o *Options) error {
	ssh, err := belvedere.SSH(ctx, o.Project, cmd.Instance, cmd.Args)
	if err != nil {
		return err
	}
	o.exit = ssh
	return nil
}

type LogsCmd struct {
	App       string        `arg:"" help:"Limit logs to the given app."`
	Release   string        `arg:"" optional:"" help:"Limit logs to the given release."`
	Instance  string        `arg:"" optional:"" help:"Limit logs to the given instance."`
	Filters   []string      `name:"filter" optional:"" help:"Limit logs with the given Stackdriver Logging filters."`
	Freshness time.Duration `default:"5m" help:"Limit logs to the last period of time."`
}

func (cmd *LogsCmd) Run(ctx context.Context, o *Options) error {
	t := time.Now().Add(-cmd.Freshness)
	logs, err := belvedere.Logs(ctx, o.Project, cmd.App, cmd.Release, cmd.Instance, t, cmd.Filters)
	if err != nil {
		return err
	}
	return o.printTable(logs)
}

type AppsListCmd struct {
}

func (AppsListCmd) Run(ctx context.Context, o *Options) error {
	apps, err := belvedere.Apps(ctx, o.Project)
	if err != nil {
		return err
	}
	return o.printTable(apps)
}

type AppsCreateCmd struct {
	App    string `arg:"" help:"The app's name."`
	Region string `arg:"" help:"The app's region."`
	Config string `arg:"" optional:"" help:"The path to the app's configuration file. Reads from STDIN if not specified."`
}

func (cmd *AppsCreateCmd) Run(ctx context.Context, o *Options) error {
	b, err := readFile(ctx, cmd.Config)
	if err != nil {
		return err
	}

	config, err := belvedere.LoadConfig(ctx, b)
	if err != nil {
		return err
	}
	return belvedere.CreateApp(ctx, o.Project, cmd.Region, cmd.App, config, o.DryRun)
}

type AppsUpdateCmd struct {
	App    string `arg:"" help:"The app's name."`
	Config string `arg:"" optional:"" help:"The path to the app's configuration file. Reads from STDIN if not specified."`
}

func (cmd *AppsUpdateCmd) Run(ctx context.Context, o *Options) error {
	b, err := readFile(ctx, cmd.Config)
	if err != nil {
		return err
	}

	config, err := belvedere.LoadConfig(ctx, b)
	if err != nil {
		return err
	}
	return belvedere.UpdateApp(ctx, o.Project, cmd.App, config, o.DryRun)
}

type AppsDeleteCmd struct {
	App   string `arg:"" help:"The app's name."`
	Async bool   `help:"Return without waiting for successful completion."`
}

func (cmd *AppsDeleteCmd) Run(ctx context.Context, o *Options) error {
	return belvedere.DeleteApp(ctx, o.Project, cmd.App, o.DryRun, cmd.Async)
}

type AppsCmd struct {
	List   AppsListCmd   `cmd:"" help:"List all apps."`
	Create AppsCreateCmd `cmd:"" help:"Create an application."`
	Update AppsUpdateCmd `cmd:"" help:"Update an application."`
	Delete AppsDeleteCmd `cmd:"" help:"Delete an application."`
}

type ReleasesListCmd struct {
	App string `optional:"" help:"Limit releases to the given app."`
}

func (cmd *ReleasesListCmd) Run(ctx context.Context, o *Options) error {
	releases, err := belvedere.Releases(ctx, o.Project, cmd.App)
	if err != nil {
		return err
	}
	return o.printTable(releases)
}

type ReleasesCreateCmd struct {
	App     string `arg:"" help:"The app's name."`
	Release string `arg:"" help:"The release's name."`
	SHA256  string `arg:"" help:"The app container's SHA256 hash."`
	Config  string `arg:"" optional:"" help:"The path to the app's configuration file. Reads from STDIN if not specified."`
	Enable  bool   `help:"Put release into service once created."`
}

func (cmd *ReleasesCreateCmd) Run(ctx context.Context, o *Options) error {
	b, err := readFile(ctx, cmd.Config)
	if err != nil {
		return err
	}

	config, err := belvedere.LoadConfig(ctx, b)
	if err != nil {
		return err
	}

	err = belvedere.CreateRelease(ctx, o.Project, cmd.App, cmd.Release, config, cmd.SHA256, o.DryRun)
	if err != nil {
		return err
	}

	if cmd.Enable {
		err = belvedere.EnableRelease(ctx, o.Project, cmd.App, cmd.Release, o.DryRun)
		if err != nil {
			return err
		}
	}
	return nil
}

type ReleasesEnableCmd struct {
	App     string `arg:"" help:"The app's name."`
	Release string `arg:"" help:"The release's name."`
}

func (cmd *ReleasesEnableCmd) Run(ctx context.Context, o *Options) error {
	return belvedere.EnableRelease(ctx, o.Project, cmd.App, cmd.Release, o.DryRun)
}

type ReleasesDisableCmd struct {
	App     string `arg:"" help:"The app's name."`
	Release string `arg:"" help:"The release's name."`
}

func (cmd *ReleasesDisableCmd) Run(ctx context.Context, o *Options) error {
	return belvedere.DisableRelease(ctx, o.Project, cmd.App, cmd.Release, o.DryRun)
}

type ReleasesDeleteCmd struct {
	App     string `arg:"" help:"The app's name."`
	Release string `arg:"" help:"The release's name."`
	Async   bool   `help:"Return without waiting for successful completion."`
}

func (cmd *ReleasesDeleteCmd) Run(ctx context.Context, o *Options) error {
	return belvedere.DeleteRelease(ctx, o.Project, cmd.App, cmd.Release, o.DryRun, cmd.Async)
}

type ReleasesCmd struct {
	List    ReleasesListCmd    `cmd:"" help:"List all releases."`
	Create  ReleasesCreateCmd  `cmd:"" help:"Create a release."`
	Enable  ReleasesEnableCmd  `cmd:"" help:"Put a release into service."`
	Disable ReleasesDisableCmd `cmd:"" help:"Remove a release from service."`
	Delete  ReleasesDeleteCmd  `cmd:"" help:"Delete a release."`
}

type SecretsListCmd struct {
}

func (*SecretsListCmd) Run(ctx context.Context, o *Options) error {
	releases, err := belvedere.Secrets(ctx, o.Project)
	if err != nil {
		return err
	}
	return o.printTable(releases)
}

type SecretsCreateCmd struct {
	Secret   string `arg:"" help:"The secret's name."`
	DataFile string `arg:"" optional:"" help:"File path from which to read secret data. Reads from STDIN if not specified."`
}

func (cmd *SecretsCreateCmd) Run(ctx context.Context, o *Options) error {
	b, err := readFile(ctx, cmd.DataFile)
	if err != nil {
		return err
	}
	return belvedere.CreateSecret(ctx, o.Project, cmd.Secret, b)
}

type SecretsGrantCmd struct {
	Secret string `arg:"" help:"The secret's name."`
	App    string `arg:"" help:"The app's name."`
}

func (cmd *SecretsGrantCmd) Run(ctx context.Context, o *Options) error {
	return belvedere.GrantSecret(ctx, o.Project, cmd.Secret, cmd.App, o.DryRun)
}

type SecretsRevokeCmd struct {
	Secret string `arg:"" help:"The secret's name."`
	App    string `arg:"" help:"The app's name."`
}

func (cmd *SecretsRevokeCmd) Run(ctx context.Context, o *Options) error {
	return belvedere.RevokeSecret(ctx, o.Project, cmd.App, cmd.Secret, o.DryRun)
}

type SecretsUpdateCmd struct {
	Secret   string `arg:"" help:"The secret's name."`
	DataFile string `arg:"" optional:"" help:"File path from which to read secret data. Reads from STDIN if not specified."`
}

func (cmd *SecretsUpdateCmd) Run(ctx context.Context, o *Options) error {
	b, err := readFile(ctx, cmd.DataFile)
	if err != nil {
		return err
	}
	return belvedere.UpdateSecret(ctx, o.Project, cmd.Secret, b)
}

type SecretsDeleteCmd struct {
	Secret string `arg:"" help:"The secret's name."`
}

func (cmd *SecretsDeleteCmd) Run(ctx context.Context, o *Options) error {
	return belvedere.DeleteSecret(ctx, o.Project, cmd.Secret)
}

type SecretsCmd struct {
	List   SecretsListCmd   `cmd:"" help:"List all secrets."`
	Create SecretsCreateCmd `cmd:"" help:"Create a secret."`
	Grant  SecretsGrantCmd  `cmd:"" help:"Grant access to a secret for an application."`
	Revoke SecretsRevokeCmd `cmd:"" help:"Revoke access to a secret for an application."`
	Update SecretsUpdateCmd `cmd:"" help:"Update a secret."`
	Delete SecretsDeleteCmd `cmd:"" help:"Delete a secret."`
}

type Options struct {
	Debug    bool             `help:"Enable debug logging." short:"d"`
	Quiet    bool             `help:"Disable all logging." short:"q"`
	DryRun   bool             `help:"Print modifications instead of performing them."`
	CSV      bool             `help:"Print tables as CSV."`
	Interval time.Duration    `help:"Specify a polling interval for long-running operations." default:"10s"`
	Timeout  time.Duration    `help:"Specify a timeout for long-running operations." default:"10m"`
	Version  kong.VersionFlag `help:"Print version information and quit."`
	Project  string           `help:"Specify a GCP project ID. Defaults to using gcloud."`

	Setup        SetupCmd        `cmd:"" help:"Initialize a GCP project for use with Belvedere."`
	Teardown     TeardownCmd     `cmd:"" help:"Remove all Belvedere resources from this project."`
	DNSServers   DNSServersCmd   `cmd:"" help:"List the DNS servers for this project."`
	MachineTypes MachineTypesCmd `cmd:"" help:"List available GCE machine types."`
	Instances    InstancesCmd    `cmd:"" help:"List running instances."`
	SSH          SSHCmd          `cmd:"" help:"SSH to an instance over IAP."`
	Logs         LogsCmd         `cmd:"" help:"Display application logs."`
	Apps         AppsCmd         `cmd:"" help:"Commands for managing apps."`
	Releases     ReleasesCmd     `cmd:"" help:"Commands for managing releases."`
	Secrets      SecretsCmd      `cmd:"" help:"Commands for managing secrets."`

	exit func() error
}

func (o *Options) rootSpan() (context.Context, context.CancelFunc, *trace.Span) {
	// Initialize a context with a timeout and an interval.
	ctx := belvedere.WithInterval(context.Background(), o.Interval)
	ctx, cancel := context.WithTimeout(ctx, o.Timeout)

	// Create a root span.
	ctx, span := trace.StartSpan(ctx, "belvedere.main")
	if hostname, err := os.Hostname(); err == nil {
		span.AddAttributes(trace.StringAttribute("hostname", hostname))
	}
	if u, err := user.Current(); err == nil {
		span.AddAttributes(trace.StringAttribute("username", u.Username))
		span.AddAttributes(trace.StringAttribute("uid", u.Uid))
	}
	return ctx, cancel, span
}

func (o *Options) enableLogging() {
	// Export all traces.
	trace.ApplyConfig(trace.Config{
		DefaultSampler: trace.AlwaysSample(),
	})

	if o.Debug {
		// Use the print exporter for debugging, as it prints everything.
		pe := &exporter.PrintExporter{}
		trace.RegisterExporter(pe)
		view.RegisterExporter(pe)
	} else if !o.Quiet {
		// Unless we're quiet, use the trace logger for more practical logging.
		trace.RegisterExporter(&traceLogger{})
	}
}

func (o *Options) detectProject(ctx context.Context) error {
	if o.Project == "" {
		p, err := belvedere.DefaultProject(ctx)
		if err != nil {
			return err
		}
		o.Project = p
	}
	return nil
}

func readFile(ctx context.Context, name string) ([]byte, error) {
	_, span := trace.StartSpan(ctx, "belvedere.readFile")
	span.AddAttributes(
		trace.StringAttribute("name", name),
	)
	defer span.End()

	// Either open the file or use STDIN.
	var r io.ReadCloser
	if name == "" {
		r = os.Stdin
	} else {
		f, err := os.Open(name)
		if err != nil {
			return nil, fmt.Errorf("error opening %s: %w", name, err)
		}

		r = f
	}
	defer func() { _ = r.Close() }()

	// Read the entire input.
	b, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("error reading from %s: %w", name, err)
	}
	return b, nil
}
