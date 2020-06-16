package main

import (
	"fmt"
	"os"
	"runtime"

	"github.com/codahale/belvedere/cmd/belvedere/internal/cli"
	"github.com/codahale/belvedere/pkg/belvedere"
	"github.com/spf13/cobra"
)

func main() {
	cobra.EnableCommandSorting = false
	version := buildVersion(version, commit, date, builtBy)
	root := newRootCmd(version).ToCobra(belvedere.NewProject, cli.NewOutput, cli.NewFs())
	if err := root.Execute(); err != nil {
		_, _ = fmt.Fprint(os.Stderr, err)
		os.Exit(1)
	}
}

func newRootCmd(version string) *cli.Command {
	return &cli.Command{
		UI: cobra.Command{
			Use:   `belvedere`,
			Short: `A small lookout tower (usually square) on the roof of a house`,
			Long: `A small lookout tower (usually square) on the roof of a house.

Belvedere provides an easy and reliable way of deploying and managing HTTP2 applications on Google
Cloud Platform. It handles load balancing, DNS, TLS, blue/green deploys, auto-scaling, CDN
configuration, access control, secret management, configuration, and more.`,
			Version: version,
		},
		Subcommands: []*cli.Command{
			newSetupCmd(),
			newTeardownCmd(),
			newDNSServersCmd(),
			newInstancesCmd(),
			newLogsCmd(),
			newMachineTypesCmd(),
			newSSHCmd(),
			newAppsCmd(),
			newReleasesCmd(),
			newSecretsCmd(),
		},
	}
}

func buildVersion(version, commit, date, builtBy string) string {
	if version == "dev" {
		return fmt.Sprintf("%s (%s)", version, runtime.Version())
	}
	return fmt.Sprintf("%s (commit %.8s, %s, %s, %s)", version, commit, date, builtBy, runtime.Version())
}

// nolint: gochecknoglobals
var (
	version = "dev"
	commit  = ""
	date    = ""
	builtBy = ""
)
