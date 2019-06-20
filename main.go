package main

import (
	"context"
	"fmt"
	"os"
	"os/user"

	"github.com/codahale/belvedere/pkg/belvedere"
	"github.com/docopt/docopt-go"
	"go.opencensus.io/examples/exporter"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/trace"
)

var (
	buildVersion = "unknown"
)

func main() {
	usage := `Belvedere: A fine place from which to survey your estate.

Usage:
  belvedere enable <project-id> [options]
  belvedere envs list <project-id> [options]
  belvedere envs create <project-id> <env name> <dns name> [options]
  belvedere envs destroy <project-id> <env name> [options]
  belvedere -h | --help
  belvedere --version

Options:
  -h --help     Show this screen.
  --version     Show version.
  --debug       Enable debug output.
  --quiet       Disable all log output.
`

	opts, err := docopt.ParseArgs(usage, nil, buildVersion)
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(-1)
	}

	trace.ApplyConfig(trace.Config{
		DefaultSampler: trace.AlwaysSample(),
	})

	if debug, err := opts.Bool("--debug"); err == nil && debug {
		pe := &exporter.PrintExporter{}
		trace.RegisterExporter(pe)
		view.RegisterExporter(pe)
	} else if quiet, err := opts.Bool("--quiet"); err != nil || !quiet {
		trace.RegisterExporter(belvedere.NewTraceLogger())
	}

	if ok, _ := opts.Bool("enable"); ok {
		projectID, _ := opts.String("<project-id>")
		if err := enable(projectID); err != nil {
			panic(err)
		}
	} else if ok, _ := opts.Bool("envs"); ok {
		projectID, _ := opts.String("<project-id>")

		if ok, _ := opts.Bool("list"); ok {
			if err := envsList(projectID); err != nil {
				panic(err)
			}
		} else if ok, _ := opts.Bool("create"); ok {
			envName, _ := opts.String("<env name>")
			dnsName, _ := opts.String("<dns name>")
			if err := envsCreate(projectID, envName, dnsName); err != nil {
				panic(err)
			}
		} else if ok, _ := opts.Bool("destroy"); ok {
			envName, _ := opts.String("<env name>")
			if err := envsDestroy(projectID, envName); err != nil {
				panic(err)
			}
		}
	} else {
		panic(fmt.Sprintf("unknown command: %v", opts))
	}
}

func enable(projectID string) error {
	ctx, span := rootSpan("belvedere.enable")
	span.AddAttributes(trace.StringAttribute("project_id", projectID))
	defer span.End()

	if err := belvedere.EnableServices(ctx, projectID); err != nil {
		return err
	}
	return belvedere.EnableDeploymentManagerIAM(ctx, projectID)
}

func envsList(projectID string) error {
	ctx, span := rootSpan("belvedere.envs.list")
	span.AddAttributes(trace.StringAttribute("project_id", projectID))
	defer span.End()

	envs, err := belvedere.ListEnvs(ctx, projectID)
	if err != nil {
		return err
	}

	for _, env := range envs {
		fmt.Printf("%+v\n", env)
	}
	return nil
}

func envsCreate(projectID, envName, dnsName string) error {
	ctx, span := rootSpan("belvedere.envs.create")
	span.AddAttributes(
		trace.StringAttribute("project_id", projectID),
		trace.StringAttribute("env_name", envName),
		trace.StringAttribute("dns_name", dnsName),
	)
	defer span.End()

	return belvedere.CreateEnv(ctx, projectID, envName, dnsName)
}

func envsDestroy(projectID, envName string) error {
	ctx, span := rootSpan("belvedere.envs.destroy")
	span.AddAttributes(
		trace.StringAttribute("project_id", projectID),
		trace.StringAttribute("env_name", envName),
	)
	defer span.End()

	return belvedere.DestroyEnv(ctx, projectID, envName)
}

func rootSpan(name string) (context.Context, *trace.Span) {
	ctx, span := trace.StartSpan(context.Background(), name)
	if hostname, err := os.Hostname(); err == nil {
		span.AddAttributes(trace.StringAttribute("hostname", hostname))
	}
	if u, err := user.Current(); err == nil {
		span.AddAttributes(trace.StringAttribute("user", u.Username))
	}
	return ctx, span
}
