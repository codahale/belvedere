package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/codahale/belvedere/cmd/internal"
	"github.com/codahale/belvedere/pkg/belvedere"
	"github.com/spf13/cobra"
	"go.opencensus.io/examples/exporter"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/trace"
	"golang.org/x/xerrors"
)

var (
	debug        bool
	quiet        bool
	dryRun       bool
	async        bool
	project      string
	interval     time.Duration
	timeout      time.Duration
	exitHandlers []func() error
	rootCtx      context.Context
	rootCancel   context.CancelFunc
	rootCmd      = &cobra.Command{
		Use:   "belvedere",
		Short: "A fine place from which to survey your estate.",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			rootCtx = belvedere.WithInterval(context.Background(), interval)
			rootCtx, rootCancel = context.WithTimeout(rootCtx, timeout)
			enableLogging()
			return findProject()
		},
		PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
			rootCancel()

			for _, f := range exitHandlers {
				if err := f(); err != nil {
					return err
				}
			}
			return nil
		},
		SilenceUsage: true,
	}
)

func init() {
	rootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "enable debug output")
	rootCmd.PersistentFlags().BoolVarP(&quiet, "quiet", "q", false, "disable all logging")
	rootCmd.PersistentFlags().BoolVar(&dryRun, "dry-run", false, "print modifications instead of performing them")
	rootCmd.PersistentFlags().StringVar(&project, "project", "", "specify a GCP project (default is gcloud core/project)")
	rootCmd.PersistentFlags().DurationVar(&interval, "interval", 10*time.Second, "specify a polling interval for long-running operations")
	rootCmd.PersistentFlags().DurationVar(&timeout, "timeout", 10*time.Minute, "specify a timeout for long-running operations")
	rootCmd.PersistentFlags().BoolVar(&async, "async", false, "return immediately rather than waiting for operations to complete")
}

func commonContext(cmd *cobra.Command) (context.Context, *trace.Span) {
	path := strings.Split(cmd.CommandPath(), " ")
	ctx, span := trace.StartSpan(rootCtx, fmt.Sprintf("belvedere.cmd.%s", strings.Join(path[1:], ".")))
	span.AddAttributes(
		trace.BoolAttribute("debug", debug),
		trace.BoolAttribute("quiet", quiet),
		trace.BoolAttribute("dry_run", dryRun),
		trace.BoolAttribute("async", async),
		trace.Int64Attribute("interval_ms", int64(interval.Round(1*time.Millisecond))),
		trace.Int64Attribute("timeout_ms", int64(timeout.Round(1*time.Millisecond))),
		trace.StringAttribute("project", project),
	)
	return ctx, span
}

func Execute(version string) {
	rootCmd.Version = version
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func enableLogging() {
	// Export all traces.
	trace.ApplyConfig(trace.Config{
		DefaultSampler: trace.AlwaysSample(),
	})

	if debug {
		// Use the print exporter for debugging, as it prints everything.
		pe := &exporter.PrintExporter{}
		trace.RegisterExporter(pe)
		view.RegisterExporter(pe)
	} else if !quiet {
		// Unless we're quiet, use the trace logger for more practical logging.
		trace.RegisterExporter(&internal.TraceLogger{})
	}
}

func findProject() error {
	if project != "" {
		return nil
	}

	b, err := exec.Command("gcloud", "config", "config-helper", "--format=json").Output()
	if err != nil {
		return xerrors.Errorf("unable to execute gcloud: %w", err)
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
		return xerrors.Errorf("unable to parse gcloud config: %w", err)
	}

	if config.Configuration.Properties.Core.Project != "" {
		project = config.Configuration.Properties.Core.Project
		return nil
	}

	return errors.New("project not found")
}

func enableUsage(f cobra.PositionalArgs) cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		if err := f(cmd, args); err != nil {
			_ = cmd.Usage()
			return err
		}
		return nil
	}
}
