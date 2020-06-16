package cli

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/user"
	"strings"

	"github.com/alessio/shellescape"
	"github.com/codahale/belvedere/pkg/belvedere"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"go.opencensus.io/examples/exporter"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/trace"
	"google.golang.org/api/option"
)

type CommandFunc func(
	ctx context.Context, project belvedere.Project, output Output, fs afero.Fs,
	args []string,
) error

type CallbackFunc func(
	ctx context.Context, project belvedere.Project, output Output, fs afero.Fs,
	args []string,
) (func() error, error)

type ProjectFactory func(ctx context.Context, name string, opts ...option.ClientOption) (belvedere.Project, error)

type OutputFactory func(w io.Writer, format string) (Output, error)

type Command struct {
	UI          cobra.Command
	Flags       func(fs *pflag.FlagSet)
	Run         CommandFunc
	RunCallback CallbackFunc
	Subcommands []*Command
}

func (c *Command) ToCobra(pf ProjectFactory, of OutputFactory, fs afero.Fs) *cobra.Command {
	cmd := c.UI

	// Populate the command's flags.
	if c.Flags != nil {
		c.Flags(cmd.Flags())
	}

	// Wrap the command's main docs.
	cmd.Long = fmtText(cmd.Long)

	// Add subcommands, if any, and return.
	if len(c.Subcommands) > 0 {
		for _, sc := range c.Subcommands {
			cmd.AddCommand(sc.ToCobra(pf, of, fs))
		}
		return &cmd
	}

	// Register the global flags for each command.
	var gf GlobalFlags
	gf.Register(cmd.Flags())

	switch {
	case c.Run != nil && c.RunCallback != nil:
		panic("both a run func and a run callback func")
	case c.Run != nil:
		// If it's a regular command, wrap it to return a nil callback func.
		cmd.RunE = runE(&gf, pf, of, fs,
			func(ctx context.Context, project belvedere.Project, output Output, fs afero.Fs, args []string) (func() error, error) {
				return nil, c.Run(ctx, project, output, fs, args)
			},
		)
	case c.RunCallback != nil:
		// Otherwise, just wrap the func.
		cmd.RunE = runE(&gf, pf, of, fs, c.RunCallback)
	default:
		panic("no subcommands or run func")
	}
	return &cmd
}

func runE(gf *GlobalFlags, pf ProjectFactory, of OutputFactory, fs afero.Fs, f CallbackFunc) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		// Wrap this in a func to make our defers work.
		callback, err := func() (func() error, error) {
			// Export all traces.
			trace.ApplyConfig(trace.Config{
				DefaultSampler: trace.AlwaysSample(),
			})

			// Enable trace logging.
			if gf.Debug {
				// Use the print exporter for debugging, as it prints everything.
				pe := &exporter.PrintExporter{}
				trace.RegisterExporter(pe)
				view.RegisterExporter(pe)
			} else if !gf.Quiet {
				// Unless we're quiet, use the trace logger for more practical logging.
				trace.RegisterExporter(NewTraceLogger(cmd.OutOrStderr()))
			}

			// Initialize a context with a timeout and an interval.
			ctx, cancel := context.WithTimeout(context.Background(), gf.Timeout)
			defer cancel()

			// Create a root span.
			ctx, span := trace.StartSpan(ctx, "belvedere.main")
			defer span.End()
			if hostname, err := os.Hostname(); err == nil {
				span.AddAttributes(trace.StringAttribute("hostname", hostname))
			}
			if u, err := user.Current(); err == nil {
				span.AddAttributes(
					trace.StringAttribute("username", u.Username),
					trace.StringAttribute("uid", u.Uid),
				)
			}

			// Create a Belvedere project.
			project, err := pf(ctx, gf.Project,
				option.WithUserAgent(fmt.Sprintf("belvedere/%s", cmd.Root().Version)))
			if err != nil {
				return nil, err
			}
			span.AddAttributes(
				trace.StringAttribute("project", project.Name()),
				trace.StringAttribute("args", escapeArgs(args)),
			)

			output, err := of(os.Stdout, gf.Format)
			if err != nil {
				return nil, err
			}
			return f(ctx, project, output, fs, args)
		}()
		if err != nil {
			return err
		}

		if callback != nil {
			cmd.PostRunE = func(cmd *cobra.Command, args []string) error {
				return callback()
			}
		}

		return nil
	}
}

func fmtText(doc string) string {
	parts := strings.Split(doc, "\n\n")
	wrapped := make([]string, len(parts))
	for i, part := range parts {
		lines, _ := tablewriter.WrapString(part, 80)
		wrapped[i] = strings.TrimSpace(strings.Join(lines, "\n"))
	}
	return strings.Join(wrapped, "\n\n")
}

func escapeArgs(args []string) string {
	escaped := make([]string, len(args))
	for i, s := range args {
		escaped[i] = shellescape.Quote(s)
	}
	return strings.Join(escaped, " ")
}
