package cmd

import "time"

type ModifyOptions struct {
	DryRun bool `help:"Print modifications instead of performing them."`
}

type LongRunningOptions struct {
	Interval time.Duration `help:"Specify a polling interval for long-running operations." default:"10s"`
}
