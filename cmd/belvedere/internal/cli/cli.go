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
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"go.opencensus.io/examples/exporter"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/trace"
	"google.golang.org/api/option"
)

type CommandFunc func(ctx context.Context, project belvedere.Project, args Args, out Output) error

type ProjectFactory func(ctx context.Context, name string, opts ...option.ClientOption) (belvedere.Project, error)

type OutputFactory func(w io.Writer, format string) (Output, error)

type Command struct {
	UI          cobra.Command
	Flags       func(fs *pflag.FlagSet)
	Run         CommandFunc
	Subcommands []*Command
}

func (c *Command) ToCobra(pf ProjectFactory, of OutputFactory) *cobra.Command {
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
			cmd.AddCommand(sc.ToCobra(pf, of))
		}

		return &cmd
	}

	// Register the global flags for each command.
	var gf GlobalFlags

	gf.Register(cmd.Flags())

	// Wrap the func, if one is provided.
	if c.Run != nil {
		cmd.RunE = runE(&gf, pf, of, c.Run)
	}

	return &cmd
}

func runE(gf *GlobalFlags, pf ProjectFactory, of OutputFactory, f CommandFunc) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, cmdArgs []string) error {
		// Enable trace logging.
		enableLogging(cmd.ErrOrStderr(), gf.Debug, gf.Quiet)

		// Initialize a context with a timeout and an interval.
		ctx, cancel := context.WithTimeout(context.Background(), gf.Timeout)
		defer cancel()

		// Create a root span.
		ctx, span := rootSpan(ctx)
		defer span.End()

		// Create a Belvedere project.
		project, err := pf(ctx, gf.Project,
			option.WithUserAgent(fmt.Sprintf("belvedere/%s", cmd.Root().Version)))
		if err != nil {
			return err
		}

		span.AddAttributes(
			trace.StringAttribute("project", project.Name()),
			trace.StringAttribute("args", escapeArgs(cmdArgs)),
		)

		// Construct args instance.
		input := &args{
			stdin: cmd.InOrStdin(),
			args:  cmdArgs,
		}

		// Construct output instance.
		output, err := of(cmd.OutOrStdout(), gf.Format)
		if err != nil {
			return err
		}

		// Execute command.
		err = f(ctx, project, input, output)
		if err != nil {
			span.SetStatus(trace.Status{
				Code:    trace.StatusCodeInternal,
				Message: err.Error(),
			})
		}

		return err
	}
}

func rootSpan(ctx context.Context) (context.Context, *trace.Span) {
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

	return ctx, span
}

func enableLogging(stderr io.Writer, debug, quiet bool) {
	// Export all traces.
	trace.ApplyConfig(trace.Config{
		DefaultSampler: trace.AlwaysSample(),
	})

	// Enable trace logging.
	if debug {
		// Use the print exporter for debugging, as it prints everything.
		pe := &exporter.PrintExporter{}
		trace.RegisterExporter(pe)
		view.RegisterExporter(pe)
	} else if !quiet {
		// Unless we're quiet, use the trace logger for more practical logging.
		trace.RegisterExporter(NewTraceLogger(stderr))
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
