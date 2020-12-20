package cli

import (
	"testing"
	"time"

	"github.com/codahale/belvedere/internal/assert"
	"github.com/spf13/pflag"
)

func TestGlobalFlags_Quiet(t *testing.T) {
	t.Parallel()

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
			t.Parallel()

			fs := pflag.NewFlagSet("test", pflag.PanicOnError)
			gf := GlobalFlags{}
			gf.Register(fs)
			if err := fs.Parse(testCase.args); err != nil {
				t.Fatal(err)
			}

			assert.Equal(t, "Quiet", testCase.quiet, gf.Quiet)
		})
	}
}

func TestGlobalFlags_Debug(t *testing.T) {
	t.Parallel()

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
			t.Parallel()

			fs := pflag.NewFlagSet("test", pflag.PanicOnError)
			gf := GlobalFlags{}
			gf.Register(fs)
			if err := fs.Parse(testCase.args); err != nil {
				t.Fatal(err)
			}

			assert.Equal(t, "Debug", testCase.debug, gf.Debug)
		})
	}
}

func TestGlobalFlags_Timeout(t *testing.T) {
	t.Parallel()

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
			t.Parallel()

			fs := pflag.NewFlagSet("test", pflag.PanicOnError)
			gf := GlobalFlags{}
			gf.Register(fs)
			if err := fs.Parse(testCase.args); err != nil {
				t.Fatal(err)
			}

			assert.Equal(t, "Timeout", testCase.timeout, gf.Timeout)
		})
	}
}

func TestGlobalFlags_Project(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		args    []string
		project string
	}{
		{
			name:    "long",
			args:    []string{"--project=example"},
			project: "example",
		},
	}

	for _, v := range tests {
		testCase := v
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			fs := pflag.NewFlagSet("test", pflag.PanicOnError)
			gf := GlobalFlags{}
			gf.Register(fs)
			if err := fs.Parse(testCase.args); err != nil {
				t.Fatal(err)
			}

			assert.Equal(t, "Project", testCase.project, gf.Project)
		})
	}
}

func TestGlobalFlags_Format(t *testing.T) {
	t.Parallel()

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
			t.Parallel()

			fs := pflag.NewFlagSet("test", pflag.PanicOnError)
			gf := GlobalFlags{}
			gf.Register(fs)
			if err := fs.Parse(testCase.args); err != nil {
				t.Fatal(err)
			}

			assert.Equal(t, "Format", testCase.format, gf.Format)
		})
	}
}

func TestModifyFlags_DryRun(t *testing.T) {
	t.Parallel()

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
			t.Parallel()

			fs := pflag.NewFlagSet("test", pflag.PanicOnError)
			mf := ModifyFlags{}
			mf.Register(fs)
			if err := fs.Parse(testCase.args); err != nil {
				t.Fatal(err)
			}

			assert.Equal(t, "DryRun", testCase.dryRun, mf.DryRun)
		})
	}
}

func TestLongRunningFlags_Interval(t *testing.T) {
	t.Parallel()

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
			t.Parallel()

			fs := pflag.NewFlagSet("test", pflag.PanicOnError)
			lrf := LongRunningFlags{}
			lrf.Register(fs)
			if err := fs.Parse(testCase.args); err != nil {
				t.Fatal(err)
			}

			assert.Equal(t, "Interval", testCase.interval, lrf.Interval)
		})
	}
}

func TestAsyncFlags_Async(t *testing.T) {
	t.Parallel()

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
			t.Parallel()

			fs := pflag.NewFlagSet("test", pflag.PanicOnError)
			af := AsyncFlags{}
			af.Register(fs)
			if err := fs.Parse(testCase.args); err != nil {
				t.Fatal(err)
			}

			assert.Equal(t, "Async", testCase.async, af.Async)
		})
	}
}
