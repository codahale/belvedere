package logscmd

import (
	"context"
	"flag"
	"strings"
	"time"

	"github.com/codahale/belvedere/cmd/belvedere/internal/cmd"
	"github.com/codahale/belvedere/cmd/belvedere/internal/rootcmd"
	"github.com/peterbourgon/ff/v2/ffcli"
)

type filters []string

func (f *filters) String() string {
	return strings.Join(*f, ",")
}

func (f *filters) Set(value string) error {
	*f = append(*f, strings.TrimSpace(value))
	return nil
}

type Config struct {
	root    *rootcmd.Config
	filters filters
	maxAge  time.Duration
}

func New(root *rootcmd.Config) *ffcli.Command {
	config := Config{root: root}

	fs := flag.NewFlagSet("belvedere logs", flag.ExitOnError)
	root.RegisterFlags(fs)
	config.RegisterFlags(fs)

	return &ffcli.Command{
		Name:       "logs",
		ShortUsage: "belvedere logs <app> [<release>] [<instance>] [--filter=<filter>...] [flags]",
		ShortHelp:  "Display application logs.",
		LongHelp: cmd.Wrap(`Display application logs.

Log entries are bounded by the -max-age parameter and filtered by the application name. They can
also be filtered by the release name, the instance name, and any additional Google Cloud Logging
filters. For more information on filter syntax, see
https://cloud.google.com/logging/docs/view/advanced-queries#advanced_logs_query_syntax.`),
		FlagSet: fs,
		Exec:    config.Exec,
	}
}

func (c *Config) RegisterFlags(fs *flag.FlagSet) {
	fs.Var(&c.filters, "filter", "limit entries to the given filter")
	fs.DurationVar(&c.maxAge, "max-age", 10*time.Minute, "limit entries by maximum age")
}

func (c *Config) Exec(ctx context.Context, args []string) error {
	if len(args) < 1 || len(args) > 3 {
		return flag.ErrHelp
	}

	app := args[0]

	var release string
	if len(args) > 1 {
		release = args[1]
	}

	var instance string
	if len(args) > 2 {
		instance = args[2]
	}

	filters := make([]string, len(c.filters)) // we're copying the filters to make gomock happy
	copy(filters, c.filters)
	entries, err := c.root.Project.Logs().List(ctx, app, release, instance, c.maxAge, filters)
	if err != nil {
		return err
	}
	return c.root.Tables.Print(entries)
}
