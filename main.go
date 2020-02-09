package main

//go:generate bash ./version.sh

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/user"
	"reflect"
	"syscall"
	"time"

	"github.com/alecthomas/kong"
	"github.com/codahale/belvedere/pkg/belvedere"
	"github.com/olekukonko/tablewriter"
	"go.opencensus.io/examples/exporter"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/trace"
	"golang.org/x/crypto/ssh/terminal"
)

func main() {
	var opts Options
	cli := kong.Parse(&opts,
		kong.Name("belvedere"),
		kong.Vars{"version": version},
		kong.Description("A small lookout tower (usually square) on the roof of a house."),
		kong.UsageOnError(),
	)
	cli.FatalIfErrorf(run(cli, opts))
}

func run(cli *kong.Context, opts Options) error {
	// Enable trace logging.
	opts.enableLogging()

	// Create a root span.
	ctx, cancel, span := opts.rootSpan()
	defer cancel()
	defer span.End()

	// If a project was not explicitly specified, detect one.
	if err := opts.detectProject(ctx); err != nil {
		return err
	}

	// Run the given command, passing in the context and options.
	cli.BindTo(ctx, (*context.Context)(nil))
	if err := cli.Run(&opts); err != nil {
		return err
	}

	// Run any post-command hook.
	if opts.exit != nil {
		return opts.exit()
	}

	return nil
}

var (
	version = "unknown" // version is injected via the go:generate statement
)

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

func (o *Options) printTable(i interface{}) error {
	t := reflect.TypeOf(i)
	if t.Kind() != reflect.Slice {
		return nil
	}

	t = t.Elem()
	if t.Kind() != reflect.Struct {
		return nil
	}

	var headers []string
	for i := 0; i < t.NumField(); i++ {
		s := t.Field(i).Tag.Get("table")
		if s == "" {
			s = t.Field(i).Name
		}
		headers = append(headers, s)
	}

	var rows [][]string
	iv := reflect.ValueOf(i)
	for i := 0; i < iv.Len(); i++ {
		var row []string
		ev := iv.Index(i)
		for j := range headers {
			f := ev.Field(j)

			if t, ok := f.Interface().(time.Time); ok {
				row = append(row, t.Format(time.Stamp))
			} else {
				row = append(row, fmt.Sprint(f.Interface()))
			}
		}
		rows = append(rows, row)
	}

	if terminal.IsTerminal(syscall.Stdout) && !o.CSV {
		tw := tablewriter.NewWriter(os.Stdout)
		tw.SetAutoFormatHeaders(false)
		tw.SetAutoWrapText(false)
		tw.SetHeader(headers)
		tw.AppendBulk(rows)
		tw.Render()
	} else {
		cw := csv.NewWriter(os.Stdout)
		_ = cw.Write(headers)
		_ = cw.WriteAll(rows)
		cw.Flush()
	}
	return nil
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
