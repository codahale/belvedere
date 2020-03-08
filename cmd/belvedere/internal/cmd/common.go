package cmd

import (
	"flag"
	"strings"
	"time"

	"github.com/olekukonko/tablewriter"
)

func Wrap(doc string) string {
	parts := strings.Split(doc, "\n\n")
	wrapped := make([]string, len(parts))
	for i, part := range parts {
		lines, _ := tablewriter.WrapString(part, 80)
		wrapped[i] = strings.TrimSpace(strings.Join(lines, "\n"))
	}
	return strings.Join(wrapped, "\n\n")
}

type ModifyOptions struct {
	DryRun bool `help:"Print modifications instead of performing them."`
}

func (o *ModifyOptions) RegisterFlags(fs *flag.FlagSet) {
	fs.BoolVar(&o.DryRun, "dry-run", false, "print modifications instead of performing them")
}

type LongRunningOptions struct {
	Interval time.Duration `help:"Specify a polling interval for long-running operations." default:"10s"`
}

func (o *LongRunningOptions) RegisterFlags(fs *flag.FlagSet) {
	fs.DurationVar(&o.Interval, "interval", 10*time.Second, "the polling interval for long-running operations")
}

type AsyncOptions struct {
	Async bool
}

func (o *AsyncOptions) RegisterFlags(fs *flag.FlagSet) {
	fs.BoolVar(&o.Async, "async", false, "return without waiting for completion")
}
