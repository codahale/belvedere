package cli

import (
	"encoding/json"
	"os/exec"
	"sync"
	"time"

	"github.com/spf13/pflag"
)

type GlobalFlags struct {
	Quiet   bool
	Debug   bool
	Timeout time.Duration
	Project string
	Format  string
}

func (gf *GlobalFlags) Register(fs *pflag.FlagSet) {
	defaultProjectOnce.Do(setDefaultProject)
	fs.BoolVarP(&gf.Quiet, "quiet", "q", false, "disable logging entirely")
	fs.BoolVar(&gf.Debug, "debug", false, "log verbose output")
	fs.DurationVar(&gf.Timeout, "timeout", 10*time.Minute, "maximum time allowed for total execution")
	fs.StringVar(&gf.Project, "project", defaultProject, "specify a Google Cloud Platform project ID")
	fs.StringVar(&gf.Format, "format", "table", "specify an output format (table, csv, json, prettyjson, yaml)")
}

type ModifyFlags struct {
	DryRun bool
}

func (m *ModifyFlags) Register(fs *pflag.FlagSet) {
	fs.BoolVar(&m.DryRun, "dry-run", false, "print modifications instead of performing them")
}

type LongRunningFlags struct {
	Interval time.Duration
}

func (l *LongRunningFlags) Register(fs *pflag.FlagSet) {
	fs.DurationVar(&l.Interval, "interval", 10*time.Second, "the polling interval for long-running operations")
}

type AsyncFlags struct {
	Async bool
}

func (a *AsyncFlags) Register(fs *pflag.FlagSet) {
	fs.BoolVar(&a.Async, "async", false, "return without waiting for completion")
}

//nolint:gochecknoglobals // project name has to be a singleton value
var (
	defaultProjectOnce sync.Once
	defaultProject     string
)

func setDefaultProject() {
	path, err := exec.LookPath("gcloud")
	if err != nil {
		return
	}

	out, err := exec.Command(path,
		"config", "get-value", "project", "-q", "--format=json",
	).CombinedOutput()
	if err != nil {
		return
	}

	_ = json.Unmarshal(out, &defaultProject)
}
