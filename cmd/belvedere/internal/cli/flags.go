package cli

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/mitchellh/go-homedir"
	"github.com/spf13/pflag"
	"gopkg.in/ini.v1"
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
	var sdkPath string

	// Find the SDK config path.
	if runtime.GOOS == "windows" {
		sdkPath = filepath.Join(os.Getenv("APPDATA"), "gcloud")
	} else {
		home, err := homedir.Dir()
		if err != nil {
			return
		}

		sdkPath = filepath.Join(home, ".config", "gcloud")
	}

	// Read the active config.
	configName, err := ioutil.ReadFile(filepath.Join(sdkPath, "active_config"))
	if err != nil {
		return
	}

	// Find the default config file.
	configPath := filepath.Join(sdkPath, "configurations",
		fmt.Sprintf("config_%s", strings.TrimSpace(string(configName))))

	// Read and parse it.
	cfg, err := ini.Load(configPath)
	if err != nil {
		return
	}

	// Find core.project, if any.
	key, err := cfg.Section("core").GetKey("project")
	if err != nil {
		return
	}

	defaultProject = key.String()
}
