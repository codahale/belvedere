package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"syscall"
	"time"

	"github.com/codahale/belvedere/pkg/belvedere"
)

type SetupCmd struct {
	DNSZone string `arg:"" required:"" help:"The DNS zone to be managed by this project."`
}

func (cmd *SetupCmd) Run(ctx context.Context, project belvedere.Project, o *Options) error {
	return project.Setup(ctx, cmd.DNSZone, o.DryRun, o.Interval)
}

type TeardownCmd struct {
	Async bool `help:"Return without waiting for successful completion."`
}

func (cmd *TeardownCmd) Run(ctx context.Context, project belvedere.Project, o *Options) error {
	return project.Teardown(ctx, o.DryRun, cmd.Async, o.Interval)
}

type DNSServersCmd struct {
}

func (cmd *DNSServersCmd) Run(ctx context.Context, project belvedere.Project, tables TableWriter) error {
	servers, err := project.DNSServers(ctx)
	if err != nil {
		return err
	}
	return tables.Print(servers)
}

type MachineTypesCmd struct {
	Region string `help:"Limit types to those available in the given region."`
}

func (cmd *MachineTypesCmd) Run(ctx context.Context, project belvedere.Project, tables TableWriter) error {
	machineTypes, err := project.MachineTypes(ctx, cmd.Region)
	if err != nil {
		return err
	}
	return tables.Print(machineTypes)
}

type InstancesCmd struct {
	App     string `arg:"" optional:"" help:"Limit instances to those running the given app."`
	Release string `arg:"" optional:"" help:"Limit instances to those running the given release."`
}

func (cmd *InstancesCmd) Run(ctx context.Context, project belvedere.Project, tables TableWriter) error {
	instances, err := project.Instances(ctx, cmd.App, cmd.Release)
	if err != nil {
		return err
	}
	return tables.Print(instances)
}

type SSHCmd struct {
	Instance string   `arg:"" required:"" help:"The instance name."`
	Args     []string `arg:"" optional:"" help:"Additional SSH arguments."`
}

func (cmd *SSHCmd) Run(o *Options) error {
	// Find gcloud on the path.
	gcloud, err := exec.LookPath("gcloud")
	if err != nil {
		return fmt.Errorf("error finding gcloud executable: %w", err)
	}

	// Concat SSH arguments.
	args := append([]string{
		gcloud, "beta", "compute", "ssh", cmd.Instance, "--tunnel-through-iap", "--",
	}, cmd.Args...)

	// Exec to gcloud as last bit of main.
	o.exit = func() error {
		return syscall.Exec(gcloud, args, os.Environ())
	}
	return nil
}

type LogsCmd struct {
	App       string        `arg:"" help:"Limit logs to the given app."`
	Release   string        `arg:"" optional:"" help:"Limit logs to the given release."`
	Instance  string        `arg:"" optional:"" help:"Limit logs to the given instance."`
	Filters   []string      `name:"filter" optional:"" help:"Limit logs with the given Stackdriver Logging filters."`
	Freshness time.Duration `default:"5m" help:"Limit logs to the last period of time."`
}

func (cmd *LogsCmd) Run(ctx context.Context, project belvedere.Project, tables TableWriter) error {
	entries, err := project.Logs().List(ctx, cmd.App, cmd.Release, cmd.Instance, cmd.Freshness, cmd.Filters)
	if err != nil {
		return err
	}
	return tables.Print(entries)
}
