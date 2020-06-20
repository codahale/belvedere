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

			want, got := testCase.quiet, gf.Quiet
			if diff := cmp.Diff(want, got); diff != "" {
				t.Errorf("Quiet mismatch (-want +got):\n%s", diff)
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

			want, got := testCase.debug, gf.Debug
			if diff := cmp.Diff(want, got); diff != "" {
				t.Errorf("Debug mismatch (-want +got):\n%s", diff)
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

			want, got := testCase.timeout, gf.Timeout
			if diff := cmp.Diff(want, got); diff != "" {
				t.Errorf("Timeout mismatch (-want +got):\n%s", diff)
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

			want, got := testCase.project, gf.Project
			if diff := cmp.Diff(want, got); diff != "" {
				t.Errorf("Project mismatch (-want +got):\n%s", diff)
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

			want, got := testCase.format, gf.Format
			if diff := cmp.Diff(want, got); diff != "" {
				t.Errorf("Format mismatch (-want +got):\n%s", diff)
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

			want, got := testCase.dryRun, mf.DryRun
			if diff := cmp.Diff(want, got); diff != "" {
				t.Errorf("DryRun mismatch (-want +got):\n%s", diff)
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

			want, got := testCase.interval, lrf.Interval
			if diff := cmp.Diff(want, got); diff != "" {
				t.Errorf("Interval mismatch (-want +got):\n%s", diff)
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

			want, got := testCase.async, af.Async
			if diff := cmp.Diff(want, got); diff != "" {
				t.Errorf("Async mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
