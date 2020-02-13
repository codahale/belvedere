package main

import (
	"context"
	"time"

	"github.com/codahale/belvedere/pkg/belvedere"
)

type SetupCmd struct {
	DNSZone string `arg:"" required:"" help:"The DNS zone to be managed by this project."`
}

func (cmd *SetupCmd) Run(ctx context.Context, o *Options) error {
	return belvedere.Setup(ctx, o.Project, cmd.DNSZone, o.DryRun, o.Interval)
}

type TeardownCmd struct {
	Async bool `help:"Return without waiting for successful completion."`
}

func (cmd *TeardownCmd) Run(ctx context.Context, o *Options) error {
	return belvedere.Teardown(ctx, o.Project, o.DryRun, cmd.Async, o.Interval)
}

type DNSServersCmd struct {
}

func (cmd *DNSServersCmd) Run(ctx context.Context, o *Options) error {
	servers, err := belvedere.DNSServers(ctx, o.Project)
	if err != nil {
		return err
	}
	return o.printTable(servers)
}

type MachineTypesCmd struct {
	Region string `help:"Limit types to those available in the given region."`
}

func (cmd *MachineTypesCmd) Run(ctx context.Context, o *Options) error {
	machineTypes, err := belvedere.MachineTypes(ctx, o.Project, cmd.Region)
	if err != nil {
		return err
	}
	return o.printTable(machineTypes)
}

type InstancesCmd struct {
	App     string `arg:"" optional:"" help:"Limit instances to those running the given app."`
	Release string `arg:"" optional:"" help:"Limit instances to those running the given release."`
}

func (cmd *InstancesCmd) Run(ctx context.Context, o *Options) error {
	instances, err := belvedere.Instances(ctx, o.Project, cmd.App, cmd.Release)
	if err != nil {
		return err
	}
	return o.printTable(instances)
}

type SSHCmd struct {
	Instance string   `arg:"" required:"" help:"The instance name."`
	Args     []string `arg:"" help:"Additional SSH arguments."`
}

func (cmd *SSHCmd) Run(ctx context.Context, o *Options) error {
	ssh, err := belvedere.SSH(ctx, o.Project, cmd.Instance, cmd.Args)
	if err != nil {
		return err
	}
	o.exit = ssh
	return nil
}

type LogsCmd struct {
	App       string        `arg:"" help:"Limit logs to the given app."`
	Release   string        `arg:"" optional:"" help:"Limit logs to the given release."`
	Instance  string        `arg:"" optional:"" help:"Limit logs to the given instance."`
	Filters   []string      `name:"filter" optional:"" help:"Limit logs with the given Stackdriver Logging filters."`
	Freshness time.Duration `default:"5m" help:"Limit logs to the last period of time."`
}

func (cmd *LogsCmd) Run(ctx context.Context, o *Options) error {
	t := time.Now().Add(-cmd.Freshness)
	logs, err := belvedere.Logs(ctx, o.Project, cmd.App, cmd.Release, cmd.Instance, t, cmd.Filters)
	if err != nil {
		return err
	}
	return o.printTable(logs)
}
