package cli

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/spf13/pflag"
)

func TestGlobalFlags_Quiet(t *testing.T) {
	tests := []struct {
		name  string
		args  []string
		quiet bool
	}{
		{
			name:  "default",
			args:  nil,
			quiet: false,
		},
		{
			name:  "short",
			args:  []string{"-q"},
			quiet: true,
		},
		{
			name:  "long",
			args:  []string{"--quiet"},
			quiet: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fs := pflag.NewFlagSet("test", pflag.PanicOnError)
			gf := GlobalFlags{}
			gf.Register(fs)
			if err := fs.Parse(test.args); err != nil {
				t.Fatal(err)
			}

			if !cmp.Equal(test.quiet, gf.Quiet) {
				t.Fatal(cmp.Diff(test.quiet, gf.Quiet))
			}
		})
	}
}

func TestGlobalFlags_Debug(t *testing.T) {
	tests := []struct {
		name  string
		args  []string
		debug bool
	}{
		{
			name:  "default",
			args:  nil,
			debug: false,
		},
		{
			name:  "long",
			args:  []string{"--debug"},
			debug: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fs := pflag.NewFlagSet("test", pflag.PanicOnError)
			gf := GlobalFlags{}
			gf.Register(fs)
			if err := fs.Parse(test.args); err != nil {
				t.Fatal(err)
			}

			if !cmp.Equal(test.debug, gf.Debug) {
				t.Fatal(cmp.Diff(test.debug, gf.Debug))
			}
		})
	}
}

func TestGlobalFlags_Timeout(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		timeout time.Duration
	}{
		{
			name:    "default",
			args:    nil,
			timeout: 10 * time.Minute,
		},
		{
			name:    "long",
			args:    []string{"--timeout=5m"},
			timeout: 5 * time.Minute,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fs := pflag.NewFlagSet("test", pflag.PanicOnError)
			gf := GlobalFlags{}
			gf.Register(fs)
			if err := fs.Parse(test.args); err != nil {
				t.Fatal(err)
			}

			if !cmp.Equal(test.timeout, gf.Timeout) {
				t.Fatal(cmp.Diff(test.timeout, gf.Timeout))
			}
		})
	}
}

func TestGlobalFlags_Project(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		project string
	}{
		{
			name:    "default",
			args:    nil,
			project: "",
		},
		{
			name:    "long",
			args:    []string{"--project=example"},
			project: "example",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fs := pflag.NewFlagSet("test", pflag.PanicOnError)
			gf := GlobalFlags{}
			gf.Register(fs)
			if err := fs.Parse(test.args); err != nil {
				t.Fatal(err)
			}

			if !cmp.Equal(test.project, gf.Project) {
				t.Fatal(cmp.Diff(test.project, gf.Project))
			}
		})
	}
}

func TestGlobalFlags_Format(t *testing.T) {
	tests := []struct {
		name   string
		args   []string
		format string
	}{
		{
			name:   "default",
			args:   nil,
			format: "table",
		},
		{
			name:   "long",
			args:   []string{"--format=json"},
			format: "json",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fs := pflag.NewFlagSet("test", pflag.PanicOnError)
			gf := GlobalFlags{}
			gf.Register(fs)
			if err := fs.Parse(test.args); err != nil {
				t.Fatal(err)
			}

			if !cmp.Equal(test.format, gf.Format) {
				t.Fatal(cmp.Diff(test.format, gf.Format))
			}
		})
	}
}

func TestModifyFlags_DryRun(t *testing.T) {
	tests := []struct {
		name   string
		args   []string
		dryRun bool
	}{
		{
			name:   "default",
			args:   nil,
			dryRun: false,
		},
		{
			name:   "long",
			args:   []string{"--dry-run"},
			dryRun: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fs := pflag.NewFlagSet("test", pflag.PanicOnError)
			mf := ModifyFlags{}
			mf.Register(fs)
			if err := fs.Parse(test.args); err != nil {
				t.Fatal(err)
			}

			if !cmp.Equal(test.dryRun, mf.DryRun) {
				t.Fatal(cmp.Diff(test.dryRun, mf.DryRun))
			}
		})
	}
}

func TestLongRunningFlags_Interval(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		interval time.Duration
	}{
		{
			name:     "default",
			args:     nil,
			interval: 10 * time.Second,
		},
		{
			name:     "long",
			args:     []string{"--interval=5m"},
			interval: 5 * time.Minute,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fs := pflag.NewFlagSet("test", pflag.PanicOnError)
			lrf := LongRunningFlags{}
			lrf.Register(fs)
			if err := fs.Parse(test.args); err != nil {
				t.Fatal(err)
			}

			if !cmp.Equal(test.interval, lrf.Interval) {
				t.Fatal(cmp.Diff(test.interval, lrf.Interval))
			}
		})
	}
}

func TestAsyncFlags_Async(t *testing.T) {
	tests := []struct {
		name  string
		args  []string
		async bool
	}{
		{
			name:  "default",
			args:  nil,
			async: false,
		},
		{
			name:  "long",
			args:  []string{"--async"},
			async: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fs := pflag.NewFlagSet("test", pflag.PanicOnError)
			af := AsyncFlags{}
			af.Register(fs)
			if err := fs.Parse(test.args); err != nil {
				t.Fatal(err)
			}

			if !cmp.Equal(test.async, af.Async) {
				t.Fatal(cmp.Diff(test.async, af.Async))
			}
		})
	}
}
