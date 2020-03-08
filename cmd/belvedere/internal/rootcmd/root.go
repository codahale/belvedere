package rootcmd

import (
	"context"
	"flag"
	"os"
	"os/user"
	"time"

	"github.com/codahale/belvedere/cmd/belvedere/internal/cmd"
	"github.com/codahale/belvedere/pkg/belvedere"
	"github.com/peterbourgon/ff/v2/ffcli"
	"go.opencensus.io/examples/exporter"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/trace"
)

type Config struct {
	// Flags
	Quiet       bool
	Debug       bool
	CSV         bool
	Timeout     time.Duration
	ProjectName string

	// Variables
	Callback func() error
	Project  belvedere.Project
	Tables   cmd.TableWriter
	Files    cmd.FileReader
}

func New() (*ffcli.Command, *Config) {
	var config Config

	fs := flag.NewFlagSet("belvedere", flag.ExitOnError)
	config.RegisterFlags(fs)

	return &ffcli.Command{
		Name:       "belvedere",
		ShortUsage: "belvedere <subcommand> [<arg>...] [flags]",
		LongHelp: cmd.Wrap(`A small lookout tower (usually square) on the roof of a house.

Belvedere provides an easy and reliable way of deploying and managing HTTP2 applications on Google
Cloud Platform. It handles load balancing, DNS, TLS, blue/green deploys, auto-scaling, CDN
configuration, access control, secret management, configuration, and more.`),
		FlagSet: fs,
		Exec:    config.Exec,
	}, &config
}

func (c *Config) RegisterFlags(fs *flag.FlagSet) {
	fs.BoolVar(&c.Quiet, "q", false, "disable logging entirely")
	fs.BoolVar(&c.Debug, "debug", false, "log verbose output")
	fs.BoolVar(&c.CSV, "csv", false, "format output as CSV")
	fs.DurationVar(&c.Timeout, "timeout", 10*time.Minute, "maximum time allowed for total execution")
	fs.StringVar(&c.ProjectName, "project", "", "specify a Google Cloud project ID")
}

func (c *Config) Exec(context.Context, []string) error {
	return flag.ErrHelp
}

func (c *Config) EnableLogging() {
	// Export all traces.
	trace.ApplyConfig(trace.Config{
		DefaultSampler: trace.AlwaysSample(),
	})

	if c.Debug {
		// Use the print exporter for debugging, as it prints everything.
		pe := &exporter.PrintExporter{}
		trace.RegisterExporter(pe)
		view.RegisterExporter(pe)
	} else if !c.Quiet {
		// Unless we're quiet, use the trace logger for more practical logging.
		trace.RegisterExporter(cmd.NewTraceLogger(os.Stderr))
	}
}

func (c *Config) RootSpan() (context.Context, context.CancelFunc, *trace.Span) {
	// Initialize a context with a timeout and an interval.
	ctx, cancel := context.WithTimeout(context.Background(), c.Timeout)

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
