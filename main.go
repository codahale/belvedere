package main

//go:generate bash ./version.sh

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/user"
	"syscall"
	"time"

	"github.com/alecthomas/kong"
	"github.com/codahale/belvedere/pkg/belvedere"
	"go.opencensus.io/examples/exporter"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/trace"
	"golang.org/x/crypto/ssh/terminal"
	"google.golang.org/genproto/googleapis/rpc/code"
)

func main() {
	// ParseConfig the command line.
	var opts CLI
	cli := kong.Parse(&opts,
		kong.Name("belvedere"),
		kong.Vars{"version": version},
		kong.Description("A small lookout tower (usually square) on the roof of a house."),
		kong.UsageOnError(),
	)

	// Run the given command.
	cli.FatalIfErrorf(run(cli, &opts))

	// Run any post-command hook.
	if opts.exit != nil {
		cli.FatalIfErrorf(opts.exit())
	}
}

func run(cli *kong.Context, opts *CLI) error {
	// Enable trace logging.
	opts.enableLogging()

	// Create a root span.
	ctx, cancel, span := opts.rootSpan()
	defer cancel()
	defer span.End()

	// Create a Belvedere project.
	project, err := belvedere.NewProject(ctx, opts.Project)
	if err != nil {
		return err
	}
	span.AddAttributes(trace.StringAttribute("project", project.Name()))

	// Run the given command, passing in the context and options.
	cli.BindTo(ctx, (*context.Context)(nil))
	cli.BindTo(project, (*belvedere.Project)(nil))
	cli.BindTo(&writer{csv: opts.CSV}, (*TableWriter)(nil))
	if err := cli.Run(&opts); err != nil {
		span.SetStatus(trace.Status{
			Code:    int32(code.Code_INTERNAL),
			Message: err.Error(),
		})
		return err
	}

	return nil
}

var (
	version = "unknown" // version is injected via the go:generate statement
)

type ModifyOptions struct {
	DryRun bool `help:"Print modifications instead of performing them."`
}

type LongRunningOptions struct {
	Interval time.Duration `help:"Specify a polling interval for long-running operations." default:"10s"`
}

type CLI struct {
	Debug   bool             `help:"Enable debug logging." short:"d"`
	Quiet   bool             `help:"Disable all logging." short:"q"`
	CSV     bool             `help:"Print tables as CSV."`
	Timeout time.Duration    `help:"Specify a timeout for long-running operations." default:"10m"`
	Version kong.VersionFlag `help:"Print version information and quit."`
	Project string           `help:"Specify a GCP project ID. Defaults to using the GCP SDK's active config'."`

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

func (cli *CLI) rootSpan() (context.Context, context.CancelFunc, *trace.Span) {
	// Initialize a context with a timeout and an interval.
	ctx, cancel := context.WithTimeout(context.Background(), cli.Timeout)

	// Create a root span.
	ctx, span := trace.StartSpan(ctx, "belvedere.main")
	if hostname, err := os.Hostname(); err == nil {
		span.AddAttributes(trace.StringAttribute("hostname", hostname))
	}
	if u, err := user.Current(); err == nil {
		span.AddAttributes(
			trace.StringAttribute("username", u.Username),
			trace.StringAttribute("uid", u.Uid),
		)
	}
	return ctx, cancel, span
}

func (cli *CLI) enableLogging() {
	// Export all traces.
	trace.ApplyConfig(trace.Config{
		DefaultSampler: trace.AlwaysSample(),
	})

	if cli.Debug {
		// Use the print exporter for debugging, as it prints everything.
		pe := &exporter.PrintExporter{}
		trace.RegisterExporter(pe)
		view.RegisterExporter(pe)
	} else if !cli.Quiet {
		// Unless we're quiet, use the trace logger for more practical logging.
		trace.RegisterExporter(&traceLogger{})
	}
}

func readFile(ctx context.Context, name string) ([]byte, error) {
	_, span := trace.StartSpan(ctx, "belvedere.readFile")
	defer span.End()

	// Either open the file or use STDIN.
	var r io.ReadCloser
	if name == "" {
		span.AddAttributes(
			trace.StringAttribute("name", "stdin"),
		)
		if terminal.IsTerminal(syscall.Stdin) {
			return nil, fmt.Errorf("can't read from stdin")
		}
		r = os.Stdin
	} else {
		span.AddAttributes(
			trace.StringAttribute("name", name),
		)
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
