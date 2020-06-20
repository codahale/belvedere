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

	for _, v := range tests {
		testCase := v
		t.Run(testCase.name, func(t *testing.T) {
			fs := pflag.NewFlagSet("test", pflag.PanicOnError)
			gf := GlobalFlags{}
			gf.Register(fs)
			if err := fs.Parse(testCase.args); err != nil {
				t.Fatal(err)
			}

			if !cmp.Equal(testCase.quiet, gf.Quiet) {
				t.Fatal(cmp.Diff(testCase.quiet, gf.Quiet))
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

	for _, v := range tests {
		testCase := v
		t.Run(testCase.name, func(t *testing.T) {
			fs := pflag.NewFlagSet("test", pflag.PanicOnError)
			gf := GlobalFlags{}
			gf.Register(fs)
			if err := fs.Parse(testCase.args); err != nil {
				t.Fatal(err)
			}

			if !cmp.Equal(testCase.debug, gf.Debug) {
				t.Fatal(cmp.Diff(testCase.debug, gf.Debug))
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

	for _, v := range tests {
		testCase := v
		t.Run(testCase.name, func(t *testing.T) {
			fs := pflag.NewFlagSet("test", pflag.PanicOnError)
			gf := GlobalFlags{}
			gf.Register(fs)
			if err := fs.Parse(testCase.args); err != nil {
				t.Fatal(err)
			}

			if !cmp.Equal(testCase.timeout, gf.Timeout) {
				t.Fatal(cmp.Diff(testCase.timeout, gf.Timeout))
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

	for _, v := range tests {
		testCase := v
		t.Run(testCase.name, func(t *testing.T) {
			fs := pflag.NewFlagSet("test", pflag.PanicOnError)
			gf := GlobalFlags{}
			gf.Register(fs)
			if err := fs.Parse(testCase.args); err != nil {
				t.Fatal(err)
			}

			if !cmp.Equal(testCase.project, gf.Project) {
				t.Fatal(cmp.Diff(testCase.project, gf.Project))
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

	for _, v := range tests {
		testCase := v
		t.Run(testCase.name, func(t *testing.T) {
			fs := pflag.NewFlagSet("test", pflag.PanicOnError)
			gf := GlobalFlags{}
			gf.Register(fs)
			if err := fs.Parse(testCase.args); err != nil {
				t.Fatal(err)
			}

			if !cmp.Equal(testCase.format, gf.Format) {
				t.Fatal(cmp.Diff(testCase.format, gf.Format))
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

	for _, v := range tests {
		testCase := v
		t.Run(testCase.name, func(t *testing.T) {
			fs := pflag.NewFlagSet("test", pflag.PanicOnError)
			mf := ModifyFlags{}
			mf.Register(fs)
			if err := fs.Parse(testCase.args); err != nil {
				t.Fatal(err)
			}

			if !cmp.Equal(testCase.dryRun, mf.DryRun) {
				t.Fatal(cmp.Diff(testCase.dryRun, mf.DryRun))
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

	for _, v := range tests {
		testCase := v
		t.Run(testCase.name, func(t *testing.T) {
			fs := pflag.NewFlagSet("test", pflag.PanicOnError)
			lrf := LongRunningFlags{}
			lrf.Register(fs)
			if err := fs.Parse(testCase.args); err != nil {
				t.Fatal(err)
			}

			if !cmp.Equal(testCase.interval, lrf.Interval) {
				t.Fatal(cmp.Diff(testCase.interval, lrf.Interval))
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

	for _, v := range tests {
		testCase := v
		t.Run(testCase.name, func(t *testing.T) {
			fs := pflag.NewFlagSet("test", pflag.PanicOnError)
			af := AsyncFlags{}
			af.Register(fs)
			if err := fs.Parse(testCase.args); err != nil {
				t.Fatal(err)
			}

			if !cmp.Equal(testCase.async, af.Async) {
				t.Fatal(cmp.Diff(testCase.async, af.Async))
			}
		})
	}
}
